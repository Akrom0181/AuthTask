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
    VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Tashkent')`

	// Execute the insert query
	_, err := r.db.Exec(ctx, query,
		id.String(),
		contact.UserID,
		contact.FirstName,
		contact.LastName,
		contact.MiddleName,
		contact.PhoneNumber)

	if err != nil {
		r.log.Error("Error creating contact", logger.Error(err))
		return nil, err
	}

	// After insert, retrieve the created_at value
	var createdAt time.Time
	selectQuery := `SELECT created_at FROM "contacts" WHERE id = $1`

	err = r.db.QueryRow(ctx, selectQuery, id.String()).Scan(&createdAt)
	if err != nil {
		r.log.Error("Error retrieving created_at", logger.Error(err))
		return nil, err
	}

	// Format the createdAt timestamp to the desired format
	formattedCreatedAt := createdAt.Format("2006-01-02 15:04:05")

	// Return the created contact with the formatted created_at value
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
		created_at sql.NullTime
		updated_at sql.NullTime
		filter     string
	)

	offset := (req.Page - 1) * req.Limit

	if req.Search != "" {
		filter += fmt.Sprintf(` AND (first_name ILIKE '%%%v%%' OR phone_number ILIKE '%%%v%%') `, req.Search, req.Search)
	}
	filter += fmt.Sprintf(" OFFSET %v LIMIT %v", offset, req.Limit)

	query := `SELECT count(id) OVER(), id, user_id, first_name, last_name, middle_name, phone_number, created_at, updated_at
		FROM "contacts" WHERE user_id=$1` + filter

	rows, err := c.db.Query(ctx, query, user_id)
	if err != nil {
		c.log.Error("Error getting all contacts", logger.Error(err))
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var contact models.Contact

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

		if created_at.Valid {
			contact.CreatedAt = created_at.Time.Format("2006-01-02 15:04:05")
		} else {
			contact.UpdatedAt = ""
		}

		if updated_at.Valid {
			contact.UpdatedAt = updated_at.Time.Local().Format("2006-01-02 15:04:05")
		} else {
			contact.UpdatedAt = ""
		}

		response.Contacts = append(response.Contacts, contact)
	}

	if err := rows.Err(); err != nil {
		c.log.Error("Error iterating over contacts", logger.Error(err))
		return nil, err
	}

	return &response, nil
}

func (c *ContactRepo) GetById(ctx context.Context, id string, userid string) (*models.Contact, error) {

	var contacts models.Contact
	var (
		created_at sql.NullTime
		updated_at sql.NullTime
	)

	query := `SELECT id, user_id, first_name, last_name, middle_name, phone_number, created_at, updated_at 
	          FROM "contacts" WHERE id = $1 AND user_id = $2`

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

		c.log.Error("Database error", logger.Error(err))
		return nil, fmt.Errorf("internal server error")
	}

	if created_at.Valid {
		contacts.CreatedAt = created_at.Time.Format("2006-01-02 15:04:05")
	} else {
		contacts.CreatedAt = ""
	}

	if updated_at.Valid {
		contacts.UpdatedAt = updated_at.Time.Local().Format("2006-01-02 15:04:05")
	} else {
		contacts.UpdatedAt = ""
	}

	c.log.Info("Contact retrieved successfully", logger.Any("contact", contacts))
	return &contacts, nil
}

func (c *ContactRepo) Update(ctx context.Context, contact *models.Contact, userID string) (*models.Contact, error) {
	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		c.log.Error("Error loading timezone", logger.Error(err))
		return nil, fmt.Errorf("failed to load timezone: %w", err)
	}

	query := `
		UPDATE "contacts" SET 
			first_name = COALESCE($1, first_name),
			last_name = COALESCE($2, last_name),
			middle_name = COALESCE($3, middle_name),
			phone_number = COALESCE($4, phone_number),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND user_id = $6
		RETURNING created_at, updated_at
	`

	var (
		createdAt sql.NullTime
		updatedAt time.Time
	)
	err = c.db.QueryRow(ctx, query,
		contact.FirstName,
		contact.LastName,
		contact.MiddleName,
		contact.PhoneNumber,
		contact.ID,
		userID,
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		c.log.Error("Error updating contact", logger.Error(err))
		return nil, fmt.Errorf("error updating contact: %w", err)
	}

	contact.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
	contact.UpdatedAt = updatedAt.In(location).Format("2006-01-02 15:04:05")

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
