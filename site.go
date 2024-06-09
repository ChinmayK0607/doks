package main

import (
	"fmt"
	"net/http"
)

// handles routes
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/export/pdf", exportPDF)
	http.HandleFunc("/export/md", exportMD)

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
