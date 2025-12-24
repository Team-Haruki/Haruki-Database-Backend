package censor

import (
	"haruki-database/utils/censor"
)

type NameRequest struct {
	Server string `json:"server"`
	UserID string `json:"userID"`
	Name   string `json:"name"`
}

type ShortBioRequest struct {
	Server  string `json:"server"`
	UserID  string `json:"userID"`
	Content string `json:"content"`
}

type CensorService struct {
	service *censor.Service
}

type CensorHandler struct {
	svc *CensorService
}
