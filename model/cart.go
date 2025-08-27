package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type ProductDetails struct {
	ProductID primitive.ObjectID `json:"product_id"`
	Quantity  int                `json:"quantity"`
}

type Cart struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Products []ProductDetails   `json:"productDetails"`
	UserId   primitive.ObjectID `json:"userId"`
}

func NewCart(product *[]ProductDetails, userId *primitive.ObjectID) *Cart {
	return &Cart{
		ID:       primitive.NewObjectID(),
		Products: *product,
		UserId:   *userId,
	}
}
