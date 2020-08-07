package vod

import (
    "errors"
    "github.com/sbofgayschool/marley/server/infra/db"
    "log"
)

/*
CREATE TABLE video
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    course INTEGER NOT NULL,
    name VARCHAR(32) NOT NULL,
    timestamp BIG INT NOT NULL,
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
    Id int
    Course int
    Name string
    Timestamp int64
}

type Media struct {
    Id int
    Video int
    Quality int
    Url int
}

func AddVideo(course int, name string, timestamp int64) error {
    stmt, err := db.DB.Prepare("INSERT INTO video(course, name, timestamp) VALUES (?, ?, ?)")
    if err != nil {
        return errors.New("database error")
    }
    defer stmt.Close()
    if _, err := stmt.Exec(course, name, timestamp); err != nil {
        return errors.New("database error")
    }
    return nil
}

func GetVideo(id int, course int, timestamp int64) (*Video, error) {
    sql := "SELECT id, course, name, timestamp FROM video WHERE id=?"
    args := []interface{}{id}
    if id < 0 {
        sql = "SELECT id, course, name, timestamp FROM video WHERE course=? AND timestamp=?"
        args = []interface{}{course, timestamp}
    }
    stmt, err := db.DB.Prepare(sql)
    if err != nil {
        return nil, errors.New("database error")
    }
    defer stmt.Close()
    res := Video{}
    if err := stmt.QueryRow(args...).Scan(&res.Id, &res.Course, &res.Name, &res.Timestamp); err != nil {
        log.Println(err)
        return nil, nil
    }
    return &res, nil
}

func SearchVideo(course int) ([]*Video, error){
    stmt, err := db.DB.Prepare("SELECT id, course, name, timestamp FROM video WHERE course=?")
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
        rows.Scan(&v.Id, &v.Course, &v.Name, &v.Timestamp)
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

func AddMedia() {
}

func SearchMedia() {
}

func AddChat() {
}

func SearchChat() {
}

func AddOpt() {
}

func SearchOpt() {
}