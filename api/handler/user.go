package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	_ "task/api/docs"
	"task/api/models"
	check "task/pkg/validation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @ID 			 create_user
// @Router       /task/api/v1/createuser [POST]
// @Summary      Create User
// @Description  Create a new user
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        User body models.CreateUser true "User"
// @Success      200 {object} models.User
// @Response     400 {object} Response{data=string} "Bad Request"
// @Failure      500 {object} Response{data=string} "Server error"
func (h *Handler) CreateUser(c *gin.Context) {
	var (
		user = models.User{}
	)

	if err := c.ShouldBindJSON(&user); err != nil {
		h.log.Error(err.Error() + " : " + "error User Should Bind Json!")
		c.JSON(http.StatusBadRequest, "Please, enter valid data!")
		return
	}

	resp, err := h.storage.User().Create(c.Request.Context(), &user)
	if err != nil {
		h.log.Error(err.Error() + ":" + "Error User Create")
		c.JSON(http.StatusInternalServerError, "Server error!")
		return
	}

	h.log.Info("User created successfully!")
	c.JSON(http.StatusCreated, resp)
}

// @Security BearerAuth
// @ID 			 update_user
// @Router       /task/api/v1/updateuser/{id} [PUT]
// @Summary      Update User
// @Description  Update an existing user
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        User body models.UpdateUser true "UpdateUserRequest"
// @Success      200 {object} models.User
// @Response     400 {object} Response{data=string} "Bad Request"
// @Failure      500 {object} Response{data=string} "Server error"
func (h *Handler) UpdateUser(c *gin.Context) {
	var updateUser models.UpdateUser

	if err := c.ShouldBindJSON(&updateUser); err != nil {
		h.log.Error(err.Error() + " : " + "error User Should Bind Json!")
		c.JSON(http.StatusBadRequest, "Please, enter valid data!")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.storage.User().GetById(c.Request.Context(), userID.(string))
	if err != nil {
		h.log.Error(err.Error() + ":" + "Error User Not Found")
		c.JSON(http.StatusBadRequest, "User not found!")
		return
	}

	// user.ID = userID.(string)
	user.FirstName = updateUser.FirstName
	user.LastName = updateUser.LastName
	// user.PhoneNumber = updateUser.PhoneNumber

	resp, err := h.storage.User().Update(c.Request.Context(), user, userID.(string))
	if err != nil {
		h.log.Error(err.Error() + ":" + "Error User Update")
		c.JSON(http.StatusInternalServerError, "Server error!")
		return
	}

	h.log.Info("User updated successfully!")
	c.JSON(http.StatusOK, resp)
}

// @ID              get_user
// @Router          /task/api/v1/getbyiduser/{id} [GET]
// @Summary         Get User by ID
// @Description     Retrieve a user by their ID
// @Tags            user
// @Accept          json
// @Produce         json
// @Param           id path string true "User ID"
// @Success         200  {object}  models.User
// @Response        400 {object} Response{data=string} "Bad Request"
// @Failure         404 {object} Response{data=string} "User not found"
// @Failure         500 {object} Response{data=string} "Server error"
func (h *Handler) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		h.log.Error("missing user id")
		c.JSON(http.StatusBadRequest, "you must fill the user ID")
		return
	}

	user, err := h.storage.User().GetById(context.Background(), id)
	if err != nil {
		// If the error is "no rows" (user not found), return 404
		if err == sql.ErrNoRows {
			h.log.Error("user not found: " + id)
			c.JSON(http.StatusNotFound, fmt.Sprintf("user with ID %s not found", id))
			return
		}
		// For other errors, return 500
		h.log.Error("Error while getting user by ID: " + err.Error())
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("server error: %v", err))
		return
	}

	h.log.Info("User retrieved successfully by ID")
	c.JSON(http.StatusOK, user)
}

