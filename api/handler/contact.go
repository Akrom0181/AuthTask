package handler

import (
	"context"
	"net/http"
	"strconv"
	_ "task/api/docs"
	"task/api/models"
	"task/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Security BearerAuth
// @ID 			 create_contact
// @Router       /task/api/v1/createcontact [POST]
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

	// Bind JSON request body to the contact model
	if err := c.ShouldBindJSON(&contact); err != nil {
		h.log.Error(err.Error() + " : " + "error Contact Should Bind Json!")
		c.JSON(http.StatusBadRequest, "Please, enter valid data!")
		return
	}

	// Extract user ID from context (JWT)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Assign userID from context to the contact's UserID field
	contact.UserID = userID.(string)

	// Create the contact in the database
	resp, err := h.storage.Contact().Create(c.Request.Context(), &contact)
	if err != nil {
		h.log.Error(err.Error() + ":" + "Error Contact Create")
		c.JSON(http.StatusInternalServerError, "Server error!")
		return
	}

	// Log success and return created contact in the response
	h.log.Info("Contact created successfully!")
	c.JSON(http.StatusCreated, Response{Data: resp})
}

// @Security BearerAuth
// @ID           update_contact
// @Router       /task/api/v1/updatecontact/{id} [PUT]
// @Summary      Update Contact
// @Description  Update an existing contact
// @Tags         contact
// @Accept       json
// @Produce      json
// @Param        id path string true "Contact ID"
// @Param        Contact body models.UpdateContact true "UpdateContactRequest"
// @Success      200 {object} models.Contact
// @Response     400 {object} Response{data=string} "Bad Request"
// @Failure      404 {object} Response{data=string} "Contact not found"
// @Failure      500 {object} Response{data=string} "Server error"
func (h *Handler) UpdateContact(c *gin.Context) {
	var updateRequest models.UpdateContact

	// Validate and bind input
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		h.log.Error("Invalid JSON payload", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Extract contact ID from path
	id := c.Param("id")
	if id == "" {
		h.log.Error("Missing contact ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contact ID is required"})
		return
	}

	// Extract user ID from context (JWT middleware ensures this is set)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
		return
	}

	// Fetch the existing contact
	existingContact, err := h.storage.Contact().GetById(c.Request.Context(), id, userIDStr)
	if err != nil {
		if err.Error() == "contact not found" {
			h.log.Error("Contact not found", logger.String("id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		} else {
			h.log.Error("Error fetching contact", logger.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching contact"})
		}
		return
	}

	// Update fields
	existingContact.FirstName = updateRequest.FirstName
	existingContact.LastName = updateRequest.LastName
	existingContact.MiddleName = updateRequest.MiddleName
	existingContact.PhoneNumber = updateRequest.PhoneNumber

	// Update the contact in the database
	updatedContact, err := h.storage.Contact().Update(c.Request.Context(), existingContact, userID.(string))
	if err != nil {
		h.log.Error("Error updating contact", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating contact"})
		return
	}

	h.log.Info("Contact updated successfully", logger.Any("contact", updatedContact))
	c.JSON(http.StatusOK, gin.H{"data": updatedContact})
}

// @Security BearerAuth
// @ID              get_Contact
// @Router          /task/api/v1/getbyidcontact/{id} [GET]
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
		h.log.Error("Error fetching contact", logger.Error(err))
		if err.Error() == "contact not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch contact"})
		}
		return
	}

	// Respond with the fetched contact
	c.JSON(http.StatusOK, gin.H{"data": contact})
}

// @Security BearerAuth
// @ID 			    get_all_Contacts
// @Router 			/task/api/v1/getallcontacts [GET]
// @Summary 		Get all Contacts
// @Description		Retrieve all Contacts
// @Tags 			contact
// @Accept 			json
// @Produce 		json
// @Param 			search query string false "Search Contacts by name or email"
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
// @Router		/task/api/v1/deletecontact/{id} [DELETE]
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

	err = h.storage.Contact().Delete(context.Background(), id, userIDStr)
	if err != nil {
		h.log.Error(err.Error() + ":" + "error while deleting Contact")
		c.JSON(http.StatusBadRequest, "please input valid data")
		return
	}

	h.log.Info("Contact deleted successfully!")
	c.JSON(http.StatusOK, Response{Data: id})
}