package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jung-kurt/gofpdf"
)

// func main() {
// 	http.HandleFunc("/", serveIndex)
// 	http.HandleFunc("/export/pdf", exportPDF)
// 	http.HandleFunc("/export/md", exportMD)
// 	http.HandleFunc("/action/copy", copyToClipboard)
// 	http.HandleFunc("/action/analyze", analyzeWithAI)

// 	fmt.Println("Server started at http://localhost:8080")
// 	http.ListenAndServe(":8080", nil)
// }

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

	// Prepare the JSON payload to send to the FastAPI server
	payload := map[string]string{"content": content}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Make the POST request to the FastAPI server
	resp, err := http.Post("http://localhost:8000/export/md", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response from the FastAPI server
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse the response
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the Markdown content to the response
	w.Header().Set("Content-Type", "text/markdown")
	w.Header().Set("Content-Disposition", "attachment; filename=export.md")
	w.Write([]byte(result["markdown"]))
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

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func summarizeContent(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	payload := map[string]string{"text": data.Text}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("http://localhost:8000/summarize", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response from the FastAPI server
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
