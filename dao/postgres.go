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
	url := os.Getenv("DATABASE_URL")
	if url != "" {
		if os.Getenv("DATABASE_SSL") == "false" {
			url = url + "?sslmode=disable"
		}
		db, _ := sql.Open("postgres", url)
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
