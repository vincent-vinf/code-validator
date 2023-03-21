package db

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vincent-vinf/code-validator/pkg/orm"

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

// GetBatchByIDWithVerifications with verifications
func GetBatchByIDWithVerifications(id int) (*orm.Batch, error) {
	v, err := GetBatchByID(id)
	if err != nil {
		return nil, err
	}
	db := getInstance()
	rows, err := db.Query("select id,name,runtime,data from verification where batch_id = ?", v.ID)
	for rows.Next() {
		vf := &orm.Verification{
			BatchID: v.ID,
		}
		if err = rows.Scan(&vf.ID, &vf.Name, &vf.Runtime, &vf.Data); err != nil {
			return nil, err
		}
		v.Verifications = append(v.Verifications, vf)
	}

	return v, nil
}

func GetBatchByID(id int) (*orm.Batch, error) {
	db := getInstance()
	rows, err := db.Query("select user_id,name,create_at from batch where id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	v := &orm.Batch{
		ID: id,
	}
	if rows.Next() {
		if err = rows.Scan(&v.UserID, &v.Name, &v.CreatedAt); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("the batch with id %d does not exist", id)
	}

	return v, nil
}

// AddBatch with verifications
func AddBatch(batch *orm.Batch) (err error) {
	db := getInstance()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	r, err := tx.Exec("insert into batch(user_id, name, create_at) values (?,?,?)", batch.UserID, batch.Name, batch.CreatedAt)
	if err != nil {
		return
	}
	id, err := r.LastInsertId()
	if err != nil {
		return
	}
	batch.ID = int(id)
	stmt, err := tx.Prepare("insert into verification(batch_id, name, runtime, data) values (?,?,?)")
	if err != nil {
		return
	}
	for _, v := range batch.Verifications {
		_, err = stmt.Exec(id, v.Name, v.Runtime, v.Data)
		if err != nil {
			return
		}
	}

	return
}

func GetVerificationByID(id int) (*orm.Verification, error) {
	db := getInstance()
	rows, err := db.Query("select batch_id,name,runtime,data from verification where id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		v := &orm.Verification{
			ID: id,
		}
		if err = rows.Scan(&v.BatchID, &v.Name, &v.Runtime, &v.Data); err != nil {
			return nil, err
		}

		return v, nil
	}

	return nil, fmt.Errorf("the verification with id %d does not exist", id)
}

// GetTaskByID with subtask
func GetTaskByID(id int) (*orm.Task, error) {
	db := getInstance()
	rows, err := db.Query("select user_id,batch_id,code,status,create_at from task where id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	task := &orm.Task{
		ID: id,
	}
	if rows.Next() {
		if err = rows.Scan(&task.UserID, &task.BatchID, &task.Code, &task.Status, &task.CreatedAt); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("the task with id %d does not exist", id)
	}
	rows, err = db.Query("select id,verification_id,status,result,message from subtask where task_id = ?", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		s := &orm.SubTask{
			TaskID: task.ID,
		}
		if err = rows.Scan(&s.ID, &s.VerificationID, &s.Status, &s.Result, &s.Message); err != nil {
			return nil, err
		}
		task.SubTasks = append(task.SubTasks, s)
	}

	return task, nil
}

func AddTask(task *orm.Task) (err error) {
	db := getInstance()
	stmt, err := db.Prepare("insert into task (user_id , batch_id, code, status, create_at) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(task.UserID, task.BatchID, task.Code, task.Status, task.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	task.ID = int(id)

	return
}

func AddSubTask(subtask *orm.SubTask) (err error) {
	db := getInstance()
	stmt, err := db.Prepare("insert into subtask (task_id , verification_id, status, result, message) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(subtask.TaskID, subtask.VerificationID, subtask.Status, subtask.Result, subtask.Message)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	subtask.ID = int(id)

	return
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
