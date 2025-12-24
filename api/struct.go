package api

type UserInfo struct {
	HarukiUserID int
	Platform     string
	UserID       string
	BanState     bool
	BanReason    string
}

type HarukiAPIResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type HarukiAPIDataResponse[T any] struct {
	HarukiAPIResponse
	Data T `json:"data,omitempty"`
}
