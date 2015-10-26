package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/bmizerany/pat"
)

var (
	templates *template.Template
)

type SMS struct {
	ID        string
	To        string
	From      string
	Body      string
	Timestamp string
}

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

	var jsonStr = []byte(`{"to":"` + r.Form.Get("to") + `","from":"` + r.Form.Get("msisdn") + `","body":"` + r.Form.Get("text") + `","timestamp":"` + timestamp.UTC().String() + `"}`)
	req, err := http.NewRequest("POST", "http://127.0.0.1:9200/messages/"+r.Form.Get("msisdn"), bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	log.Println(string(body))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

}

func dashboard(w http.ResponseWriter, r *http.Request) {

	var jsonStr = []byte(`{"query":{"match_all":{}},"aggs":{"top-types":{"terms":{"field":"_type"},"aggs":{"top_docs":{"top_hits":{"sort":[{"_score":{"order":"desc"}}],"size":1}}}}}}`)
	req, err := http.NewRequest("POST", "http://127.0.0.1:9200/messages/_search?search_type=count", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	log.Println(string(body))

	data := struct {
		Aggregations struct {
			Top struct {
				Buckets []struct {
					Top struct {
						Hits struct {
							Hits []struct {
								Index  string `json:"_index"`
								Type   string `json:"_type"`
								ID     string `json:"_id"`
								Source struct {
									To        string `json:"to"`
									From      string `json:"from"`
									Body      string `json:"body"`
									Timestamp string `json:"timestamp"`
								} `json:"_source"`
							} `json:"hits"`
						} `json:"hits"`
					} `json:"top_docs"`
				} `json:"buckets"`
			} `json:"top-types"`
		} `json:"aggregations"`
	}{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	messages := []SMS{}

	for _, bucket := range data.Aggregations.Top.Buckets {

		messages = append(messages, SMS{
			bucket.Top.Hits.Hits[0].ID,
			bucket.Top.Hits.Hits[0].Source.To,
			bucket.Top.Hits.Hits[0].Source.From,
			bucket.Top.Hits.Hits[0].Source.Body,
			bucket.Top.Hits.Hits[0].Source.Timestamp,
		})

	}

	log.Println(messages)

	err = templates.ExecuteTemplate(w, "index.html", struct{ Messages []SMS }{messages})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
		return
	}

}

func main() {

	var err error

	templates, err = template.ParseGlob("templates/*")
	if err != nil {
		panic(err)
	}

	router := pat.New()

	router.Get("/hooks/nexmo", http.HandlerFunc(nexmo))
	router.Get("/", http.HandlerFunc(dashboard))

	log.Fatal(http.ListenAndServe(":8080", router))

}
