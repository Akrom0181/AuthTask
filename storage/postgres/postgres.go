package postgres

import (
	"context"
	"fmt"
	"task/config"
	"task/pkg/logger"
	"task/storage"
	"task/storage/redis"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Store struct {
	Pool    *pgxpool.Pool
	log     logger.LoggerI
	user    *UserRepo
	device  *DeviceRepo
	contact *ContactRepo
	cfg     config.Config
}

func (s *Store) CloseDB() {
	s.Pool.Close()
}

func NewConnectionPostgres(cfg *config.Config) (storage.IStorage, error) {
	connect, err := pgxpool.ParseConfig(fmt.Sprintf(
		"host=%s user=%s dbname=%s password=%s port=%d ",
		cfg.PostgresHost,
		cfg.PostgresUser,
		cfg.PostgresDatabase,
		cfg.PostgresPassword,
		cfg.PostgresPort,
	))

	if err != nil {
		return nil, err
	}
	connect.MaxConns = 100

	pgxpool, err := pgxpool.ConnectConfig(context.Background(), connect)
	if err != nil {
		return nil, err
	}

	// Optional: Check database connectivity
	if err := pgxpool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}

	var loggerLevel = new(string)
	log := logger.NewLogger("app", *loggerLevel)

	return &Store{
		Pool: pgxpool,
		log:  log,
		cfg:  *cfg,
	}, nil
}

func (s *Store) User() storage.IUserStorage {
	if s.user == nil {
		s.user = &UserRepo{
			db:  s.Pool,
			log: s.log,
		}
	}
	return s.user
}

func (s *Store) Contact() storage.IContactStorage {
	if s.contact == nil {
		s.contact = &ContactRepo{
			db:  s.Pool,
			log: s.log,
		}
	}
	return s.contact
}

func (s *Store) Device() storage.IDeviceStorage {
	if s.device == nil {
		s.device = &DeviceRepo{
			db:  s.Pool,
			log: s.log,
		}
	}
	return s.device
}

func (s *Store) Redis() storage.IRedisStorage {
	return redis.New(s.cfg)
}
