package db
import (
    "github.com/joho/godotenv"
    "log"
    "os"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

var dbh *sql.DB

func Db() *sql.DB {
    if dbh != nil {
        return dbh
    }

    godotenv.Load()
    var dbc = os.Getenv("DB_CONNECTION")

    log.Println("connecting " + dbc)

    dbh, err := sql.Open("mysql", dbc)

    if err != nil {
        panic(err)
    }

    dbh.SetMaxOpenConns(10)
    dbh.SetMaxIdleConns(10)
    dbh.SetConnMaxIdleTime(5000000000)
    return dbh
}
