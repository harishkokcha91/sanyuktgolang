package domain

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"log"

	"sanyuktgolang/errs"
	"sanyuktgolang/logger"

	"github.com/jmoiron/sqlx"
)

type AuthRepository interface {
	FindBy(username string, password string) (*Login, *errs.AppError)
	VerifyOtp(mobile string, otp string) (*Users, *errs.AppError)
	FindByMobile(mobile string) (*Users, *errs.AppError)
	GenerateAndSaveRefreshTokenToStore(authToken AuthToken) (string, *errs.AppError)
	RefreshTokenExists(refreshToken string) *errs.AppError
}

type AuthRepositoryDb struct {
	client *sqlx.DB
}

func (d AuthRepositoryDb) RefreshTokenExists(refreshToken string) *errs.AppError {
	sqlSelect := "select refresh_token from refresh_token_store where refresh_token = ?"
	var token string
	err := d.client.Get(&token, sqlSelect, refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.NewAuthenticationError("refresh token not registered in the store")
		} else {
			logger.Error("Unexpected database error: " + err.Error())
			return errs.NewUnexpectedError("unexpected database error")
		}
	}
	return nil
}

func (d AuthRepositoryDb) GenerateAndSaveRefreshTokenToStore(authToken AuthToken) (string, *errs.AppError) {
	// generate the refresh token
	var appErr *errs.AppError
	var refreshToken string
	if refreshToken, appErr = authToken.newRefreshToken(); appErr != nil {
		return "", appErr
	}

	// store it in the store
	sqlInsert := "insert into refresh_token_store (refresh_token) values (?)"
	_, err := d.client.Exec(sqlInsert, refreshToken)
	if err != nil {
		logger.Error("unexpected database error: " + err.Error())
		return "", errs.NewUnexpectedError("unexpected database error")
	}
	return refreshToken, nil
}

func (d AuthRepositoryDb) FindBy(username, password string) (*Login, *errs.AppError) {
	var login Login

	sqlVerify := `SELECT username, customer_id, role FROM users WHERE username = ? and password = ?`
	logger.Debug(fmt.Sprintf("Sql %s: ...", sqlVerify))
	err := d.client.Get(&login, sqlVerify, username, password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.NewAuthenticationError("invalid credentials")
		} else {
			logger.Error("Error while verifying login request from database: " + err.Error())
			return nil, errs.NewUnexpectedError("Unexpected database error")
		}
	}
	return &login, nil
}

func (d AuthRepositoryDb) VerifyOtp(mobile, otp string) (*Users, *errs.AppError) {
	var login Users

	sqlVerify := `SELECT user_id,otp_verified FROM users_otp WHERE otp_verified=0 and user_mobile = ? and user_otp = ?`
	logger.Debug(fmt.Sprintf("Sql %s: ...", sqlVerify))
	err := d.client.Get(&login, sqlVerify, mobile, otp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.NewAuthenticationError("Invalid Otp")
		} else {
			logger.Error("Error while verifying login request from database: " + err.Error())
			return nil, errs.NewUnexpectedError("Unexpected database error")
		}
	}
	_, error := d.updateUserVerified(mobile)
	if error != nil {
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	return &login, nil
}

func (d AuthRepositoryDb) updateUserVerified(mobile string) (bool, *errs.AppError) {
	sql := `UPDATE users_otp SET otp_verified=1,updated_on=now() where user_mobile=?`
	insertResult, err := d.client.ExecContext(context.Background(), sql, mobile)
	logger.Error(sql)
	if err != nil {
		logger.Error(err.Error())
		return false, errs.NewAuthenticationError("Unable to update otp")
	}
	id, err := insertResult.LastInsertId()
	if err != nil {
		log.Fatalf("impossible to retrieve last inserted otp id: %s", err)
		return false, errs.NewAuthenticationError("Impossible to retrieve last inserted otp id")
	}
	log.Printf("inserted id: %d", id)

	return true, nil
}

