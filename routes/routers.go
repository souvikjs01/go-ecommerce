package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/souvikjs01/go-ecommerce/handlers"
	"github.com/souvikjs01/go-ecommerce/middlewares"
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
	userService := services.NewUserService(db)
	productService := services.NewProductService(db)
	orderService := services.NewOrderService(db)
	cartService := services.NewCartService(db)

	// handlers
	authhandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	productHandler := handlers.NewProductHandler(productService)
	orderHandler := handlers.NewOrderHandler(orderService)
	cartHandler := handlers.NewCartHandler(cartService)

	// Public Routes  -- *** Modification ***
	publicAuthRoute := router.Group("/api/v1/auth")
	publicAuthRoute.Use(middlewares.Rate_lim())
	{
		publicAuthRoute.POST("/signup", authhandler.Signup)
		publicAuthRoute.POST("/login", authhandler.Login)
		publicAuthRoute.GET("/logout", authhandler.Logout)
	}

	// Private Routes    user routes
	user_private_routes := router.Group("/api/v1/user")
	user_private_routes.Use(middlewares.RequireAuth())
	user_private_routes.Use(middlewares.Rate_lim())
	{
		user_private_routes.GET("/me", userHandler.GetMyProfile)
		user_private_routes.PUT("/update_me", userHandler.UpdateUserProfile)
		user_private_routes.DELETE("/delete_me", userHandler.DeleteUserProfile)
		user_private_routes.GET("/user_info/:userID", userHandler.GetUserFromUserID)
		// // TODO := to fix and Work this and also chk this
		// user_private_routes.GET("/random_users", userHandler.GetRandomUsersHandler)
		// user_private_routes.GET("/recent_users", userHandler.GetRecentUsers)
		user_private_routes.GET("/query_user", userHandler.SearchForUsers)
	}

	// Product Routes
	public_product_routes := router.Group("/api/v1/product")
	public_product_routes.Use(middlewares.Rate_lim())
	{
		public_product_routes.GET("/query", productHandler.ProductByQuery)
		public_product_routes.GET("/:productId", productHandler.GetProductDetailsByID)
		public_product_routes.GET("/latest", productHandler.LatestProducts)
		public_product_routes.GET("/all", productHandler.AllProducts)
	}

	private_product_routes := router.Group("/api/v1/products")
	private_product_routes.Use(middlewares.RequireAuth())
	private_product_routes.Use(middlewares.Rate_lim())
	{
		private_product_routes.POST("/create-product", productHandler.CreateProductHandler)
		private_product_routes.PUT("/update-product/:productId", productHandler.UpdateProductHandler)
		private_product_routes.DELETE("/delete-product/:productId", productHandler.DeleteProduct)
	}

	// order routes
	order_Routes := router.Group("/api/v1/orders")
	order_Routes.Use(middlewares.RequireAuth())
	order_Routes.Use(middlewares.Rate_lim())
	{
		order_Routes.POST("/create-order", orderHandler.CreateOrderHandler)
		order_Routes.GET("/user-orders", orderHandler.GetUserOrdersHandler)
		// order_Routes.GET("/orders", orderHandler.GetOrdersHandler)
		// order_Routes.DELETE("/order/:orderId", orderHandler.DeleteOrderHandler)
		// order_Routes.PUT("/order/:orderId", orderHandler.UpdateOrderHandler)
	}

	// cart routes
	cart_routes := router.Group("/api/v1/cart")
	cart_routes.Use(middlewares.RequireAuth())
	cart_routes.Use(middlewares.Rate_lim())
	{
		cart_routes.POST("/add-to-cart", cartHandler.AddToCartHandler)
		cart_routes.GET("/my-cart", cartHandler.GetMyCart)
		cart_routes.DELETE("/delete-my-cart/:cartId", cartHandler.DeleteCartHandler)
		cart_routes.GET("/all-carts", cartHandler.GetCartshandler)
		cart_routes.PUT("/update-cart/:cartID", cartHandler.UpdateCarthandler)
	}

	return router
}
