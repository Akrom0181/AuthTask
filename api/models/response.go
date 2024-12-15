package models

type Response struct {
	StatusCode  int         `json:"status_code"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
	Error       interface{} `json:"error"`
}
