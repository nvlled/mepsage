
package main

import (
    "net/http"
    "fmt"
    //"strings"
    "log"
    "strconv"
    "math"
    "math/rand"
    "html/template"
    "github.com/gorilla/mux"
    "os"
    "time"
    "crypto/sha256"
)

type MessageId string
type HashKey [sha256.Size]byte
type Message struct{
    Text string
    Id MessageId
}

const (
    Port = "7000"
    ResourcesDir = "static/"
    messagePath = "/m/"
    MaxRecent = 10
)

var (
    hashes map[HashKey]MessageId
    messages map[MessageId]string
    recentMessages []Message
    templ map[string]*template.Template
)

func init() {
    rand.Seed(time.Now().Unix())
    hashes = make(map[HashKey]MessageId)
    messages = make(map[MessageId]string)
    templ = make(map[string]*template.Template)
    initTemplates()
}

func initTemplates() {
    t, err := template.ParseFiles("pages/message.html")
    if err != nil { panic(err) }
    templ["message"] = t

    t, err = template.ParseFiles("pages/about.html")
    if err != nil { panic(err) }
    templ["about"] = t
}

func main() {
    handler := buildRoutes()
    port := os.Getenv("PORT")
    if port == "" {
        port = Port
    }

    log.Println("Server started at " + port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}

func buildRoutes() http.Handler {
    fileServer := http.FileServer(http.Dir(ResourcesDir))
    router := mux.NewRouter()
    router.StrictSlash(true)

    router.HandleFunc("/", indexPage)
    router.HandleFunc("/about", aboutPage)
    router.HandleFunc("/submit", submit)
    router.HandleFunc("/{id}", messagePage)

    sub := router.PathPrefix("/"+ResourcesDir)
    sub.Handler(http.StripPrefix("/"+ResourcesDir, fileServer))

    return router
}

func messagePage(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    msg, ok := messages[MessageId(id)]

    if !ok {
        msg = "404 Page Not Found"
    }
    render(w, "message", map[string]interface{}{
        "Message": msg,
    })
}

func aboutPage(w http.ResponseWriter, r *http.Request) {
    render(w, "about", map[string]interface{}{
        "Messages": recentMessages,
    })
}

func indexPage(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "pages/index.html")
}

func render(w http.ResponseWriter, name string, data map[string]interface{}) {
    t, ok := templ[name]
    if !ok {
        panic("template not found")
    }
    t.Execute(w, data)
}

func submit(w http.ResponseWriter, r *http.Request) {
    msg := r.FormValue("msg")

    key := HashKey(sha256.Sum256([]byte(msg)))
    if id, ok := hashes[key]; ok {
        log.Println("Re-using message with id", id)
        fmt.Fprint(w, "/"+id)
    } else {
        id := generateId()
        hashes[key] = id
        messages[id] = msg
        addToRecentMessages(msg, id)
        fmt.Fprint(w, "/"+id)
    }
}

func addToRecentMessages(text string, id MessageId) {
    recentMessages = append(recentMessages, Message{
        Text: text,
        Id: id,
    })
    if len(recentMessages) > MaxRecent {
        recentMessages = recentMessages[1:]
    }
}

func generateId() MessageId {
    var min int64 = int64(math.Pow(35, 5))
    var max int64 = int64(math.Pow(35, 6))
    return MessageId(strconv.FormatInt(min + rand.Int63n(max - min), 35))
}
