package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/jung-kurt/gofpdf"
	"github.com/shurcooL/github_flavored_markdown"
)

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
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

	// Process the content with GitHub Flavored Markdown
	mdContent_test := github_flavored_markdown.Markdown([]byte(content))

	reBoldOpen := regexp.MustCompile(`(?i)<b>`)
	reBoldClose := regexp.MustCompile(`(?i)\s*</b>`)
	reItalicOpen := regexp.MustCompile(`(?i)<i>`)
	reItalicClose := regexp.MustCompile(`(?i)\s*</i>`)
	reRemoveTags := regexp.MustCompile(`(?i)<[^/b|/i][^>]*>`)

	mdContent := reBoldOpen.ReplaceAllString(string(mdContent_test), "**")
	mdContent = reBoldClose.ReplaceAllString(mdContent, "**")
	mdContent = reItalicOpen.ReplaceAllString(mdContent, "*")
	mdContent = reItalicClose.ReplaceAllString(mdContent, "*")

	mdContent = reRemoveTags.ReplaceAllString(mdContent, "")
	w.Header().Set("Content-Type", "text/markdown")
	w.Header().Set("Content-Disposition", "attachment; filename=export.md")
	fmt.Println(mdContent)
	w.Write([]byte(mdContent))
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
	resp, err := http.Post("https://api.example.com/analyze", "application/json", bytes.NewBufferString(content))
	if err != nil {
		return "Error calling LLM"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

// func main() {
// 	http.HandleFunc("/", serveIndex)
// 	http.HandleFunc("/export/pdf", exportPDF)
// 	http.HandleFunc("/export/md", exportMD)
// 	http.HandleFunc("/copy", copyToClipboard)
// 	http.HandleFunc("/analyze", analyzeWithAI)

// 	http.ListenAndServe(":8080", nil)
// }
