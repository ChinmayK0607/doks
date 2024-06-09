package main

import (
	"bytes"
	"net/http"

	"github.com/jung-kurt/gofpdf"
	"github.com/shurcooL/github_flavored_markdown"
)

// TODO: function to handle llm.
func ai() {

}

// handles functions
// TODO: doesnt really work, exports html??
func exportPDF(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	content := r.FormValue("content")

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 10, content, "", "", false)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
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

	md := github_flavored_markdown.Markdown([]byte(content))

	w.Header().Set("Content-Type", "text/markdown")
	w.Header().Set("Content-Disposition", "attachment; filename=export.md")
	w.Write(md)
}
