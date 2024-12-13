package handler

import (
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
		handleResponseLog(c, h.log, "error while binding body", http.StatusBadRequest, err.Error())
		c.JSON(http.StatusBadRequest, Response{Status: http.StatusBadRequest, Description: "error while binding body", Data: &loginReq, Error: err.Error()})
		return
	}

	if err := check.ValidatePhoneNumber(loginReq.MobilePhone); err != nil {
		handleResponseLog(c, h.log, "error while validating phone number: "+loginReq.MobilePhone, http.StatusBadRequest, err.Error())
		c.JSON(http.StatusBadRequest, Response{Status: http.StatusBadRequest,
			Description: "error validating phone number",
			Data:        loginReq.MobilePhone,
			Error:       err.Error()})
		return
	}

	otp, err := h.service.Auth().UserRegister(c.Request.Context(), loginReq)
	if err != nil {
		handleResponseLog(c, h.log, "error while sending sms code to "+loginReq.MobilePhone, http.StatusInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, Response{
			Status:      http.StatusInternalServerError,
			Description: "error while sending otp code",
			Data:        otp,
			Error:       err.Error(),
		})
		return
	}

	handleResponseLog(c, h.log, "Otp sent successfull", http.StatusOK, otp)
	c.JSON(http.StatusOK, Response{Status: http.StatusOK, Description: "sending otp code successfully", Data: otp})

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
		c.JSON(http.StatusBadRequest, Response{Status: http.StatusBadRequest, Description: "error while binding body", Data: &req, Error: err.Error()})
		return
	}

	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		handleResponseLog(c, h.log, "error validating phone number", http.StatusBadRequest, err.Error())
		c.JSON(http.StatusBadRequest, Response{Status: http.StatusBadRequest,
			Description: "error validating phone number",
			Data:        req.MobilePhone,
			Error:       err.Error()})
		return
	}

	// Call service layer to confirm registration
	confResp, err := h.service.Auth().UserRegisterConfirm(c.Request.Context(), req)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while confirming user registration")
		c.JSON(http.StatusInternalServerError, Response{
			Status:      http.StatusInternalServerError,
			Description: "Failed to confirm user registration",
			Data:        confResp,
			Error:       err.Error(),
		})
		return
	}

	handleResponseLog(c, h.log, "User registration confirmed successfully", http.StatusOK, confResp)
	c.JSON(http.StatusOK, Response{Status: http.StatusOK, Description: "User registration confirmed successfully!", Data: confResp})
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
		h.log.Error("error while binding body: " + err.Error())
		c.JSON(http.StatusBadRequest, Response{Status: http.StatusBadRequest, Description: "error while binding body", Data: &loginReq, Error: err.Error()})
		return
	}

	if err := check.ValidatePhoneNumber(loginReq.MobilePhone); err != nil {
		h.log.Error("error while validating phone number: " + loginReq.MobilePhone + err.Error())
		c.JSON(http.StatusBadRequest, Response{Status: http.StatusBadRequest, Description: "error validating phone number", Data: loginReq.MobilePhone, Error: err.Error()})
		return
	}

	loginResp, err := h.service.Auth().UserLoginSendOTP(c.Request.Context(), loginReq)
	if err != nil {
		h.log.Error("error while sending otp" + err.Error())
		c.JSON(http.StatusInternalServerError, Response{Status: http.StatusInternalServerError, Description: "error while sending otp", Error: err.Error()})
		return
	}

	h.log.Info("successfully sent otp")
	c.JSON(http.StatusOK, Response{Status: http.StatusOK, Description: "successfully sent otp", Data: loginResp})
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

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("error while binding request body: " + err.Error())
		c.JSON(http.StatusBadRequest, Response{
			Status:      http.StatusBadRequest,
			Description: "error while binding request body",
			Data:        &req,
			Error:       err.Error(),
		})
		return
	}

	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		h.log.Error("error while validating phone number: " + req.MobilePhone + err.Error())
		c.JSON(http.StatusBadRequest, Response{
			Status:      http.StatusBadRequest,
			Description: "error validating phone number",
			Data:        req.MobilePhone,
			Error:       err.Error()})
		return
	}

	user, err := h.storage.User().GetByPhone(c.Request.Context(), req.MobilePhone)
	if err != nil {
		h.log.Error("error fetching user by phone number: " + err.Error())
		c.JSON(http.StatusUnauthorized, Response{
			Status:      http.StatusUnauthorized,
			Description: "User not found or phone number not registered",
			Error:       err.Error(),
		})
		return
	}

	deviceCount, err := h.storage.Device().GetDeviceCount(c.Request.Context(), user.ID)
	if err != nil {
		h.log.Error("error getting device count: " + err.Error())
		c.JSON(http.StatusInternalServerError, Response{
			Status:      http.StatusInternalServerError,
			Description: "Error checking device count",
			Error:       err.Error(),
		})
		return
	}

	if deviceCount >= 3 {
		h.log.Error("User has exceeded the device limit")
		c.JSON(http.StatusBadRequest, Response{
			Status:      http.StatusBadRequest,
			Description: "You have exceeded the device limit. Please delete one of your devices to proceed.",
		})
		return
	}

	resp, err := h.service.Auth().UserLoginByPhoneConfirm(c.Request.Context(), req)
	if err != nil {
		h.log.Error("error getting user in auth")
		c.JSON(http.StatusInternalServerError, Response{Status: http.StatusInternalServerError, Description: "error while sending otp", Error: err.Error()})
		if err.Error() == "OTP code not found or expired time" || err.Error() == "Incorrect OTP code" || err.Error() == "OTP data not found in Redis" {
			c.JSON(http.StatusBadRequest, Response{Status: http.StatusBadRequest, Description: "error on otp code", Error: err})

		}

		h.log.Error("error in UserLoginByPhoneConfirm: " + err.Error())
		c.JSON(http.StatusInternalServerError, Response{
			Status:      http.StatusInternalServerError,
			Description: "error in user loggin by phoneconfirm",
			Error:       err.Error(),
		})
		return
	}

	h.log.Info("Successfully logged in by phone number")
	c.JSON(http.StatusOK, Response{Status: http.StatusOK, Description: "successfully logged in by phone number", Data: resp})
}
