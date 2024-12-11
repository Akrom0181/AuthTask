package handler

import (
	"fmt"
	"task/api/models"
	check "task/pkg/validation"

	"net/http"

	"github.com/gin-gonic/gin"
)

// UserRegister godoc
// @Router       /task/api/v1/user/registerrequest [POST]
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

	// Bind incoming JSON to the request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		handleResponseLog(c, h.log, "error while binding body", http.StatusBadRequest, err.Error())
		return
	}

	// Log the request
	// h.log.Debug("Received registration request", "mobile_phone", req.MobilePhone)

	// Validate phone number
	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		// h.log.Error()
		handleResponseLog(c, h.log, "error validating phone number", http.StatusBadRequest, err.Error())
		return
	}

	// Call service layer to confirm registration
	confResp, err := h.service.Auth().UserRegisterConfirm(c.Request.Context(), req)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while confirming user registration")
		c.JSON(http.StatusInternalServerError, Response{
			Status:      http.StatusInternalServerError,
			Description: "Failed to confirm user registration",
			Data:        err.Error() + ":" + "error while confirming user registration",
		})
		return
	}

	// Log and return success response
	handleResponseLog(c, h.log, "User registration confirmed successfully", http.StatusOK, confResp)
}

/*
	h.log.Error("missing device id")
	c.JSON(http.StatusBadRequest, "fill the gap with id")
*/

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
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode:  http.StatusBadRequest,
			Description: "You have exceeded the device limit. Please delete one of your devices to proceed.",
		})
		return
	}

	// Proceed with login process after confirming OTP
	resp, err := h.service.Auth().UserLoginByPhoneConfirm(c.Request.Context(), req)
	if err != nil {
		StatusCode := http.StatusInternalServerError
		message := "INTERNAL_SERVER_ERROR" + err.Error()

		if err.Error() == "OTP code not found or expired time" || err.Error() == "Incorrect OTP code" || err.Error() == "OTP data not found in Redis" {
			StatusCode = http.StatusBadRequest
			message = err.Error()
		}

		h.log.Error("error in UserLoginByPhoneConfirm: " + err.Error())
		c.JSON(StatusCode, models.Response{
			StatusCode:  StatusCode,
			Description: message,
		})
		return
	}

	h.log.Info("Successfully logged in by phone")
	c.JSON(http.StatusOK, resp)
}
