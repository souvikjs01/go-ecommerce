package model

import (
	"encoding/json"
	"time"

	"github.com/souvikjs01/go-ecommerce/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
	Other  Gender = "other"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id" json:"_id"`
	Username     string             `json:"username"`
	FirstName    string             `json:"firstName"`
	LastName     string             `json:"lastName"`
	Email        string             `json:"email"`
	Gender       Gender             `json:"gender"`
	ProfileImage *string            `json:"profileImage"`
	Password     string             `json:"password"`
	IsAdmin      bool               `json:"isAdmin"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"UpdatedAt"`
}

type UpdateUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Gender   string `json:"gender"`
}

func NewUser(username, firstName, lastName, email, gender, profileImage, password *string) *User {
	hash, _ := utils.HashPassword(*password)
	return &User{
		ID:           primitive.NewObjectID(),
		Username:     *username,
		FirstName:    *firstName,
		LastName:     *lastName,
		Email:        *email,
		Gender:       Gender(*gender),
		ProfileImage: profileImage,
		IsAdmin:      false,
		Password:     hash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func (u *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}
