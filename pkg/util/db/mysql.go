package db

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vincent-vinf/code-validator/pkg/orm"
	"github.com/vincent-vinf/code-validator/pkg/vo"

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

func ListBatchWithUserName() ([]vo.Batch, error) {
	db := getInstance()
	rows, err := db.Query("SELECT b.id,u.id,u.username,b.name,b.runtime,b.info,b.create_at FROM batch b LEFT JOIN user u ON b.user_id = u.id")
	if err != nil {
		return nil, err
	}
	var res []vo.Batch
	defer rows.Close()
	for rows.Next() {
		v := vo.Batch{}
		if err = rows.Scan(&v.ID, &v.UserID, &v.Username, &v.Name, &v.Runtime, &v.Describe, &v.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, v)
	}

	return res, nil
}

// GetBatchByIDWithVerifications with verifications
func GetBatchByIDWithVerifications(id int) (*orm.Batch, error) {
	v, err := GetBatchByID(id)
	if err != nil {
		return nil, err
	}
	db := getInstance()
	rows, err := db.Query("select id,name,runtime,data from verification where batch_id = ?", v.ID)
	if err != nil {
		return nil, err
	}
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
	rows, err := db.Query("select user_id,name,runtime,info,create_at from batch where id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	v := &orm.Batch{
		ID: id,
	}
	if rows.Next() {
		if err = rows.Scan(&v.UserID, &v.Name, &v.Runtime, &v.Describe, &v.CreatedAt); err != nil {
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
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	r, err := tx.Exec("insert into batch(user_id, name, runtime, info, create_at) values (?,?,?,?,?)", batch.UserID, batch.Name, batch.Runtime, batch.Describe, batch.CreatedAt)
	if err != nil {
		return
	}
	id, err := r.LastInsertId()
	if err != nil {
		return
	}
	batch.ID = int(id)
	stmt, err := tx.Prepare("insert into verification(batch_id, name, runtime, data) values (?,?,?,?)")
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

func GetTaskByID(id int) (*orm.Task, error) {
	db := getInstance()
	rows, err := db.Query("select user_id,batch_id,create_at from task where id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	task := &orm.Task{
		ID: id,
	}
	if rows.Next() {
		if err = rows.Scan(&task.UserID, &task.BatchID, &task.CreatedAt); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("the task with id %d does not exist", id)
	}

	return task, nil
}

// GetTaskInfoByID with subtask
func GetTaskInfoByID(id int) (*vo.Task, error) {
	db := getInstance()
	rows, err := db.Query("SELECT u.id,u.username,t.batch_id,b.`name`,b.runtime,t.create_at FROM task t LEFT JOIN user u ON t.user_id = u.id LEFT JOIN batch b ON t.batch_id = b.id where t.id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	task := &vo.Task{}
	task.ID = id
	if rows.Next() {
		if err = rows.Scan(&task.UserID, &task.Username, &task.BatchID, &task.BatchName, &task.Runtime, &task.CreatedAt); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("the task with id %d does not exist", id)
	}
	rows, err = db.Query("select subtask.id,verification_id,v.`name`,status,result,message from subtask LEFT JOIN verification v ON v.id = verification_id where task_id = ?", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		s := &vo.SubTask{}
		s.TaskID = task.ID
		if err = rows.Scan(&s.ID, &s.VerificationID, &s.VerificationName, &s.Status, &s.Result, &s.Message); err != nil {
			return nil, err
		}
		task.SubTasks = append(task.SubTasks, s)
	}

	return task, nil
}

func ListTasks(batchID, userID int) ([]vo.Task, error) {
	db := getInstance()
	var (
		rows *sql.Rows
		err  error
	)
	if batchID != 0 {
		rows, err = db.Query("SELECT t.id,u.id,u.username,t.batch_id,b.`name`,b.runtime,t.create_at FROM task t LEFT JOIN user u ON t.user_id = u.id LEFT JOIN batch b ON t.batch_id = b.id where t.batch_id = ?", batchID)
	} else {
		rows, err = db.Query("SELECT t.id,u.id,u.username,t.batch_id,b.`name`,b.runtime,t.create_at FROM task t LEFT JOIN user u ON t.user_id = u.id LEFT JOIN batch b ON t.batch_id = b.id where t.user_id = ?", userID)
	}
	if err != nil {
		return nil, err
	}
	var res []vo.Task
	defer rows.Close()
	for rows.Next() {
		v := vo.Task{}
		if err = rows.Scan(&v.ID, &v.UserID, &v.Username, &v.BatchID, &v.BatchName, &v.Runtime, &v.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, v)
	}

	return res, nil
}

func AddTask(task *orm.Task) (err error) {
	db := getInstance()
	stmt, err := db.Prepare("insert into task (user_id , batch_id, create_at) VALUES (?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(task.UserID, task.BatchID, task.CreatedAt)
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

func UpdateSubTask(subtask *orm.SubTask) (err error) {
	db := getInstance()
	stmt, err := db.Prepare("UPDATE subtask SET status = ?, result = ?, message = ?  WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(subtask.Status, subtask.Result, subtask.Message, subtask.ID)
	if err != nil {
		return err
	}

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
