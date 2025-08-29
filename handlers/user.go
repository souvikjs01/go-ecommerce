package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/souvikjs01/go-ecommerce/model"
	"github.com/souvikjs01/go-ecommerce/services"
)

type UserHandlerStruct struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandlerStruct {
	return &UserHandlerStruct{
		service: service,
	}
}

// Get My Profile
func (h *UserHandlerStruct) GetMyProfile(ctx *gin.Context) {
	uId := ctx.GetString("userId")

	user_chan := make(chan *model.User, 32)
	err_chan := make(chan error, 32)

	go func() {
		defer func() {
			close(user_chan)
			close(err_chan)
		}()

		user, err := h.service.GetMyProfile(uId)
		if err != nil {
			err_chan <- err
			return
		}
		user_chan <- user

	}()

	select {
	case user := <-user_chan:
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message": "User Profile", "data": user,
			},
		)
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error fetching user profile", "error": err.Error(),
			},
		)
	}

}

// Update the UserProfile
func (h *UserHandlerStruct) UpdateUserProfile(ctx *gin.Context) {
	var update_user model.UpdateUser
	if err := ctx.ShouldBindJSON(&update_user); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":   err.Error(),
				"success": false,
			},
		)
		return
	}

	userId := ctx.GetString("userId")
	user, err := h.service.UpdateMyProfile(&update_user, &userId)

	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error updating user profile",
				"error":   err.Error(),
			},
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    user,
		},
	)
}

// Delete the Userprofile
func (h *UserHandlerStruct) DeleteUserProfile(ctx *gin.Context) {
	userId := ctx.GetString("userId")

	deleted_user_info_chan := make(chan *model.User, 32)
	err_chan := make(chan error, 32)

	go func() {
		user, err := h.service.DeleteMyProfile(userId)
		if err != nil {
			err_chan <- err
			return
		}

		deleted_user_info_chan <- user
	}()

	select {
	case user := <-deleted_user_info_chan:
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message": "User profile deleted successfully",
				"data":    *user,
			},
		)
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error deleting user profile",
				"error":   err.Error(),
			},
		)
	}
}

// Get the userProfile from given ID parameter
func (h *UserHandlerStruct) GetUserFromUserID(ctx *gin.Context) {
	userID := ctx.Param("userID")

	get_user_data := make(chan *model.User, 32)
	err_data := make(chan error, 32)

	go func() {
		user, err := h.service.GetUserProfile(userID)
		if err != nil {
			err_data <- err
			return
		}
		get_user_data <- user
	}()

	select {
	case user := <-get_user_data:
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message": "User profile",
				"data":    user,
			})
	case err := <-err_data:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error fetching user profile",
				"error":   err.Error(),
			})
	}

}

// TODO := to fix and Work this and also chk this

// Getting the n number of users
func (h *UserHandlerStruct) GetRandomUsersHandler(ctx *gin.Context) {
	userNum := 2

	get_random_users := make(chan *[]model.User, 32)
	err_chan := make(chan error, 32)
	go func() {
		users, err := h.service.GetRandomUsers(userNum)

		if err != nil {
			err_chan <- err
			return
		}
		get_random_users <- users
	}()

	select {
	case users := <-get_random_users:
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message": "Random users",
				"data":    users,
			},
		)
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error fetching random users",
				"error":   err.Error(),
			},
		)
	}
}

// Getting the Recently joined users
func (h *UserHandlerStruct) GetRecentUsers(ctx *gin.Context) {
	userNum := 2
	userId := ctx.GetString("userId")
	get_recently_joined_users := make(chan *[]model.User, 32)
	err_chan := make(chan error, 32)

	go func() {
		users, err := h.service.GetRecentlyJoinedUsers(userNum, userId)

		if err != nil {
			err_chan <- err
			return
		}
		get_recently_joined_users <- users
	}()

	select {
	case users := <-get_recently_joined_users:
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"message": "Random users",
				"data":    users,
			},
		)
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error fetching recently joined users",
				"error":   err.Error(),
			},
		)
	}
}

// Search for user using its username email (not me)
func (h *UserHandlerStruct) SearchForUsers(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	query := ctx.Query("query")

	users_chan := make(chan *[]model.User, 32)
	err_chan := make(chan error, 32)

	go func() {
		users, err := h.service.SearchUser(query, userId)

		if err != nil {
			err_chan <- err
			return
		}
		users_chan <- users
	}()

	select {
	case users := <-users_chan:
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"data": users,
			},
		)
	case err := <-err_chan:
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Error fetching finding users",
				"error":   err.Error(),
			},
		)
	}

}
