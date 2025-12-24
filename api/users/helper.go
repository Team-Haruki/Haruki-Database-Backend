package users

import (
	"crypto/rand"
	"math/big"

	"haruki-database/database/schema/users"
)

func NewUserService(client *users.Client) *UserService {
	return &UserService{client: client}
}

func NewUserHandler(svc *UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func generateUserID() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + 100000, nil
}
