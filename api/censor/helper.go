package censor

import (
	"haruki-database/utils/censor"
)

func NewCensorService(service *censor.Service) *CensorService {
	return &CensorService{service: service}
}

func NewCensorHandler(svc *CensorService) *CensorHandler {
	return &CensorHandler{svc: svc}
}
