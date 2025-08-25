package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/response"
	"github.com/souvikjs01/go-ecommerce/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService interface {
	SignUpService(user *model.User) (*model.User, error)
	LoginService(payload response.LoginResponse) (*model.User, string, error)
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
func (a *AuthServiceStruct) SignUpService(user *model.User) (*model.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	if user.Username == "" || user.Email == "" || user.Gender == "" || user.Password == "" {
		return nil, errors.New("missing required fields")
	}
	newUser := model.NewUser(&user.Username, &user.Email, &user.Gender, &user.Password)

	errChan := make(chan error, 32)
	userChan := make(chan model.User, 32)
	find_user := bson.M{
		"username": user.Username,
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
func (a *AuthServiceStruct) LoginService(payload response.LoginResponse) (*model.User, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	find_user := bson.M{
		"username": payload.Username,
	}
	var user model.User

	loggedIn_user_chan := make(chan *model.User, 32)
	err_chan := make(chan error, 32)

	if payload.Password == "" || payload.Username == "" {
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
		fmt.Printf("user Password %v\n", user.Password)
		fmt.Printf("login_payload Password %v\n", payload.Password)
		isValidPassword := utils.VerifyPassword(payload.Password, user.Password)
		if !isValidPassword {
			err_chan <- fmt.Errorf("invalid credentials")
			return
		}
		redis_client := utils.GetRedis()

		redRes1, err := redis_client.Set("login_info:username", user.Username, 0).Result()
		if err != nil {
			err_chan <- err
			return
		}
		redRes2, err := redis_client.Set("login_info:user_id", user.ID.Hex(), 0).Result()
		if err != nil {
			err_chan <- err
			return
		}
		redRes3, err := redis_client.Set("login_info:email", user.Email, 0).Result()
		if err != nil {
			err_chan <- err
			return
		}
		redRes4, err := redis_client.Set("login_info:isAdmin", user.IsAdmin, 0).Result()

		fmt.Println("red_res")
		fmt.Println(redRes1)
		fmt.Println(redRes2)
		fmt.Println(redRes3)
		fmt.Println(redRes4)

		if err != nil {
			err_chan <- err
			return
		}

		redis_login_result := redis_client.Get("login_info")
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
