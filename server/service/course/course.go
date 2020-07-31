package course

import (
	"errors"
	"github.com/sbofgayschool/marley/server/infra/db"
	"github.com/sbofgayschool/marley/server/utils"
	"log"
)

/*
CREATE TABLE course
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50) NOT NULL,
    owner INTEGER NOT NULL,
    tag TEXT NOT NULL,
    note TEXT NOT NULL,
    FOREIGN KEY(owner) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE,
    UNIQUE(name, owner)
);
CREATE TABLE relation
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    course INTEGER NOT NULL,
    user INTEGER NOT NULL,
    relation INTEGER NOT NULL,
    FOREIGN KEY(user) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY(course) REFERENCES course(id) ON UPDATE CASCADE ON DELETE CASCADE,
    UNIQUE(user, course)
);
CREATE TABLE comment
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user INTEGER NOT NULL,
    course INTEGER NOT NULL,
    rate INTEGER NOT NULL,
    comment TEXT NOT NULL,
    time INTEGER NOT NULL,
    FOREIGN KEY(user) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY(course) REFERENCES course(id) ON UPDATE CASCADE ON DELETE CASCADE,
    UNIQUE(user, course)
);
*/

type Course struct {
	Id            int     `json:"Id"`
	Name          string  `json:"Name"`
	Owner         int     `json:"Owner"`
	OwnerUsername string  `json:"OwnerUsername"`
	Tag           string  `json:"Tag"`
	Note          string  `json:"Note"`
	Rate          float64 `json:"Rate"`
	Relation      int     `json:"Relation"`
}

type Relation struct {
	Course   int    `json:"Course"`
	User     int    `json:"User"`
	Username string `json:"Username"`
	Relation int    `json:"Relation"`
}

type Comment struct {
	Course   int    `json:"Course"`
	User     int    `json:"User"`
	Username string `json:"Username"`
	Rate     int    `json:"Rate"`
	Comment  string `json:"Comment"`
	Time     int64  `json:"Time"`
}

func AddCourse(name string, owner int, tag string, note string) (int64, error) {
	stmt, err := db.DB.Prepare("INSERT INTO course(name, owner, tag, note) VALUES (?, ?, ?, ?)")
	if err != nil {
		return -1, errors.New("database error")
	}
	defer stmt.Close()
	if res, err := stmt.Exec(name, owner, tag, note); err != nil {
		return -1, errors.New("database error")
	} else if id, err := res.LastInsertId(); err != nil {
		return -1, errors.New("database error")
	} else {
		return id, nil
	}
}

func GetCourse(id int, user int) (*Course, error) {
	stmt, err := db.DB.Prepare("SELECT course.id, course.name, course.owner, user.username, course.tag, course.note, IFNULL((SELECT AVG(rate) FROM comment WHERE course=course.id), 0), IFNULL(relation.relation, 0) FROM (course JOIN user ON course.owner=user.id) LEFT JOIN (SELECT course, relation FROM relation AS relation WHERE user=?) AS relation ON course.id=relation.course WHERE course.id=?")
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	res := Course{}
	if err := stmt.QueryRow(user, id).Scan(&res.Id, &res.Name, &res.Owner, &res.OwnerUsername, &res.Tag, &res.Note, &res.Rate, &res.Relation); err != nil {
		log.Println(err)
		return nil, nil
	}
	return &res, nil
}

func SearchCourse(user int) ([]*Course, error) {
	stmt, err := db.DB.Prepare("SELECT course.id, course.name, course.owner, user.username, course.tag, course.note, IFNULL((SELECT AVG(rate) FROM comment WHERE course=course.id), 0), IFNULL(relation.relation, 0) FROM (course JOIN user ON course.owner=user.id) LEFT JOIN (SELECT course, relation FROM relation AS relation WHERE user=?) AS relation ON course.id=relation.course")
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	rows, err := stmt.Query(user)
	if err != nil {
		return nil, errors.New("database error")
	}
	defer rows.Close()
	var res []*Course
	for rows.Next() {
		c := Course{}
		_ = rows.Scan(&c.Id, &c.Name, &c.Owner, &c.OwnerUsername, &c.Tag, &c.Note, &c.Rate, &c.Relation)
		res = append(res, &c)
	}
	return res, nil
}

func SetCourse(id int, tag string, note string) error {
	stmt, err := db.DB.Prepare("UPDATE user SET tag=?, note=? WHERE id=?")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(tag, note, id); err != nil {
		return errors.New("database error")
	}
	return nil
}

func SearchRelation(course int) ([]*Relation, error) {
	stmt, err := db.DB.Prepare("SELECT relation.course, relation.user, user.username, relation.relation FROM relation JOIN user ON relation.user=user.id")
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	rows, err := stmt.Query(course)
	if err != nil {
		return nil, errors.New("database error")
	}
	defer rows.Close()
	var res []*Relation
	for rows.Next() {
		r := Relation{}
		rows.Scan(&r.Course, &r.User, &r.Username, &r.Relation)
		res = append(res, &r)
	}
	return res, nil
}

func SetRelation(course int, user int, relation int) error {
	sql := "INSERT INTO relation(course, user, relation) VALUES (?, ?, ?)"
	var args = []interface{}{course, user, relation}
	if relation == 0 {
		sql = "DELETE FROM relation WHERE course=? AND user=?"
		args = []interface{}{course, user}
	}
	stmt, err := db.DB.Prepare(sql)
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(args...); err != nil {
		return errors.New("database error")
	}
	return nil
}

func AddComment(course int, user int, rate int, comment string) error {
    stmt, err := db.DB.Prepare("INSERT INTO comment(course, user, rate, comment, time) VALUES (?, ?, ?, ?, ?)")
    if err != nil {
        return errors.New("database error")
    }
    defer stmt.Close()
    if _, err := stmt.Exec(course, user, rate, comment, utils.UnixMillion()); err != nil {
        return errors.New("database error")
    }
    return nil
}

func SearchComment(course int) ([]*Comment, error){
    stmt, err := db.DB.Prepare("SELECT comment.course, comment.user, user.username, comment.rate, comment.comment, comment.time FROM comment JOIN user ON comment.user=user.id WHERE comment.course=?")
    if err != nil {
        return nil, errors.New("database error")
    }
    defer stmt.Close()
    rows, err := stmt.Query(course)
    if err != nil {
        return nil, errors.New("database error")
    }
    defer rows.Close()
    var res []*Comment
    for rows.Next() {
        c := Comment{}
        rows.Scan(&c.Course, &c.User, &c.Username, &c.Rate, &c.Comment, &c.Time)
        res = append(res, &c)
    }
    return res, nil
}

func DeleteComment(course int, user int) error {
    stmt, err := db.DB.Prepare("DELETE FROM comment WHERE course=? AND user=?")
    if err != nil {
        return errors.New("database error")
    }
    defer stmt.Close()
    if _, err := stmt.Exec(course, user); err != nil {
        return errors.New("database error")
    }
    return nil
}