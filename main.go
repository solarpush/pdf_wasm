//go:build !wasip1

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"pdf_wasm/internal/template"
)

// Structure pour accepter template + variables
type InputData struct {
	PdfTemplate *template.Template     `json:"pdf_template,omitempty"`
	PdfVars     map[string]interface{} `json:"pdfVars,omitempty"`
	// Compatibilité avec l'ancien format (template direct)
	template.Template
}

func main() {
	// Lire JSON depuis stdin
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read stdin:", err)
		os.Exit(1)
	}

	var input InputData
	if err := json.Unmarshal(in, &input); err != nil {
		fmt.Fprintln(os.Stderr, "bad json:", err)
		os.Exit(1)
	}

	var pdfBytes []byte

	// Nouveau format avec template + variables
	if input.PdfTemplate != nil {
		// Convertir le template en JSON puis le traiter avec les variables
		templateBytes, err := json.Marshal(input.PdfTemplate)
		if err != nil {
			fmt.Fprintln(os.Stderr, "template marshal error:", err)
			os.Exit(1)
		}

		// Traiter le template avec les variables
		pdfBytes, err = template.GeneratePDFFromContent(templateBytes, input.PdfVars)
		if err != nil {
			fmt.Fprintln(os.Stderr, "pdf generation error:", err)
			os.Exit(1)
		}
	} else {
		// Ancien format (template direct) - compatibilité ascendante
		pdfBytes, err = template.GeneratePDF(input.Template)
		if err != nil {
			fmt.Fprintln(os.Stderr, "pdf error:", err)
			os.Exit(1)
		}
	}

	// Écrire le PDF binaire sur stdout

	if _, err := os.Stdout.Write(pdfBytes); err != nil {
		fmt.Fprintf(os.Stderr, "write stdout: %v\n", err)
		os.Exit(1)
	}
}
