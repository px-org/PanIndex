package dao

import (
	"fmt"
	gorm_logrus "github.com/onrik/gorm-logrus"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func init() {
	RegisterDb("sqlite", &Sqlite{})
}

type Sqlite struct{}

func (s Sqlite) CreateDb(dsn string) {
	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gorm_logrus.New(),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
	} else {
		log.Info("[Boot]Sqlite >> Database connection succeeded")
	}
	DB = database
	InitDb()
}
