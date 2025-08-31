package main

import (
	"encoding/json"
	"fmt"
	"os"
	"pdf_wasm/internal/template"
)

func main() {
	fmt.Println("🎯 Test du support YAML pour les templates PDF...")

	// 1. Charger les variables
	variables := make(map[string]interface{})
	variablesData, err := os.ReadFile("../../variables_dynamic.json")
	if err != nil {
		fmt.Printf("❌ Erreur lors du chargement des variables: %v\n", err)
		return
	}

	if err := json.Unmarshal(variablesData, &variables); err != nil {
		fmt.Printf("❌ Erreur lors du parsing des variables: %v\n", err)
		return
	}

	// 2. Test de conversion JSON vers YAML
	fmt.Println("📄 Conversion du template JSON en YAML...")
	jsonContent, err := os.ReadFile("../../template_dynamic.json")
	if err != nil {
		fmt.Printf("❌ Erreur lors du chargement du JSON: %v\n", err)
		return
	}

	yamlContent, err := template.ConvertJSONToYAML(jsonContent)
	if err != nil {
		fmt.Printf("❌ Erreur lors de la conversion JSON->YAML: %v\n", err)
		return
	}

	if err := os.WriteFile("../../template_converted.pdftpl", yamlContent, 0644); err != nil {
		fmt.Printf("❌ Erreur lors de la sauvegarde YAML: %v\n", err)
		return
	}

	fmt.Println("✅ Template converti en YAML: template_converted.pdftpl")

	// 3. Test de génération PDF depuis YAML
	fmt.Println("📄 Génération PDF depuis template YAML...")
	pdfBytes, err := template.GeneratePDFFromYAMLFile("../../template_simple.pdftpl", variables)
	if err != nil {
		fmt.Printf("❌ Erreur lors de la génération PDF depuis YAML: %v\n", err)
		return
	}

	pdfPath := "../../yaml_invoice.pdf"
	if err := os.WriteFile(pdfPath, pdfBytes, 0644); err != nil {
		fmt.Printf("❌ Erreur lors de la sauvegarde PDF: %v\n", err)
		return
	}

	fmt.Printf("✅ PDF généré depuis YAML: %s\n", pdfPath)

	// 4. Test de détection de format
	fmt.Println("🔍 Test de détection de format...")
	
	formats := map[string]string{
		"../../template_dynamic.json":   template.DetectTemplateFormat("../../template_dynamic.json"),
		"../../template_simple.pdftpl": template.DetectTemplateFormat("../../template_simple.pdftpl"),
		"../../template_converted.pdftpl": template.DetectTemplateFormat("../../template_converted.pdftpl"),
	}

	for file, format := range formats {
		fmt.Printf("   📁 %s → %s\n", file, format)
	}

	// 5. Test de conversion YAML vers JSON
	fmt.Println("📄 Test conversion YAML vers JSON...")
	yamlContent2, err := os.ReadFile("../../template_simple.pdftpl")
	if err != nil {
		fmt.Printf("❌ Erreur lors du chargement YAML: %v\n", err)
		return
	}

	jsonContent2, err := template.ConvertYAMLToJSON(yamlContent2)
	if err != nil {
		fmt.Printf("❌ Erreur lors de la conversion YAML->JSON: %v\n", err)
		return
	}

	if err := os.WriteFile("../../template_yaml_to_json.json", jsonContent2, 0644); err != nil {
		fmt.Printf("❌ Erreur lors de la sauvegarde JSON: %v\n", err)
		return
	}

	fmt.Println("✅ Template converti en JSON: template_yaml_to_json.json")

	fmt.Println("\n🎉 Tests YAML terminés avec succès!")
	fmt.Println("📋 Fichiers générés:")
	fmt.Printf("   - %s (PDF depuis YAML)\n", pdfPath)
	fmt.Println("   - ../../template_converted.pdftpl (JSON→YAML)")
	fmt.Println("   - ../../template_yaml_to_json.json (YAML→JSON)")
	
	fmt.Println("\n✨ Fonctionnalités YAML disponibles:")
	fmt.Println("   • Autocomplétion VS Code avec .pdftpl")
	fmt.Println("   • Validation de schéma en temps réel")
	fmt.Println("   • Conversion bidirectionnelle JSON ↔ YAML")
	fmt.Println("   • Détection automatique de format")
	fmt.Println("   • Support complet des variables et boucles")
}
