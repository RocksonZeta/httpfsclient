package httpfsclient

import (
	"database/sql"
	"os"
	"time"
)

const (
	StateOk    = 0
	StateError = 1
)

type JsonResult struct {
	State int
	Data  interface{}
	Err   error
}

func (r *JsonResult) Error() string {
	return r.Err.Error()
}

type FileInfo struct {
	Name     string
	Size     int64
	Mode     os.FileMode
	ModeTime time.Time
	IsDir    bool
	RawName  string
}

type File struct {
	Id         int
	ServerId   int
	Path       string
	Bytes      int64
	State      int
	CreateTime int64
	Mime       sql.NullString
	RawName    sql.NullString
	Backup1    sql.NullInt64
	Backup2    sql.NullInt64
}

// type Server struct {
// 	Id         int
// 	Server     string
// 	Proxy      string
// 	Root       string
// 	Ready      bool
// 	RatedSpace int
// 	UsedSpace  int64
// 	Backup1    sql.NullInt64
// 	Backup2    sql.NullInt64
// }

type Server struct {
	ClusterId, ServerId string
	Local, Proxy        string
	Ut                  int64 //update time
	RatedSpace          int   //MB
	FreeSpace           int   //MB
	Cpu                 int   //
	Mem                 int   //MB
	MemFree             int
	LoadAverage         int
}

type RedisService interface {
	HMGetAll(key string, result interface{}) error
}
