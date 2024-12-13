package storage

import (
	"context"
	"task/api/models"
	"time"
)

type IStorage interface {
	CloseDB()
	User() IUserStorage
	Contact() IContactStorage
	Device() IDeviceStorage
	Redis() IRedisStorage
}

type IUserStorage interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	Update(ctx context.Context, user *models.User, user_id string) (*models.User, error)
	GetAll(ctx context.Context, req *models.GetAllUsersRequest) (*models.GetAllUsersResponse, error)
	GetById(ctx context.Context, id string) (*models.User, error)
	Delete(ctx context.Context, id string) error
	GetByLogin(ctx context.Context, login string) (models.User, error)
	CheckPhoneNumberExist(ctx context.Context, id string) (models.User, error)
	GetByPhone(ctx context.Context, number string) (*models.User, error)
	UpdatePhoneNumber(ctx context.Context, userid string, number string) (string, error)
}

type IContactStorage interface {
	Create(ctx context.Context, contact *models.Contact) (*models.Contact, error)
	Update(ctx context.Context, contact *models.Contact, user_id string) (*models.Contact, error)
	GetAll(ctx context.Context, req *models.GetAllContactsRequest, user_id string) (*models.GetAllContactsResponse, error)
	GetById(ctx context.Context, id string, userid string) (*models.Contact, error)
	Delete(ctx context.Context, id string, userid string) error
}

type IDeviceStorage interface {
	Insert(ctx context.Context, device *models.Device) (*models.Device, error)
	GetAll(ctx context.Context, id string) (*[]models.Device, error)
	GetByID(ctx context.Context, id, user_id string) (*models.Device, error)
	GetDeviceCount(ctx context.Context, userId string) (int, error)
	Delete(ctx context.Context, id string, userid string) error
	Remove(ctx context.Context, id string) error
	GetByIdRemove(ctx context.Context, id string) (*models.Device, error)
}

type IRedisStorage interface {
	SetX(ctx context.Context, key string, value interface{}, duration time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Del(ctx context.Context, key string) error
}
