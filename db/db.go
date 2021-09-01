package db
import (
    "github.com/joho/godotenv"
    "log"
    "os"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

var dbh *sql.DB

func init() {
    godotenv.Load()
    var dbc = os.Getenv("DB_CONNECTION")
    var err error

    log.Println("connecting " + dbc)

    dbh, err = sql.Open("mysql", dbc)

    if err != nil {
        panic(err)
    }

    dbh.SetMaxOpenConns(10)
    dbh.SetMaxIdleConns(10)
    dbh.SetConnMaxIdleTime(5000000000)
}

func Db() *sql.DB {
    return dbh
}

func Query(sql string, params ...interface{}) (*sql.Rows, func()) {
    rows, err := dbh.Query(sql, params...)

    if err != nil {
        panic(err)
    }

    return rows, func() {
        if err := rows.Close(); err != nil {
            panic(err)
        }

        if err := rows.Err(); err != nil {
            panic(err)
        }
    }
}

func QueryRow(query string, params ...interface{}) *sql.Row {
    row := dbh.QueryRow(query, params...)

    if err := row.Err(); err != nil {
        panic(err)
    }

    return row
}

func Exec(query string, params ...interface{}) *sql.Result {
    var result sql.Result
    var err error

    if result, err = dbh.Exec(query, params...); err != nil {
        panic(err)
    }
    return &result
}

func ScanRows(rows *sql.Rows, params ...interface{}) {
    if err := rows.Scan(params...); err != nil {
        panic(err)
    }
}

func ScanRow(row *sql.Row, params ...interface{}) {
    if err := row.Scan(params...); err != nil {
        panic(err)
    }
}
