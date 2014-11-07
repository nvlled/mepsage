
package main

import (
    "net/http"
    //"strings"
    "log"
    "html/template"
    "github.com/gorilla/mux"
    "os"
)

const (
    Port = "7000"
)

var templ map[string]*template.Template

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
