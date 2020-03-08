package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	//ERRORES
	//----------------------------------------------

	//ErrInvalideEmail is error of user invalid
	ErrInvalideEmail = errors.New("Error email invalido")

	//ErrInvalideUsername is error of user is incorrect.
	ErrInvalideUsername = errors.New("Error usuario invalido")

	//ErrInvalidUser is error of user exist in database.
	ErrInvalidUser = errors.New("Usuario ya existe")

	// ErrUserNotFound denotes a not found user.
	ErrUserNotFound = errors.New("user not found")

	//ErrUserOk es el mensaje para el usuario entrante y guardado correctamente
	ErrUserOk = errors.New("Usuario ingresado Correctamente")

	//ErrInvalidPassword is error of scriture of simboles.
	ErrInvalidPassword = errors.New("Ingrese una contraseña válida")

	//Regular Expressions
	//-----------------------------------------------
	//For insert correo electronico
	rxEmail = regexp.MustCompile("^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$")

	//For insert username.
	rxUsername = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]{0,17}$")

	// ErrUnauthenticated cuando el usuario no está autenticado
	ErrUnauthenticated = errors.New("Usted no está autenticado")

	//ErrForbiddenFollow para evitar el autoseguido de usuario
	ErrForbiddenFollow = errors.New("No se puede autoseguir")
)

