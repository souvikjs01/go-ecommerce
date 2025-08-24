package model

import (
	"encoding/json"
	"time"

	"github.com/souvikjs01/go-ecommerce/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	Gender    string             `json:"gender"`
	Password  string             `json:"password"`
	IsAdmin   bool               `json:"isAdmin"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"UpdatedAt"`
}

type UpdateUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Gender   string `json:"gender"`
}

func NewUser(username, email, gender, password *string) *User {
	hash, _ := utils.HashPassword(*password)
	return &User{
		ID:        primitive.NewObjectID(),
		Username:  *username,
		Email:     *email,
		Gender:    *gender,
		IsAdmin:   false,
		Password:  hash,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (u *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}
