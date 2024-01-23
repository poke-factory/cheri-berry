package requests

import "time"

type LoginRequest struct {
	Id       string        `json:"_id"`
	Name     string        `json:"name"`
	Password string        `json:"password"`
	Type     string        `json:"type"`
	Roles    []interface{} `json:"roles"`
	Date     time.Time     `json:"date"`
}
