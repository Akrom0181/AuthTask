package models

type Contact struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id,omitempty"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	MiddleName  string `json:"middle_name"`
	PhoneNumber string `json:"phone_number"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	DeletedAt   string `json:"deleted_at,omitempty"`
}

type CreateContact struct {
	// UserID      string `json:"user_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	MiddleName  string `json:"middle_name"`
	PhoneNumber string `json:"phone_number"`
}

type UpdateContact struct {
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	MiddleName  *string `json:"middle_name,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
}

type GetContact struct {
	ID string `json:"id"`
	// UserID      string `json:"user_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	MiddleName  string `json:"middle_name"`
	PhoneNumber string `json:"phone_number"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type DeleteContact struct {
	ID     string `json:"id"`
	UserId string `json:"user_id"`
}

type GetAllContactsRequest struct {
	Search string `json:"search"`
	Page   uint64 `json:"page"`
	Limit  uint64 `json:"limit"`
}

type GetAllContactsResponse struct {
	Contacts []Contact `json:"contacts"`
	Count    int64     `json:"count"`
}
