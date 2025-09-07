package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/request"
	"github.com/souvikjs01/go-ecommerce/services"
)

type ProductHandlerStruct struct {
	service services.ProductService
}

func NewProductHandler(service services.ProductService) *ProductHandlerStruct {
	return &ProductHandlerStruct{
		service: service,
	}
}

func (h *ProductHandlerStruct) CreateProductHandler(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	userRole := ctx.GetBool("isAdmin")

	if !userRole {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"success": false,
				"error":   "Unauthorized",
			},
		)
		return
	}

	var product *request.ProductPayload
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error":   err.Error(),
				"success": false,
			},
		)
		return
	}

	product_chan := make(chan *model.Product, 32)
	err_chan := make(chan error, 32)

	go func() {
		product, err := h.service.CreateProduct(userId, product)
		if err != nil {
			err_chan <- err
			return
		}
		product_chan <- product
	}()

	for {
		select {
		case <-ctx.Done():
			ctx.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error":   "request canceled by client",
			})
			return
		case product := <-product_chan:
			ctx.JSON(
				http.StatusOK,
				gin.H{
					"success": true,
					"data":    product,
				},
			)
			return
		case err := <-err_chan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	}
}

func (h *ProductHandlerStruct) UpdateProductHandler(ctx *gin.Context) {

	prod_chan := make(chan *model.Product, 32)
	err_chan := make(chan error, 32)

	userRole := ctx.GetBool("isAdmin")

	if !userRole {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"success": false,
				"error":   "Unauthorized",
			},
		)
		return
	}

	var update_product request.UpdateProductPayload
	if err := ctx.ShouldBindJSON(&update_product); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":   err.Error(),
				"success": false,
			},
		)
		return
	}

	prod_id := ctx.Param("productId")

	go func() {
		prod, err := h.service.UpdateProductsDetails(&prod_id, &update_product)
		if err != nil {
			err_chan <- err
		}
		prod_chan <- prod
	}()

	for {
		select {
		case product := <-prod_chan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    product,
			})
			return
		case err := <-err_chan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	}
}

func (h *ProductHandlerStruct) GetProductDetailsByID(ctx *gin.Context) {
	productId := ctx.Param("productId")

	prodChan := make(chan *model.Product, 32)
	errChan := make(chan error, 32)

	go func() {
		prod, err := h.service.GetProductDetailsByID(productId)

		if err != nil {
			errChan <- err
			return
		}
		prodChan <- prod
	}()

	for {
		select {
		case <-ctx.Done():
			ctx.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error":   "request time out",
			})
			return
		case product := <-prodChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    product,
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

func (h *ProductHandlerStruct) DeleteProduct(ctx *gin.Context) {
	prodChan := make(chan *model.Product, 32)
	errChan := make(chan error, 32)

	prod_id := ctx.Param("productId")
	userId := ctx.GetString("userId")
	userRole := ctx.GetBool("isAdmin")

	if userId == "" {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"success": false,
				"error":   "session expired, login again",
			},
		)
		return
	}

	if !userRole {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"success": false,
				"error":   "Unauthorized",
			},
		)
		return
	}

	go func() {
		product, err := h.service.DeleteProductsDetails(&prod_id)
		if err != nil {
			errChan <- err
			return
		}
		prodChan <- product
	}()

	for {
		select {
		case <-ctx.Done():
			ctx.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error":   "request time out",
			})
			return
		case product := <-prodChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    product,
			})
			return
		case err := <-errChan:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
	}
}

func (h *ProductHandlerStruct) LatestProducts(ctx *gin.Context) {

	productsChan := make(chan *[]model.Product, 32)
	errChan := make(chan error, 32)

	go func() {
		products, err := h.service.GetLatestProducts()
		if err != nil {
			errChan <- err
			return
		}
		productsChan <- products
	}()

	for {
		select {
		case <-ctx.Done():
			ctx.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error":   "request timeout",
			})
			return
		case products := <-productsChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    products,
			})
			return
		case err := <-errChan:
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"success": false,
					"error":   err.Error(),
				},
			)
			return
		}
	}

}

func (h *ProductHandlerStruct) AllProducts(ctx *gin.Context) {

	productsChan := make(chan *[]model.Product, 32)
	errChan := make(chan error, 32)

	go func() {
		products, err := h.service.GetAllProduct()
		if err != nil {
			errChan <- err
			return
		}
		productsChan <- products
	}()

	for {
		select {
		case <-ctx.Done():
			ctx.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error":   "request timeout",
			})
			return
		case products := <-productsChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    products,
			},
			)
			return
		case err := <-errChan:
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"success": false,
					"error":   err.Error(),
				},
			)
			return
		}
	}

}

func (h *ProductHandlerStruct) ProductByQuery(ctx *gin.Context) {
	query := ctx.Query("query")
	prodChan := make(chan *[]model.Product, 32)
	errChan := make(chan error, 32)

	go func() {
		prods, err := h.service.GetProductsByQuery(query)

		if err != nil {
			errChan <- err
			return
		}
		prodChan <- prods
	}()

	for {
		select {
		case <-ctx.Done():
			ctx.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error":   "request timeout",
			})
			return
		case products := <-prodChan:
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    products,
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
