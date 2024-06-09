package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/jung-kurt/gofpdf"
)

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func exportPDF(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	content := r.FormValue("content")

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 10, content, "", "", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=export.pdf")
	w.Write(buf.Bytes())
}

func exportMD(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	content := r.FormValue("content")

	w.Header().Set("Content-Type", "text/markdown")
	w.Header().Set("Content-Disposition", "attachment; filename=export.md")
	w.Write([]byte(content))
}

func copyToClipboard(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Copied to clipboard"})
}

func analyzeWithAI(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	analysis := callLLM(data.Content)
	json.NewEncoder(w).Encode(map[string]string{"analysis": analysis})
}

func callLLM(content string) string {
	// This is a placeholder. Replace with actual LLM call.
	// Example: Make an HTTP request to an LLM API
	resp, err := http.Post("https://api.example.com/analyze", "application/json", bytes.NewBufferString(content))
	if err != nil {
		return "Error calling LLM"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

