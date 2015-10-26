package main

import (
	"log"
	"net/http"

	"github.com/bmizerany/pat"
)

type sms struct {
	to      string
	from    string
	message string
}

func nexmoWebhook(w http.ResponseWriter, r *http.Request) {

	msg := sms{
		r.FormValue("to"),
		r.FormValue("msisdn"),
		r.FormValue("text"),
	}

	log.Println(msg)

}

func main() {

	router := pat.New()

	router.Post("/hooks/nexmo", http.HandlerFunc(nexmoWebhook))

	log.Fatal(http.ListenAndServe(":8080", router))

}
