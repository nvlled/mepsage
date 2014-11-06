
package main

import (
    "net/http"
    "fmt"
    //"strings"
    "log"
    "html/template"
    "github.com/gorilla/mux"
    "os"
    "github.com/nvlled/mepsage/db"
)

const (
    Port = "7000"
    ResourcesDir = "static/"
    messagePath = "/m/"
)

var templ map[string]*template.Template

func init() {
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

    log.Println("server started at " + port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}

func buildRoutes() http.Handler {
    fileServer := http.FileServer(http.Dir(ResourcesDir))
    router := mux.NewRouter()
    router.StrictSlash(true)

    router.HandleFunc("/", indexPage)
    router.HandleFunc("/about", aboutPage)
    router.HandleFunc("/submit", submitPage)
    router.HandleFunc("/submit-async", submitAsync)
    router.HandleFunc("/{id}", messagePage)

    sub := router.PathPrefix("/"+ResourcesDir)
    sub.Handler(http.StripPrefix("/"+ResourcesDir, fileServer))

    return router
}

func messagePage(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    msg, ok := db.GetMessage(db.MessageId(id))

    if !ok {
        msg = "404 Page Not Found"
    }
    render(w, "message", map[string]interface{}{
        "Message": msg,
    })
}

func aboutPage(w http.ResponseWriter, r *http.Request) {
    render(w, "about", map[string]interface{}{
        "Messages": db.RecentMessages(),
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

func submitPage(w http.ResponseWriter, r *http.Request) {
    msg := r.FormValue("msg")
    id := db.AddMessage(msg)
    w.Header().Set("Location", "/"+string(id))
    w.WriteHeader(301)
    fmt.Fprint(w, "/"+id)
}

func submitAsync(w http.ResponseWriter, r *http.Request) {
    msg := r.FormValue("msg")
    id := db.AddMessage(msg)
    fmt.Fprint(w, "/"+id)
}
