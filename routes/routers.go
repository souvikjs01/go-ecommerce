package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/souvikjs01/go-ecommerce/handlers"
	"github.com/souvikjs01/go-ecommerce/services"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(db *mongo.Client) *gin.Engine {
	router := gin.Default()
	// CORS Setup
	conf := cors.DefaultConfig()
	conf.AllowAllOrigins = true
	conf.AllowCredentials = true
	conf.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}

	router.Use(cors.New(conf))

	// Setup Prometheus
	// middlewares.PrometheusInit()

	// services
	authService := services.NewAuthService(db)

	// handlers
	authhandler := handlers.NewAuthHandler(authService)

	// Public Routes  -- *** Modification ***
	publicAuthRoute := router.Group("/api/v1/auth")
	// publicAuthRoute.Use(middlewares.Rate_lim())
	{
		publicAuthRoute.POST("/signup", authhandler.Signup)
		publicAuthRoute.POST("/login", authhandler.Login)
		publicAuthRoute.GET("/logout", authhandler.Logout)
	}

	return router
}
