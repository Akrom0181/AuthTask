package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"task/api/models"
	"task/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserRepo struct {
	db  *pgxpool.Pool
	log logger.LoggerI
}

func NewUserRepo(db *pgxpool.Pool, log logger.LoggerI) *UserRepo {
	return &UserRepo{
		db:  db,
		log: log,
	}
}

func (u *UserRepo) Create(ctx context.Context, user *models.User) (*models.User, error) {
	id := uuid.New()
	query := `INSERT INTO users (id, first_name, last_name, phone_number, created_at) 
			  VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
			  RETURNING created_at`

	var createdAt time.Time

	err := u.db.QueryRow(ctx, query,
		id.String(),
		user.FirstName,
		user.LastName,
		user.PhoneNumber,
	).Scan(&createdAt)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return &models.User{}, err
	}

	// Load the timezone (e.g., Asia/Tashkent)
	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		u.log.Error("Error loading timezone", logger.Error(err))
		return nil, err
	}

	formattedCreatedAt := createdAt.In(location).Format("2006-01-02 15:04:05 MST")

	return &models.User{
		ID:          id.String(),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
		CreatedAt:   formattedCreatedAt,
	}, nil
}

// GetByID retrieves a user by their ID
func (r *UserRepo) GetById(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	var (
		created_at time.Time
		updated_at sql.NullTime
	)

	query := `SELECT id,  first_name, last_name, phone_number, created_at, updated_at 
	          FROM "users" WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.PhoneNumber,
		&created_at,
		&updated_at,
	)

	if err != nil {
		r.log.Error("Error getting user by ID", logger.Error(err))
		return nil, err
	}

	// Load the timezone (e.g., Asia/Tashkent)
	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		r.log.Error("Error loading timezone", logger.Error(err))
		return nil, err
	}

	// Format created_at to the desired format and timezone
	formattedCreatedAt := created_at.In(location).Format("2006-01-02 15:04:05")

	// Format updated_at if it is not null
	var formattedUpdatedAt string
	if updated_at.Valid {
		formattedUpdatedAt = updated_at.Time.In(location).Format("2006-01-02 15:04:05")
	} else {
		formattedUpdatedAt = ""
	}

	// Assign formatted values to user model
	user.CreatedAt = formattedCreatedAt
	user.UpdatedAt = formattedUpdatedAt

	return &user, nil
}

// GetAll retrieves all users with optional filters (e.g., search, pagination)
func (u *UserRepo) GetAll(ctx context.Context, req *models.GetAllUsersRequest) (*models.GetAllUsersResponse, error) {
	var response models.GetAllUsersResponse
	var (
		created_at time.Time
		updated_at sql.NullTime
	)
	offset := (req.Page - 1) * req.Limit
	filter := ""

	if req.Search != "" {
		filter = fmt.Sprintf(` WHERE (first_name ILIKE '%%%v%%' OR phone_number ILIKE '%%%v%%' )`, req.Search, req.Search)
	}

	query := fmt.Sprintf(`SELECT count(id) OVER(), id, first_name, last_name, phone_number, created_at, updated_at
		FROM "users" %s OFFSET $1 LIMIT $2`, filter)

	rows, err := u.db.Query(ctx, query, offset, req.Limit)
	if err != nil {
		u.log.Error("Error getting all users", logger.Error(err))
		return nil, err
	}

	// Load the timezone (e.g., Asia/Tashkent)
	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		u.log.Error("Error loading timezone", logger.Error(err))
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var user models.User
		// Scan into time.Time first for created_at and updated_at
		err := rows.Scan(
			&response.Count,
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.PhoneNumber,
			&created_at,
			&updated_at,
		)
		if err != nil {
			u.log.Error("Error scanning user", logger.Error(err))
			return nil, err
		}
		var formattedUpdatedAt string
		if updated_at.Valid {
			formattedUpdatedAt = updated_at.Time.In(location).Format("2006-01-02 15:04:05")
		} else {
			formattedUpdatedAt = ""
		}
		// Convert time.Time to the desired format and timezone
		user.CreatedAt = created_at.In(location).Format("2006-01-02 15:04:05")
		user.UpdatedAt = formattedUpdatedAt
		response.Users = append(response.Users, user)
	}

	if err := rows.Err(); err != nil {
		u.log.Error("Error iterating over users", logger.Error(err))
		return nil, err
	}

	return &response, nil
}

func (u *UserRepo) Update(ctx context.Context, user *models.User, user_id string) (*models.User, error) {
	var (
		created_at, updated_at time.Time
	)
	query := `UPDATE "users" SET 
		first_name = $1,
		last_name = $2,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING created_at, updated_at`

	err := u.db.QueryRow(ctx, query,
		user.FirstName,
		user.LastName,
		user_id,
	).Scan(&created_at, &updated_at)

	if err != nil {
		u.log.Error("Error updating user", logger.Error(err))
		return nil, err
	}
	// Load the timezone (e.g., Asia/Tashkent)
	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		u.log.Error("Error loading timezone", logger.Error(err))
		return nil, err
	}

	// Format created_at to the desired format and timezone
	formattedCreatedAt := created_at.In(location).Format("2006-01-02 15:04:05")
	formattedUpdatedAt := updated_at.In(location).Format("2006-01-02 15:04:05")

	return &models.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: formattedCreatedAt,
		UpdatedAt: formattedUpdatedAt,
	}, nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.log.Error("Error starting transaction", logger.Error(err))
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	deleteDevicesQuery := `DELETE FROM "devices" WHERE user_id = $1`
	_, err = tx.Exec(ctx, deleteDevicesQuery, id)
	if err != nil {
		r.log.Error("Error deleting devices", logger.Error(err))
		tx.Rollback(ctx)
		return fmt.Errorf("failed to delete devices: %w", err)
	}

	deleteUserQuery := `DELETE FROM "users" WHERE id = $1`
	_, err = tx.Exec(ctx, deleteUserQuery, id)
	if err != nil {
		r.log.Error("Error deleting user", logger.Error(err))
		tx.Rollback(ctx)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	deleteContactQuery := `DELETE FROM "contacts" WHERE user_id = $1`
	_, err = tx.Exec(ctx, deleteContactQuery, id)
	if err != nil {
		r.log.Error("Error deleting contact", logger.Error(err))
		tx.Rollback(ctx)
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.log.Error("Error committing transaction", logger.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.log.Info("User and associated devices deleted successfully", logger.String("userID", id))
	return nil
}

func (c *UserRepo) CheckPhoneNumberExist(ctx context.Context, id string) (models.User, error) {

	resp := models.User{}

	query := ` SELECT id FROM "users" WHERE phone_number = $1 `

	err := c.db.QueryRow(ctx, query, id).Scan(&resp.ID)
	if err != nil {
		return models.User{}, err
	}

	return resp, nil
}

func (c *UserRepo) GetByLogin(ctx context.Context, login string) (models.User, error) {
	var (
		first_name sql.NullString
		last_name  sql.NullString
		phone      sql.NullString
		createdat  sql.NullString
		updatedat  sql.NullString
	)

	query := `SELECT id, first_name, last_name, phone_number, created_at, updated_at
		FROM "users" WHERE phone_number = $1`

	row := c.db.QueryRow(ctx, query, login)

	user := models.User{}

	err := row.Scan(
		&user.ID,
		&first_name,
		&last_name,
		&phone,
		&createdat,
		&updatedat,
	)

	if err != nil {
		c.log.Error("failed to scan user by LOGIN from database", logger.Error(err))
		return models.User{}, err
	}

	user.FirstName = first_name.String
	user.LastName = last_name.String
	user.PhoneNumber = phone.String
	user.CreatedAt = createdat.String
	user.UpdatedAt = updatedat.String

	return user, nil
}

func (u *UserRepo) GetByPhone(ctx context.Context, number string) (*models.User, error) {
	var (
		user       = models.User{}
		created_at sql.NullString
		updated_at sql.NullString
	)
	if err := u.db.QueryRow(context.Background(), `SELECT id, first_name, last_name,
	 phone_number, created_at, 
	    updated_at FROM "users"
		  WHERE phone_number = $1`, number).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.PhoneNumber,
		&created_at,
		&updated_at,
	); err != nil {
		return &models.User{}, err
	}
	return &models.User{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
		CreatedAt:   created_at.String,
		UpdatedAt:   updated_at.String,
	}, nil
}

func (u *UserRepo) UpdatePhoneNumber(ctx context.Context, userID string, number string) (string, error) {
	query := `UPDATE "users" 
	          SET phone_number = $1, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $2`
	_, err := u.db.Exec(ctx, query, number, userID)
	if err != nil {
		u.log.Error("Error updating user phone number", logger.Error(err))
		return "", err
	}
	return "Phone number updated successfully", nil
}
