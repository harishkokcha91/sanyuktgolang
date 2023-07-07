package domain

type UserProfile struct {
	UserName  string `json:"user_name"`
	UserAge   string `json:"user_age"`
	Mobile    string `json:"user_mobile"`
	Otp       string `json:"user_otp"`
	Role      string `json:"user_role"`
	CreatedOn string `json:"created_on"`
	UpdateOn  string `json:"updated_on"`
}
