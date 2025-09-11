package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/request"
	"github.com/souvikjs01/go-ecommerce/services"
)

type CartHandlerStruct struct {
	service services.CartService
}

func NewCartHandler(service services.CartService) *CartHandlerStruct {
	return &CartHandlerStruct{
		service: service,
	}
}

func (h *CartHandlerStruct) AddToCartHandler(ctx *gin.Context) {
	var cart request.AddToCartPayload
	userId := ctx.GetString("userId")

	if err := ctx.ShouldBindJSON(&cart); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"success": false,
				"err":     fmt.Errorf("error in geting the cart data"),
			},
		)
		return
	}

	cartChan := make(chan *model.Cart, 32)
	errChan := make(chan error, 32)

	go func() {
		cart, err := h.service.AddToCart(&cart, userId)
		if err != nil {
			errChan <- err
			return
		}
		cartChan <- cart
	}()

	for {
		select {
		case <-ctx.Done():
			ctx.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error":   "request time out",
			})
			return
		case cart := <-cartChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"cart":    cart,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"err":     err.Error(),
			})
			return
		}
	}

}

func (h *CartHandlerStruct) GetMyCart(ctx *gin.Context) {
	userId := ctx.GetString("userId")

	cartChan := make(chan *model.Cart, 32)
	errChan := make(chan error, 32)

	go func() {
		cart, err := h.service.GetCartDetails(userId)
		if err != nil {
			errChan <- err
			return
		}
		cartChan <- cart
	}()

	for {
		select {
		case cart := <-cartChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"cart":    cart,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"err":     err.Error(),
			})
			return
		}
	}

}

func (h *CartHandlerStruct) DeleteCartHandler(ctx *gin.Context) {
	cartID := ctx.Param("cartId")
	userId := ctx.GetString("userId")

	cartDetailsChan := make(chan *model.Cart, 32)
	errChan := make(chan error, 32)

	go func() {
		cart, err := h.service.DeleteCart(userId, cartID)
		if err != nil {
			errChan <- err
			return
		}
		cartDetailsChan <- cart
	}()

	for {
		select {
		case cart := <-cartDetailsChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"cart":    cart,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"err":     err.Error(),
			})
			return
		}
	}
}

func (h *CartHandlerStruct) GetCartshandler(ctx *gin.Context) {
	cartsChan := make(chan *[]model.Cart, 32)
	errChan := make(chan error, 32)

	go func() {
		carts, err := h.service.GetAllCarts()
		if err != nil {
			errChan <- err
			return
		}
		cartsChan <- carts
	}()

	for {
		select {
		case carts := <-cartsChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"cart":    carts,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"err":     err.Error(),
			})
			return
		}
	}
}

func (h *CartHandlerStruct) UpdateCarthandler(ctx *gin.Context) {

	cartChan := make(chan *model.Cart, 32)
	errChan := make(chan error, 32)

	userId := ctx.GetString("userId")
	cartID := ctx.Param("cartID")
	var new_cart model.Cart

	if err := ctx.ShouldBindJSON(&new_cart); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success": false,
				"err":     fmt.Errorf("error in getting the cart data"),
			},
		)
		return
	}
	go func() {

		updated_cart, err := h.service.UpdateCart(&new_cart, userId, cartID)
		if err != nil {
			errChan <- err
			return
		}
		cartChan <- updated_cart
	}()

	for {
		select {
		case cart := <-cartChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"cart":    cart,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"err":     err.Error(),
			})
			return
		}
	}
}
