package ddd

import (
	"database/sql"
	_ "errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

var (
	Hostname = ""
	Port     = 5432
	Username = ""
	Password = ""
	Database = ""
)

type UserData struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

func openConnection() (*sql.DB, error) {
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		Hostname, Port, Username, Password, Database)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CleanupDbTables() error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()
	_, err = db.Exec("TRUNCATE userdata, users;")
	return err
}

func GetUserByUsername(username string) (*UserData, error) {
	const GET_USER = "SELECT id, username from users where username = $1"
	const GET_USER_DATA = "SELECT name, surname, description from userdata where userid = $1"
	username = strings.ToLower(username)
	db, err := openConnection()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	userData := &UserData{}
	err = db.QueryRow(GET_USER, username).Scan(&(userData.ID), &(userData.Username))
	if err != nil {
		return nil, err
	}
	err = db.QueryRow(GET_USER_DATA, userData.ID).Scan(&(userData.Name), &(userData.Surname), &(userData.Description))
	if err != nil {
		return nil, err
	}
	return userData, nil
}

func AddUser(d *UserData) error {
	const INSERT_USER_QUERY = "INSERT INTO users(username) VALUES ($1) RETURNING id"
	const INSERT_USER_DATA_QUERY = `INSERT INTO userdata(userid, name, surname, description) VALUES ($1, $2, $3, $4)`

	db, err := openConnection()
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()
	d.Username = strings.ToLower(d.Username)

	err = db.QueryRow(INSERT_USER_QUERY, d.Username).Scan(&d.ID)
	if err != nil {
		return err
	}

	db.QueryRow(INSERT_USER_DATA_QUERY, d.ID, d.Name, d.Surname, d.Description)
	return nil
}
