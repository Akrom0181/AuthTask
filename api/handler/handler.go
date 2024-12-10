package handler

import (
	"net/http"
	"task/api/models"
	"task/config"
	"task/service"

	// "food/pkg/jwt"
	"strconv"
	"task/pkg/logger"
	"task/storage"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	log     logger.LoggerI
	storage storage.IStorage
	service service.IServiceManager
	cfg     *config.Config
}

type Response struct {
	Status      int         `json:"status"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
	Error       interface{} `json:"error"`
}

type ErrorResponse struct {
	Error interface{} `json:"error"`
}

func NewStrg(log logger.LoggerI, strg storage.IStorage, cfg *config.Config, service service.IServiceManager) *Handler {
	return &Handler{
		log:     log,
		storage: strg,
		service: service,
		cfg:     cfg,
	}
}

func handleResponseLog(c *gin.Context, log logger.LoggerI, msg string, statusCode int, data interface{}) {
	resp := models.Response{}

	if statusCode >= 100 && statusCode <= 199 {
		resp.Description = config.ERR_INFORMATION
	} else if statusCode >= 200 && statusCode <= 299 {
		resp.Description = config.SUCCESS
		log.Info("REQUEST SUCCEEDED", logger.Any("msg: ", msg), logger.Int("status: ", statusCode))

	} else if statusCode >= 300 && statusCode <= 399 {
		resp.Description = config.ERR_REDIRECTION
	} else if statusCode >= 400 && statusCode <= 499 {
		resp.Description = config.ERR_BADREQUEST
		log.Error("!!!!!!!! BAD REQUEST !!!!!!!!", logger.Any("error: ", msg), logger.Int("status: ", statusCode))
	} else {
		resp.Description = config.ERR_INTERNAL_SERVER
		log.Error("!!!!!!!! ERR_INTERNAL_SERVER !!!!!!!!", logger.Any("error: ", msg), logger.Int("status: ", statusCode))
	}

	resp.StatusCode = statusCode
	resp.Data = data

	c.JSON(resp.StatusCode, resp)
}

func Welcome(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found"})
		return
	}

	switch role {
	case "admin":
		c.JSON(http.StatusOK, gin.H{"message": "Welcome, Admin"})
	case "customer":
		c.JSON(http.StatusOK, gin.H{"message": "Hello, Customer"})
	case "barber":
		c.JSON(http.StatusOK, gin.H{"message": "Greetings, Barber"})
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
	}
}

func ParsePageQueryParam(c *gin.Context) (uint64, error) {
	pageStr := c.Query("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, err := strconv.ParseUint(pageStr, 10, 30)
	if err != nil {
		return 0, err
	}

	if page == 0 {
		return 1, nil
	}

	return page, nil
}

func ParseLimitQueryParam(c *gin.Context) (uint64, error) {
	limitStr := c.Query("limit")
	if limitStr == "" {
		limitStr = "2"
	}

	limit, err := strconv.ParseUint(limitStr, 10, 30)
	if err != nil {
		return 0, err
	}

	if limit == 0 {
		return 2, nil
	}

	return limit, nil
}

// func getAuthInfo(c *gin.Context) (models.AuthInfo, error) {
// 	accessToken := c.GetHeader("Authorization")
// 	if accessToken == "" {
// 		return models.AuthInfo{}, errors.New("unauthorized")
// 	}

// 	m, err := jwt.ExtractClaims(accessToken)
// 	if err != nil {
// 		return models.AuthInfo{}, err
// 	}

// 	role := m["user_role"].(string)
// 	if !(role == config.ADMIN_ROLE || role == config.USER_ROLE) {
// 		return models.AuthInfo{}, errors.New("unauthorized")
// 	}

// 	return models.AuthInfo{
// 		UserID:   m["user_id"].(string),
// 		UserRole: role,
// 	}, nil
// }
