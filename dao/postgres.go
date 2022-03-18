package dao

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	gorm_logrus "github.com/onrik/gorm-logrus"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"os"
)

func init() {
	RegisterDb("postgres", &Postgres{})
}

type Postgres struct{}

func (driver Postgres) CreateDb(dsn string) {
	dialector := postgres.Open(dsn)
	if os.Getenv("DATABASE_URL") != "" {
		db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
		dialector = postgres.New(postgres.Config{
			Conn: db,
		})
	}
	database, err := gorm.Open(dialector, &gorm.Config{
		Logger: gorm_logrus.New(),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
	} else {
		log.Info("[Boot]Postgres >> Database connection succeeded")
	}
	DB = database
	InitDb()
}
