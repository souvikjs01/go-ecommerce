package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/response"
	"github.com/souvikjs01/go-ecommerce/services"
	"github.com/souvikjs01/go-ecommerce/utils"
)

type AuthInterface interface{}

// Dependency injection of Services
type AuthHandlerStruct struct {
	services services.AuthService
}

func NewAuthHandler(services services.AuthService) *AuthHandlerStruct {
	return &AuthHandlerStruct{
		services: services,
	}
}

// Signup handler
func (h *AuthHandlerStruct) Signup(ctx *gin.Context) {
	var user *model.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error":   errors.New("invalid credentials"),
			},
		)
		return
	}

	user_chan := make(chan *model.User, 32)
	err_chan := make(chan error, 32)

	go func() {
		user, err := h.services.SignUpService(user)
		if err != nil {
			err_chan <- err
			return
		}
		user_chan <- user
	}()

	select {
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error":   err.Error(),
			},
		)
	case result_user := <-user_chan:
		ctx.JSON(
			http.StatusAccepted,
			gin.H{
				"success": true,
				"data":    result_user,
			},
		)
	}
}

// Login Handler
func (h *AuthHandlerStruct) Login(ctx *gin.Context) {
	var payload response.LoginResponse
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error":   errors.New("invalid credentials"),
			},
		)
		return
	}

	// res_chan := make(chan *response.LoginResponse, 32)
	user_chan := make(chan *model.User, 32)
	err_chan := make(chan error, 32)

	go func() {
		res, token, err := h.services.LoginService(payload)
		if err != nil {
			err_chan <- err
			return
		}
		ctx.SetCookie("authCookie_golang", token, 3600, "/", "localhost", false, true) // 3600 in seconds
		user_chan <- res
	}()
	select {
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success": false,
				"error":   err.Error(),
			},
		)
	case res_response := <-user_chan:
		ctx.JSON(
			http.StatusAccepted,
			gin.H{
				"success": true,
				"data":    res_response,
			},
		)
	}
}

// Logout Hanler
func (h *AuthHandlerStruct) Logout(ctx *gin.Context) {
	// clear the cookie
	ctx.SetCookie("authCookie_golang", "", -1, "/", "localhost", false, true)

	// clear data from Redis database
	redis_client := utils.GetRedis()
	redis_client.Del("login_info:username")
	redis_client.Del("login_info:user_id")
	redis_client.Del("login_info:email")
	redis_client.Del("login_info:isAdmin")

	// and then return from that
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged Out Successfully!",
	})
}
