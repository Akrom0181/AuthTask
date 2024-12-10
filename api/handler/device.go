package handler

import (
	"net/http"
	"task/api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Security BearerAuth
// @ID 			 create_device
// @Router       /task/api/v1/devicecreate [POST]
// @Summary      Create device
// @Description  Create a new device
// @Tags         device
// @Accept       json
// @Produce      json
// @Param        Device body models.CreateDevice true "Device"
// @Success      200 {object} Response{data=string} "Successfully created device"
// @Response     400 {object} Response{data=string} "Bad Request"
// @Failure      500 {object} Response{data=string} "Server error"
func (h *Handler) CreateDevice(c *gin.Context) {
	var device models.Device

	if err := c.ShouldBindJSON(&device); err != nil {
		h.log.Error(err.Error() + " : " + "error device Should Bind Json!")
		c.JSON(http.StatusBadRequest, "Please, enter valid data!")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	device.UserID = userID.(string)

	resp, err := h.storage.Device().Insert(c.Request.Context(), &device)
	if err != nil {
		h.log.Error(err.Error() + ":" + "Error Device Create")
		c.JSON(http.StatusInternalServerError, "Server error!")
		return
	}

	h.log.Info("Device created successfully!")
	c.JSON(http.StatusCreated, Response{Data: resp})
}

// @Security BearerAuth
// @ID 			    get_all_Devices
// @Router 			/task/api/v1/getlistdevices [GET]
// @Summary 		Get all devices
// @Description		Retrieve all devices
// @Tags 			device
// @Accept 			json
// @Produce 		json
// @Success 		200 {object} Response{data=string} "Successfully retrieved devices"
// @Response        400 {object} Response{data=string} "Bad Request"
// @Failure         404 {object} Response{data=string} "Contact not found"
// @Failure         500 {object} Response{data=string} "Server error"
func (h *Handler) GetAllDevices(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
		return
	}

	devices, err := h.storage.Device().GetAll(c.Request.Context(), userIDStr)
	if err != nil {
		h.log.Error(err.Error() + ":" + "Error while getting all devices")
		c.JSON(http.StatusInternalServerError, "Error while getting all devices")
		return
	}

	h.log.Info("Devices retrieved successfully")
	c.JSON(http.StatusOK, Response{Data: devices})
}

// @Security BearerAuth
// @ID 			delete_device
// @Router		/task/api/v1/deletedevice/{id} [DELETE]
// @Summary		Delete device by ID
// @Description Delete device by its ID
// @Tags		device
// @Accept		json
// @Produce		json
// @Param		id path string true "Device ID"
// @Success     200 {object} Response{data=string} "Success Request"
// @Response    400 {object} Response{data=string} "Bad Request"
// @Failure     500 {object} Response{data=string} "Server error"
func (h *Handler) DeleteDevice(c *gin.Context) {
	id := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	if id == "" {
		h.log.Error("missing device id")
		c.JSON(http.StatusBadRequest, "fill the gap with id")
		return
	}

	err := uuid.Validate(id)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while validating id")
		c.JSON(http.StatusBadRequest, "please enter a valid id")
		return
	}

	err = h.storage.Device().Delete(c.Request.Context(), id, userIDStr)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while deleting device")
		c.JSON(http.StatusBadRequest, "please input valid data")
		return
	}

	h.log.Info("Device deleted successfully!")
	c.JSON(http.StatusOK, id)
}

// @ID 			remove_device
// @Router		/task/api/v1/removedevice/{id} [DELETE]
// @Summary		Remove device by ID
// @Description Remove device by its ID
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param		id path string true "Device ID"
// @Success     200 {object} Response{data=string} "Success Request"
// @Response    400 {object} Response{data=string} "Bad Request"
// @Failure     500 {object} Response{data=string} "Server error"
func (h *Handler) RemoveDevice(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		h.log.Error("missing device id")
		c.JSON(http.StatusBadRequest, "fill the gap with id")
		return
	}

	err := uuid.Validate(id)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while validating id")
		c.JSON(http.StatusBadRequest, "please enter a valid id")
		return
	}

	err = h.storage.Device().Remove(c.Request.Context(), id)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while deleting device")
		c.JSON(http.StatusBadRequest, "please input valid data")
		return
	}

	h.log.Info("Device deleted successfully!")
	c.JSON(http.StatusOK, id)
}
