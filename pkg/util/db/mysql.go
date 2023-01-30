package db

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
)

var (
	db   *sql.DB
	once sync.Once
	cfg  config.Config
)

func Init(config config.Config) {
	cfg = config
}

func getInstance() *sql.DB {
	if db == nil {
		once.Do(func() {
			c := cfg.Mysql
			source := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.User, c.Passwd, c.Host, c.Port, c.Database)
			var err error
			db, err = sql.Open("mysql", source)
			if err != nil {
				panic(err)
			}
			db.SetConnMaxLifetime(time.Minute * 3)
			db.SetMaxOpenConns(10)
			db.SetMaxIdleConns(10)
		})
	}
	return db
}

func Close() {
	if db != nil {
		_ = db.Close()
	}
}

func Register(username, email, passwd string) (int, error) {
	db := getInstance()
	stmt, err := db.Prepare("insert into user (username , email, passwd) VALUES (?,?,?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(username, email, passwd)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func Login(email, passwd string) (int, error) {
	db := getInstance()
	query := "select id from user where email = ? and passwd = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(email, passwd)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return 0, err
		}
		return id, nil
	}
	return 0, errors.New("does not exist")
}

func IsExistEmail(email string) (bool, error) {
	db := getInstance()
	stmt, err := db.Prepare("select id from user where email = ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(email)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

//func GetUserById(id string) (*orm.User, error) {
//	db := getInstance()
//	stmt, err := db.Prepare("select username,email,id_number,work_status,age from user where id = ?")
//	if err != nil {
//		return nil, err
//	}
//	defer stmt.Close()
//	rows, err := stmt.Query(id)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//	u := &orm.User{}
//	u.ID = id
//	if rows.Next() {
//		err := rows.Scan(&u.Username, &u.email, &u.IDNumber, &u.WorkStatus, &u.Age)
//		if err != nil {
//			return nil, err
//		}
//		return u, nil
//	}
//	return nil, nil
//}
