package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type key string

var (
	//TokenLifespan calcule the time of life of the Token.
	TokenLifespan = time.Hour * 24 * 14

	//KeyAuthUser es la Clave para ayutorizar al usuuario en el contexto
	KeyAuthUser key = "auth_user_id"

	// KeyUnauthenticated es cuando el usuario no está autenticado en el contexto
	KeyUnauthenticated = errors.New("unauthenticated")
)

//LoginOutput respuesta del servidor.
type LoginOutput struct {
	Token     string    `json:"token,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	AuthUser  User      `json:"auth_user,omitempty"`
}

//Login implementa la seguridad
func (s *Service) Login(ctx context.Context, email, password string) (LoginOutput, error) {
	var out LoginOutput

	email = strings.TrimSpace(email)
	if !rxEmail.MatchString(email) {
		return out, ErrInvalideEmail
	}

	password = strings.TrimSpace(password)
	if password == "" {
		return out, ErrInvalidPassword
	}

	query := "SELECT id, username, password from user where email=?"

	var key string

	err := s.db.QueryRowContext(ctx, query, email).Scan(&out.AuthUser.ID, &out.AuthUser.Username, &key)

	if err == sql.ErrNoRows {
		return out, ErrUserNotFound
	}

	if err != nil {
		return out, fmt.Errorf("No se encontro ningun registro user: %v", err)
	}

	hashedPasswordFromDatabase := []byte(key)
	val := bcrypt.CompareHashAndPassword(hashedPasswordFromDatabase, []byte(password))

	if val != nil {
		return out, ErrUserNotFound
	}

	out.Token, err = s.codec.EncodeToString(strconv.FormatInt(out.AuthUser.ID, 10))

	if err != nil {
		return out, fmt.Errorf("No se pudo generar el token: %v", err)
	}

	out.ExpiresAt = time.Now().Add(TokenLifespan)

	return out, nil
}

//AuthUserID Evaluar token
func (s *Service) AuthUserID(token string) (int64, error) {
	str, err := s.codec.DecodeToString(token)

	if err != nil {
		return 0, fmt.Errorf("No se puede decodificar el token: %v", err)
	}

	i, err := strconv.ParseInt(str, 10, 64)

	if err != nil {
		return 0, fmt.Errorf("No se puede obtener el id del usuario en el token: %v", err)
	}

	return i, nil
}

// AuthUser crea una consulta sobre el contexto
func (s *Service) AuthUser(ctx context.Context) (User, error) {
	var u User
	uid, ok := ctx.Value(KeyAuthUser).(int64)

	if !ok {
		return u, ErrUnauthenticated
	}
	query := "SELECT username FROM user WHERE id=?"
	err := s.db.QueryRowContext(ctx, query, uid).Scan(&u.Username)

	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("Error en la consulta verifique: %v", err)
	}

	u.ID = uid
	return u, nil

}
