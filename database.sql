DROP database IF EXISTS network;
CREATE DATABASE IF NOT EXISTS network;
USE network;

CREATE TABLE IF NOT EXISTS user(
	id int auto_increment primary key,
    email varchar(50) not null unique, 
    username varchar(50) not null unique,
    password varchar(75) not null,
    followers_count int not null default 0 check(followers_count>=0),
    followees_count int not null default 0 check(followers_count>=0)
);

CREATE TABLE IF NOT exists follows(
	follower_id int not null,
    followee_id int not null,
    primary key(follower_id, followee_id)
);


alter table user
add check (followees_count >= 0)
add check (followers_count >= 0);

delimiter $
CREATE PROCEDURE `addfollowers` (
in _id int
)
BEGIN
DECLARE EXIT HANDLER FOR SQLEXCEPTION
 BEGIN
  SHOW ERRORS LIMIT 1;
 ROLLBACK;
 END;
 DECLARE EXIT HANDLER FOR SQLWARNING
 BEGIN
 SHOW WARNINGS LIMIT 1;
 ROLLBACK;
 END;
START TRANSACTION;
UPDATE user SET followers_count = followers_count + 1 WHERE id = _id;
SELECT  followers_count FROM user where id = _id;
COMMIT;
END $


delimiter $
CREATE PROCEDURE `subfollowers` (
in _id int
)
BEGIN
DECLARE EXIT HANDLER FOR SQLEXCEPTION
 BEGIN
  SHOW ERRORS LIMIT 1;
 ROLLBACK;
 END;
 DECLARE EXIT HANDLER FOR SQLWARNING
 BEGIN
 SHOW WARNINGS LIMIT 1;
 ROLLBACK;
 END;
START TRANSACTION;
UPDATE user SET followers_count = followers_count - 1 WHERE id = _id;
SELECT  followers_count FROM user where id = _id;
COMMIT;
END $