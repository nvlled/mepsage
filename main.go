
package main

import (
    "net/http"
    "fmt"
    "strings"
    "log"
    "strconv"
    "math"
    "math/rand"
    "html/template"
    //"os"
    "time"
)

const (
    Port = "7000"
    ResourcesDir = "static/"
    messagePath = "/m/"
)

var (
    messages map[string]string
)

func init() {
    rand.Seed(time.Now().Unix())
    messages = make(map[string]string)
}

func main() {
    addRoutes()
    log.Println("Server started at " + Port)
    log.Fatal(http.ListenAndServe(":"+Port, nil))
}

func addRoutes() {
    fileServer := http.FileServer(http.Dir(ResourcesDir))
    http.Handle("/"+ResourcesDir, http.StripPrefix("/"+ResourcesDir, fileServer))

    http.HandleFunc("/", indexPage)
    http.HandleFunc("/submit", submit)
    http.HandleFunc(messagePath, getMessage)
}

func getMessage(w http.ResponseWriter, r *http.Request) {
    id := strings.TrimPrefix(r.URL.Path,messagePath)
    msg, ok := messages[id]

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

func submit(w http.ResponseWriter, r *http.Request) {
    msg := r.FormValue("msg")
    id := generateId()
    messages[id] = msg
    fmt.Fprint(w,messagePath+id)
}

func generateId() string {
    var min int64 = int64(math.Pow(35, 5))
    var max int64 = int64(math.Pow(35, 6))
    return strconv.FormatInt(min + rand.Int63n(max - min), 35)
}

