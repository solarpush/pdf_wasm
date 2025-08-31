# Système de Templating avec Variables

Le système PDF Template supporte maintenant les **variables dynamiques** qui permettent de générer des PDFs personnalisés à partir d'un seul template.

## 🎯 Fonctionnalités

### Variables simples

```json
{
  "type": "text",
  "content": "Hello {{name}}!"
}
```

### Variables imbriquées

```json
{
  "type": "text",
  "content": "Email: {{user.email}}"
}
```

### Variables dans les styles

```json
{
  "style": {
    "color": "{{theme.primary}}",
    "bgColor": "{{theme.secondary}}"
  }
}
```

## 📝 Utilisation

### 1. Créer un template avec variables

**template_invoice.json:**

```json
{
  "page": { "format": "A4", "orientation": "portrait" },
  "fonts": { "default": "DejaVu" },
  "elements": [
    {
      "type": "text",
      "content": "FACTURE N° {{invoice.number}}",
      "style": { "size": 20, "bold": true, "color": "{{theme.primary}}" }
    },
    {
      "type": "text",
      "content": "Client: {{buyer.name}}\nAdresse: {{buyer.address}}"
    }
  ]
}
```

### 2. Définir les variables

**variables.json:**

```json
{
  "invoice": {
    "number": "FAC-2025-001"
  },
  "buyer": {
    "name": "ACME Corp",
    "address": "123 Business St, 75001 Paris"
  },
  "theme": {
    "primary": "#2563EB"
  }
}
```

### 3. Générer le PDF

#### En Go (programmatique)

```go
import "pdf_wasm/internal/template"

// Charger variables depuis fichier
variables := make(map[string]interface{})
data, _ := os.ReadFile("variables.json")
json.Unmarshal(data, &variables)

// Générer PDF avec templating
pdfBytes, err := template.GeneratePDFFromFile("template_invoice.json", variables)

// Ou depuis contenu en mémoire
pdfBytes, err := template.GeneratePDFFromContent(templateContent, variables)
```

#### Avec les outils de test

```bash
# Lancer les tests de templating
make test-templating

# Générer les exemples
make examples-templating
```

## 🔧 API du Système de Templating

### Structures principales

```go
// Processeur de template
type TemplateProcessor struct {
    variables map[string]interface{}
}

// Créer un processeur
processor := template.NewTemplateProcessor(variables)

// Traiter un template
processed, err := processor.ProcessTemplate(templateContent)
```

### Fonctions utilitaires

```go
// Depuis fichier + variables
template.GeneratePDFFromFile(templatePath, variables)

// Depuis contenu + variables
template.GeneratePDFFromContent(templateContent, variables)

// Traiter template seul
template.ProcessTemplateFile(templatePath, variables)
template.ProcessTemplateContent(templateContent, variables)
```

## 📋 Syntaxe des Variables

### Formats supportés

- `{{variable}}` - Variable simple
- `{{object.field}}` - Champ d'objet
- `{{object.nested.field}}` - Imbrication profonde
- Espaces autorisés: `{{ variable }}`, `{{ object.field }}`

### Types de données supportés

- **String**: `"Hello {{name}}"`
- **Nombres**: `{{count}}` → `42`
- **Booléens**: `{{enabled}}` → `true`/`false`
- **Objets**: `{{user.email}}` → `"user@example.com"`

### Échappement automatique

Les caractères spéciaux JSON sont automatiquement échappés :

- `\n` → `\\n`
- `\"` → `\\\"`
- `\\` → `\\\\`

## 📂 Structure des Fichiers

```
├── template_with_variables.json      # Template d'exemple
├── variables_example.json            # Variables d'exemple
├── cmd/
│   └── test_full_templating/         # Tests complets
├── templated_invoice_example.pdf     # Résultat avec variables_example.json
├── templated_invoice_custom.pdf      # Résultat avec variables personnalisées
└── debug_processed_template.json     # Template traité (debug)
```

## 🚀 Exemples Concrets

### Facture Dynamique

Voir `template_with_variables.json` + `variables_example.json`

### Variables Personnalisées

```go
variables := map[string]interface{}{
    "company": "Mon Entreprise",
    "client": map[string]interface{}{
        "name": "Client VIP",
        "discount": 15,
    },
    "items": []map[string]interface{}{
        {"name": "Service A", "price": 100},
        {"name": "Service B", "price": 200},
    },
}
```

### Thèmes Dynamiques

```json
{
  "themes": {
    "corporate": { "primary": "#1F2937", "accent": "#3B82F6" },
    "modern": { "primary": "#DC2626", "accent": "#059669" },
    "elegant": { "primary": "#7C3AED", "accent": "#F59E0B" }
  }
}
```

## ✅ Avantages

1. **Réutilisabilité**: Un template → Plusieurs PDFs
2. **Flexibilité**: Variables dans contenu ET styles
3. **Sécurité**: Échappement automatique des caractères
4. **Performance**: Traitement en mémoire
5. **Simplicité**: Syntaxe `{{variable}}` intuitive
6. **Intégration**: Compatible avec le système existant

## 🔄 Workflow Recommandé

1. **Concevoir** le template avec placeholders `{{...}}`
2. **Tester** avec `make test-templating`
3. **Débugger** avec le fichier `debug_processed_template.json`
4. **Intégrer** dans votre application avec l'API Go
5. **Déployer** en WASM pour usage web/Node.js

---

_Le système de templating étend les capacités du générateur PDF en permettant la création de documents personnalisés à grande échelle._
