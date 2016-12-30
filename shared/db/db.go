package dbutil

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

type dbConfig struct {
	Host     string
	Username string
	Password string
}

var dbConfigs = map[string]dbConfig{
	"alpha": dbConfig{
		Host:     "52.186.123.148",
		Username: "superanswer",
		Password: "supermeme2",
	},
}

func InitDb(dbName string) {
	dbConfig, ok := dbConfigs[dbName]
	if !ok {
		glog.Fatalf("dbName: %s not registered in dbConfigs", dbName)
	}
	// datasourceName format
	// username:password@protocol(address)/dbname?param=value
	datasourceName := fmt.Sprintf("%s:%s@(%s:3306)/%s?parseTime=true",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbName)
	var err error
	db, err = sqlx.Connect("mysql", datasourceName)
	if err != nil {
		glog.Fatal(err)
	}
}

func DbContext() *sqlx.DB {
	return db
}
