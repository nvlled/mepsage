
package main

import (
    "net/http"
    "github.com/gorilla/mux"
    "github.com/nvlled/mepsage/db"
    "fmt"
    "math/rand"
)

const (
    ResourcesDir = "static/"
)

func indexPage(w http.ResponseWriter, r *http.Request) {
    messages := db.RecentMessages(4)
    n := len(messages)

    var location string
    if n > 0 {
        i := rand.Intn(n-1)
        msg := messages[i]
        location = "/"+string(msg.Id)
    } else {
        location = "/create"
    }

    w.Header().Set("Location", location)
    w.WriteHeader(302)
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

func createPage(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "pages/create.html")
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
