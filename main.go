
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
    //"os"
    "time"
    "crypto/sha256"
)

type MessageId string
type HashKey [sha256.Size]byte

const (
    Port = "7000"
    ResourcesDir = "static/"
    messagePath = "/m/"
)

var (
    hashes map[HashKey]MessageId
    messages map[MessageId]string
)

func init() {
    rand.Seed(time.Now().Unix())
    hashes = make(map[HashKey]MessageId)
    messages = make(map[MessageId]string)
}

func main() {
    handler := buildRoutes()

    log.Println("Server started at " + Port)
    log.Fatal(http.ListenAndServe(":"+Port, handler))
}

func buildRoutes() http.Handler {
    fileServer := http.FileServer(http.Dir(ResourcesDir))
    router := mux.NewRouter()
    router.StrictSlash(true)

    router.HandleFunc("/", indexPage)
    router.HandleFunc("/wat", watPage)
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
    renderMessage(w, msg)
}

func renderMessage(w http.ResponseWriter, msg string) {
    t, _ := template.ParseFiles("pages/message.html")
    t.Execute(w, map[string] string {
        "Message": msg,
    })
}

func indexPage(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        renderMessage(w, "You must be lost.")
    } else {
        http.ServeFile(w, r, "pages/index.html")
    }
}

func watPage(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "pages/wat.html")
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
        fmt.Fprint(w, "/"+id)
    }
}

func generateId() MessageId {
    var min int64 = int64(math.Pow(35, 5))
    var max int64 = int64(math.Pow(35, 6))
    return MessageId(strconv.FormatInt(min + rand.Int63n(max - min), 35))
}
