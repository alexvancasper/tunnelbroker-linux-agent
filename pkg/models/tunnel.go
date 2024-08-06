package models

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type Tunnel struct {
	gorm.Model
	Configured         bool   `json:"configured"`
	UserID             uint   `json:"userid"`
	IPv6ClientEndpoint string `json:"ipv6clientendpoint"`
	IPv6ServerEndpoint string `json:"ipv6serverendpoint"`
	PD                 string `json:"pd"`
	IPv4Remote         string `json:"ipv4remote"`
	IPv4Local          string `json:"ipv4local"`
	TunnelName         string `json:"tunnelname"`
}

func (t Tunnel) Marshal() ([]byte, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("Tunnel Marshal error %w", err)
	}
	return data, nil
}
