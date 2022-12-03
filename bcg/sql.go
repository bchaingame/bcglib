package bcg

import (
	"database/sql"
	"os"
)

var db *sql.DB

func OpenSqlite(path, dbname string) (*sql.DB, error) {
	if path == "" {
		path = "./data"
	}
	err := os.MkdirAll(path, 0777)
	if CheckError(err) {
		return nil, err
	}
	return sql.Open("sqlite3", path+"/"+dbname)
}

func OpenMysql(user, pass, host, dbname string) {
	var err error
	db, err = sql.Open("mysql", user+":"+pass+"@tcp("+host+")/"+dbname+"?charset=utf8")
	CheckError(err)
}

func GetDb() *sql.DB {
	return db
}

type QueryCall func(rows *sql.Rows)
type QueryFunc func(string, QueryCall, ...interface{}) error

// Query 查询需要返回数据的语句，数据从回调函数的 rows 里获取，无需执行 rows 的 Close 函数，
// 这个设计的目的是减少遗忘 Close 的可能，因为遗忘 Close 不会对程序有立即的影响，直到 Mysql
// 资源被耗尽，对于海量的查询语句来说，定位哪里忘记 Close 是非常困难的。
func Query(sqlCase string, qc QueryCall, v ...interface{}) error {
	rows, err := db.Query(sqlCase, v...)
	if CheckErrTrace(err, 2) {
		return err
	}
	qc(rows)
	_ = rows.Close()
	return nil
}
func Exec(sqlCase string, v ...interface{}) (sql.Result, error) {
	return db.Exec(sqlCase, v...)
}
