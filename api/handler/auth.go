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
		h.log.Error("error while binding body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadRequest, Description: "error while binding body", Data: &loginReq, Error: err.Error()})
		return
	}

	if err := check.ValidatePhoneNumber(loginReq.MobilePhone); err != nil {
		h.log.Error("error while validating phone number: " + loginReq.MobilePhone)
		c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadRequest,
			Description: "error validating phone number",
			Data:        loginReq.MobilePhone,
			Error:       err.Error()})
		return
	}

	otp, err := h.service.Auth().UserRegister(c.Request.Context(), loginReq)
	if err != nil {
		h.log.Error("error while sending sms code to " + loginReq.MobilePhone)
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode:  http.StatusInternalServerError,
			Description: "error while sending otp code",
			Data:        otp,
			Error:       err.Error(),
		})
		return
	}

	h.log.Info("Otp sent successfull")
	c.JSON(http.StatusOK, models.Response{StatusCode: http.StatusOK, Description: "sending otp code successfully", Data: otp})

}

// UserRegisterConfirm godoc
// @Router       /task/api/v1/user/registerconfirm [POST]
// @Summary      User register confirmation
// @Description  Registering to System
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        register body models.UserRegisterConfRequest true "register"
// @Success      201  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      404  {object}  models.Response
// @Failure      500  {object}  models.Response
func (h *Handler) Register(c *gin.Context) {
	req := models.UserRegisterConfRequest{}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("error while binding body" + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadRequest, Description: "error while binding body", Data: &req, Error: err.Error()})
		return
	}

	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		h.log.Error("error while checking phone number" + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadRequest,
			Description: "error validating phone number",
			Data:        req.MobilePhone,
			Error:       err.Error()})
		return
	}

	confResp, err := h.service.Auth().UserRegisterConfirm(c.Request.Context(), req)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while confirming user registration")
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode:  http.StatusInternalServerError,
			Description: "Failed to confirm user registration",
			Data:        confResp,
			Error:       err.Error(),
		})
		return
	}

	h.log.Info("User registration confirmed successfully")
	c.JSON(http.StatusOK, models.Response{StatusCode: http.StatusOK, Description: "User registration confirmed successfully!", Data: confResp})
}

// UserLogin godoc
// @Router       /task/api/v1/user/loginrequest [POST]
// @Summary      User login request
// @Description  Login to System
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        login body models.UserLoginRequest true "login"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      401  {object}  models.Response
// @Failure      404  {object}  models.Response
// @Failure      500  {object}  models.Response
func (h *Handler) UserLogin(c *gin.Context) {
	req := models.UserLoginRequest{}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("error while binding body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadRequest, Description: "error while binding body", Data: &req, Error: err.Error()})
		return
	}

	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		h.log.Error("error while validating phone number: " + req.MobilePhone + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadRequest, Description: "error validating phone number", Data: req.MobilePhone, Error: err.Error()})
		return
	}

	isExist, err := h.storage.User().UserExists(c.Request.Context(), req.MobilePhone)
	if err != nil {
		h.log.Error("error while checking if user exists: " + req.MobilePhone)
		c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "user does not exist", Data: req.MobilePhone, Error: err.Error()})
		return
	}

	if !isExist {
		h.log.Error("user does not exist: " + req.MobilePhone)
		c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "user does not exist", Data: req.MobilePhone})
		return
	}

	loginResp, err := h.service.Auth().UserLoginSendOTP(c.Request.Context(), req)
	if err != nil {
		h.log.Error("error while sending otp" + err.Error())
		c.JSON(http.StatusInternalServerError, models.Response{StatusCode: http.StatusInternalServerError, Description: "error while sending otp", Error: err.Error()})
		return
	}

	h.log.Info("successfully sent otp")
	c.JSON(http.StatusOK, models.Response{StatusCode: http.StatusOK, Description: "successfully sent otp", Data: loginResp})
}

// UserLoginByPhoneConfirm godoc
// @Router       /task/api/v1/user/loginconfirm [POST]
// @Summary      User login by phone confirmation
// @Description  Login to the system using phone number and OTP
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        login body models.UserLoginPhoneConfirmRequest true "login"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      401  {object}  models.Response
// @Failure      500  {object}  models.Response
func (h *Handler) UserLoginByPhoneConfirm(c *gin.Context) {
	var req models.UserLoginPhoneConfirmRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("error while binding request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode:  http.StatusBadRequest,
			Description: "error while binding request body",
			Data:        &req,
			Error:       err.Error(),
		})
		return
	}

	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		h.log.Error("error while validating phone number: " + req.MobilePhone + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode:  http.StatusBadRequest,
			Description: "error validating phone number",
			Data:        req.MobilePhone,
			Error:       err.Error()})
		return
	}

	user, err := h.storage.User().GetByPhone(c.Request.Context(), req.MobilePhone)
	if err != nil {
		h.log.Error("error fetching user by phone number: " + err.Error())
		c.JSON(http.StatusUnauthorized, models.Response{
			StatusCode:  http.StatusUnauthorized,
			Description: "user not found",
			Data:        "please register your account",
			Error:       err.Error(),
		})
		return
	}

	deviceCount, err := h.storage.Device().GetDeviceCount(c.Request.Context(), user.ID)
	if err != nil {
		h.log.Error("error getting device count: " + err.Error())
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode:  http.StatusInternalServerError,
			Description: "Error checking device count",
			Error:       err.Error(),
		})
		return
	}

	if deviceCount >= 3 {
		h.log.Error("User has exceeded the device limit")
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode:  http.StatusBadRequest,
			Description: "You have exceeded the device limit. Please delete one of your devices to proceed.",
		})
		return
	}

	resp, err := h.service.Auth().UserLoginByPhoneConfirm(c.Request.Context(), req)
	if err != nil {
		h.log.Error("error getting user in auth")
		c.JSON(http.StatusInternalServerError, models.Response{StatusCode: http.StatusInternalServerError, Description: "error while sending otp", Error: err.Error()})
		if err.Error() == "OTP code not found or expired time" || err.Error() == "Incorrect OTP code" || err.Error() == "OTP data not found in Redis" {
			c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadRequest, Description: "error on otp code", Error: err})

		}

		h.log.Error("error in UserLoginByPhoneConfirm: " + err.Error())
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode:  http.StatusInternalServerError,
			Description: "error in user login by phone",
			Error:       err.Error(),
		})
		return
	}

	h.log.Info("Successfully logged in by phone number")
	c.JSON(http.StatusOK, models.Response{StatusCode: http.StatusOK, Description: "successfully logged in by phone number", Data: resp})
}
