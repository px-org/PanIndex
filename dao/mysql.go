package dao

import (
	"fmt"
	gorm_logrus "github.com/onrik/gorm-logrus"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func init() {
	RegisterDb("mysql", &Mysql{})
}

type Mysql struct{}

func (driver Mysql) CreateDb(dsn string) {
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gorm_logrus.New(),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
	} else {
		log.Info("[Boot]Mysql >> Database connection succeeded")
	}
	DB = database
	InitDb()
}
