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
)

// // @ID 			 create_user
// // @Router       /task/api/v1/user/create [POST]
// // @Summary      Create User
// // @Description  Create a new user
// // @Tags         user
// // @Accept       json
// // @Produce      json
// // @Param        User body models.CreateUser true "User"
// // @Success      200 {object} models.User
// // @Response     400 {object} models.Response{data=string} "Bad Request"
// // @Failure      500 {object} models.Response{data=string} "Server error"
// func (h *Handler) CreateUser(c *gin.Context) {
// 	var (
// 		user = models.User{}
// 	)

// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		h.log.Error(err.Error() + " : " + "error User Should Bind Json!")
// 		c.JSON(http.StatusBadRequest, "Please, enter valid data!")
// 		return
// 	}

// 	resp, err := h.storage.User().Create(c.Request.Context(), &user)
// 	if err != nil {
// 		h.log.Error(err.Error() + ":" + "Error User Create")
// 		c.JSON(http.StatusInternalServerError, "Server error!")
// 		return
// 	}

// 	h.log.Info("User created successfully!")
// 	c.JSON(http.StatusCreated, resp)
// }

// @Security BearerAuth
// @ID 			 update_user
// @Router       /task/api/v1/user/update/{id} [PUT]
// @Summary      Update User
// @Description  Update an existing user
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        User body models.UpdateUser true "UpdateUserRequest"
// @Success      200 {object} models.Response{data=string} "Success"
// @Response     400 {object} models.Response{data=string} "Bad Request"
// @Failure      500 {object} models.Response{data=string} "Server error"
func (h *Handler) UpdateUser(c *gin.Context) {
	var updateUser models.UpdateUser

	if err := c.ShouldBindJSON(&updateUser); err != nil {
		h.log.Error(err.Error() + " : " + "error User Should Bind Json!")
		c.JSON(http.StatusBadRequest, "Please, enter valid data!")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "unauthorized"})
		return
	}

	user, err := h.storage.User().GetById(context.Background(), userID.(string))
	if err != nil {
		if err == sql.ErrNoRows {
			h.log.Error("user not found: " + userID.(string))
			c.JSON(http.StatusNotFound, models.Response{StatusCode: http.StatusNotFound, Description: "user not found", Error: err})
			return
		}
		h.log.Error("Error while getting user by ID" + err.Error())
		c.JSON(http.StatusInternalServerError, models.Response{StatusCode: http.StatusInternalServerError, Description: "error while getting user by id", Error: err})
		return
	}

	// user.ID = userID.(string)
	user.FirstName = updateUser.FirstName
	user.LastName = updateUser.LastName
	// user.PhoneNumber = updateUser.PhoneNumber

	resp, err := h.storage.User().Update(c.Request.Context(), user, userID.(string))
	if err != nil {
		h.log.Error(err.Error() + ":" + "Error User Update")
		c.JSON(http.StatusInternalServerError, models.Response{StatusCode: http.StatusInternalServerError, Description: "error while updating user by id", Error: err})
		return
	}

	h.log.Info("User updated successfully!")
	c.JSON(http.StatusOK, models.Response{StatusCode: http.StatusOK, Description: "User updated successfully", Data: resp})
}

// @ID              get_user
// @Router          /task/api/v1/user/getbyid/{id} [GET]
// @Summary         Get User by ID
// @Description     Retrieve a user by their ID
// @Tags            user
// @Accept          json
// @Produce         json
// @Param           id path string true "User ID"
// @Success         200 {object} models.Response{data=string} "Success"
// @Response        400 {object} models.Response{data=string} "Bad Request"
// @Failure         404 {object} models.Response{data=string} "User not found"
// @Failure         500 {object} models.Response{data=string} "Server error"
func (h *Handler) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		h.log.Error("missing user id")
		c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadRequest, Description: "you must fill the user id"})
		return
	}

	user, err := h.storage.User().GetById(context.Background(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			h.log.Error("user not found: " + id)
			c.JSON(http.StatusNotFound, fmt.Sprintf("user with ID %s not found", id))
			return
		}
		h.log.Error("Error while getting user by ID" + err.Error())
		c.JSON(http.StatusInternalServerError, models.Response{StatusCode: http.StatusInternalServerError, Description: "error while getting user by id", Error: err})
		return
	}

	h.log.Info("User retrieved successfully by ID")
	c.JSON(http.StatusOK, models.Response{StatusCode: http.StatusOK, Description: "user retrieved successfully", Data: user})
}

// @ID 			    get_all_users
// @Router 			/task/api/v1/user/getall [GET]
// @Summary 		Get All Users
// @Description		Retrieve all users
// @Tags 			user
// @Accept 			json
// @Produce 		json
// @Param 			search query string false "Search users by first_name or phone_number"
// @Param 			page   query uint64 false "Page number"
// @Param 			limit  query uint64 false "Limit number of results per page"
// @Success 		200 {object} models.Response{data=string} "Success"
// @Response        400 {object} models.Response{data=string} "Bad Request"
// @Response        401 {object} models.Response{data=string} "Unauthorized"
// @Failure         404 {object} models.Response{data=string} "Contact not found"
// @Failure         500 {object} models.Response{data=string} "Server error"
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
		c.JSON(http.StatusInternalServerError, models.Response{StatusCode: http.StatusInternalServerError, Description: "error while getting all users", Error: err})
		return
	}

	h.log.Info("Users retrieved successfully")
	c.JSON(http.StatusOK, models.Response{StatusCode: http.StatusOK, Description: "users retrieved successfully", Data: users})
}

