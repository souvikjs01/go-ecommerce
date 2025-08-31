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
	SignUpService(req *request.SignupRequest) (*model.User, error)
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
func (a *AuthServiceStruct) SignUpService(req *request.SignupRequest) (*model.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	validGenders := map[string]bool{"male": true, "female": true, "other": true}
	if !validGenders[req.Gender] {
		return nil, fmt.Errorf("invalid gender: must be male, female, or other")
	}

	if req.Username == "" || req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return nil, errors.New("missing required fields")
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

	errChan := make(chan error, 32)
	userChan := make(chan model.User, 32)

	find_user := bson.M{
		"username": req.Username,
		"email":    req.Email,
	}

	go func() {
		defer func() {
			close(errChan)
			close(userChan)
		}()

		err := a.db.Database("go-ecomm").Collection("users").FindOne(ctx, find_user).Decode(&newUser)

		if err == nil {
			errChan <- fmt.Errorf("user already exists")
			return
		}

		// It means that User is not exists in our Database So we need to create a user
		insert_res, err := a.db.Database("go-ecomm").Collection("users").InsertOne(ctx, newUser)
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("insert_res")
		fmt.Println(insert_res)

		userChan <- *newUser

	}()
	select {
	case err := <-errChan:
		return nil, err
	case res_user := <-userChan:
		return &res_user, nil
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
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

		fmt.Println("inside the GoRoutine......")
		// Store inside the Redis

		// DataBse ddata
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
