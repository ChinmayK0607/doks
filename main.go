package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/export/pdf", exportPDF)
	http.HandleFunc("/export/md", exportMD)
	http.HandleFunc("/action/copy", copyToClipboard)
	http.HandleFunc("/action/analyze", analyzeWithAI)
	http.HandleFunc("/action/summarize", summarizeContent)
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
