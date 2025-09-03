package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID         primitive.ObjectID `bson:"_id" json:"_id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Title      string             `json:"title"`
	Desc       string             `json:"desc"`
	Img        string             `json:"img"`
	Categories []string           `json:"categories"`
	Size       []string           `json:"size"`
	Color      []string           `json:"color"`
	Price      int                `json:"price"`
	InStock    bool               `json:"instock"`
}

func NewProduct(title *string, description *string, image *string, categories *[]string, size *[]string, color *[]string, price *int, inStock *bool, userId *primitive.ObjectID) *Product {
	return &Product{
		ID:         primitive.NewObjectID(),
		Title:      *title,
		Desc:       *description,
		Img:        *image,
		Categories: *categories,
		Size:       *size,
		Color:      *color,
		Price:      *price,
		InStock:    *inStock,
		UserID:     *userId,
	}
}
