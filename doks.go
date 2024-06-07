package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

var tmpl = template.Must(template.ParseFiles("index.html"))

var users = map[string]string{
	"athu":  "athu",
	"karan": "karan",
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || users[username] != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func main() {
	http.HandleFunc("/", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	}))

	http.HandleFunc("/render", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		r.ParseForm()
		markdownText := r.FormValue("markdown")
		unsafeHTML := blackfriday.Run([]byte(markdownText))
		safeHTML := bluemonday.UGCPolicy().SanitizeBytes(unsafeHTML)

		response := map[string]string{"html": string(safeHTML)}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	log.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

