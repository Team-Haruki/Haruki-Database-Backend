package censor

import (
	"haruki-database/database/schema/users"
	"haruki-database/utils/censor"

	"github.com/redis/go-redis/v9"
)

type NameRequest struct {
	Server       string `json:"server"`
	UserID       string `json:"user_id"`
	Name         string `json:"name"`
	HarukiUserID int    `json:"haruki_user_id"`
}

type ShortBioRequest struct {
	Server       string `json:"server"`
	UserID       string `json:"user_id"`
	Content      string `json:"content"`
	HarukiUserID int    `json:"haruki_user_id"`
}

type CensorService struct {
	service *censor.Service
}

type CensorHandler struct {
	svc         *CensorService
	usersClient *users.Client
	redisClient *redis.Client
}
