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

type DeviceRepo struct {
	db  *pgxpool.Pool
	log logger.LoggerI
}

func NewDeviceRepo(db *pgxpool.Pool, log logger.LoggerI) *DeviceRepo {
	return &DeviceRepo{
		db:  db,
		log: log,
	}
}

func (s *DeviceRepo) Insert(ctx context.Context, device *models.Device) (*models.Device, error) {
	id := uuid.New()
	var created_at time.Time
	query := `
		INSERT INTO "devices" (id, user_id, name, notification_key, type, os_version, app_version, remember_me, ad_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP)
		RETURNING id, user_id, name, notification_key, type, os_version, app_version, remember_me, ad_id, created_at;
	`
	row := s.db.QueryRow(ctx, query,
		id.String(),
		device.UserID,
		device.Name,
		device.NotificationKey,
		device.Type,
		device.OsVersion,
		device.AppVersion,
		device.RememberMe,
		device.AdId)

	err := row.Scan(
		&device.ID,
		&device.UserID,
		&device.Name,
		&device.NotificationKey,
		&device.Type,
		&device.OsVersion,
		&device.AppVersion,
		&device.RememberMe,
		&device.AdId,
		&created_at,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert device: %w", err)
	}
	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		s.log.Error("Error loading timezone", logger.Error(err))
		return nil, fmt.Errorf("timezone error: %w", err)
	}
	device.CreatedAt = created_at.In(location).Format("2006-01-02 15:04:05")

	return device, nil
}

func (s *DeviceRepo) GetAll(ctx context.Context, userId string) (*[]models.Device, error) {

	query := `
		SELECT id, user_id, name, notification_key, type, os_version, app_version, remember_me, ad_id, created_at
		FROM "devices"
		WHERE user_id = $1
		ORDER BY created_at ASC;
	`

	rows, err := s.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var device models.Device
		var created_at time.Time
		err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.Name,
			&device.NotificationKey,
			&device.Type,
			&device.OsVersion,
			&device.AppVersion,
			&device.RememberMe,
			&device.AdId,
			&created_at,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		device.CreatedAt = created_at.Format("2006-01-02 15:04:05")

		devices = append(devices, device)
	}

	return &devices, nil
}

func (s *DeviceRepo) GetDeviceCount(ctx context.Context, userId string) (int, error) {
	query := `
		SELECT COUNT(*) FROM "devices"
		WHERE user_id = $1;
	`

	var count int
	err := s.db.QueryRow(ctx, query, userId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get device count: %w", err)
	}

	return count, nil
}

func (s *DeviceRepo) Delete(ctx context.Context, deviceId, userid string) error {
	query := `
		DELETE FROM "devices"
		WHERE id = $1 AND user_id = $2;
	`

	_, err := s.db.Exec(ctx, query, deviceId, userid)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	return nil
}

func (s *DeviceRepo) Remove(ctx context.Context, deviceId string) error {
	query := `
		DELETE FROM "devices"
		WHERE id = $1;
	`

	_, err := s.db.Exec(ctx, query, deviceId)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	return nil
}

func (s *DeviceRepo) GetByID(ctx context.Context, id, user_id string) (*models.Device, error) {
	query := `Select id, user_id FROM "devices" WHERE id = $1 AND user_id = $2`

	var deviceID, deviceUserID string

	err := s.db.QueryRow(ctx, query, id, user_id).Scan(&deviceID, &deviceUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
		}
		s.log.Error("Error executing query", logger.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &models.Device{
		ID:     id,
		UserID: user_id,
	}, nil
}

func (s *DeviceRepo) GetByIdRemove(ctx context.Context, id string) (*models.Device, error) {
	query := `Select id FROM "devices" WHERE id = $1`

	var deviceID string

	err := s.db.QueryRow(ctx, query, id).Scan(&deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
		}
		s.log.Error("Error executing query", logger.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &models.Device{
		ID: id,
	}, nil
}
