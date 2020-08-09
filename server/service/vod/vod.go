package vod

import (
	"errors"
	"github.com/sbofgayschool/marley/server/infra/db"
	"github.com/sbofgayschool/marley/server/service/common"
	"log"
)

/*
CREATE TABLE video
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    course INTEGER NOT NULL,
    name VARCHAR(32) NOT NULL,
    timestamp BIG INT NOT NULL,
    pdf TEXT NOT NULL,
    FOREIGN KEY(course) REFERENCES course(id) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE TABLE media
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    video INTEGER NOT NULL,
    quality INTEGER NOT NULL,
    url TEXT NOT NULL,
    FOREIGN KEY(video) REFERENCES video(id) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE TABLE chat
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    video INTEGER NOT NULL,
    user INTEGER NOT NULL,
    msg_type TEXT NOT NULL,
    message TEXT NOT NULL,
    source TEXT NOT NULL,
    elapsed_time BIG INT NOT NULL,
    FOREIGN KEY(video) REFERENCES video(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY(user) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE TABLE operation
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    video INTEGER NOT NULL,
    opt TEXT NOT NULL,
    elapsed_time BIG INT NOT NULL,
    FOREIGN KEY(video) REFERENCES video(id) ON UPDATE CASCADE ON DELETE CASCADE
);
*/

type Video struct {
	Id        int
	Course    int
	Name      string
	Timestamp int64
	Pdf       string
}

type Media struct {
	Id      int
	Video   int
	Quality int
	Url     string
}

func AddVideo(course int, name string, timestamp int64, pdf string) (int64, error) {
	stmt, err := db.DB.Prepare("INSERT INTO video(course, name, timestamp, pdf) VALUES (?, ?, ?, ?)")
	if err != nil {
		return -1, errors.New("database error")
	}
	defer stmt.Close()
	if res, err := stmt.Exec(course, name, timestamp, pdf); err != nil {
		return -1, errors.New("database error")
	} else if id, err := res.LastInsertId(); err != nil {
		return -1, errors.New("database error")
	} else {
		return id, nil
	}
}

func GetVideo(id int, course int, timestamp int64) (*Video, error) {
	sql := "SELECT id, course, name, timestamp, pdf FROM video WHERE id=?"
	args := []interface{}{id}
	if id < 0 {
		sql = "SELECT id, course, name, timestamp, pdf FROM video WHERE course=? AND timestamp=?"
		args = []interface{}{course, timestamp}
	}
	stmt, err := db.DB.Prepare(sql)
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	res := Video{}
	if err := stmt.QueryRow(args...).Scan(&res.Id, &res.Course, &res.Name, &res.Timestamp, &res.Pdf); err != nil {
		log.Println(err)
		return nil, nil
	}
	return &res, nil
}

func SearchVideo(course int) ([]*Video, error) {
	stmt, err := db.DB.Prepare("SELECT id, course, name, timestamp, pdf FROM video WHERE (SELECT COUNT(*) FROM media WHERE media.video=video.id)>0 AND course=?")
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	rows, err := stmt.Query(course)
	if err != nil {
		return nil, errors.New("database error")
	}
	defer rows.Close()
	var res []*Video
	for rows.Next() {
		v := Video{}
		rows.Scan(&v.Id, &v.Course, &v.Name, &v.Timestamp, &v.Pdf)
		res = append(res, &v)
	}
	return res, nil
}

func SetVideo(id int, name string) error {
	stmt, err := db.DB.Prepare("UPDATE video SET name=? WHERE id=?")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(name, id); err != nil {
		return errors.New("database error")
	}
	return nil
}

func DeleteVideo(id int) error {
	stmt, err := db.DB.Prepare("DELETE FROM video WHERE id=?")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(id); err != nil {
		return errors.New("database error")
	}
	return nil
}

func AddMedia(video int64, quality int, url string) error {
	stmt, err := db.DB.Prepare("INSERT INTO media(video, quality, url) VALUES (?, ?, ?)")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(video, quality, url); err != nil {
		return errors.New("database error")
	}
	return nil
}

func SearchMedia(video int) ([]*Media, error) {
	stmt, err := db.DB.Prepare("SELECT id, video, quality, url FROM media WHERE video=?")
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	rows, err := stmt.Query(video)
	if err != nil {
		return nil, errors.New("database error")
	}
	defer rows.Close()
	var res []*Media
	for rows.Next() {
		m := Media{}
		rows.Scan(&m.Id, &m.Video, &m.Quality, &m.Url)
		res = append(res, &m)
	}
	return res, nil
}

func AddChat(video int64, c *common.Chat) error {
	stmt, err := db.DB.Prepare("INSERT INTO chat(video, user, msg_type, message, source, elapsed_time) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(video, c.Uid, c.MsgType, c.Message, c.Source, c.ElapsedTime); err != nil {
		return errors.New("database error")
	}
	return nil
}

func SearchChat(video int) ([]*common.Chat, error) {
	stmt, err := db.DB.Prepare("SELECT chat.user, user.username, chat.msg_type, chat.message, chat.source, chat.elapsed_time FROM chat JOIN user ON chat.user=user.id WHERE video=? ORDER BY chat.elapsed_time ASC")
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	rows, err := stmt.Query(video)
	if err != nil {
		return nil, errors.New("database error")
	}
	defer rows.Close()
	var res []*common.Chat
	for rows.Next() {
		c := common.Chat{}
		rows.Scan(&c.Uid, &c.Username, &c.MsgType, &c.Message, &c.Source, &c.ElapsedTime)
		res = append(res, &c)
	}
	return res, nil
}

func AddOperation(video int64, o *common.Operation) error {
	stmt, err := db.DB.Prepare("INSERT INTO operation(video, opt, elapsed_time) VALUES (?, ?, ?)")
	if err != nil {
		return errors.New("database error")
	}
	defer stmt.Close()
	if _, err := stmt.Exec(video, o.Opt, o.ElapsedTime); err != nil {
		return errors.New("database error")
	}
	return nil
}

func SearchOperation(video int) ([]*common.Operation, error) {
	stmt, err := db.DB.Prepare("SELECT opt, elapsed_time FROM operation WHERE video=? ORDER BY elapsed_time ASC")
	if err != nil {
		return nil, errors.New("database error")
	}
	defer stmt.Close()
	rows, err := stmt.Query(video)
	if err != nil {
		return nil, errors.New("database error")
	}
	defer rows.Close()
	var res []*common.Operation
	for rows.Next() {
		o := common.Operation{}
		rows.Scan(&o.Opt, &o.ElapsedTime)
		res = append(res, &o)
	}
	return res, nil
}
