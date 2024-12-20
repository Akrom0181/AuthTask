package config

import "os"

const (
	ERR_INFORMATION     = "The server has received the request and is continuing the process"
	SUCCESS             = "The request was successful"
	ERR_REDIRECTION     = "You have been redirected and the completion of the request requires further action"
	ERR_BADREQUEST      = "Bad request"
	ERR_INTERNAL_SERVER = "While the request appears to be valid, the server could not complete the request"
	USER_ROLE           = "user"
	ADMIN_ROLE          = "admin"
	Status_NEW          = "new"
	Status_IN_PROCESS   = "in-process"
	Status_FINISHED     = "finished"
	Status_CANCELED     = "canceled"
	SmtpServer          = "smtp.gmail.com"
	SmtpPort            = "587"
	SmtpUsername        = "akromjonotaboyev@gmail.com"
	SmtpPassword        = "xgap tptk zutm ueep"
	ErrOtpExpired       = "otp has expired. Please request a new one"
	ErrOtpMismatch      = "incorrect OTP. Please try again"
	ErrOtpInvalidFormat = "invalid OTP data format"
	ErrUserCreation     = "failed to create user"
	ErrDeviceInsertion  = "failed to insert device"
	ErrTokenGeneration  = "failed to generate tokens"
)

var SignedKey = []byte(os.Getenv("SECRET_KEY_JWT"))

const (
	DebugMode   = "debug"
	TestMode    = "test"
	ReleaseMode = "release"
)
