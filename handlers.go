package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nvlled/mepsage/db"
	"net/http"
)

const (
	ResourcesDir = "static/"
)

func indexPage(w http.ResponseWriter, r *http.Request) {
	var location string
	msg := db.RandomMessage()
	if msg.Id == "" {
		location = "/create"
	} else {
		location = "/" + string(msg.Id)
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
