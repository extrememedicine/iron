package main

import (
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/bmizerany/pat"
	"github.com/gocql/gocql"
)

var (
	templates *template.Template
	cluster   *gocql.ClusterConfig
)

func nexmo(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}

	if r.Form.Get("type") != "text" {
		log.Println("Invalid message type")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
		return
	}

	// YYYY-MM-DD HH:MM:SS
	timestamp, err := time.Parse("2006-01-02 15:04:05", r.Form.Get("message-timestamp"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
		return
	}

	session, err := cluster.CreateSession()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
		return
	}

	defer session.Close()

	err = session.Query(`INSERT INTO messages (id, receiver, sender, body, time) VALUES (?, ?, ?, ?, ?)`, gocql.TimeUUID(), r.Form.Get("to"), r.Form.Get("msisdn"), r.Form.Get("text"), timestamp.UTC()).Exec()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

}

func dashboard(w http.ResponseWriter, r *http.Request) {

	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
	}

}

func main() {

	var err error

	templates, err = template.ParseGlob("templates/*")
	if err != nil {
		panic(err)
	}

	cluster = gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "iron"

	router := pat.New()

	router.Get("/hooks/nexmo", http.HandlerFunc(nexmo))
	router.Get("/", http.HandlerFunc(dashboard))

	log.Fatal(http.ListenAndServe(":8080", router))

}