//User model.
type User struct {
	ID       int64  `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

//UserProfile model.
type UserProfile struct {
	User
	Email          string `json:"email,omitempty"`
	FollowersCount int    `json:"followers_count"`
	FolloweesCount int    `json:"followees_count"`
	Me             bool   `json:"me"`
	Following      bool   `json:"following"`
	Followeed      bool   `json:"followed"`
}

//ToggleFollowOutput es la estructura para los seguidores
type ToggleFollowOutput struct {
	Following      bool `json:"following,omitempty"`
	FollowersCount int  `json:"followers_count,omitempty"`
}

//CreateUser insert a user in the database.
func (s *Service) CreateUser(ctx context.Context, email, username, password string) error {
	//db := mysql.Connect()

	email = strings.TrimSpace(email)
	if !rxEmail.MatchString(email) {
		return ErrInvalideEmail
	}

	username = strings.TrimSpace(username)
	if !rxUsername.MatchString(username) {
		return ErrInvalideUsername
	}

	password = strings.TrimSpace(password)
	if password == "" {
		return ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Password do not hashed")
	}

	query := "INSERT INTO user (email, username, password) VALUES (?, ?, ?)"
	_, err = s.db.ExecContext(ctx, query, email, username, string(hashedPassword))

	if err != nil {
		return ErrInvalidUser
	}

	return ErrUserOk
}

//User selecciona el usuario de la base de datos
func (s *Service) User(ctx context.Context, username string) (UserProfile, error) {
	var u UserProfile

	username = strings.TrimSpace(username)
	if !rxUsername.MatchString(username) {
		return u, ErrInvalideUsername
	}

	uid, auth := ctx.Value(KeyAuthUser).(int64)
	args := []interface{}{uid}
	dest := []interface{}{&u.ID, &u.Email, &u.FolloweesCount, &u.FollowersCount}
	query := "SELECT  id, email, followers_count, followees_count "
	if auth {
		query += ", " +
			"followers.follower_id IS NOT NULL AS following, " +
			"followees.followee_id IS NOT NULL AS followeed "
		dest = append(dest, &u.Following, &u.Followeed)
	}

	query += "FROM user "
	if auth {
		query += "LEFT JOIN follows AS followers ON followers.follower_id = ? AND followers.followee_id = user.id " +
			"LEFT JOIN follows AS followees ON followees.follower_id = user.id AND followees.followee_id = ? "

		args = append(args, username, username)
	}

	query += "WHERE username = ? "
	err := s.db.QueryRowContext(ctx, query, args...).Scan(dest...)
	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("El query de perfil de usuario a fallado: %v", err)
	}

	u.Username = username
	u.Me = auth && uid == u.ID

	if !u.Me {
		u.ID = 0
		u.Email = ""
	}

	return u, nil
}

//ToggleFollow para seguirse entre dos usuarios
func (s *Service) ToggleFollow(ctx context.Context, username string) (ToggleFollowOutput, error) {
	var out ToggleFollowOutput

	followerID, ok := ctx.Value(KeyAuthUser).(int64)

	if !ok {
		return out, ErrUnauthenticated
	}

	username = strings.TrimSpace(username)
	if !rxUsername.MatchString(username) {
		return out, ErrInvalideUsername
	}

	//inicio de una transacción
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, fmt.Errorf("no se pudo iniciar la transaccion: %v", err)
	}

	defer tx.Rollback()
	//fin de la transacción

	var followeeID int64

	query := "SELECT id FROM user WHERE username=?"
	err = tx.QueryRowContext(ctx, query, username).Scan(&followeeID)
	if err == sql.ErrNoRows {
		return out, ErrUserNotFound
	}

	if err != nil {
		return out, fmt.Errorf("No se puede realizar la consulta-- select id from user where username: %v ", err)
	}

	if followeeID == followerID {
		return out, ErrForbiddenFollow
	}

	query = "SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id=? AND followee_id=?)"
	if err = tx.QueryRowContext(ctx, query, followerID, followeeID).Scan(&out.Following); err != nil {
		return out, fmt.Errorf("No se pudo realizar la consulta de seguidor: %v", err)
	}

	//Para cuando un usario esté siguiendo y quiera dejar de seguir

	if out.Following {
		query = "DELETE FROM follows WHERE follower_id=? AND followee_id=?"

		if _, err := tx.ExecContext(ctx, query, followerID, followeeID); err != nil {
			return out, fmt.Errorf("No se pudo borrar los seguidores: %v", err)
		}

		query = "UPDATE user SET followees_count = followees_count - 1 WHERE id=?"
		if _, err = tx.ExecContext(ctx, query, followerID); err != nil {
			return out, fmt.Errorf("no se pudo actualizar el contador de seguidos: %v", err)
		}

		query = "call subfollowers(?)"
		if err = tx.QueryRowContext(ctx, query, followeeID).Scan(&out.FollowersCount); err != nil {
			return out, fmt.Errorf("No se pudo actualizar el contador de seguidores: %v", err)
		}

	} else { //cuando un usario quiera seguir a otro usuario
		//inserta el usuario seguido
		query = "INSERT INTO follows(follower_id, followee_id) VALUES (?, ?)"
		if _, err = tx.ExecContext(ctx, query, followerID, followeeID); err != nil {
			return out, fmt.Errorf("No se pudo insertar usuarios seguidos: %v", err)
		}

		//actualiza el contador de seguidores
		query = "UPDATE user SET followees_count = followees_count + 1 WHERE id=?"
		if _, err = tx.ExecContext(ctx, query, followerID); err != nil {
			return out, fmt.Errorf("No se pudo actualizar el contador de seguidos: %v", err)
		}

		query = "call addfollowers(?)"
		if err = tx.QueryRowContext(ctx, query, followeeID).Scan(&out.FollowersCount); err != nil {
			return out, fmt.Errorf("No se pudo actualizar el contador de seguidoores: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return out, fmt.Errorf("No se realizo un commit al toogle de seguir: %v", err)
	}

	out.Following = !out.Following

	if out.Following {
		//TODO: notificacion de seguidores
	}
	return out, nil
}

func (s *Service) Users(ctx context.Context, search string, first int, after string) ([]UserProfile, error) {

	search = strings.TrimSpace(search)
	after = strings.TrimSpace(after)
	first = normalizePageSize(first)

	uid, auth := ctx.Value(KeyAuthUser).(int64)

	query, args, err := buildQuery(`
		SELECT id, email, username, followers_count, followees_count 
		{{if .auth}}
		,followers.follower_id IS NOT NULL AS following  
		,followees.followee_id IS NOT NULL AS followeed
		{{end}}
		FROM user 
		{{if .auth}}
		LEFT JOIN follows AS followers.follower_id = @uid AND followers.followee_id = user.id
		LEFT JOIN follows AS followees.follower_id = user.id AND followees.followee_id = @uid
		{{end}}
		{{if or .search .after}}WHERE{{end}} 
		{{if .search}} username LIKE '%' || @search || '%'{{end}}
		{{if and .search .after}}AND{{end}}
		{{if .after}}username > @after {{end}}
		ORDER BY username ASC
		LIMIT @first`, map[string]interface{}{
		"auth":   auth,
		"uid":    uid,
		"search": search,
		"first":  first,
		"after":  after,
	})

	if err != nil {
		return nil, fmt.Errorf("No se puede construir el query: %v", err)
	}

	log.Printf("users query: %s \nargs: %v\n", query, args)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("No se pudo completar el query seleccionar usuarios: %v", err)
	}

	defer rows.Close()
	uu := make([]UserProfile, 0, first)
	for rows.Next() {
		var u UserProfile
		dest := []interface{}{&u.ID, &u.Email, &u.Username, &u.FollowersCount, &u.FolloweesCount}
		if auth {
			dest = append(dest, &u.Following, &u.Followeed)
		}

		if err = rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("No se pudo escanear el query usuarios: %v", err)

		}
		u.Me = auth && uid == u.ID

		if !u.Me {
			u.ID = 0
			u.Email = ""

			uu = append(uu, u)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pueden iterar las filas: %v", err)
	}
	return uu, nil
}
