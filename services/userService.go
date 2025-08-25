package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService interface {
	GetMyProfile(userId string) (*model.User, error)
	UpdateMyProfile(updateUserData *model.UpdateUser, userId *string) (*model.User, error)
	DeleteMyProfile(userId string) (*model.User, error)
	GetUserProfile(userID string) (*model.User, error)
	GetRandomUsers(userNum int) (*[]model.User, error)
	GetRecentlyJoinedUsers(userNum int, userId string) (*[]model.User, error)
	SearchUser(query, userId string) (*[]model.User, error)
}

type UserServiceStruct struct {
	db *mongo.Client
}

func NewUserService(db *mongo.Client) *UserServiceStruct {
	return &UserServiceStruct{
		db: db,
	}
}

// Get MyProfile
func (u *UserServiceStruct) GetMyProfile(userId string) (*model.User, error) {
	redis_client := utils.GetRedis()
	if redis_client == nil {
		return nil, fmt.Errorf("redis connection failed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	userInfo := make(chan model.User, 32)
	errChan := make(chan error, 32)

	var user model.User

	// Chk the UserId is valid or not
	user_id := redis_client.Get("login_info:user_id")
	if user_id.Val() != userId {
		errChan <- fmt.Errorf("invalid")
	}

	go func() {
		defer func() {
			close(userInfo)
			close(errChan)
		}()

		// converting the UserID to monogo ObjectID
		obj_Id, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			errChan <- err
		}

		// fmt.Println("obj_Id")
		// fmt.Println(obj_Id)

		get_user_id := bson.M{
			"_id": obj_Id,
		}

		// chk in Redis
		user_info_val := redis_client.Get("user_info" + userId).Val()
		fmt.Println("user_info_val")
		fmt.Println(user_info_val)

		if user_info_val == "" {
			fmt.Println("We didn;t get any value")
			err = u.db.Database("go-ecomm").Collection("users").FindOne(ctx, get_user_id).Decode(&user)
			if err != nil {
				errChan <- err
				return
			}

			// Store this information to Redis
			to_store_user_in_redis := &model.User{
				ID:        user.ID,
				Username:  user.Username,
				Email:     user.Email,
				Gender:    user.Gender,
				IsAdmin:   user.IsAdmin,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			}
			err = redis_client.Set("user_info"+user.ID.Hex(), to_store_user_in_redis, time.Second*10).Err()
			if err != nil {
				errChan <- err
			}
			userInfo <- user
		} else {
			got_data_from_redis := redis_client.Get("user_info" + userId)
			fmt.Println("got_data_from_redis")
			fmt.Println(got_data_from_redis.Val())
			// Decode Redis data into user struct
			err = json.Unmarshal([]byte(got_data_from_redis.Val()), &user)
			if err != nil {
				errChan <- err
				return
			}
			userInfo <- user
		}
	}()

	select {
	case err := <-errChan:
		return nil, err
	case user := <-userInfo:
		return &user, nil
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}
