//go:build !wasip1

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"pdf_wasm/internal/template"
)

func main() {
	// Lire JSON depuis stdin
	in, err := os.ReadFile("/dev/stdin")
	if err != nil {
		fmt.Fprintln(os.Stderr, "read stdin:", err)
		os.Exit(1)
	}

	var tmpl template.Template
	if err := json.Unmarshal(in, &tmpl); err != nil {
		fmt.Fprintln(os.Stderr, "bad json:", err)
		os.Exit(1)
	}

	// Générer le PDF depuis le template
	pdfBytes, err := template.GeneratePDF(tmpl)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pdf error:", err)
		os.Exit(1)
	}

	// Écrire le PDF binaire sur stdout
	os.Stdout.Write(pdfBytes)
}
