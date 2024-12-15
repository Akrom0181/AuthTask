package models

type UserLoginRequest struct {
	MobilePhone string `json:"phone_number"`
}

type UserLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthInfo struct {
	UserID   string `json:"user_id"`
	UserRole string `json:"user_role"`
}

type UserRegisterRequest struct {
	MobilePhone string      `json:"phone_number"`
	User        *CreateUser `json:"user"`
}

type UserRegisterConfRequest struct {
	MobilePhone string       `json:"phone_number"`
	Otp         string       `json:"otp"`
	Identifier  string       `json:"identifier"`
	DeviceInfo  CreateDevice `json:"device_info"`
}

type UserLoginPhoneConfirmRequest struct {
	MobilePhone string       `json:"phone_number"`
	SmsCode     string       `json:"otp"`
	Identifier  string       `json:"identifier"`
	DeviceInfo  CreateDevice `json:"device_info"`
}

type UserChangePhone struct {
	MobilePhone string `json:"phone_number"`
}

type UserChangePhoneConfirm struct {
	MobilePhone string `json:"phone_number"`
	SmsCode     string `json:"smscode"`
}

type SUserRegisterConfRequest struct {
	MobilePhone string `json:"phone_number"`
	Otp         string `json:"otp"`
}
