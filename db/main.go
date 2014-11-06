
package db

import (
    "os"
    "io/ioutil"
    "encoding/json"
    "database/sql"
    _ "github.com/lib/pq"
    "log"
)

const (
    DB_PATH = "DBCRED"
)

var DbName string
var DbSource  string
var Db   *sql.DB

var Store API

func AddMessage(message string) MessageId {
    return Store.AddMessage(message)
}

func GetMessage(id MessageId) (string, bool) {
    return Store.GetMessage(id)
}

func RecentMessages() []Message {
    return Store.RecentMessages()
}

func initSqlStore() (API, error) {
    DbName = os.Getenv("DBNAME")
    DbSource = os.Getenv("DATABASE_URL")

    if DbSource == "" {
        file, err := os.Open(DB_PATH)
        if err != nil { return nil, err }

        bytes, err := ioutil.ReadAll(file)
        var db struct{
            Name string
            DataSource string
        }
        err = json.Unmarshal(bytes, &db)
        if err != nil { return nil, err }
        DbName =  db.Name
        DbSource =  db.DataSource

    }
    if DbName == "" {
        DbName = "postgres"
    }

    db, err := sql.Open(DbName, DbSource)
    if err != nil { return nil, err }
    err = db.Ping()
    if err != nil { return nil, err }
    Db = db

    return NewSqlStore(), nil
}

func createTables() {
    if os.Getenv("DROPDB") == "YES" {
        log.Println("Dropping messages table...")
        Db.Exec(`DROP TABLE messages`)
    }

    log.Println("Creating messages table...")
    _, err := Db.Exec(`CREATE Table messages(
        id char(10) PRIMARY KEY,
        content text,
        hash bytea,
        time timestamp
    )`)
    if err != nil {
        log.Println("messages table already exists")
    } else {
        log.Println("message table created")
    }
}

func init() {
    store, err := initSqlStore()
    if err != nil {
        log.Printf("%v", err)
        log.Println("using Memstore...")
        store = NewMemStore()
    } else {
        createTables()
    }
    Store = store
}
