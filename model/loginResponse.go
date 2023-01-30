package model

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UserName     string `json:"user_name,omitempty"`
	UserMobile   string `json:"user_mobile,omitempty"`
}
