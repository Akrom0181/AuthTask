package models

type Device struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	Name            string `json:"name"`
	NotificationKey string `json:"notificationKey"`
	Type            string `json:"type"`
	OsVersion       string `json:"osVersion"`
	AppVersion      string `json:"appVersion"`
	RememberMe      bool   `json:"remember_me"`
	AdId            string `json:"adId"`
	CreatedAt       string `json:"created_at"`
}

type CreateDevice struct {
	Name            string `json:"name"`
	NotificationKey string `json:"notificationKey"`
	Type            string `json:"type"`
	OsVersion       string `json:"osVersion"`
	AppVersion      string `json:"appVersion"`
	RememberMe      bool   `json:"remember_me"`
	AdId            string `json:"adId"`
}

type GetDevice struct {
	Name            string `json:"name"`
	NotificationKey string `json:"notificationKey"`
	Type            string `json:"type"`
	OsVersion       string `json:"osVersion"`
	AppVersion      string `json:"appVersion"`
	RememberMe      bool   `json:"remember_me"`
	AdId            string `json:"adId"`
	CreatedAt       string `json:"createdAt"`
}
