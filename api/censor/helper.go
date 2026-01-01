package censor

import (
	"haruki-database/database/schema/users"
	"haruki-database/utils/censor"

	"github.com/redis/go-redis/v9"
)

func NewCensorService(service *censor.Service) *CensorService {
	return &CensorService{service: service}
}

func NewCensorHandler(svc *CensorService, usersClient *users.Client, redisClient *redis.Client) *CensorHandler {
	return &CensorHandler{
		svc:         svc,
		usersClient: usersClient,
		redisClient: redisClient,
	}
}
