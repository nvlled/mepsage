
package db

import (
    "log"
    "crypto/sha256"
    "database/sql"
    "math"
    "math/rand"
    "strconv"
    "time"
)

const (
    MaxRecent = 10
)

type MessageId string
type HashKey [sha256.Size]byte
type Message struct{
    Text string
    Id MessageId
}

type API interface {
    AddMessage(message string) MessageId
    GetMessage(id MessageId) (string, bool)
    RecentMessages() []Message
}

func generateId() MessageId {
    var min int64 = int64(math.Pow(35, 5))
    var max int64 = int64(math.Pow(35, 6))
    return MessageId(strconv.FormatInt(min + rand.Int63n(max - min), 35))
}

type MemStore struct {
    hashes map[HashKey]MessageId
    messages map[MessageId]string
    recentMessages []Message
}

func NewMemStore() *MemStore {
    return &MemStore{
        hashes : make(map[HashKey]MessageId),
        messages : make(map[MessageId]string),
    }
}

func (store *MemStore) AddMessage(msg string) MessageId {
    var messageId MessageId
    key := HashKey(sha256.Sum256([]byte(msg)))
    if id, ok := store.hashes[key]; ok {
        log.Println("Re-using message with id", id)
        messageId = id
    } else {
        id := generateId()
        store.hashes[key] = id
        store.messages[id] = msg
        store.addToRecentMessages(msg, id)
        messageId = id
    }
    return messageId
}

func (store *MemStore) GetMessage(id MessageId) (string, bool) {
    msg, ok := store.messages[id]
    return msg, ok
}

func (store *MemStore) RecentMessages() []Message {
    return store.recentMessages
}

func (store *MemStore) addToRecentMessages(text string, id MessageId) {
    messages := store.recentMessages
    messages = append(messages, Message{
        Text: text,
        Id: id,
    })
    if len(messages) > MaxRecent {
        messages = messages[1:]
    }
    store.recentMessages = messages
}

type SqlStore struct { }

func NewSqlStore() *SqlStore {
    if Db == nil {
        panic("No db connection available")
    }
    return &SqlStore{}
}

func (store *SqlStore) AddMessage(message string) MessageId {
    var id string
    bytes := sha256.Sum256([]byte(message))
    hash := string(bytes[:])

    row := Db.QueryRow(`SELECT id FROM messages where hash = $1`, hash)
    err := row.Scan(&id)
    if err == nil {
        log.Println("Re-using message with id", id)
        return MessageId(id)
    }
    if err != sql.ErrNoRows { panic(err) }

    id = string(generateId())
    stmt, err := Db.Prepare(`INSERT INTO messages(id, content, hash, time) VALUES ($1, $2, $3, now())`)

    if err != nil { panic(err) }
    _, err = stmt.Exec(string(id), message, hash)
    if err != nil { panic(err) }

    return MessageId(id)
}

func (store *SqlStore) GetMessage(id MessageId) (string, bool) {
    var content string
    row := Db.QueryRow(`
        SELECT content FROM messages
            WHERE id = $1
            ORDER BY time DESC
        `, string(id))
    err := row.Scan(&content)

    if err != nil {
        return "", false
    }
    return content, true
}

func (store *SqlStore) RecentMessages() (messages []Message) {
    rows, err := Db.Query(`
        SELECT content, id FROM messages
            ORDER BY time DESC
            LIMIT $1
        `, MaxRecent)

    if err != nil { panic(err) }
    for rows.Next() {
        var content, id string
        rows.Scan(&content, &id)
        messages = append(messages, Message{content, MessageId(id)})
    }

    return
}

func init() {
    rand.Seed(time.Now().Unix())
}