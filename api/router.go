package api

import (
	"net/http"
	"strings"
	_ "task/api/docs"
	"task/api/handler"
	"task/api/models"
	"task/config"
	"task/pkg/jwt"
	"task/pkg/logger"
	"task/service"
	"task/storage"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// NewApi initializes and configures the API routes
// @title Swagger Example API
// @version 1.0
// @description This is a sample server.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func NewApi(r *gin.Engine, cfg *config.Config, storage storage.IStorage, logger logger.LoggerI, service service.IServiceManager) {
	// Set trusted proxies to avoid the warning about trusting all proxies
	err := r.SetTrustedProxies([]string{
		"13.228.225.19",  
		"18.142.128.26",  
		"54.254.162.138", 
	})
	if err != nil {
		logger.Fatal("Failed to set trusted proxies", zap.Error(err))
	}

	h := handler.NewStrg(logger, storage, cfg, service)
	// Apply CORS middleware globally
	r.Use(customCORSMiddleware())

	// Define API routes
	v1 := r.Group("/task/api/v1")

	// Auth-related routes
	v1.POST("/user/registerrequest", h.SendCode)
	v1.POST("/user/registerconfirm", h.Register)
	v1.POST("/user/loginrequest", h.UserLogin)
	v1.POST("/user/loginconfirm", h.UserLoginByPhoneConfirm)
	v1.POST("/user/sendotp", JWTAuthMiddleware(), h.RequestOTP)
	v1.POST("/user/confirmotp", JWTAuthMiddleware(), h.ConfirmOTPAndUpdatePhoneNumber)

	// User CRUD
	// v1.POST("/user/create", h.CreateUser)
	v1.GET("/user/getbyid/:id", h.GetUserByID)
	v1.GET("/user/getall", h.GetAllUsers)
	v1.PUT("/user/update/:id", JWTAuthMiddleware(), h.UpdateUser)
	v1.DELETE("/user/delete", JWTAuthMiddleware(), h.DeleteUser)
	v1.DELETE("/user/logout", JWTAuthMiddleware(), h.Logout)

	// Contact CRUD
	v1.POST("/contact/create", JWTAuthMiddleware(), h.CreateContact)
	v1.GET("/contact/getbyid/:id", JWTAuthMiddleware(), h.GetContactsById)
	v1.GET("/contact/getall", JWTAuthMiddleware(), h.GetAllContacts)
	v1.PUT("/contact/update/:id", JWTAuthMiddleware(), h.UpdateContact)
	v1.DELETE("/contact/delete/:id", JWTAuthMiddleware(), h.DeleteContact)

	// Device CRUD
	// v1.POST("/device/create", JWTAuthMiddleware(), h.CreateDevice)
	v1.GET("/device/getlist", JWTAuthMiddleware(), h.GetAllDevices)
	v1.DELETE("/device/delete/:id", JWTAuthMiddleware(), h.DeleteDevice)
	v1.DELETE("/device/remove/:id", h.RemoveDevice)

	// Swagger Documentation Route
	url := ginSwagger.URL("swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
}

// customCORSMiddleware allows cross-origin requests
func customCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE, HEAD")
		c.Header("Access-Control-Allow-Headers", "Platform-Id, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// JWTAuthMiddleware validates JWT tokens and extracts user information
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "missing authorization header"})
			c.Abort()
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "invalid authorization header format", Data: "token should start with Bearer "})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		claims, err := jwt.VerifyJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "invalid or expired token", Data: nil})
			c.Abort()
			return
		}

		// Extract user_id directly as a string (not as a map)
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "invalid token claims: user_id is not a valid string", Data: userID})
			c.Abort()
			return
		}

		c.Set("user_id", userID)

		// Extract device_id from claims and set it if exists
		deviceID, ok := claims["device_id"].(string)
		if ok && deviceID != "" {
			c.Set("device_id", deviceID)
		}

		c.Next()
	}
}