func (d AuthRepositoryDb) FindByMobile(mobile string) (*Users, *errs.AppError) {
	var user Users

	sqlVerify := `SELECT user_id,user_name,user_mobile,user_role FROM sanyukt_users WHERE user_mobile = ?`
	logger.Debug(fmt.Sprintf("Sql %s: ...", sqlVerify))
	err := d.client.Get(&user, sqlVerify, mobile)
	if err != nil {
		if err == sql.ErrNoRows {
			//Create user
			user, err := d.CreateUser(mobile)
			if err != nil {
				return nil, err
			} else {
				_, err := d.GenerateOtp(mobile, user.Id)
				if err != nil {
					return nil, errs.NewAuthenticationError("invalid otp credentials")
				} else {
					return user, nil
				}
			}
		} else {
			logger.Error("Error while verifying login request from database: " + err.Error())
			return nil, errs.NewUnexpectedError("Unexpected database error")
		}
	}

	_, errs := d.GenerateOtp(mobile, user.Id)
	if errs != nil {
		return nil, errs
	} else {
		return &user, nil
	}

}
func (d AuthRepositoryDb) GenerateOtp(mobile string, userId int64) (bool, *errs.AppError) {
	isPresent, err := d.isUserOtpPresent(mobile)
	if err != nil {
		return false, err
	}
	if isPresent {
		isComplete, err := d.updateOtpForUser(mobile)
		return isComplete, err
	} else {
		sql := `INSERT INTO users_otp  (user_mobile,user_otp,otp_verified,user_id) VALUES (?,?,false,?)`
		insertResult, err := d.client.ExecContext(context.Background(), sql, mobile, getRandomSixDigit(), userId)
		logger.Error(sql)
		if err != nil {
			logger.Error(err.Error())
			return false, errs.NewAuthenticationError("Unable to insert otp")
		}
		id, err := insertResult.LastInsertId()
		if err != nil {
			log.Fatalf("impossible to retrieve last inserted otp id: %s", err)
			return false, errs.NewAuthenticationError("Impossible to retrieve last inserted otp id")
		}
		log.Printf("inserted id: %d", id)

		return true, nil
	}
}

func (d AuthRepositoryDb) updateOtpForUser(mobile string) (bool, *errs.AppError) {
	sql := `UPDATE users_otp SET user_otp = ?,otp_verified=0,updated_on=now() where user_mobile=?`
	insertResult, err := d.client.ExecContext(context.Background(), sql, getRandomSixDigit(), mobile)
	logger.Error(sql)
	if err != nil {
		logger.Error(err.Error())
		return false, errs.NewAuthenticationError("Unable to insert otp")
	}
	id, err := insertResult.LastInsertId()
	if err != nil {
		log.Fatalf("impossible to retrieve last inserted otp id: %s", err)
		return false, errs.NewAuthenticationError("Impossible to retrieve last inserted otp id")
	}
	log.Printf("inserted id: %d", id)

	return true, nil
}

func (d AuthRepositoryDb) isUserOtpPresent(mobile string) (bool, *errs.AppError) {
	var login Users

	sqlVerify := `SELECT user_mobile FROM users_otp WHERE user_mobile = ?`
	logger.Info(fmt.Sprintf("Sql %s: ...", sqlVerify))
	err := d.client.Get(&login, sqlVerify, mobile)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		} else {
			logger.Error("Error while verifying login request from database: " + err.Error())
			return false, errs.NewUnexpectedError("Unexpected database error")
		}
	}
	return true, nil
}

func (d AuthRepositoryDb) CreateUser(mobile string) (*Users, *errs.AppError) {
	sql := `INSERT INTO sanyukt_users  (user_mobile,user_role) VALUES (?,"user")`
	insertResult, err := d.client.ExecContext(context.Background(), sql, mobile)
	if err != nil {
		logger.Error(err.Error())
		return nil, errs.NewAuthenticationError("Unable to create user")
	}
	id, err := insertResult.LastInsertId()
	if err != nil {
		log.Fatalf("impossible to retrieve last inserted id: %s", err)
		return nil, errs.NewAuthenticationError("Impossible to retrieve last inserted id")
	}
	log.Printf("inserted id: %d", id)

	return &Users{Id: id, Mobile: mobile}, nil
}

func getRandomSixDigit() string {

	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	max := 6
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

func NewAuthRepository(client *sqlx.DB) AuthRepositoryDb {
	return AuthRepositoryDb{client}
}
