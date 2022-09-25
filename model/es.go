package model

import (
	"encoding/json"
)

type ES struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Version   string `json:"version"`
	NodeSets  []NS   `json:"nodesets"`
	// EntryPoint string `json:"entrypoint"`
	// Token      string `json:"token"`
}

type NS struct {
	Name         string      `json:"name"`
	NodeRole     string      `json:"roles"`
	Count        interface{} `json:"count"`
	DiskSize     string      `json:"diskSize"`
	StorageClass string      `json:"storageClass"`
}

// NewLoginDto is constructor.
func NewES() *ES {
	return &ES{}
}

// ToString is return string of object
func (l *ES) ToString() (string, error) {
	bytes, err := json.Marshal(l)
	return string(bytes), err
}

type ESEntryPoint struct {
	Name       string `json:"name"`
	Namespace  string `json:"name"`
	EntryPoint string `json:"entrypiont"`
	UserName   string `json:username`
	Password   string `json:password`
	// EntryPoint string `json:"entrypoint"`
	// Token      string `json:"token"`
}
