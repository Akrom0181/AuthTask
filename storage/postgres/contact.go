package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"task/api/models"
	"task/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ContactRepo struct {
	db  *pgxpool.Pool
	log logger.LoggerI
}

func NewContactRepo(db *pgxpool.Pool, log logger.LoggerI) *ContactRepo {
	return &ContactRepo{
		db:  db,
		log: log,
	}
}

func (r *ContactRepo) Create(ctx context.Context, contact *models.Contact) (*models.Contact, error) {
	id := uuid.New()

	query := `INSERT INTO "contacts" (
        id,
        user_id,
        first_name, 
        last_name, 
        middle_name,
        phone_number, 
        created_at)
    VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Tashkent')
        RETURNING created_at`

	var createdAt time.Time

	err := r.db.QueryRow(ctx, query,
		id.String(),
		contact.UserID,
		contact.FirstName,
		contact.LastName,
		contact.MiddleName,
		contact.PhoneNumber,
	).Scan(&createdAt)

	if err != nil {
		r.log.Error("Error creating contact", logger.Error(err))
		return nil, err
	}

	formattedCreatedAt := createdAt.Format("2006-01-02 15:04:05 MST")

	return &models.Contact{
		ID:          id.String(),
		UserID:      contact.UserID,
		FirstName:   contact.FirstName,
		LastName:    contact.LastName,
		MiddleName:  contact.MiddleName,
		PhoneNumber: contact.PhoneNumber,
		CreatedAt:   formattedCreatedAt,
	}, nil
}

func (c *ContactRepo) GetAll(ctx context.Context, req *models.GetAllContactsRequest, user_id string) (*models.GetAllContactsResponse, error) {
	var response models.GetAllContactsResponse
	var (
		created_at time.Time
		updated_at sql.NullTime
	)

	// Calculate offset based on page and limit
	offset := (req.Page - 1) * req.Limit

	// Prepare the SQL query without the filter
	query := `SELECT count(id) OVER(), id, user_id, first_name, last_name, middle_name, phone_number, created_at, updated_at
		FROM "contacts" WHERE user_id=$1 OFFSET $2 LIMIT $3`

	// Execute the query with the necessary parameters
	rows, err := c.db.Query(ctx, query, user_id, offset, req.Limit)
	if err != nil {
		c.log.Error("Error getting all contacts", logger.Error(err))
		return nil, err
	}

	// Load the timezone (e.g., Asia/Tashkent)
	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		c.log.Error("Error loading timezone", logger.Error(err))
		return nil, err
	}

	defer rows.Close()

	// Iterate over the query result rows
	for rows.Next() {
		var contact models.Contact

		// Scan the row into the contact struct
		err := rows.Scan(
			&response.Count,
			&contact.ID,
			&contact.UserID,
			&contact.FirstName,
			&contact.LastName,
			&contact.MiddleName,
			&contact.PhoneNumber,
			&created_at,
			&updated_at,
		)
		if err != nil {
			c.log.Error("Error scanning contact", logger.Error(err))
			return nil, err
		}

		// Format the dates according to the timezone
		var formattedUpdatedAt string
		if updated_at.Valid {
			formattedUpdatedAt = updated_at.Time.In(location).Format("2006-01-02 15:04:05")
		} else {
			formattedUpdatedAt = ""
		}

		contact.CreatedAt = created_at.In(location).Format("2006-01-02 15:04:05")
		contact.UpdatedAt = formattedUpdatedAt

		// Append the contact to the response
		response.Contacts = append(response.Contacts, contact)
	}

	// Check if there were any issues during iteration
	if err := rows.Err(); err != nil {
		c.log.Error("Error iterating over contacts", logger.Error(err))
		return nil, err
	}

	return &response, nil
}

func (c *ContactRepo) GetById(ctx context.Context, id string, userid string) (*models.Contact, error) {
	var contacts models.Contact
	var (
		created_at time.Time
		updated_at sql.NullTime
	)

	// Define the SQL query
	query := `SELECT id, user_id, first_name, last_name, middle_name, phone_number, created_at, updated_at 
	          FROM "contacts" WHERE id = $1 AND user_id = $2`

	// Log the query and parameters for debugging
	c.log.Info("Executing query", logger.String("query", query), logger.String("contactID", id), logger.String("userID", userid))

	// Execute the query and scan the result
	err := c.db.QueryRow(ctx, query, id, userid).Scan(
		&contacts.ID,
		&userid,
		&contacts.FirstName,
		&contacts.LastName,
		&contacts.MiddleName,
		&contacts.PhoneNumber,
		&created_at,
		&updated_at,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.log.Warn("Contact not found", logger.String("contactID", id), logger.String("userID", userid))
			return nil, fmt.Errorf("contact not found")
		}
		c.log.Error("Error executing query", logger.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Format the `created_at` and `updated_at` timestamps
	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		c.log.Error("Error loading timezone", logger.Error(err))
		return nil, fmt.Errorf("failed to load timezone: %w", err)
	}

	contacts.CreatedAt = created_at.In(location).Format("2006-01-02 15:04:05 MST")
	if updated_at.Valid {
		contacts.UpdatedAt = updated_at.Time.In(location).Format("2006-01-02 15:04:05 MST")
	} else {
		contacts.UpdatedAt = ""
	}

	c.log.Info("Contact retrieved successfully", logger.Any("contact", contacts))
	return &contacts, nil
}

func (c *ContactRepo) Update(ctx context.Context, contact *models.Contact, user_id string) (*models.Contact, error) {
	var createdAt, updatedAt time.Time

	query := `UPDATE "contacts" SET 
		first_name = $1,
		last_name = $2,
		middle_name = $3,
		phone_number = $4,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND user_id = $6
		RETURNING created_at, updated_at`

	// Execute query
	err := c.db.QueryRow(ctx, query,
		contact.FirstName,   
		contact.LastName,    
		contact.MiddleName,  
		contact.PhoneNumber, 
		contact.ID,          
		user_id,             
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		c.log.Error("Error updating contact", logger.Error(err))
		return nil, fmt.Errorf("error updating contact: %w", err)
	}

	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		c.log.Error("Error loading timezone", logger.Error(err))
		return nil, fmt.Errorf("timezone error: %w", err)
	}
	contact.CreatedAt = createdAt.In(location).Format("2006-01-02 15:04:05 MST")
	contact.UpdatedAt = updatedAt.In(location).Format("2006-01-02 15:04:05 MST")

	return contact, nil
}

func (c *ContactRepo) Delete(ctx context.Context, id, userid string) error {
	query := `DELETE FROM "contacts" WHERE id = $1 AND user_id = $2`
	_, err := c.db.Exec(ctx, query, id, userid)
	if err != nil {
		c.log.Error("Error deleting contacts", logger.Error(err))
		return err
	}
	return nil
}
