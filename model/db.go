package model

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type AppDb interface {
	DbStore() *sqlx.DB
	Close() error
	NewUser(string, string)
	Authenticate(string, string)
}

type appDb struct {
	db *sqlx.DB
}

func (db *appDb) DbStore() *sqlx.DB {
	return db.db
}

func (db *appDb) Close() error {
	return db.db.Close()
}

// TODO: Add better error messages. See Authenticate()
func (db *appDb) NewUser(name string, password string) (user User, err error) {
	stmt, err := db.db.Preparex(
		`INSERT INTO user(name, password_hash) ` +
			`VALUES (?, ?);`,
	)
	if err != nil {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 5)
	if err != nil {
		return
	}
	_, err = stmt.Exec(name, string(hash))
	if err != nil {
		return
	}
	user, err = db.findUser(name)
	return
}

func (db *appDb) Authenticate(name string, password string) (User, error) {
	user, err := db.findUser(name)
	if err != nil {
		return user, UserAuthenticationError{"user not found"}
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		err = UserAuthenticationError{"wrong password"}
		user = User{}
	}
	return user, err
}

func (db *appDb) findUser(name string) (User, error) {
	user := User{}
	err := db.db.Get(&user, "SELECT * FROM user WHERE name=?", name)
	return user, err
}

type User struct {
	Name         string `form:"name" json:"name"`
	PasswordHash string `db:"password_hash"`
}

type UserAuthenticationError struct {
	msg string
}

func (e UserAuthenticationError) Error() string {
	return e.msg
}

func CreateDb(conn string) (*appDb, error) {
	db, err := sqlx.Open("sqlite3", conn)
	a := new(appDb)
	a.db = db

	if err != nil {
		return a, err
	}

	err = a.createUserTable()

	return a, err
}

func (db *appDb) createUserTable() (err error) {
	_, err = db.db.Exec(
		`CREATE TABLE user (` +
			`name            string NOT NULL, ` +
			`password_hash   string NOT NULL); ` +
			`CREATE UNIQUE INDEX user_idx ON user (name);`,
	)
	return err
}
