package request

import (
	"github.com/souvikjs01/go-ecommerce/model"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignupRequest struct {
	Username     string  `json:"username" binding:"required"`
	FirstName    string  `json:"firstName" binding:"required"`
	LastName     string  `json:"lastName" binding:"required"`
	Email        string  `json:"email" binding:"required"`
	Password     string  `json:"password" binding:"required,min=4"`
	Gender       string  `json:"gender" binding:"required,oneof=male female other"`
	ProfileImage *string `json:"profileImage"`
}

type UpdateRequest struct {
	Username     *string `json:"username"`
	FirstName    *string `json:"firstName"`
	LastName     *string `json:"lastName"`
	ProfileImage *string `json:"profileImage"`
}

type ProductPayload struct {
	Title      string   `json:"title" binding:"required"`
	Desc       string   `json:"desc" binding:"required"`
	Img        string   `json:"img" binding:"required"`
	Categories []string `json:"categories" binding:"required"`
	Size       []string `json:"size" binding:"required"`
	Color      []string `json:"color" binding:"required"`
	Price      int      `json:"price" binding:"required"`
	InStock    bool     `json:"instock" binding:"required"`
}

type UpdateProductPayload struct {
	Title      *string   `json:"title"`
	Desc       *string   `json:"desc"`
	Img        *string   `json:"img"`
	Categories *[]string `json:"categories"`
	Size       *[]string `json:"size"`
	Color      *[]string `json:"color"`
	Price      *int      `json:"price"`
	InStock    *bool     `json:"instock"`
}

type CreateOrderPayload struct {
	Products []model.ProductInfo `json:"products" binding:"required"`
	Address  string              `json:"address" binding:"required,min=4,max=20"`
	Status   string              `json:"status" binding:"required"`
}

type AddToCartPayload struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}
