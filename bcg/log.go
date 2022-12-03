// Package bcg
// log 相关函数，用于打印调试信息，信息自动包含时间和在源文件的调用位置，在 IDEA 的输出窗口，
// 可以通过点击输出窗口迅速定位到编辑器的代码行。
// 可以设置日志保存到数据库，只需要在 SetLogParam 函数传入一个数据库实例。
// To 版本的函数支持保存到不同的表，输出特性和没有 To 的函数相同
// LogF 版本的函数可以格式化日志信息 log 字段
package bcg

import (
	"database/sql"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// 指明数据库的类型，目前支持Mysq和SQLite两种，其它数据库可能运行不正常，未测试
const (
	DbTypeMysql  = 0
	DbTypeSqlite = 1
)

//这些颜色在某些 console 窗口可能并不起作用
const (
	TextBlack = iota + 30
	TextRed
	TextGreen
	TextYellow
	TextBlue
	TextMagenta
	TextCyan
	TextWhite
)

// LogParam
// LogDb: A valid database value, it can't be nil, other feilds can be default
// DbType: Mysql:0 or SQLite:1, other database maybe not work, default is Mysql
// LogTable: table name of log, default is jsuse_log
// SaveToLog: log whether save to database, default is false
// ShowOnConsole: log whether show on console, default is true
type LogParam struct {
	DbType        int
	LogDb         *sql.DB
	LogTable      string
	SaveToLog     bool
	ShowOnConsole bool
	MaxLogCount   int64
}

var logParam = LogParam{
	LogDb:         nil,
	DbType:        DbTypeMysql,
	LogTable:      "jsuse_log",
	SaveToLog:     false,
	ShowOnConsole: true,
	MaxLogCount:   1000,
}

// SetLogParam Set log parameter
func SetLogParam(param LogParam) {
	if param.LogDb == nil {
		err := "LogDb can not be nil"
		outputLogTrace(TextRed, 1, err)
		return
	}
	logParam = param
	if logParam.LogTable == "" {
		logParam.LogTable = "jsuse_log"
	}
	createLogTable(logParam.LogTable)
}
func createLogTable(table string) bool {
	var createCase string
	if logParam.DbType == DbTypeSqlite {
		createCase = `CREATE TABLE IF NOT EXISTS ` + table + `(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		log TEXT NOT NULL,
		trace VARCHAR(255) NOT NULL,
		color int,
		created_at TIMESTAMP DEFAULT (DATETIME('now', 'localtime'))
	);`
	} else if logParam.DbType == DbTypeMysql {
		createCase = `CREATE TABLE IF NOT EXISTS ` + table + `(
		id INTEGER PRIMARY KEY AUTO_INCREMENT,
		log TEXT NOT NULL,
		trace VARCHAR(255) NOT NULL,
		color int,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	}
	_, err := logParam.LogDb.Exec(createCase)
	checkLogError(err)
	return err == nil
}

type LogInfo struct {
	Id        int    `json:"id"`
	Color     int    `json:"color"`
	Log       string `json:"log"`
	Trace     string `json:"trace"`
	CreatedAt string `json:"created_at"`
}

//DeleteLog 删除默认的日志记录，idStart 到 idStop 的log都会被删除，包含这两个 id，要删除一个id 设置 idStart = id = idStop
func DeleteLog(idStart, idStop int64) int64 {
	return DeleteLogTo(logParam.LogTable, idStart, idStop)
}

//DeleteLogTo 删除指定表的日志记录，idStart 到 idStop 的log都会被删除，包含这两个 id，要删除一个id 设置 idStart = id = idStop
func DeleteLogTo(table string, idStart, idStop int64) int64 {
	ret, err := logParam.LogDb.Exec("DELETE FROM "+table+" WHERE id>=? AND id<=?", idStart, idStop)
	if err != nil {
		outputLogTrace(TextRed, 1, err)
		return 0
	}
	count, err := ret.RowsAffected()
	return count
}

//ClearLog 清空数据库中的所有记录
func ClearLog() {
	ClearLogTo(logParam.LogTable)
}

//ClearLogTo 清空指定日志数据库中的所有记录
func ClearLogTo(table string) {
	if logParam.LogDb == nil {
		err := "log database not set"
		outputLogTrace(TextRed, 1, err)
		return
	}
	sqlCase := "TRUNCATE " + table
	_, err := logParam.LogDb.Exec(sqlCase)
	if err != nil {
		outputLogTrace(TextRed, 1, err.Error())
	}
}

// GetLog 读取数据库中保存的日志
// page 第几页，count 读取的日志数量，本质上函数读取的范围是 count * page 到 count * (page+1)
// 所以如果读取开头，page 应设置为 0
// file 指明要获取某一个源文件的日志，空表示获取全部日志
// 返回值为日志数据，和日志总数
func GetLog(page, count int, file string) ([]*LogInfo, int) {
	return GetLogTo(logParam.LogTable, page, count, file)
}
func GetLogTo(table string, page, count int, file string) ([]*LogInfo, int) {
	start := count * page
	lis := make([]*LogInfo, 0, count)
	if logParam.LogDb == nil {
		return lis, 0
	}
	var sqlCase string
	if file == "" {
		sqlCase = "SELECT id,log,trace,color,created_at FROM " + table + " ORDER BY id DESC LIMIT ?,?"
	} else {
		sqlCase = "SELECT id,log,trace,color,created_at FROM " + table + " WHERE trace LIKE '" + file + "%' ORDER BY id DESC LIMIT ?,?"
	}
	rows, err := logParam.LogDb.Query(sqlCase, start, count)
	if checkLogError(err) {
		return lis, 0
	}
	for rows.Next() {
		var li LogInfo
		err = rows.Scan(&li.Id, &li.Log, &li.Trace, &li.Color, &li.CreatedAt)
		if err == nil {
			lis = append(lis, &li)
		} else {
			LogTrace(TextRed, 1, false, err.Error())
		}
	}
	_ = rows.Close()

	var total int
	sqlCase = "SELECT count(*) FROM " + table
	rows, err = logParam.LogDb.Query(sqlCase)
	if checkLogError(err) {
		return lis, 0
	}
	for rows.Next() {
		err = rows.Scan(&total)
		CheckError(err)
	}
	_ = rows.Close()
	return lis, total
}
func saveLog(log, trace string, color int) {
	if logParam.LogDb == nil {
		return
	}
	_, err := logParam.LogDb.Exec("INSERT INTO "+logParam.LogTable+" (log,trace,color) VALUES (?,?,?)",
		log, trace, color)
	checkLogError(err)
}
func saveLogTo2(table, log, trace string, color int) {
	if logParam.LogDb == nil {
		return
	}
	for i := 0; i < 2; i++ {
		query := "INSERT INTO " + table + " (log,trace,color,created_at) VALUES (?,?,?,?)"
		_, err := logParam.LogDb.Exec(query, log, trace, color, GetNowDate())
		if err == nil {
			break
		} else {
			if strings.Index(err.Error(), "1146") != -1 {
				createLogTable(table)
			}
		}
	}
}
func SaveLogTo(tabName, log, trace string, color int) {
	if logParam.LogDb == nil {
		return
	}
	sqlCase := "SELECT MAX(id) FROM " + tabName
	rows, err := logParam.LogDb.Query(sqlCase)
	if checkCreateDeviceLogTable(err, tabName) {
		return
	}

	var id int64
	if rows != nil {
		if rows.Next() {
			_ = rows.Scan(&id)
		}
		_ = rows.Close()
	}
	if id < logParam.MaxLogCount || logParam.MaxLogCount < 0 {
		sqlCase = "INSERT INTO " + tabName + " (log,trace,color,created_at) VALUES (?,?,?,?)"
		_, err = logParam.LogDb.Exec(sqlCase, log, trace, color, GetNowDate())
	} else {
		sqlCase = "UPDATE " + tabName + " SET log=?,trace=?,color=?,created_at=? ORDER BY created_at LIMIT 1"
		_, err = logParam.LogDb.Exec(sqlCase, log, trace, color, GetNowDate())
	}
	CheckError(err)
}
func checkCreateDeviceLogTable(err error, tabName string) bool {
	if err != nil {
		if strings.Index(err.Error(), "1146") != -1 {
			if !createLogTable(tabName) {
				return true
			}
		} else {
			tr := GetTrace(0)
			outPutColor(err.Error(), tr[5], TextRed)
			return true
		}
	}
	return false
}

// SaveLogsTo 一次保存多条日志数据到数据库，使用批量语句优化数据库性能
func SaveLogsTo(table string, logs []*LogInfo) {
	tx, err := logParam.LogDb.Begin()
	if checkLogError(err) {
		return
	}
	query := "INSERT INTO " + table + " (log,trace,color,created_at) VALUES (?,?,?,?)"
	for i, log := range logs {
		date := log.CreatedAt
		if date == "" {
			date = GetNowDate()
		}
		_, err := tx.Exec(query, log.Log, log.Trace, log.Color, date)
		if err != nil {
			if strings.Index(err.Error(), "Error 1146: Table") == 0 {
				createLogTable(table)
			}
			break
		}
		if i > 100 {
			break
		}
	}
	checkLogError(tx.Commit())
}

func getTrace() []string {
	buf := make([]byte, 10240)
	n := runtime.Stack(buf, true)
	str := string(buf[0:n])
	arr := strings.Split(str, "\n")
	if len(arr) > 6 {
		str := arr[6]
		index := strings.LastIndex(str, " ")
		if index != -1 {
			arr[6] = strings.TrimSpace(str[:index])
		}
	}
	return arr
}

//GetTrace
//trace: 获取堆栈层数，default is 3
func GetTrace(trace int) []string {
	trace *= 2
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, true)
	str := string(buf[0:n])
	arr := strings.Split(str, "\n")
	if len(arr) > trace {
		str := arr[trace]
		index := strings.LastIndex(str, " ")
		if index != -1 {
			arr[trace] = strings.TrimSpace(str[:index])
		}
	}
	return arr
}

//FilterIgnoreDiction 过滤字典，有些日志不需要，统一过滤掉。规则是如果一个日志字串包含字典里的任何一个关键字，都会被忽略掉
var FilterIgnoreDiction = map[string]bool{}

func textColor(color int, str string) string {
	switch color {
	case TextBlack:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextBlack, str)
	case TextRed:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextRed, str)
	case TextGreen:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextGreen, str)
	case TextYellow:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextYellow, str)
	case TextBlue:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextBlue, str)
	case TextMagenta:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextMagenta, str)
	case TextCyan:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextCyan, str)
	case TextWhite:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextWhite, str)
	default:
		return str
	}
}
func filterLog(str string) bool {
	for key := range FilterIgnoreDiction {
		if strings.Index(str, key) != -1 {
			return true
		}
	}
	return false
}
func logFColor(str, trace string, color int) {
	//去掉两端的中括号
	if filterLog(str) {
		return
	}
	if logParam.LogDb != nil && logParam.SaveToLog {
		//只显示文件名
		pos := strings.LastIndex(trace, "/")
		if pos != -1 {
			trace = trace[pos+1:]
		}
		saveLog(str, trace, color)
	}
	if logParam.ShowOnConsole {
		outPutColor(str, trace, color)
	}
}
func logColor(str, trace string, color int) {
	//去掉两端的中括号
	strlen := len(str)
	if strlen >= 2 {
		str = str[1 : strlen-1]
	}
	if filterLog(str) {
		return
	}
	if logParam.LogDb != nil && logParam.SaveToLog {
		//只显示文件名
		pos := strings.LastIndex(trace, "/")
		if pos != -1 {
			trace = trace[pos+1:]
		}
		saveLog(str, trace, color)
	}
	if logParam.ShowOnConsole {
		outPutColor(str, trace, color)
	}
}
func logColorTo(tabId, str, trace string, color int) {
	//去掉两端的中括号
	strlen := len(str)
	if strlen >= 2 {
		str = str[1 : strlen-1]
	}
	if filterLog(str) {
		return
	}
	if logParam.LogDb != nil && logParam.SaveToLog {
		//只显示文件名
		pos := strings.LastIndex(trace, "/")
		if pos != -1 {
			trace = trace[pos+1:]
		}
		SaveLogTo(tabId, str, trace, color)
	}
	if logParam.ShowOnConsole {
		outPutColor(str, trace, color)
	}
}
func outPutColor(str, trace string, color int) {
	ts := time.Now().Format("15:04:05")
	str = ts + " " + trace + " " + str
	str = textColor(color, str)
	fmt.Println(str)
}

func OutputColor(color int, v ...interface{}) {
	vs := fmt.Sprint(v)
	ts := time.Now().Format("15:04:05")
	strLen := len(vs)
	if strLen >= 4 {
		vs = vs[2 : strLen-2]
	} else if strLen >= 2 {
		vs = vs[1 : strLen-1]
	}
	str := ts + " " + vs
	str = textColor(color, str)
	fmt.Println(str)
}

func LogBlack(v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColor(vs, arr[6], TextBlack)
}
func LogWhite(v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColor(vs, arr[6], TextWhite)
}
func LogMagenta(v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColor(vs, arr[6], TextMagenta)
}
func LogCyan(v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColor(vs, arr[6], TextCyan)
}
func LogBlue(v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColor(vs, arr[6], TextBlue)
}
func LogRed(v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColor(vs, arr[6], TextRed)
}
func LogGreen(v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColor(vs, arr[6], TextGreen)
}
func LogYellow(v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColor(vs, arr[6], TextYellow)
}

func LogBlackTo(tabId string, v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColorTo(tabId, vs, arr[6], TextBlack)
}
func LogWhiteTo(tabId string, v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColorTo(tabId, vs, arr[6], TextWhite)
}
func LogMagentaTo(tabId string, v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColorTo(tabId, vs, arr[6], TextMagenta)
}
func LogCyanTo(tabId string, v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColorTo(tabId, vs, arr[6], TextCyan)
}
func LogBlueTo(tabId string, v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColorTo(tabId, vs, arr[6], TextBlue)
}
func LogRedTo(tabId string, v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColorTo(tabId, vs, arr[6], TextRed)
}
func LogGreenTo(tabId string, v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColorTo(tabId, vs, arr[6], TextGreen)
}
func LogYellowTo(tabId string, v ...interface{}) {
	vs := fmt.Sprint(v)
	arr := getTrace()
	logColorTo(tabId, vs, arr[6], TextYellow)
}

func LogFGreen(fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logFColor(vs, arr[6], TextGreen)
}
func LogFRed(fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logFColor(vs, arr[6], TextRed)
}
func LogFYellow(fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logFColor(vs, arr[6], TextYellow)
}
func LogFBlue(fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logFColor(vs, arr[6], TextBlue)
}
func LogFCyan(fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logFColor(vs, arr[6], TextCyan)
}
func LogFMagenta(fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logFColor(vs, arr[6], TextMagenta)
}
func LogFWhite(fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logFColor(vs, arr[6], TextWhite)
}
func LogFBlack(fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logFColor(vs, arr[6], TextBlack)
}

func LogFGreenTo(table, fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logColorTo(table, vs, arr[6], TextGreen)
}
func LogFRedTo(table, fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logColorTo(table, vs, arr[6], TextRed)
}
func LogFYellowTo(table, fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logColorTo(table, vs, arr[6], TextYellow)
}
func LogFBlueTo(table, fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logColorTo(table, vs, arr[6], TextBlue)
}
func LogFCyanTo(table, fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logColorTo(table, vs, arr[6], TextCyan)
}
func LogFMagentaTo(table, fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logColorTo(table, vs, arr[6], TextMagenta)
}
func LogFWhiteTo(table, fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logColorTo(table, vs, arr[6], TextWhite)
}
func LogFBlackTo(table, fs string, v ...interface{}) {
	vs := fmt.Sprintf(fs, v...)
	arr := getTrace()
	logColorTo(table, vs, arr[6], TextBlack)
}

//LogTrace 打印调用堆栈信息，一般用于追踪函数调用
//color: 颜色
//trace: 0, 则记录当前位置，1 是上级函数调用位置，依次类推
func LogTrace(color int, trace uint, v ...interface{}) {
	trace += 3
	vs := fmt.Sprint(v)
	arr := GetTrace(int(trace))
	logColor(vs, arr[trace*2], color)
}
func outputLogTrace(color int, trace uint, v ...interface{}) {
	trace += 3
	vs := fmt.Sprint(v)
	arr := GetTrace(int(trace))
	logColor(vs, arr[trace*2], color)
}

//checkLogError 这个函数仅输出，不会保存到数据库，用于输出 log 数据库异常
func checkLogError(err error) bool {
	if err != nil {
		vs := fmt.Sprint(err.Error())
		trace := 3
		arr := GetTrace(int(trace))
		outPutColor(vs, arr[trace*2], TextRed)
		return true
	}
	return false
}
