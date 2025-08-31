# PDF Template Generator with Dynamic Loops

Ce projet fournit un système de génération de PDF basé sur des templates JSON avec support des **boucles dynamiques**, conçu pour être compilé en WebAssembly (WASM) et utilisé depuis Node.js.

## 🚀 Fonctionnalités

- **Templates JSON** : Configuration flexible des PDFs
- **Boucles dynamiques** : Templating avec syntaxe `{{#array}}...{{/array}}`
- **Variables contextuelles** : Support de `{{variable}}` et `{{object.field}}`
- **Système de grille** : Positionnement précis des éléments  
- **Support UTF-8** : Polices DejaVu intégrées
- **Styles avancés** : Couleurs, marges, padding, bordures
- **WASM Ready** : Compilation pour Node.js

## 🎯 Boucles Dynamiques

### Template avec boucles

```json
{
  "elements": [
    {
      "type": "table",
      "rows": [
        {{#items}}
        {
          "cells": [
            "{{description}}",
            "{{quantity}}",
            "{{price}} {{currency}}"
          ]
        }
        {{/items}}
      ]
    }
  ]
}
```

### Variables avec tableau

```json
{
  "items": [
    {"description": "Item 1", "quantity": "2", "price": "100.00"},
    {"description": "Item 2", "quantity": "1", "price": "200.00"},
    {"description": "Item 3", "quantity": "3", "price": "50.00"}
  ],
  "currency": "€"
}
```

### Résultat généré

Le template génère automatiquement une ligne pour chaque item du tableau, sans limitation de nombre.

## 📋 Utilisation

### Génération simple

```bash
echo '{"page":{"format":"A4"},"elements":[{"type":"text","content":"Hello World"}]}' | ./pdf-template > output.pdf
```

### Avec boucles dynamiques (API Go)

```go
import "pdf_wasm/internal/template"

variables := map[string]interface{}{
    "items": []map[string]interface{}{
        {"name": "Product A", "price": 100},
        {"name": "Product B", "price": 200},
    },
    "currency": "€",
}

pdfBytes, err := template.GeneratePDFFromFile("template_dynamic.json", variables)
```

## 🛠️ Développement

### Build et test

```bash
make all          # Build + tests
make test-dynamic # Test des boucles dynamiques
make examples     # Générer les exemples
```

### Compilation WASM

```bash
make build-wasm
```

## 📖 Documentation

- [Guide du système de templating](README_TEMPLATING.md) - Documentation complète

## 🔧 Structure du Projet

```
├── main.go                      # Point d'entrée principal
├── internal/template/           # Moteur de templating avec boucles
├── template_dynamic.json        # Template d'exemple avec {{#items}}
├── variables_dynamic.json       # Variables d'exemple avec tableau
├── cmd/test_dynamic_loops/      # Tests des boucles dynamiques
└── output/                      # PDFs générés
```

## ⚡ Quick Start

1. **Cloner et builder**

   ```bash
   git clone <repo>
   cd goPdf
   make all
   ```

2. **Tester les boucles dynamiques**

   ```bash
   make test-dynamic
   ```

3. **Voir les résultats**

   - `output/dynamic_invoice.pdf` - PDF avec 4 items
   - `output/dynamic_invoice_extended.pdf` - PDF avec 6 items

4. **Intégrer dans votre code**

   ```go
   // Template avec boucles
   pdfBytes, err := template.GeneratePDFFromFile("template.json", variables)
   ```

## ✨ Avantages des Boucles Dynamiques

- **🔄 Nombre illimité** : Autant d'items que nécessaire
- **📝 Template unique** : Un seul template pour tous les cas
- **🛠️ Maintenance facile** : Pas de duplication de code
- **🎯 Syntaxe claire** : `{{#array}}...{{/array}}`
- **📊 Variables contextuelles** : `{{index}}`, `{{index1}}`

Le système remplace avantageusement les templates statiques avec des variables fixes (`{{item1}}`, `{{item2}}`, etc.).
