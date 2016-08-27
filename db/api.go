package db

import (
	"crypto/sha256"
	"database/sql"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"
)

const (
	MaxRecent = 8
)

type MessageId string
type HashKey [sha256.Size]byte
type Message struct {
	Text string
	Id   MessageId
}

type API interface {
	AddMessage(message string) MessageId
	GetMessage(id MessageId) (string, bool)
	RecentMessages(limit ...int) []Message
	RandomMessage() Message
}

func generateId() MessageId {
	var min int64 = int64(math.Pow(35, 5))
	var max int64 = int64(math.Pow(35, 6))
	return MessageId(strconv.FormatInt(min+rand.Int63n(max-min), 35))
}

type MemStore struct {
	hashes         map[HashKey]MessageId
	messages       map[MessageId]string
	recentMessages []Message
}

func NewMemStore() *MemStore {
	return &MemStore{
		hashes:   make(map[HashKey]MessageId),
		messages: make(map[MessageId]string),
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

func (store *MemStore) RecentMessages(limitArg ...int) []Message {
	limit := MaxRecent
	if len(limitArg) > 0 {
		limit = limitArg[0]
	}
	messages := store.recentMessages
	n := len(messages)
	messages = messages[max(0, n-limit):n]
	// TODO: reverse order
	return messages
}

func (store *MemStore) RandomMessage() Message {
	n := len(store.messages)
	if n == 0 {
		return Message{}
	}
	if n == 1 {
		for k, v := range store.messages {
			return Message{v, k}
		}
	}
	var huh []MessageId
	for k := range store.messages {
		huh = append(huh, k)
	}
	k := huh[rand.Intn(n-1)]
	return Message{store.messages[k], k}
}

func (store *MemStore) addToRecentMessages(text string, id MessageId) {
	messages := store.recentMessages
	messages = append(messages, Message{
		Text: text,
		Id:   id,
	})
	if len(messages) > MaxRecent {
		messages = messages[1:]
	}
	store.recentMessages = messages
}

type SqlStore struct{}

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
	if err != sql.ErrNoRows {
		panic(err)
	}

	id = string(generateId())
	stmt, err := Db.Prepare(`INSERT INTO messages(id, content, hash, time) VALUES ($1, $2, $3, now())`)

	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec(string(id), message, hash)
	if err != nil {
		panic(err)
	}

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

func (store *SqlStore) RecentMessages(limitArg ...int) (messages []Message) {
	limit := MaxRecent
	if len(limitArg) > 0 {
		limit = limitArg[0]
	}

	rows, err := Db.Query(`
        SELECT content, id FROM messages
            ORDER BY time DESC
            LIMIT $1
        `, limit)

	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var content, id string
		rows.Scan(&content, &id)
		messages = append(messages, Message{content, MessageId(id)})
	}

	return
}

func (store *SqlStore) RandomMessage() Message {
	rows, err := Db.Query(`
        SELECT content, id FROM messages
            ORDER BY RAND()
            LIMIT 1
	`)

	if err != nil {
		panic(err)
	}
	if rows.Next() {
		var content, id string
		rows.Scan(&content, &id)
		return Message{content, MessageId(id)}
	}
	return Message{}
}

func max(x int, y int) int {
	return int(math.Max(float64(x), float64(y)))
}

func init() {
	rand.Seed(time.Now().Unix())
}
