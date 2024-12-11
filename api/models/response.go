package models

type Response struct {
	StatusCode  int         `json:"Status_code"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
}
