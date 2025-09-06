package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/request"
	"github.com/souvikjs01/go-ecommerce/services"
)

type OrderHandlerStruct struct {
	services services.OrderService
}

func NewOrderHandler(service services.OrderService) *OrderHandlerStruct {
	return &OrderHandlerStruct{
		services: service,
	}
}

func (h *OrderHandlerStruct) CreateOrderHandler(ctx *gin.Context) {
	var order request.CreateOrderPayload

	userId := ctx.GetString("userId")
	orderChan := make(chan *model.Order, 32)
	errChan := make(chan error, 32)

	if err := ctx.ShouldBindJSON(&order); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	go func() {
		new_order, err := h.services.CreateOrder(&order, userId)
		if err != nil {
			errChan <- err
			return
		}
		orderChan <- new_order
	}()

	for {
		select {
		case <-ctx.Done():
			ctx.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error":   "request time error",
			})
			return
		case order := <-orderChan:
			ctx.JSON(http.StatusCreated, gin.H{
				"success": true,
				"order":   order,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	}
}

func (h *OrderHandlerStruct) GetUserOrdersHandler(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	orderChan := make(chan *[]model.Order, 32)
	errChan := make(chan error, 32)

	go func() {
		user_order, err := h.services.GetUserOrders(userID)
		if err != nil {
			errChan <- err
		}
		orderChan <- user_order
	}()

	for {
		select {
		case <-ctx.Done():
			ctx.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error":   "request time out",
			})
			return
		case order := <-orderChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"orders":  order,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	}
}

// func (h *OrderHandlerStruct) GetOrdersHandler(ctx *gin.Context) {
// 	orderChan := make(chan *[]model.Order, 32)
// 	errChan := make(chan error, 32)

// 	go func() {
// 		orders, err := h.services.GetAllOrders()
// 		if err != nil {
// 			errChan <- err
// 			return
// 		}
// 		orderChan <- orders
// 	}()

// 	for {
// 		select {
// 		case orders := <-orderChan:
// 			ctx.JSON(http.StatusOK, gin.H{
// 				"success": true,
// 				"orders":  orders,
// 			})
// 			return
// 		case err := <-errChan:
// 			ctx.JSON(http.StatusInternalServerError, gin.H{
// 				"success": false,
// 				"error":   err.Error(),
// 			})
// 			return
// 		}
// 	}
// }

// func (h *OrderHandlerStruct) DeleteOrderHandler(ctx *gin.Context) {
// 	orderChan := make(chan *model.Order, 32)
// 	errChan := make(chan error, 32)

// 	userId := ctx.GetString("userId")
// 	orderId := ctx.Param("orderId")

// 	go func() {
// 		order, err := h.services.DeleteUserOrder(userId, orderId)
// 		if err != nil {
// 			errChan <- err
// 			return
// 		}
// 		orderChan <- order
// 	}()

// 	for {
// 		select {
// 		case order := <-orderChan:
// 			ctx.JSON(http.StatusOK, gin.H{
// 				"success": true,
// 				"order":   order,
// 			})
// 			return
// 		case err := <-errChan:
// 			ctx.JSON(http.StatusInternalServerError, gin.H{
// 				"error":   err.Error(),
// 				"success": false,
// 			})
// 			return
// 		}
// 	}
// }

// func (h *OrderHandlerStruct) UpdateOrderHandler(ctx *gin.Context) {
// 	var order model.Order
// 	userId := ctx.GetString("userId")
// 	orderID := ctx.Param("orderId")

// 	orderChan := make(chan *model.Order, 32)
// 	errChan := make(chan error, 32)

// 	if err := ctx.ShouldBindJSON(&order); err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error":   err.Error(),
// 			"success": false,
// 		})
// 		return
// 	}

// 	go func() {
// 		updated_order, err := h.services.UpdateOrderDetails(&order, userId, orderID)
// 		if err != nil {
// 			errChan <- err
// 			return
// 		}
// 		orderChan <- updated_order
// 	}()

// 	for {
// 		select {
// 		case order := <-orderChan:
// 			ctx.JSON(http.StatusCreated, gin.H{
// 				"success": true,
// 				"order":   order,
// 			})
// 			return
// 		case err := <-errChan:
// 			ctx.JSON(http.StatusInternalServerError, gin.H{
// 				"error":   err.Error(),
// 				"success": false,
// 			})
// 			return
// 		}
// 	}
// }
