package handler

import (
	"context"
	"net/http"
	"strconv"
	_ "task/api/docs"
	"task/api/models"
	"task/pkg/logger"
	check "task/pkg/validation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Security BearerAuth
// @ID 			 create_contact
// @Router       /task/api/v1/contact/create [POST]
// @Summary      Create Contact
// @Description  Create a new contact
// @Tags         contact
// @Accept       json
// @Produce      json
// @Param        Contact body models.CreateContact true "Contact"
// @Success      200 {object} Response{data=string} "Successfully created contact"
// @Response     400 {object} Response{data=string} "Bad Request"
// @Failure      500 {object} Response{data=string} "Server error"
func (h *Handler) CreateContact(c *gin.Context) {
	var contact models.Contact

	if err := c.ShouldBindJSON(&contact); err != nil {
		h.log.Error(err.Error() + " : " + "error Contact Should Bind Json!")
		c.JSON(http.StatusBadRequest, "Please, enter valid data!")
		return
	}

	if err := check.ValidatePhoneNumber(contact.PhoneNumber); err != nil {
		handleResponseLog(c, h.log, "error validating phone number", http.StatusBadRequest, err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	contact.UserID = userID.(string)

	resp, err := h.storage.Contact().Create(c.Request.Context(), &contact)
	if err != nil {
		h.log.Error(err.Error() + ":" + "Error Contact Create")
		c.JSON(http.StatusInternalServerError, "Server error!")
		return
	}

	h.log.Info("Contact created successfully!")
	c.JSON(http.StatusCreated, Response{Data: resp})
}

// @Security BearerAuth
// @ID           update_contact
// @Router       /task/api/v1/contact/update/{id} [PUT]
// @Summary      Update Contact
// @Description  Update an existing contact
// @Tags         contact
// @Accept       json
// @Produce      json
// @Param        id path string true "Contact ID"
// @Param        Contact body models.UpdateContact true "UpdateContactRequest"
// @Success      200 {object} Response{data=models.Contact}
// @Response     400 {object} Response{data=string} "Bad Request"
// @Response     401 {object} Response{data=string} "Unauthorized"
// @Response     404 {object} Response{data=string} "Contact not found"
// @Response     500 {object} Response{data=string} "Server error"
func (h *Handler) UpdateContact(c *gin.Context) {
	var updateRequest models.UpdateContact

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		h.log.Error("Invalid JSON payload", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := check.ValidatePhoneNumber(*updateRequest.PhoneNumber); err != nil {
		handleResponseLog(c, h.log, "error validating phone number", http.StatusBadRequest, err.Error())
		return
	}

	id := c.Param("id")
	if id == "" {
		h.log.Error("Missing contact ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contact ID is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	existingContact, err := h.storage.Contact().GetById(c.Request.Context(), id, userIDStr)
	if err != nil {
		if err.Error() == "contact not found" {
			h.log.Error("Contact not found", logger.String("id", id), logger.String("user_id", userIDStr))
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		} else {
			h.log.Error("Error fetching contact", logger.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching contact"})
		}
		return
	}

	if updateRequest.FirstName != nil {
		existingContact.FirstName = *updateRequest.FirstName
	}
	if updateRequest.LastName != nil {
		existingContact.LastName = *updateRequest.LastName
	}
	if updateRequest.MiddleName != nil {
		existingContact.MiddleName = *updateRequest.MiddleName
	}
	if updateRequest.PhoneNumber != nil {
		existingContact.PhoneNumber = *updateRequest.PhoneNumber
	}

	updatedContact, err := h.storage.Contact().Update(c.Request.Context(), existingContact, userIDStr)
	if err != nil {
		h.log.Error("Error updating contact", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating contact"})
		return
	}

	h.log.Info("Contact updated successfully", logger.String("id", id), logger.String("user_id", userIDStr))
	c.JSON(http.StatusOK, gin.H{"data": updatedContact})
}

// @Security BearerAuth
// @ID              get_Contact
// @Router          /task/api/v1/contact/getbyid/{id} [GET]
// @Summary         Get contact by ID
// @Description     Retrieve a contact by their ID
// @Tags            contact
// @Accept          json
// @Produce         json
// @Param           id path string true "Contact ID"
// @Success         200 {object} Response{data=string} "Successfully retrieved contact"
// @Response        400 {object} Response{data=string} "Bad Request"
// @Failure         404 {object} Response{data=string} "Contact not found"
// @Failure         500 {object} Response{data=string} "Server error"
func (h *Handler) GetContactsById(c *gin.Context) {
	// Extract user ID from context
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

	// Extract contact ID from the path
	contactID := c.Param("id")
	if contactID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contact ID is required"})
		return
	}

	// Query the database to fetch the contact
	contact, err := h.storage.Contact().GetById(c.Request.Context(), contactID, userIDStr)
	if err != nil {
		h.log.Error("Error: ", logger.Error(err))
		if err.Error() == "database error: no rows in result set" {
			c.JSON(http.StatusNotFound, Response{Status: http.StatusNotFound, Description: "contact not found"})
		} else {
			c.JSON(http.StatusInternalServerError, Response{Status: http.StatusInternalServerError, Description: err.Error()})
		}
		return
	}

	// Respond with the fetched contact
	c.JSON(http.StatusOK, gin.H{"data": contact})
}

// @Security BearerAuth
// @ID 			    get_all_Contacts
// @Router 			/task/api/v1/contact/getall [GET]
// @Summary 		Get all Contacts
// @Description		Retrieve all Contacts
// @Tags 			contact
// @Accept 			json
// @Produce 		json
// @Param 			search query string false "Search contacts by firstname or phonenumber"
// @Param 			page   query uint64 false "Page number"
// @Param 			limit  query uint64 false "Limit number of results per page"
// @Success 		200 {object} Response{data=string} "Successfully retrieved contacts"
// @Response        400 {object} Response{data=string} "Bad Request"
// @Failure         404 {object} Response{data=string} "Contact not found"
// @Failure         500 {object} Response{data=string} "Server error"
func (h *Handler) GetAllContacts(c *gin.Context) {
	var req = &models.GetAllContactsRequest{}

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

	// Extract contact ID from the path
	// contactID := c.Param("id")
	// if contactID == "" {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "contact ID is required"})
	// 	return
	// }

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

	Contacts, err := h.storage.Contact().GetAll(context.Background(), req, userIDStr)
	if err != nil {
		h.log.Error(err.Error() + ":" + "Error while getting all contacts")
		c.JSON(http.StatusInternalServerError, "Error while getting all contacts")
		return
	}

	h.log.Info("Contacts retrieved successfully")
	c.JSON(http.StatusOK, Response{Data: Contacts})
}

// @Security BearerAuth
// @ID 			delete_Contact
// @Router		/task/api/v1/contact/delete/{id} [DELETE]
// @Summary		Delete contact by ID
// @Description Delete a contact by its ID
// @Tags		contact
// @Accept		json
// @Produce		json
// @Param		id path string true "Contact ID"
// @Success     200 {object} Response{data=string} "Successfully deleted"
// @Response    400 {object} Response{data=string} "Bad Request"
// @Failure     500 {object} Response{data=string} "Server error"
func (h *Handler) DeleteContact(c *gin.Context) {
	id := c.Param("id")

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

	if id == "" {
		h.log.Error("missing Contact id")
		c.JSON(http.StatusBadRequest, "fill the gap with id")
		return
	}

	err := uuid.Validate(id)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while validating id")
		c.JSON(http.StatusBadRequest, "please enter a valid id")
		return
	}

	contact, err := h.storage.Contact().GetById(c.Request.Context(), id, userIDStr)
	if err != nil {
		h.log.Error("Error: ", logger.Error(err))
		if err.Error() == "database error: no rows in result set" {
			c.JSON(http.StatusNotFound, Response{Status: http.StatusNotFound, Description: "contact not found"})
		} else {
			c.JSON(http.StatusInternalServerError, Response{Status: http.StatusInternalServerError, Description: err.Error()})
		}
		return
	}

	err = h.storage.Contact().Delete(context.Background(), contact.ID, userIDStr)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while deleting Contact")
		c.JSON(http.StatusBadRequest, "please input valid data")
		return
	}

	h.log.Info("Contact deleted successfully!")
	c.JSON(http.StatusOK, Response{Status: http.StatusOK, Description: "contact deleted successfully!", Data: id})
}
