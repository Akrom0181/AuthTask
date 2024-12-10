package postgres

import (
	"context"
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
		INSERT INTO "devices" (id, user_id, device_info, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		RETURNING id, user_id, device_info, created_at;
	`
	row := s.db.QueryRow(ctx, query, id.String(), device.UserID, device.DeviceInfo)

	var createdDevice models.Device
	err := row.Scan(&createdDevice.ID, &createdDevice.UserID, &createdDevice.DeviceInfo, &created_at)
	if err != nil {
		return nil, fmt.Errorf("failed to insert device: %w", err)
	}
	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		s.log.Error("Error loading timezone", logger.Error(err))
		return nil, fmt.Errorf("timezone error: %w", err)
	}
	createdDevice.CreatedAt = created_at.In(location).Format("2006-01-02 15:04:05 MST")

	return &createdDevice, nil
}

// GetAll retrieves all devices for a given user ID
func (s *DeviceRepo) GetAll(ctx context.Context, userId string) (*[]models.Device, error) {

	query := `
		SELECT id, user_id, device_info, created_at
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
		err := rows.Scan(&device.ID, &device.UserID, &device.DeviceInfo, &created_at)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		location, err := time.LoadLocation("Asia/Tashkent")

		if err != nil {
			s.log.Error("Error loading timezone", logger.Error(err))
			return nil, fmt.Errorf("timezone error: %w", err)
		}

		device.CreatedAt = created_at.In(location).Format("2006-01-02 15:04:05 MST")
		devices = append(devices, device)
	}

	return &devices, nil
}

// GetDeviceCount returns the number of devices associated with a given user
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

// Delete deletes a device by its ID
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

// Remove removes a device by its ID
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
