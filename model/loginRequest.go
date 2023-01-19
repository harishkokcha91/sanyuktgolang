package model

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Mobile   string `json:"user_mobile"`
	Otp      string `json:"user_otp"`
}
