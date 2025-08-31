package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/request"
	"github.com/souvikjs01/go-ecommerce/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService interface {
	GetMyProfile(userId string) (*model.User, error)
	UpdateMyProfile(updateUserData *request.UpdateRequest, userId *string) (*model.User, error)
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	userInfo := make(chan model.User, 32)
	errChan := make(chan error, 32)

	var user model.User

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

		get_user_id := bson.M{
			"_id": obj_Id,
		}

		// chk in Redis
		user_info_val := redis_client.Get("user_info" + userId).Val()

		if user_info_val == "" {
			fmt.Println("We didn;t get any value")
			err = u.db.Database("go-ecomm").Collection("users").FindOne(ctx, get_user_id).Decode(&user)
			if err != nil {
				errChan <- err
				return
			}

			// Store this information to Redis
			to_store_user_in_redis := &model.User{
				ID:           user.ID,
				Username:     user.Username,
				Email:        user.Email,
				Gender:       user.Gender,
				IsAdmin:      user.IsAdmin,
				CreatedAt:    user.CreatedAt,
				UpdatedAt:    user.UpdatedAt,
				FirstName:    user.FirstName,
				LastName:     user.LastName,
				ProfileImage: user.ProfileImage,
			}
			err = redis_client.Set("user_info"+user.ID.Hex(), to_store_user_in_redis, 1*time.Hour).Err()
			if err != nil {
				errChan <- err
			}
			userInfo <- user
		} else {
			got_data_from_redis := redis_client.Get("user_info" + userId)
			fmt.Println("got_data_from_redis")

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

// Update the UserProfile
func (u *UserServiceStruct) UpdateMyProfile(user_data *request.UpdateRequest, userId *string) (*model.User, error) {

	user_chan := make(chan *model.User, 32)
	err_chan := make(chan error, 32)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// search Id of user in Redis
	redis_client := utils.GetRedis()
	if redis_client == nil {
		return nil, fmt.Errorf("redis connection failed")
	}

	go func() {
		defer func() {
			close(user_chan)
			close(err_chan)
		}()

		loginKey := fmt.Sprintf("login_info:%s", *userId)

		r_data, err := redis_client.HGetAll(loginKey).Result()
		if err != nil {
			err_chan <- fmt.Errorf("error fetching user from redis: %v", err)
			return
		}

		if r_data["user_id"] != *userId {
			err_chan <- fmt.Errorf("invalid userId provided")
			return
		}

		objID, err := primitive.ObjectIDFromHex(*userId)
		if err != nil {
			err_chan <- fmt.Errorf("invalid objectId: %v", err)
			return
		}

		// find the user in MongoDB
		var user model.User
		filter := bson.M{"_id": objID}
		err = u.db.Database("go-ecomm").Collection("users").FindOne(ctx, filter).Decode(&user)
		if err != nil {
			err_chan <- fmt.Errorf("user not found: %v", err)
			return
		}

		// update fields if present
		updateData := bson.M{}
		if user_data.Username != nil {
			user.Username = *user_data.Username
			updateData["username"] = *user_data.Username
		}
		if user_data.FirstName != nil {
			user.FirstName = *user_data.FirstName
			updateData["firstName"] = *user_data.FirstName
		}
		if user_data.LastName != nil {
			user.LastName = *user_data.LastName
			updateData["lastName"] = *user_data.LastName
		}
		if user_data.ProfileImage != nil {
			user.ProfileImage = user_data.ProfileImage
			updateData["profileImage"] = *user_data.ProfileImage
		}
		user.UpdatedAt = time.Now()
		updateData["updatedAt"] = user.UpdatedAt

		// apply update in MongoDB
		_, err = u.db.Database("go-ecomm").Collection("users").
			UpdateOne(ctx, filter, bson.M{"$set": updateData})
		if err != nil {
			err_chan <- fmt.Errorf("failed to update user in DB: %v", err)
			return
		}

		// update Redis hash
		redisUpdate := map[string]interface{}{}
		for k, v := range updateData {
			redisUpdate[k] = v
		}
		if len(redisUpdate) > 0 {
			if err := redis_client.HMSet(loginKey, redisUpdate).Err(); err != nil {
				err_chan <- fmt.Errorf("failed to update redis: %v", err)
				return
			}
		}

		user_chan <- &user
	}()

	select {
	case user := <-user_chan:
		return user, nil
	case err := <-err_chan:
		return nil, err
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

// Delete the UserProfile
func (u *UserServiceStruct) DeleteMyProfile(userId string) (*model.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	user_chan := make(chan *model.User, 32)
	err_chan := make(chan error, 32)

	// find and chk the userid is exis or not
	redis_client := utils.GetRedis()
	if redis_client == nil {
		err_chan <- fmt.Errorf("redis connection failed")
	}

	loginKey := fmt.Sprintf("login_info:%s", userId)

	user_Object_id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		err_chan <- err
	}
	var user *model.User

	var wg sync.WaitGroup
	wg.Add(1)
	// to delete the user details from the redis
	go func() {
		defer func() {
			wg.Done()
		}()

		// delete the detilas from the Redis
		err := redis_client.Del(loginKey).Err()
		if err != nil {
			err_chan <- fmt.Errorf("error deleting user data: %v", err)
			return
		}

		fmt.Println("User data deleted successfully from redis")
	}()

	wg.Add(1)
	// to delete the user details from the MongoDb
	go func() {
		defer func() {
			wg.Done()
		}()

		fmt.Println("Inside the mongo deletion")
		delete_user_filter := bson.M{
			"_id": user_Object_id,
		}

		err := u.db.Database("go-ecomm").Collection("users").FindOneAndDelete(ctx, delete_user_filter).Decode(&user)
		if err != nil {
			err_chan <- err
			return
		}
		fmt.Println("mongo_del_res")

		user_chan <- user

	}()
	wg.Wait()

	select {
	case user_data := <-user_chan:
		return user_data, nil
	case err := <-err_chan:
		return nil, err
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

// Get any User Profile    -----******
func (u *UserServiceStruct) GetUserProfile(userID string) (*model.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	userChan := make(chan *model.User, 32)
	errChan := make(chan error, 32)

	redisClient := utils.GetRedis()
	if redisClient == nil {
		return nil, fmt.Errorf("redis connection failed")
	}

	userKey := fmt.Sprintf("user_info:%s", userID)

	// Check if data exists in Redis
	exist, err := redisClient.Exists(userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("redis exists check failed: %v", err)
	}

	if exist == 0 {
		// Cache miss → fetch from MongoDB
		go func() {
			defer close(userChan)
			defer close(errChan)

			userObjectID, err := primitive.ObjectIDFromHex(userID)
			if err != nil {
				errChan <- fmt.Errorf("invalid user ID: %v", err)
				return
			}

			toFindUser := bson.M{"_id": userObjectID}
			var user model.User

			err = u.db.Database("go-ecomm").
				Collection("users").
				FindOne(ctx, toFindUser).
				Decode(&user)
			if err != nil {
				errChan <- err
				return
			}

			// Convert to JSON for Redis
			userBytes, err := json.Marshal(user)
			if err != nil {
				errChan <- fmt.Errorf("failed to marshal user: %v", err)
				return
			}

			if err := redisClient.Set(userKey, userBytes, 1*time.Hour).Err(); err != nil {
				errChan <- fmt.Errorf("failed to set user in redis: %v", err)
				return
			}

			userChan <- &user
		}()
	} else {
		// Cache hit → read from Redis
		go func() {
			defer close(userChan)
			defer close(errChan)

			userBytes, err := redisClient.Get(userKey).Bytes()
			if err != nil {
				errChan <- fmt.Errorf("failed to get user from redis: %v", err)
				return
			}

			var user model.User
			if err := json.Unmarshal(userBytes, &user); err != nil {
				errChan <- fmt.Errorf("failed to unmarshal user: %v", err)
				return
			}

			fmt.Println("From Redis Cache")
			userChan <- &user
		}()
	}

	select {
	case userData := <-userChan:
		return userData, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}

}

// Get Random n no. of users
func (u *UserServiceStruct) GetRandomUsers(userNum int) (*[]model.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	users_chan := make(chan []model.User, 32)
	err_chan := make(chan error, 32)

	// search in Redis if not exists then  go for MongoDB
	redis_client := utils.GetRedis()
	var users []model.User
	if redis_client == nil {
		err_chan <- fmt.Errorf("redis connection failed")
	}

	// Search in Redis
	getUsersFromRedis := redis_client.LLen("random_users").Val()
	fmt.Println("getUsersFromRedis")
	fmt.Println(getUsersFromRedis)

	if getUsersFromRedis > 0 {
		go func() {
			fmt.Println("From Redis")
			getUsersFromRedisArr, err := redis_client.LRange("random_users", 0, int64(userNum)).Result()
			if err != nil {
				err_chan <- err
				return
			}
			fmt.Println("getUsersFromRedisArr")
			fmt.Println(getUsersFromRedisArr)

			for _, user := range getUsersFromRedisArr {
				var get_user model.User
				err := json.Unmarshal([]byte(user), &get_user)
				if err != nil {
					err_chan <- err
					return
				}
				fmt.Println("Get Usrere user")
				fmt.Println(get_user)
				users = append(users, get_user)
			}
			fmt.Println("*users")
			fmt.Println(users)
			users_chan <- users
		}()
	} else {
		fmt.Println("From MongoDB")

		to_search_random_users := bson.M{
			"$sample": bson.M{
				"size": userNum,
			},
		}

		go func() {
			cur, err := u.db.Database("go-ecomm").Collection("users").Aggregate(ctx, bson.A{
				to_search_random_users,
			})

			if cur.Err() != nil {
				err_chan <- cur.Err()
				return
			}
			defer cur.Close(ctx)

			if err != nil {
				fmt.Println("Error from get users from the database MongoDB")
				err_chan <- err
				return
			}

			isKeExists, err := redis_client.Exists("random_users").Result()
			if err != nil {
				err_chan <- err
				return
			}

			fmt.Printf("isKeyeEissts Res :- %v\n", isKeExists)

			del_res, err := redis_client.Del("random_users").Result()
			if err != nil {
				err_chan <- err
				return
			}

			fmt.Printf("Del Res :- %v\n", del_res)

			for cur.Next(ctx) {
				var user model.User
				err := cur.Decode(&user)
				if err != nil {
					err_chan <- err
				}
				json_user, err := json.Marshal(user)
				if err != nil {
					err_chan <- err
				}
				lpush_result, err := redis_client.LPush("random_users", string(json_user)).Result()
				if err != nil {
					err_chan <- err
				}
				fmt.Printf("LPush Res :- %v\n", lpush_result)
				redis_client.LPush("random_users", user)
				(users) = append(users, user)
			}

			// Set TTl for List
			err = redis_client.Expire("random_users", time.Second*10).Err()
			if err != nil {
				err_chan <- err
				return
			}

			users_chan <- users
		}()
	}

	select {
	case users_data := <-users_chan:
		fmt.Println("Data received: ", users_data)
		return &users_data, nil
	case err := <-err_chan:
		return nil, err
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

// Get Recently joined Users
func (u *UserServiceStruct) GetRecentlyJoinedUsers(userNum int, userId string) (*[]model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	users_chan := make(chan []model.User, 32)
	err_chan := make(chan error, 32)

	// search in Redis if not exists then  go for MongoDB
	redis_client := utils.GetRedis()
	var users []model.User
	if redis_client == nil {
		err_chan <- fmt.Errorf("redis connection failed")
	}

	// Search in Redis
	joined_users, err := redis_client.SCard("recently_joined_users").Result()
	if err != nil {
		err_chan <- fmt.Errorf("getting data from redis is failed: %v", err)
	}
	fmt.Println("joined_users")
	fmt.Println(joined_users)

	user_obj_id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		err_chan <- fmt.Errorf("parsing user id from hex is failed: %v", err)
	}

	if joined_users > 0 {
		go func() {
			fmt.Println("From Redis")
			getUsersFromRedisArr, err := redis_client.SMembers("recently_joined_users").Result()
			if err != nil {
				err_chan <- err
				return
			}
			for _, user := range getUsersFromRedisArr {
				var get_user model.User
				err := json.Unmarshal([]byte(user), &get_user)
				if err != nil {
					err_chan <- err
					return
				}
				(users) = append(users, get_user)
			}
			users_chan <- users
		}()
	} else {
		fmt.Println("From MongoDB")

		// not me
		// sort the users by createdAt : Desc
		to_find_latest_joined_users := bson.A{
			bson.M{
				"$match": bson.M{
					"_id": bson.M{
						"$ne": user_obj_id,
					},
				},
			},
			bson.M{
				"$sort": bson.M{"createdAt": -1},
			},
			bson.M{
				"$limit": userNum,
			},
		}

		go func() {
			cur, err := u.db.Database("go-ecomm").Collection("users").Aggregate(ctx, to_find_latest_joined_users)
			if err != nil {
				fmt.Println("Error from get users from the database MongoDB")
				err_chan <- err
				return
			}

			for cur.Next(ctx) {
				var user *model.User
				err := cur.Decode(&user)
				if err != nil {
					err_chan <- err
				}
				json_user, err := json.Marshal(*user)
				if err != nil {
					err_chan <- err
				}
				if err != nil {
					fmt.Println("Error in Decoding the User: ", err)
					err_chan <- err
					return
				}

				// Adding the user to our set
				res, err := redis_client.SAdd("recently_joined_users", string(json_user)).Result()
				if err != nil {
					err_chan <- fmt.Errorf("error in adding user to redis set: %v", err)
					return
				}
				fmt.Printf("SAdd Res :- %v\n", res)
				(users) = append(users, *user)

				// Setting the TTL for Reently users Set in Redis
				_, err = redis_client.Expire("recently_joined_users", time.Second*6).Result()
				if err != nil {
					err_chan <- fmt.Errorf("error in setting expiration for redis set: %v", err)
					return
				}
				users_chan <- users
			}
		}()
	}

	for {
		select {
		case users_data := <-users_chan:
			return &users_data, nil
		case err := <-err_chan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}
}

// search for user using its username email (not me)
func (u *UserServiceStruct) SearchUser(query, userId string) (*[]model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	users_chan := make(chan []model.User, 32)
	err_chan := make(chan error, 32)

	var users []model.User

	// not me
	obj_user_id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		err_chan <- fmt.Errorf("parsing user id from hex is failed: %v", err)
	}

	toQueryUser := bson.M{
		"$and": bson.A{
			bson.M{"_id": bson.M{"$ne": obj_user_id}}, // exclude current user
			bson.M{"$or": bson.A{
				bson.M{"username": bson.M{"$regex": query, "$options": "i"}},
				bson.M{"email": bson.M{"$regex": query, "$options": "i"}},
				bson.M{"firstname": bson.M{"$regex": query, "$options": "i"}},
				bson.M{"lastname": bson.M{"$regex": query, "$options": "i"}},
				bson.M{"gender": query},
			}},
		},
	}

	// search in Direct mongoDb
	go func() {
		cur, err := u.db.Database("go-ecomm").Collection("users").Find(ctx, toQueryUser)
		if err != nil {
			fmt.Println("Error from get users from the database MongoDB")
			err_chan <- err
			return
		}

		defer cur.Close(ctx)

		for cur.Next(ctx) {
			var user model.User
			err := cur.Decode(&user)
			if err != nil {
				err_chan <- err
				return
			}
			users = append(users, user)
		}

		users_chan <- users
	}()

	for {
		select {
		case users_data := <-users_chan:
			return &users_data, nil
		case err := <-err_chan:
			return nil, err
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}

}
