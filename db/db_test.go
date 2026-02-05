package db_test

import (
	"os"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var dbTest *sqlx.DB
var getDbOnce sync.Once

func getDBTest() *sqlx.DB {
	getDbOnce.Do(func() {
		var err error
		dbTest, err = sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
		if err != nil {
			panic(err)
		}
	})
	return dbTest
}
