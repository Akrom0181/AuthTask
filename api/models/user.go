package models

type User struct {
	ID          string    `json:"id,omitempty"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	// Contacts    []Contact `json:"contacts"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at,omitempty"`
}

type CreateUser struct {
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	// PhoneNumber string    `json:"phone_number"`
	// Contacts    []Contact `json:"contacts"`
}

type UpdateUser struct {
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	// PhoneNumber string    `json:"phone_number"`
	// Contacts    []Contact `json:"contacts"`
}

type GetUser struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	PhoneNumber string    `json:"phone_number"`
	// Contacts    []Contact `json:"contacts"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at,omitempty"`
}

type GetAllUsersRequest struct {
	Search string `json:"search"`
	Page   uint64 `json:"page"`
	Limit  uint64 `json:"limit"`
}

type GetAllUsersResponse struct {
	Users []User `json:"users"`
	Count int64  `json:"count"`
}
