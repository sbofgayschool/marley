package forum

import (
	"errors"
	"github.com/sbofgayschool/marley/server/infra/db"
	"github.com/sbofgayschool/marley/server/utils"
	"log"
)

/*
CREATE TABLE forum
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user INTEGER NOT NULL,
    course INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    time INTEGER NOT NULL,
    FOREIGN KEY(user) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY(course) REFERENCES course(id) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE TABLE reply
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user INTEGER NOT NULL,
    forum INTEGER NOT NULL,
    content TEXT NOT NULL,
    time INTEGER NOT NULL,
    FOREIGN KEY(user) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY(forum) REFERENCES forum(id) ON UPDATE CASCADE ON DELETE CASCADE
);
*/

type Forum struct {
	Id       int    `json:"Id"`
	User     int    `json:"User"`
	Username string `json:"Username"`
	Course   int    `json:"Course"`
	Title    string `json:"Title"`
	Content  string `json:"Content"`
	Time     int    `json:"Time"`
}

type Reply struct {
	Id       int    `json:"Id"`
	User     int    `json:"User"`
	Username string `json:"Username"`
	Forum    int    `json:"Forum"`
	Content  string `json:"Content"`
	Time     int    `json:"Time"`
}

func AddForum(user int, course int, title string, content string) (int64, error) {
	stmt, err := db.DB.Prepare("INSERT INTO forum(user, course, title, content, time) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return -1, errors.New("database error")
	}
	defer stmt.Close()
	if res, err := stmt.Exec(user, course, title, content, utils.UnixMillion()); err != nil {
		return -1, errors.New("database error")
	} else if id, err := res.LastInsertId(); err != nil {
		return -1, errors.New("database error")
	} else {
		return id, nil
	}
}

func GetForum(id int) (*Forum, error) {
	stmt, err := db.DB.Prepare("SELECT forum.id, forum.user, user.username, forum.course, forum.title, forum.content, forum.time FROM forum JOIN user ON forum.user=user.id WHERE forum.id=?")
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	res := Forum{}
	if err := stmt.QueryRow(id).Scan(&res.Id, &res.User, &res.Username, &res.Course, &res.Title, &res.Content, &res.Time); err != nil {
		log.Println(err)
		return nil, nil
	}
	return &res, nil
}

func SearchForum(course int) ([]*Forum, error) {
	stmt, err := db.DB.Prepare("SELECT forum.id, forum.user, user.username, forum.course, forum.title, forum.content, forum.time FROM forum JOIN user ON forum.user=user.id WHERE forum.course=?")
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	rows, err := stmt.Query(course)
	if err != nil {
		return nil, errors.New("database error")
	}
	defer rows.Close()
	var res []*Forum
	for rows.Next() {
		f := Forum{}
		_ = rows.Scan(&f.Id, &f.User, &f.Username, &f.Course, &f.Title, &f.Content, &f.Time)
		res = append(res, &f)
	}
	return res, nil
}

func DeleteForum(id int) error {
	stmt, err := db.DB.Prepare("DELETE FROM forum WHERE id=?")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(id); err != nil {
		return errors.New("database error")
	}
	return nil
}

func AddReply(user int, forum int, content string) error {
	stmt, err := db.DB.Prepare("INSERT INTO reply(user, forum, content, time) VALUES (?, ?, ?, ?)")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(user, forum, content, utils.UnixMillion()); err != nil {
		return errors.New("database error")
	}
	return nil
}

func GetReply(id int) (*Reply, error) {
	stmt, err := db.DB.Prepare("SELECT reply.id, reply.user, user.username, reply.forum, reply.content, reply.time FROM reply JOIN user ON reply.user=user.id WHERE reply.id=?")
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	res := Reply{}
	if err := stmt.QueryRow(id).Scan(&res.Id, &res.User, &res.Username, &res.Forum, &res.Content, &res.Time); err != nil {
		log.Println(err)
		return nil, nil
	}
	return &res, nil
}

func SearchReply(forum int) ([]*Reply, error) {
	stmt, err := db.DB.Prepare("SELECT reply.id, reply.user, user.username, reply.forum, reply.content, reply.time FROM reply JOIN user ON reply.user=user.id WHERE reply.forum=?")
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	rows, err := stmt.Query(forum)
	if err != nil {
		return nil, errors.New("database error")
	}
	defer rows.Close()
	var res []*Reply
	for rows.Next() {
		r := Reply{}
		_ = rows.Scan(&r.Id, &r.User, &r.Username, &r.Forum, &r.Content, &r.Time)
		res = append(res, &r)
	}
	return res, nil
}

func DeleteReply(id int) error {
	stmt, err := db.DB.Prepare("DELETE FROM reply WHERE id=?")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(id); err != nil {
		return errors.New("database error")
	}
	return nil
}
