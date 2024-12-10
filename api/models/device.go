package models

type Device struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	DeviceInfo string `json:"device_info"`
	CreatedAt  string `json:"created_at"`
}

type CreateDevice struct {
	DeviceInfo string `json:"device_info"`
}

type GetDevice struct {
	DeviceInfo string `json:"device_info"`
	CreatedAt  string `json:"created_at"`
}


