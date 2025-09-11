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
