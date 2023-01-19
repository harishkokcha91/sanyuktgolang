package domain

import (
	"database/sql"
)

type Users struct {
	Id        int64          `db:"user_id"`
	Name      sql.NullString `db:"user_name,omitempty"`
	Mobile    string         `db:"user_mobile"`
	Otp       string         `db:"user_otp"`
	Role      string         `db:"user_role"`
	CreatedOn string         `db:"created_on"`
	UpdateOn  string         `db:"UpdatedOn"`
}
