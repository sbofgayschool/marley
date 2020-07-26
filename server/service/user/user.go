package user

import (
	"errors"
	"github.com/sbofgayschool/marley/server/infra/db"
)

type SockUser struct {
	Uid      int
	Username string
	Teacher  bool
}

/*
CREATE TABLE user
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(40) NOT NULL UNIQUE,
    password VARCHAR(40) NOT NULL,
    teacher INTEGER NOT NULL DEFAULT 0,
    note TEXT NOT NULL
);
*/

type User struct {
	Id int
	Username string
	Teacher int
	Note string
}

func AddUser(username string, password string, teacher int, note string) error {
	stmt, err := db.DB.Prepare("INSERT INTO user(username, password, teacher, note) VALUES (?, ?, ?, ?)")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(username, password, teacher, note); err != nil {
		return errors.New("database error")
	}
	return nil
}

func GetUser(id int) (*User, string, error) {
	stmt, err := db.DB.Prepare("SELECT (id, username, password, teacher, note) FROM user WHERE id=?")
	if err != nil {
		return nil, "", errors.New("database error")
	}
	defer stmt.Close()
	res := User{}
	password := ""
	if err := stmt.QueryRow(id).Scan(&res.Id, &res.Username, &password, &res.Teacher, &res.Note); err != nil {
		return nil, "", nil
	}
	return &res, password, nil
}

func SearchUser(username string, teacher int) ([]*User, error) {
	condition := " WHERE 1=1"
	var args []interface{}
	if username != "" {
		condition += " AND username=?"
		args = append(args, username)
	}
	if teacher != -1 {
		condition += " AND teacher=?"
		args = append(args, teacher)
	}
	stmt, err := db.DB.Prepare("SELECT (id, username, teacher, note) FROM user" + condition)
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, errors.New("database error")
	}
	defer rows.Close()
	var res []*User
	for rows.Next() {
		u := User{}
		rows.Scan(&u.Id, &u.Username, &u.Teacher, &u.Note)
		res = append(res, &u)
	}
	return res, nil
}

func SetPassword(id int, password string) error {
	stmt, err := db.DB.Prepare("UPDATE user SET password=? WHERE id=?")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(id, password); err != nil {
		return errors.New("database error")
	}
	return nil
}

func SetNote(id int, note string) error {
	stmt, err := db.DB.Prepare("UPDATE user SET note=? WHERE id=?")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(id, note); err != nil {
		return errors.New("database error")
	}
	return nil
}
