package service

import (
	"database/sql"

	"github.com/hako/branca"
)

//Service es el core de la aplicación
type Service struct {
	db    *sql.DB
	codec *branca.Branca
}

//New create a new service of connection
func New(db *sql.DB, codec *branca.Branca) *Service {
	return &Service{
		db:    db,
		codec: codec,
	}
}
