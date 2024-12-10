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
	STATUS_NEW          = "new"
	STATUS_IN_PROCESS   = "in-process"
	STATUS_FINISHED     = "finished"
	STATUS_CANCELED     = "canceled"
	SmtpServer          = "smtp.gmail.com"
	SmtpPort            = "587"
	SmtpUsername        = "akromjonotaboyev@gmail.com"
	SmtpPassword        = "duriexakadbzalxw"
)

var SignedKey = []byte(os.Getenv("SECRET_KEY"))

const (
	DebugMode   = "debug"
	TestMode    = "test"
	ReleaseMode = "release"
)