// @ID 			    get_all_users
// @Router 			/task/api/v1/getallusers [GET]
// @Summary 		Get All Users
// @Description		Retrieve all users
// @Tags 			user
// @Accept 			json
// @Produce 		json
// @Param 			search query string false "Search users by name or email"
// @Param 			page   query uint64 false "Page number"
// @Param 			limit  query uint64 false "Limit number of results per page"
// @Success 		200  {object} models.GetAllUsersResponse
// @Response        400 {object} Response{data=string} "Bad Request"
// @Failure         500 {object} Response{data=string} "Server error"
func (h *Handler) GetAllUsers(c *gin.Context) {
	var req = &models.GetAllUsersRequest{}

	req.Search = c.Query("search")

	page, err := strconv.ParseUint(c.DefaultQuery("page", "1"), 10, 64)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while parsing page")
		c.JSON(http.StatusBadRequest, "BadRequest at paging")
		return
	}

	limit, err := strconv.ParseUint(c.DefaultQuery("limit", "10"), 10, 64)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while parsing limit")
		c.JSON(http.StatusInternalServerError, "Internal server error while parsing limit")
		return
	}

	req.Page = page
	req.Limit = limit

	users, err := h.storage.User().GetAll(context.Background(), req)
	if err != nil {
		h.log.Error(err.Error() + ":" + "Error while getting all users")
		c.JSON(http.StatusInternalServerError, "Error while getting all users")
		return
	}

	h.log.Info("Users retrieved successfully")
	c.JSON(http.StatusOK, users)
}

// @Security BearerAuth
// @ID 			delete_user
// @Router		/task/api/v1/deleteuser/{id} [DELETE]
// @Summary		Delete User by ID
// @Description Delete a user by their ID
// @Tags		user
// @Accept		json
// @Produce		json
// @Param		id path string true "User ID"
// @Success     200 {object} Response{data=string} "Success Request"
// @Response    400 {object} Response{data=string} "Bad Request"
// @Failure     500 {object} Response{data=string} "Server error"
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		h.log.Error("missing user id")
		c.JSON(http.StatusBadRequest, "fill the gap with id")
		return
	}

	err := uuid.Validate(id)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while validating id")
		c.JSON(http.StatusBadRequest, "please enter a valid id")
		return
	}

	err = h.storage.User().Delete(context.Background(), id)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while deleting user")
		c.JSON(http.StatusBadRequest, "please input valid data")
		return
	}

	h.log.Info("User deleted successfully!")
	c.JSON(http.StatusOK, id)
}

// @Security BearerAuth
// @ID 			update_user_phone_number
// @Router		/task/api/v1/user/sendotp [POST]
// @Summary		update user phone_number by otp
// @Description update user phone_number by using otp
// @Tags		user
// @Accept		json
// @Produce		json
// @Param       User body models.UserChangePhone true "User"
// @Success     200 {object} Response{data=string} "Success Request"
// @Response    400 {object} Response{data=string} "Bad Request"
// @Failure     500 {object} Response{data=string} "Server error"
func (h *Handler) RequestOTP(c *gin.Context) {
	var req models.UserLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		handleResponseLog(c, h.log, "error while validating phone number: "+req.MobilePhone, http.StatusBadRequest, err.Error())
		return
	}

	otpMsg, err := h.service.Auth().OTPForChangingNumber(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": otpMsg})
}

// @Security BearerAuth
// @ID 			update_user_phoneNumber
// @Router		/task/api/v1/user/confirmotp [POST]
// @Summary		Confirm user phone_number by otp
// @Description Confirm user phone_number by otp to update phone_number
// @Tags		user
// @Accept		json
// @Produce		json
// @Param       User body models.UserChangePhoneConfirm true "User"
// @Success     200 {object} Response{data=string} "Success Request"
// @Response    400 {object} Response{data=string} "Bad Request"
// @Failure     500 {object} Response{data=string} "Server error"
func (h *Handler) ConfirmOTPAndUpdatePhoneNumber(c *gin.Context) {
	var req models.UserChangePhoneConfirm

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		handleResponseLog(c, h.log, "error while validating phone number: "+req.MobilePhone, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.Auth().ConfirmOTPAndUpdatePhoneNumber(c.Request.Context(), req.MobilePhone, req.SmsCode, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Phone number updated successfully"})
}

// @Security BearerAuth
// @ID 			logout
// @Router		/task/api/v1/user/logout [DELETE]
// @Summary		Logout
// @Description Logout for user
// @Tags		user
// @Accept		json
// @Produce		json
// @Success     200 {object} Response{data=string} "Success Request"
// @Response    400 {object} Response{data=string} "Bad Request"
// @Failure     500 {object} Response{data=string} "Server error"
func (h *Handler) Logout(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	deviceID, exists := c.Get("device_id")
	log.Println(deviceID)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device ID is required"})
		return
	}

	err := h.storage.Device().Delete(c.Request.Context(), deviceID.(string), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log out"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
