package api

import (
	"fmt"
	"net/http"
	"strings"
	_ "task/api/docs"
	"task/api/handler"
	"task/config"
	"task/pkg/jwt"
	"task/pkg/logger"
	"task/service"
	"task/storage"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewApi initializes and configures the API routes
// @title Swagger Example API
// @version 1.0
// @description This is a sample server.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func NewApi(r *gin.Engine, cfg *config.Config, storage storage.IStorage, logger logger.LoggerI, service service.IServiceManager) {
	h := handler.NewStrg(logger, storage, cfg, service)

	// Apply CORS middleware globally
	r.Use(customCORSMiddleware())

	// Define API routes
	v1 := r.Group("/task/api/v1")

	// Auth-related routes
	v1.POST("/user/sendcode", h.SendCode)
	v1.POST("/user/registerconfirm", h.Register)
	v1.POST("/user/loginrequest", h.UserLogin)
	v1.POST("/user/loginconfirm", h.UserLoginByPhoneConfirm)
	v1.POST("/user/sendotp", JWTAuthMiddleware(), h.RequestOTP)
	v1.POST("/user/confirmotp", JWTAuthMiddleware(), h.ConfirmOTPAndUpdatePhoneNumber)

	// User CRUD
	v1.POST("/createuser", h.CreateUser)
	v1.GET("/getbyiduser/:id", h.GetUserByID)
	v1.GET("/getallusers", h.GetAllUsers)
	v1.PUT("/updateuser/:id", JWTAuthMiddleware(), h.UpdateUser)
	v1.DELETE("/deleteuser/:id", JWTAuthMiddleware(), h.DeleteUser)
	v1.DELETE("/user/logout", JWTAuthMiddleware(), h.Logout)

	// Contact CRUD
	v1.POST("/createcontact", JWTAuthMiddleware(), h.CreateContact)
	v1.GET("/getbyidcontact/:id", JWTAuthMiddleware(), h.GetContactsById)
	v1.GET("/getallcontacts", JWTAuthMiddleware(), h.GetAllContacts)
	v1.PUT("/updatecontact/:id", JWTAuthMiddleware(), h.UpdateContact)
	v1.DELETE("/deletecontact/:id", JWTAuthMiddleware(), h.DeleteContact)

	// Device CRUD
	v1.POST("/devicecreate", JWTAuthMiddleware(), h.CreateDevice)
	v1.GET("/getlistdevices", JWTAuthMiddleware(), h.GetAllDevices)
	v1.DELETE("/deletedevice/:id", JWTAuthMiddleware(), h.DeleteDevice)
	v1.DELETE("/removedevice/:id", h.RemoveDevice)

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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		claims, err := jwt.VerifyJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token", "details": err.Error()})
			c.Abort()
			return
		}

		// Debug: Print claims for inspection
		fmt.Printf("Extracted claims: %+v\n", claims)

		// Extract user_id from the nested structure
		userIDMap, ok := claims["user_id"].(map[string]interface{})
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims: user_id is not a valid map"})
			c.Abort()
			return
		}

		// Extract the id from the nested user_id map
		userID, ok := userIDMap["id"].(string)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims: id not found in user_id"})
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
