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

func toUserResponse(u *users.User) UserResponse {
	return UserResponse{
		ID:                     u.ID,
		Platform:               u.Platform,
		UserID:                 u.UserID,
		BanState:               u.BanState,
		BanReason:              u.BanReason,
		PjskBanState:           u.PjskBanState,
		PjskBanReason:          u.PjskBanReason,
		ChunithmBanState:       u.ChunithmBanState,
		ChunithmBanReason:      u.ChunithmBanReason,
		PjskMainBanState:       u.PjskMainBanState,
		PjskMainBanReason:      u.PjskMainBanReason,
		PjskRankingBanState:    u.PjskRankingBanState,
		PjskRankingBanReason:   u.PjskRankingBanReason,
		PjskAliasBanState:      u.PjskAliasBanState,
		PjskAliasBanReason:     u.PjskAliasBanReason,
		PjskMysekaiBanState:    u.PjskMysekaiBanState,
		PjskMysekaiBanReason:   u.PjskMysekaiBanReason,
		ChunithmMainBanState:   u.ChunithmMainBanState,
		ChunithmMainBanReason:  u.ChunithmMainBanReason,
		ChunithmAliasBanState:  u.ChunithmAliasBanState,
		ChunithmAliasBanReason: u.ChunithmAliasBanReason,
	}
}
