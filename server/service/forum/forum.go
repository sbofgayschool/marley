package forum

/*
CREATE TABLE forum
(
    id INT NOT NULL AUTO_INCREMENT,
    user INT NOT NULL,
    course INT NOT NULL,
    video INT NOT NULL,
    content TEXT NOT NULL,
    time INTEGER NOT NULL,
    FOREIGN KEY(user) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY(course) REFERENCES course(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY(video) REFERENCES video(id) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE TABLE reply
(
    id INT NOT NULL AUTO_INCREMENT,
    user INT NOT NULL,
    forum INT NOT NULL
    content TEXT NOT NULL,
    time INTEGER NOT NULL,
    FOREIGN KEY(user) REFERENCES user(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY(forum) REFERENCES forum(id) ON UPDATE CASCADE ON DELETE CASCADE,
);
*/
