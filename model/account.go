package model

import (
	"encoding/json"
)

type Account struct {
	Name  string `json:"name"`
	Token string `json:"token"`
	Host  string `json:"host"`
	// Cfg   *rest.Config
}

// NewLoginDto is constructor.
func NewAccount() *Account {
	return &Account{}
}

// ToString is return string of object
func (l *Account) ToString() (string, error) {
	bytes, err := json.Marshal(l)
	return string(bytes), err
}
