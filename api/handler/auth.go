package handler

import (
	"fmt"
	"task/api/models"
	check "task/pkg/validation"

	"net/http"

	"github.com/gin-gonic/gin"
)

// UserRegister godoc
// @Router       /task/api/v1/user/sendcode [POST]
// @Summary      Sending otp to register
// @Description  Registering to System
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        register body models.UserRegisterRequest true "register"
// @Success      201  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      404  {object}  models.Response
// @Failure      500  {object}  models.Response
func (h *Handler) SendCode(c *gin.Context) {
	loginReq := models.UserRegisterRequest{}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		handleResponseLog(c, h.log, "error while binding body", http.StatusBadRequest, err)
		return
	}
	fmt.Println("loginReq: ", loginReq)

	if err := check.ValidatePhoneNumber(loginReq.MobilePhone); err != nil {
		handleResponseLog(c, h.log, "error while validating phone number: "+loginReq.MobilePhone, http.StatusBadRequest, err.Error())
		return
	}

	otp, err := h.service.Auth().UserRegister(c.Request.Context(), loginReq)
	if err != nil {
		handleResponseLog(c, h.log, "error while sending sms code to "+loginReq.MobilePhone, http.StatusInternalServerError, err)
		return
	}

	handleResponseLog(c, h.log, "Otp sent successfull", http.StatusOK, otp)
}

// UserRegisterConfirm godoc
// @Router       /task/api/v1/user/registerconfirm [POST]
// @Summary      User register confirmation
// @Description  Registering to System
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        register body models.UserRegisterConfRequest true "register"
// @Success      201  {object}  models.UserLoginResponse
// @Failure      400  {object}  models.Response
// @Failure      404  {object}  models.Response
// @Failure      500  {object}  models.Response
func (h *Handler) Register(c *gin.Context) {
	req := models.UserRegisterConfRequest{}

	if err := c.ShouldBindJSON(&req); err != nil {
		handleResponseLog(c, h.log, "error while binding body", http.StatusBadRequest, err)
		return
	}
	fmt.Println("req: ", req)

	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		handleResponseLog(c, h.log, "error validating phone number", http.StatusBadRequest, err.Error())
		return
	}

	confResp, err := h.service.Auth().UserRegisterConfirm(c.Request.Context(), req)
	if err != nil {
		handleResponseLog(c, h.log, "error while confirming", http.StatusUnauthorized, err.Error())
		return
	}

	handleResponseLog(c, h.log, "Success", http.StatusOK, confResp)

}

// UserLogin godoc
// @Router       /task/api/v1/user/loginrequest [POST]
// @Summary      User login requst
// @Description  Login to System
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        login body models.UserLoginRequest true "login"
// @Success      201  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      404  {object}  models.Response
// @Failure      500  {object}  models.Response
func (h *Handler) UserLogin(c *gin.Context) {
	loginReq := models.UserLoginRequest{}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		handleResponseLog(c, h.log, "error while binding body", http.StatusInternalServerError, err)
		return
	}

	if err := check.ValidatePhoneNumber(loginReq.MobilePhone); err != nil {
		handleResponseLog(c, h.log, "error while validating phone number: "+loginReq.MobilePhone, http.StatusBadRequest, err.Error())
		return
	}

	loginResp, err := h.service.Auth().UserLoginSendOTP(c.Request.Context(), loginReq)
	if err != nil {
		handleResponseLog(c, h.log, "unauthorized", http.StatusUnauthorized, err)
		return
	}

	handleResponseLog(c, h.log, "Success", http.StatusOK, loginResp)

}

// UserLoginByPhoneConfirm godoc
// @Router       /task/api/v1/user/loginconfirm [POST]
// @Summary      Customer login by phone confirmation
// @Description  Login to the system using phone number and OTP
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        login body models.UserLoginPhoneConfirmRequest true "login"
// @Success      200  {object}  models.UserLoginResponse
// @Failure      400  {object}  models.Response
// @Failure      401  {object}  models.Response
// @Failure      500  {object}  models.Response
func (h *Handler) UserLoginByPhoneConfirm(c *gin.Context) {
	var req models.UserLoginPhoneConfirmRequest

	// Bind the incoming request body to req
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("error while binding request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode:  http.StatusBadRequest,
			Description: err.Error(),
		})
		return
	}

	// Validate the phone number
	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		handleResponseLog(c, h.log, "error validating phone number", http.StatusBadRequest, err.Error())
		return
	}

	// Fetch the UserID based on the phone number
	user, err := h.storage.User().GetByPhone(c.Request.Context(), req.MobilePhone)
	if err != nil {
		h.log.Error("error fetching user by phone number: " + err.Error())
		c.JSON(http.StatusUnauthorized, models.Response{
			StatusCode:  http.StatusUnauthorized,
			Description: "User not found or phone number not registered",
		})
		return
	}

	// Check the count of devices associated with the user
	deviceCount, err := h.storage.Device().GetDeviceCount(c.Request.Context(), user.ID)
	if err != nil {
		h.log.Error("error getting device count: " + err.Error())
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode:  http.StatusInternalServerError,
			Description: "Error checking device count",
		})
		return
	}

	// If the user has more than 3 devices, prompt them to delete one
	if deviceCount >= 3 {
		// Get all devices for the user
		devices, err := h.storage.Device().GetAll(c.Request.Context(), user.ID)
		if err != nil {
			h.log.Error("error fetching devices: " + err.Error())
			c.JSON(http.StatusInternalServerError, models.Response{
				StatusCode:  http.StatusInternalServerError,
				Description: "Error fetching devices",
			})
			return
		}

		// Return the devices and ask the user to choose which one to delete
		c.JSON(http.StatusConflict, gin.H{
			"message": "Too many devices. Please remove one device to continue.",
			"devices": devices,
		})
		return
	}

	// Proceed with login process after confirming OTP
	resp, err := h.service.Auth().UserLoginByPhoneConfirm(c.Request.Context(), req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "INTERNAL_SERVER_ERROR"

		if err.Error() == "OTP code not found or expired time" || err.Error() == "Incorrect OTP code" {
			statusCode = http.StatusUnauthorized
			message = err.Error()
		}

		h.log.Error("error in UserLoginByPhoneConfirm: " + err.Error())
		c.JSON(statusCode, models.Response{
			StatusCode:  statusCode,
			Description: message,
		})
		return
	}
	// dev := models.Device{}
	// // Get the device info from the request (for the new device)
	// device := models.Device{
	// 	UserID:     user.ID,
	// 	DeviceInfo: dev.DeviceInfo, // Assuming the client sends the device info
	// }

	// Insert the new device into the database
	// _, err = h.storage.Device().Insert(c.Request.Context(), &device)
	// if err != nil {
	// 	h.log.Error("error inserting device: " + err.Error())
	// 	c.JSON(http.StatusInternalServerError, models.Response{
	// 		StatusCode:  http.StatusInternalServerError,
	// 		Description: "Failed to log device",
	// 	})
	// 	return
	// }

	h.log.Info("Successfully logged in by phone")
	c.JSON(http.StatusOK, resp)
}