// @Security BearerAuth
// @ID 			delete_user
// @Router		/task/api/v1/user/delete [DELETE]
// @Summary		Delete User by ID
// @Description Delete a user by their ID
// @Tags		user
// @Accept		json
// @Produce		json
// @Success     200 {object} models.Response{data=string} "Success Request"
// @Response    400 {object} models.Response{data=string} "Bad Request"
// @Response    401 {object} models.Response{data=string} "Unauthorized"
// @Failure     404 {object} models.Response{data=string} "Contact not found"
// @Failure     500 {object} models.Response{data=string} "Server error"
func (h *Handler) DeleteUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "invalid user_id"})
		return
	}

	err := h.storage.User().Delete(context.Background(), userIDStr)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while deleting user")
		c.JSON(http.StatusInternalServerError, models.Response{StatusCode: http.StatusInternalServerError, Description: "error while deleting user", Error: err})
		return
	}

	h.log.Info("User deleted successfully!")
	c.JSON(http.StatusOK, models.Response{StatusCode: http.StatusOK, Data: userIDStr, Description: "user deleted successfully"})
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
// @Success     200 {object} models.Response{data=string} "Success Request"
// @Response    400 {object} models.Response{data=string} "Bad Request"
// @Response    401 {object} models.Response{data=string} "Unauthorized"
// @Failure     404 {object} models.Response{data=string} "User not found"
// @Failure     500 {object} models.Response{data=string} "Server error"
func (h *Handler) RequestOTP(c *gin.Context) {
	var req models.UserLoginRequest

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "invalid user_id"})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("error binding body" + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadRequest, Description: "invalid request", Error: err})
		return
	}

	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		h.log.Error("error while validating phone number: " + req.MobilePhone + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadRequest, Description: "please input valid phone number", Error: err})
		return
	}

	otpMsg, err := h.service.Auth().OTPForChangingNumber(c.Request.Context(), req, userIDStr)
	if err != nil {
		h.log.Error("Failed to generate OTP" + err.Error())
		c.JSON(http.StatusInternalServerError, models.Response{StatusCode: http.StatusInternalServerError, Description: "Failed to generate OTP", Error: err})
		return
	}

	// handleResponseLog(c, h.log, "successfully generated otp", http.StatusOK, err)
	c.JSON(http.StatusOK, models.Response{StatusCode: http.StatusOK, Description: "successfully generated otp", Data: otpMsg, Error: err})
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
// @Success     200 {object} models.Response{data=string} "Success Request"
// @Response    400 {object} models.Response{data=string} "Bad Request"
// @Response    401 {object} models.Response{data=string} "Unauthorized"
// @Failure     404 {object} models.Response{data=string} "User not found"
// @Failure     500 {object} models.Response{data=string} "Server error"
func (h *Handler) ConfirmOTPAndUpdatePhoneNumber(c *gin.Context) {
	var req models.UserChangePhoneConfirm

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("error binding json" + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadGateway, Description: "invalid request", Error: err})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "please authorize"})
		return
	}

	if err := check.ValidatePhoneNumber(req.MobilePhone); err != nil {
		h.log.Error("error while validating phone number: " + req.MobilePhone + err.Error())
		c.JSON(http.StatusBadRequest, models.Response{StatusCode: http.StatusBadRequest, Description: "error validating phone number", Data: req.MobilePhone, Error: err})
		return
	}

	err := h.service.Auth().ConfirmOTPAndUpdatePhoneNumber(c.Request.Context(), req.MobilePhone, req.SmsCode, userID.(string))
	if err != nil {
		h.log.Error("error while confirming phone number" + err.Error())
		c.JSON(http.StatusInternalServerError, models.Response{StatusCode: http.StatusInternalServerError, Description: "error while confirming phone number"})
		return
	}

	h.log.Info("Phone number updated successfully" + req.MobilePhone)
	c.JSON(http.StatusOK, models.Response{StatusCode: http.StatusOK, Description: "Phone number updated successfully", Data: req.MobilePhone})
}

// @Security BearerAuth
// @ID 			logout
// @Router		/task/api/v1/user/logout [DELETE]
// @Summary		Logout
// @Description Logout for user
// @Tags		user
// @Accept		json
// @Produce		json
// @Success     200 {object} models.Response{data=string} "Success Request"
// @Response    400 {object} models.Response{data=string} "Bad Request"
// @Response    401 {object} models.Response{data=string} "Unauthorized"
// @Response    409 {object} models.Response{data=string} "Conflict"
// @Failure     500 {object} models.Response{data=string} "Server error"
func (h *Handler) Logout(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{StatusCode: http.StatusUnauthorized, Description: "invalid request"})
		return
	}

	deviceID, exists := c.Get("device_id")
	log.Println(deviceID)
	if !exists {
		c.JSON(http.StatusConflict, models.Response{StatusCode: http.StatusConflict, Description: "device id not found"})
		return
	}

	err := h.storage.Device().Delete(c.Request.Context(), deviceID.(string), userID.(string))
	if err != nil {
		h.log.Error("error while logging out" + err.Error())
		c.JSON(http.StatusInternalServerError, models.Response{StatusCode: http.StatusInternalServerError, Description: "failed to log out", Error: err})
		return
	}

	h.log.Info("logged out successfully")
	c.JSON(http.StatusOK, models.Response{StatusCode: http.StatusOK, Description: "logged out successfully"})
}
