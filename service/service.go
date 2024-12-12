package service

import (
	"task/pkg/logger"
	"task/storage"
)

type IServiceManager interface {
	Auth() authService
}

type Service struct {
	auth   authService
	logger logger.LoggerI
}

func New(storage storage.IStorage, log logger.LoggerI, redis storage.IRedisStorage) Service {
	return Service{
		auth:   NewAuthService(storage, log, redis),
		logger: log,
	}
}

func (s Service) Auth() authService {
	return s.auth
}
