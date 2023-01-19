package domain

import "github.com/jmoiron/sqlx"

type LoginRepositoryDb struct {
	client *sqlx.DB
}

func NewLoginRepositoryDb(dbClient *sqlx.DB) LoginRepositoryDb {
	return LoginRepositoryDb{dbClient}
}
