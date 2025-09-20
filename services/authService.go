package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/request"
	"github.com/souvikjs01/go-ecommerce/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService interface {
	SignUpService(req *request.SignupRequest) (*model.User, string, error)
	LoginService(payload request.LoginRequest) (*model.User, string, error)
}

type AuthServiceStruct struct {
	db *mongo.Client
}

func NewAuthService(Db *mongo.Client) *AuthServiceStruct {
	return &AuthServiceStruct{
		db: Db,
	}
}

// Signup Service
func (a *AuthServiceStruct) SignUpService(req *request.SignupRequest) (*model.User, string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	validGenders := map[string]bool{"male": true, "female": true, "other": true}
	if !validGenders[req.Gender] {
		return nil, "", fmt.Errorf("invalid gender: must be male, female, or other")
	}

	if req.Username == "" || req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return nil, "", errors.New("missing required fields")
	}
	newUser := model.NewUser(
		&req.Username,
		&req.FirstName,
		&req.LastName,
		&req.Email,
		&req.Gender,
		req.ProfileImage,
		&req.Password,
	)

	errChan := make(chan model.ErrMsg, 32)
	userChan := make(chan model.User, 32)

	go func() {
		defer func() {
			close(errChan)
			close(userChan)
		}()

		var existUser *model.User
		err := a.db.Database("go-ecomm").Collection("users").FindOne(ctx, bson.M{
			"email": req.Email,
		}).Decode(&existUser)

		if err == nil {
			errChan <- model.ErrMsg{
				Err:  fmt.Errorf("user with this email already exists"),
				Code: 400,
			}
			return
		}

		err = a.db.Database("go-ecomm").Collection("users").FindOne(ctx, bson.M{
			"username": req.Username,
		}).Decode(&existUser)

		if err == nil {
			errChan <- model.ErrMsg{
				Err:  fmt.Errorf("user with this username already exist"),
				Code: 400,
			}
			return
		}

		// It means that User is not exists in our Database So we need to create a user
		_, err = a.db.Database("go-ecomm").Collection("users").InsertOne(ctx, newUser)
		if err != nil {
			errChan <- model.ErrMsg{
				Err:  err,
				Code: 500,
			}
			return
		}

		userChan <- *newUser

	}()
	select {
	case err := <-errChan:
		return nil, "", err
	case res_user := <-userChan:
		token, err := utils.CreateJWTToken(res_user.ID.Hex(), res_user.Username, res_user.IsAdmin)
		if err != nil {
			return nil, "", err
		}
		return &res_user, token, nil
	case <-ctx.Done():
		return nil, "", context.DeadlineExceeded
	}
}

// Login Handler
func (a *AuthServiceStruct) LoginService(payload request.LoginRequest) (*model.User, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	find_user := bson.M{
		"username": payload.Username,
	}
	var user model.User

	loggedIn_user_chan := make(chan *model.User, 32)
	err_chan := make(chan error, 32)

	if strings.TrimSpace(payload.Password) == "" || strings.TrimSpace(payload.Username) == "" {
		return nil, "", errors.New("invalid credentials")
	}

	go func() {
		defer func() {
			close(loggedIn_user_chan)
			close(err_chan)
		}()

		err := a.db.Database("go-ecomm").Collection("users").FindOne(ctx, find_user).Decode(&user)
		if err != nil {
			err_chan <- err
			return
		}

		isValidPassword := utils.VerifyPassword(payload.Password, user.Password)
		if !isValidPassword {
			err_chan <- fmt.Errorf("invalid credentials %s", err)
			return
		}
		redis_client := utils.GetRedis()

		loginKey := fmt.Sprintf("login_info:%s", user.ID.Hex())
		userMap := map[string]interface{}{
			"username":  user.Username,
			"user_id":   user.ID.Hex(),
			"email":     user.Email,
			"isAdmin":   user.IsAdmin,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
		}
		if user.ProfileImage != nil {
			userMap["profileImage"] = *user.ProfileImage
		}

		if err := redis_client.HMSet(loginKey, userMap).Err(); err != nil {
			err_chan <- err
			return
		}

		if err := redis_client.Expire(loginKey, 1*time.Hour).Err(); err != nil {
			err_chan <- err
			return
		}

		redis_login_result := redis_client.Get(loginKey)
		fmt.Println("redis_login_result")
		fmt.Println(redis_login_result)
		// fmt.Println(user)
		loggedIn_user_chan <- &user
	}()

	select {
	case err := <-err_chan:
		return nil, "", err
	case user_details := <-loggedIn_user_chan:
		// creating A JWT Token
		token, err := utils.CreateJWTToken(user.ID.Hex(), user.Username, user.IsAdmin)
		if err != nil {
			return nil, "", err
		}
		return user_details, token, nil
	case <-ctx.Done():
		return nil, "", context.DeadlineExceeded
	}
}
