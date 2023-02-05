package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/vincent-vinf/code-validator/pkg/orm"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
)

var (
	db   *sql.DB
	once sync.Once
	cfg  config.Mysql
)

func Init(config config.Mysql) {
	cfg = config
}

func getInstance() *sql.DB {
	if db == nil {
		once.Do(func() {
			source := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.User, cfg.Passwd, cfg.Host, cfg.Port, cfg.Database)
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

func GetVerificationByID(id int) (*orm.Verification, error) {
	db := getInstance()
	rows, err := db.Query("select batch_id,name,data from verification where id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		v := &orm.Verification{
			ID: id,
		}
		if err = rows.Scan(&v.BatchID, &v.Name, &v.Data); err != nil {
			return nil, err
		}

		return v, nil
	}

	return nil, fmt.Errorf("the verification with id %d does not exist", id)
}

func GetTaskByID(id int) (*orm.Task, error) {
	db := getInstance()
	rows, err := db.Query("select user_id,batch_id,code,create_at from task where id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		v := &orm.Task{
			ID: id,
		}
		if err = rows.Scan(&v.UserID, &v.BatchID, &v.Code, &v.CreateAt); err != nil {
			return nil, err
		}

		return v, nil
	}

	return nil, fmt.Errorf("the verification with id %d does not exist", id)
}

func AddBatch(batch *orm.Batch) error {
	db := getInstance()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	r, err := tx.Exec("insert into batch(user_id, name, create_at) values (?,?,?)", batch.UserID, batch.Name, batch.CreateAt)
	if err != nil {
		tx.Rollback()
		return err
	}
	id, err := r.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}
	stmt, err := tx.Prepare("insert into verification(batch_id, name, data) values (?,?,?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, v := range batch.Verifications {
		_, err := stmt.Exec(id, v.Name, v.Data)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
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
