package template

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// Embarquer les polices dans le binaire
//
//go:embed fonts/DejaVuSans.ttf
var dejaVuSansFont []byte

//go:embed fonts/DejaVuSans-Bold.ttf
var dejaVuSansBoldFont []byte

//go:embed fonts/DejaVuSans-Oblique.ttf
var dejaVuSansObliqueFont []byte

//go:embed fonts/DejaVuSans-BoldOblique.ttf
var dejaVuSansBoldObliqueFont []byte

// --- Structures du template JSON ---

type PageConfig struct {
	Format      string    `json:"format"`      // "A4", "Letter", etc.
	Orientation string    `json:"orientation"` // "portrait", "landscape"
	Margins     []float64 `json:"margins"`     // [left, top, right, bottom]
}

type FontConfig struct {
	Default string            `json:"default"`
	Paths   map[string]string `json:"paths"`
}

type Style struct {
	Font    string  `json:"font,omitempty"`
	Size    float64 `json:"size,omitempty"`
	Bold    bool    `json:"bold,omitempty"`
	Italic  bool    `json:"italic,omitempty"`
	Color   string  `json:"color,omitempty"`   // hex color
	BgColor string  `json:"bgColor,omitempty"` // hex background color
	Align   string  `json:"align,omitempty"`   // "left", "center", "right"
	Border  string  `json:"border,omitempty"`  // "0", "1", "LTR", etc.
	Fill    bool    `json:"fill,omitempty"`    // remplir la cellule
	Width   float64 `json:"width,omitempty"`   // largeur spécifique
	Height  float64 `json:"height,omitempty"`  // hauteur spécifique

	// Espacement
	Margin  []float64 `json:"margin,omitempty"`  // [top, right, bottom, left] ou [vertical, horizontal] ou [all]
	Padding []float64 `json:"padding,omitempty"` // [top, right, bottom, left] ou [vertical, horizontal] ou [all]
}

type Element struct {
	Type     string      `json:"type"`               // "text", "table", "grid", "space", "line", "image"
	Content  interface{} `json:"content,omitempty"`  // contenu variable selon le type
	Style    *Style      `json:"style,omitempty"`    // style pour cet élément
	Children []Element   `json:"children,omitempty"` // pour les grilles

	// Spécifique aux tableaux
	Columns []TableColumn `json:"columns,omitempty"`
	Rows    interface{}   `json:"rows,omitempty"` // Peut être []TableRow ou string (pour templates avec boucles)

	// Spécifique aux grilles
	GridColumns int `json:"gridColumns,omitempty"`

	// Spécifique aux lignes
	Length float64 `json:"length,omitempty"`
}

type TableColumn struct {
	Header string  `json:"header"`
	Width  float64 `json:"width"`
	Align  string  `json:"align,omitempty"`
}

type TableRow struct {
	Cells []string `json:"cells"`
	Style *Style   `json:"style,omitempty"`
}

type Template struct {
	Page     PageConfig `json:"page"`
	Fonts    FontConfig `json:"fonts"`
	Elements []Element  `json:"elements"`
}

// --- Wrapper PDF ---

type PDFBuilder struct {
	pdf     *gofpdf.Fpdf
	config  Template
	margins struct{ left, top, right, bottom float64 }
}

func NewPDFBuilder(template Template) *PDFBuilder {
	// Configuration par défaut
	if template.Page.Format == "" {
		template.Page.Format = "A4"
	}
	if template.Page.Orientation == "" {
		template.Page.Orientation = "portrait"
	}
	if len(template.Page.Margins) != 4 {
		template.Page.Margins = []float64{15, 12, 15, 20}
	}
	if template.Fonts.Default == "" {
		template.Fonts.Default = "Arial"
	}

	// Créer le PDF
	orientation := "P"
	if template.Page.Orientation == "landscape" {
		orientation = "L"
	}

	pdf := gofpdf.New(orientation, "mm", template.Page.Format, "")

	builder := &PDFBuilder{
		pdf:    pdf,
		config: template,
	}

	// Marges
	builder.margins.left = template.Page.Margins[0]
	builder.margins.top = template.Page.Margins[1]
	builder.margins.right = template.Page.Margins[2]
	builder.margins.bottom = template.Page.Margins[3]

	pdf.SetMargins(builder.margins.left, builder.margins.top, builder.margins.right)
	pdf.SetAutoPageBreak(true, builder.margins.bottom)

	return builder
}

func (b *PDFBuilder) setupFonts() {
	// Utiliser les polices embarquées
	embeddedFonts := map[string]struct {
		data  []byte
		style string
	}{
		"DejaVu":    {dejaVuSansFont, ""},
		"DejaVu-B":  {dejaVuSansBoldFont, "B"},
		"DejaVu-I":  {dejaVuSansObliqueFont, "I"},
		"DejaVu-BI": {dejaVuSansBoldObliqueFont, "BI"},
	}

	// Ajouter les polices embarquées
	for name, fontInfo := range embeddedFonts {
		fontName := strings.Split(name, "-")[0]

		// Utiliser AddUTF8FontFromBytes pour les polices embarquées
		b.pdf.AddUTF8FontFromBytes(fontName, fontInfo.style, fontInfo.data)
	}

	// Ajouter les polices personnalisées (si définies)
	for name, path := range b.config.Fonts.Paths {
		// Essayer de charger depuis le fichier pour les polices personnalisées
		if _, err := os.Stat(path); err == nil {
			b.pdf.AddUTF8Font(name, "", path)
		} else {
			fmt.Printf("Warning: Custom font %s not found at %s\n", name, path)
		}
	}
}

func (b *PDFBuilder) applyStyle(style *Style) {
	if style == nil {
		b.pdf.SetFont(b.config.Fonts.Default, "", 10)
		b.pdf.SetTextColor(0, 0, 0)
		return
	}

	// Police
	font := b.config.Fonts.Default
	if style.Font != "" {
		font = style.Font
	}

	fontStyle := ""
	if style.Bold {
		fontStyle += "B"
	}
	if style.Italic {
		fontStyle += "I"
	}

	size := 10.0
	if style.Size > 0 {
		size = style.Size
	}

	b.pdf.SetFont(font, fontStyle, size)

	// Couleur du texte
	if style.Color != "" {
		r, g, blue := hexToRGB(style.Color)
		b.pdf.SetTextColor(r, g, blue)
	} else {
		b.pdf.SetTextColor(0, 0, 0)
	}

	// Couleur de fond
	if style.BgColor != "" {
		r, g, blue := hexToRGB(style.BgColor)
		b.pdf.SetFillColor(r, g, blue)
	}
}

func (b *PDFBuilder) renderElement(element Element) {
	// Appliquer les marges universelles pour tous les éléments
	b.applyMargin(element.Style)

	switch element.Type {
	case "text":
		b.renderText(element)
	case "table":
		b.renderTable(element)
	case "grid":
		b.renderGrid(element)
	case "space":
		b.renderSpace(element)
	case "line":
		b.renderLine(element)
	case "image":
		b.renderImage(element)
	}

	// Appliquer la marge du bas pour tous les éléments
	if element.Style != nil && len(element.Style.Margin) > 0 {
		margin := parseSpacing(element.Style.Margin)
		if margin.Bottom > 0 {
			b.pdf.Ln(margin.Bottom)
		}
	}
}

func (b *PDFBuilder) renderText(element Element) {
	b.applyStyle(element.Style)

	content := ""
	if str, ok := element.Content.(string); ok {
		content = str
	}

	align := "L"
	if element.Style != nil && element.Style.Align != "" {
		switch element.Style.Align {
		case "center":
			align = "C"
		case "right":
			align = "R"
		}
	}

	height := 8.0
	if element.Style != nil && element.Style.Height > 0 {
		height = element.Style.Height
	}

	fill := false
	if element.Style != nil && element.Style.Fill {
		fill = true
	}

	// Calculer la largeur effective avec marges et padding
	pageWidth, _ := b.pdf.GetPageSize()
	availableWidth := pageWidth - b.margins.left - b.margins.right
	contentWidth, leftOffset := b.getContentArea(availableWidth, element.Style)

	// Ajuster la position si nécessaire
	if leftOffset > 0 {
		currentX := b.pdf.GetX()
		b.pdf.SetX(currentX + leftOffset)
	}

	// Si le contenu contient des retours à la ligne, utiliser MultiCell
	if strings.Contains(content, "\n") {
		border := ""
		if element.Style != nil && element.Style.Border != "" {
			border = element.Style.Border
		}
		b.pdf.MultiCell(contentWidth, height, content, border, align, fill)
	} else {
		b.pdf.CellFormat(contentWidth, height, content, "", 1, align, fill, 0, "")
	}
}

func (b *PDFBuilder) renderTable(element Element) {
	if len(element.Columns) == 0 {
		return
	}

	// Calculer l'alignement du tableau
	tableAlign := "L"
	if element.Style != nil && element.Style.Align != "" {
		switch element.Style.Align {
		case "center":
			tableAlign = "C"
		case "right":
			tableAlign = "R"
		}
	}

	// Calculer la largeur totale du tableau
	totalWidth := 0.0
	for _, col := range element.Columns {
		totalWidth += col.Width
	}

	// Calculer la position X selon l'alignement
	pageWidth, _ := b.pdf.GetPageSize()
	availableWidth := pageWidth - b.margins.left - b.margins.right

	var startX float64
	switch tableAlign {
	case "C":
		startX = b.margins.left + (availableWidth-totalWidth)/2
	case "R":
		startX = b.margins.left + (availableWidth - totalWidth)
	default:
		startX = b.margins.left
	}

	// Sauvegarder la position actuelle et se placer pour le tableau
	currentY := b.pdf.GetY()
	b.pdf.SetXY(startX, currentY)

	// Gérer les différents types de Rows
	var rows []TableRow
	switch r := element.Rows.(type) {
	case []TableRow:
		rows = r
	case []interface{}:
		// Conversion depuis l'interface générique (venant du JSON)
		for _, item := range r {
			if rowMap, ok := item.(map[string]interface{}); ok {
				row := TableRow{}
				if cells, exists := rowMap["cells"]; exists {
					if cellsSlice, ok := cells.([]interface{}); ok {
						for _, cell := range cellsSlice {
							if cellStr, ok := cell.(string); ok {
								row.Cells = append(row.Cells, cellStr)
							}
						}
					}
				}

				// Gérer le style de la ligne
				if styleData, exists := rowMap["style"]; exists {
					if styleMap, ok := styleData.(map[string]interface{}); ok {
						style := &Style{}

						if bold, exists := styleMap["bold"]; exists {
							if b, ok := bold.(bool); ok {
								style.Bold = b
							}
						}
						if size, exists := styleMap["size"]; exists {
							if s, ok := size.(float64); ok {
								style.Size = s
							}
						}
						if color, exists := styleMap["color"]; exists {
							if c, ok := color.(string); ok {
								style.Color = c
							}
						}
						if bgColor, exists := styleMap["bgColor"]; exists {
							if bg, ok := bgColor.(string); ok {
								style.BgColor = bg
							}
						}

						row.Style = style
					}
				}

				rows = append(rows, row)
			}
		}
	case string:
		// Template string avec boucles - ne rien faire ici car c'est géré par le template processor
		return
	default:
		return
	}

	if len(rows) == 0 {
		return
	}

	// En-têtes
	if element.Style != nil {
		b.applyStyle(element.Style)
	}

	for _, col := range element.Columns {
		align := "L"
		if col.Align == "center" {
			align = "C"
		} else if col.Align == "right" {
			align = "R"
		}

		fill := element.Style != nil && element.Style.Fill
		border := "1"
		if element.Style != nil && element.Style.Border != "" {
			border = element.Style.Border
		}

		b.pdf.CellFormat(col.Width, 8, col.Header, border, 0, align, fill, 0, "")
	}

	// Nouvelle ligne en gardant la position X
	currentY = b.pdf.GetY()
	b.pdf.SetXY(startX, currentY+8) // Lignes de données
	for _, row := range rows {
		if row.Style != nil {
			b.applyStyle(row.Style)
		} else {
			b.pdf.SetFont(b.config.Fonts.Default, "", 10)
			b.pdf.SetTextColor(0, 0, 0)
		}

		for i, cell := range row.Cells {
			if i >= len(element.Columns) {
				break
			}

			col := element.Columns[i]
			align := "L"
			if col.Align == "center" {
				align = "C"
			} else if col.Align == "right" {
				align = "R"
			}

			b.pdf.CellFormat(col.Width, 8, cell, "1", 0, align, false, 0, "")
		}

		// Nouvelle ligne en gardant la position X
		currentY = b.pdf.GetY()
		b.pdf.SetXY(startX, currentY+8)
	}
}

func (b *PDFBuilder) renderGrid(element Element) {
	if element.GridColumns <= 0 || len(element.Children) == 0 {
		return
	}

	// Calculer la largeur de chaque colonne
	pageWidth, _ := b.pdf.GetPageSize()
	availableWidth := pageWidth - b.margins.left - b.margins.right
	columnWidth := availableWidth / float64(element.GridColumns)

	// Traiter les éléments par ligne complète
	for i := 0; i < len(element.Children); i += element.GridColumns {
		startY := b.pdf.GetY()
		maxHeight := 0.0

		// Premier passage : calculer la hauteur maximale de la ligne
		for col := 0; col < element.GridColumns && i+col < len(element.Children); col++ {
			x := b.margins.left + float64(col)*columnWidth

			// Simuler le rendu pour calculer la hauteur
			b.pdf.SetXY(x, startY)
			beforeY := b.pdf.GetY()

			// Sauvegarder l'état avant le rendu temporaire
			savedX, savedY := b.pdf.GetXY()

			// Rendre l'élément temporairement pour mesurer
			b.renderElementInWidth(element.Children[i+col], columnWidth-2) // -2 pour marge

			afterY := b.pdf.GetY()
			height := afterY - beforeY
			if height > maxHeight {
				maxHeight = height
			}

			// Restaurer la position
			b.pdf.SetXY(savedX, savedY)
		}

		// Deuxième passage : rendre tous les éléments de la ligne avec la même hauteur
		for col := 0; col < element.GridColumns && i+col < len(element.Children); col++ {
			x := b.margins.left + float64(col)*columnWidth
			b.pdf.SetXY(x, startY)

			// Rendre l'élément dans sa colonne
			b.renderElementInWidth(element.Children[i+col], columnWidth-2)
		}

		// Avancer à la ligne suivante
		b.pdf.SetXY(b.margins.left, startY+maxHeight+2)
	}
}

// Nouvelle fonction pour rendre un élément dans une largeur spécifique
func (b *PDFBuilder) renderElementInWidth(element Element, maxWidth float64) {
	switch element.Type {
	case "text":
		b.renderTextInWidth(element, maxWidth)
	default:
		b.renderElement(element) // Fallback pour les autres types
	}
}

func (b *PDFBuilder) renderTextInWidth(element Element, maxWidth float64) {
	b.applyStyle(element.Style)

	content := ""
	if str, ok := element.Content.(string); ok {
		content = str
	}

	align := "L"
	if element.Style != nil && element.Style.Align != "" {
		switch element.Style.Align {
		case "center":
			align = "C"
		case "right":
			align = "R"
		}
	}

	height := 5.0
	if element.Style != nil && element.Style.Height > 0 {
		height = element.Style.Height
	}

	fill := false
	if element.Style != nil && element.Style.Fill {
		fill = true
	}

	border := ""
	if element.Style != nil && element.Style.Border != "" {
		border = element.Style.Border
	}

	// Si le contenu contient des retours à la ligne, utiliser MultiCell
	if strings.Contains(content, "\n") {
		b.pdf.MultiCell(maxWidth, height, content, border, align, fill)
	} else {
		b.pdf.CellFormat(maxWidth, height*1.5, content, border, 1, align, fill, 0, "")
	}
}

func (b *PDFBuilder) renderSpace(element Element) {
	height := 5.0
	if element.Style != nil && element.Style.Height > 0 {
		height = element.Style.Height
	}
	b.pdf.Ln(height)
}

func (b *PDFBuilder) renderLine(element Element) {
	length := 0.0 // 0 = toute la largeur
	if element.Length > 0 {
		length = element.Length
	}

	thickness := 0.1
	if element.Style != nil && element.Style.Height > 0 {
		thickness = element.Style.Height
	}

	// Couleur de la ligne
	if element.Style != nil && element.Style.Color != "" {
		r, g, blue := hexToRGB(element.Style.Color)
		b.pdf.SetDrawColor(r, g, blue)
	}

	b.pdf.SetLineWidth(thickness)

	currentX := b.pdf.GetX()
	currentY := b.pdf.GetY()

	endX := currentX + length
	if length == 0 {
		pageWidth, _ := b.pdf.GetPageSize()
		endX = pageWidth - b.margins.right
	}

	b.pdf.Line(currentX, currentY, endX, currentY)
	b.pdf.Ln(2)

	// Remettre la couleur par défaut
	b.pdf.SetDrawColor(0, 0, 0)
}

func (b *PDFBuilder) renderImage(element Element) {
	// Pour l'instant, on ne gère que les images base64
	if str, ok := element.Content.(string); ok && strings.HasPrefix(str, "data:image/") {
		// Extraire les données base64
		parts := strings.Split(str, ",")
		if len(parts) == 2 {
			imageData, err := base64.StdEncoding.DecodeString(parts[1])
			if err == nil {
				// Créer un nom temporaire
				imageName := fmt.Sprintf("temp_image_%d", len(imageData))

				// Type d'image
				imageType := "PNG"
				if strings.Contains(parts[0], "jpeg") || strings.Contains(parts[0], "jpg") {
					imageType = "JPG"
				}

				// Enregistrer temporairement l'image
				b.pdf.RegisterImageReader(imageName, imageType, bytes.NewReader(imageData))

				// Dimensions
				width := 50.0
				height := 0.0 // auto
				if element.Style != nil {
					if element.Style.Width > 0 {
						width = element.Style.Width
					}
					if element.Style.Height > 0 {
						height = element.Style.Height
					}
				}

				b.pdf.ImageOptions(imageName, b.pdf.GetX(), b.pdf.GetY(), width, height, false, gofpdf.ImageOptions{}, 0, "")

				if height == 0 {
					height = width * 0.75 // ratio par défaut
				}
				b.pdf.Ln(height + 2)
			}
		}
	}
}

func (b *PDFBuilder) Build() ([]byte, error) {
	b.pdf.AddPage()
	b.setupFonts()

	// Rendre tous les éléments
	for _, element := range b.config.Elements {
		b.renderElement(element)
	}

	var buf bytes.Buffer
	if err := b.pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// --- Utilitaires pour marges et padding ---

type Spacing struct {
	Top, Right, Bottom, Left float64
}

func parseSpacing(values []float64) Spacing {
	if len(values) == 0 {
		return Spacing{0, 0, 0, 0}
	}
	if len(values) == 1 {
		// [all]
		return Spacing{values[0], values[0], values[0], values[0]}
	}
	if len(values) == 2 {
		// [vertical, horizontal]
		return Spacing{values[0], values[1], values[0], values[1]}
	}
	if len(values) >= 4 {
		// [top, right, bottom, left]
		return Spacing{values[0], values[1], values[2], values[3]}
	}
	// [top, horizontal, bottom]
	return Spacing{values[0], values[1], values[2], values[1]}
}

func (b *PDFBuilder) applyMargin(style *Style) {
	if style == nil || len(style.Margin) == 0 {
		return
	}
	margin := parseSpacing(style.Margin)

	// Appliquer la marge du haut
	if margin.Top > 0 {
		b.pdf.Ln(margin.Top)
	}

	// Les marges gauche/droite seront gérées lors du rendu des éléments
}

func (b *PDFBuilder) getEffectiveWidth(baseWidth float64, style *Style) float64 {
	if style == nil || len(style.Margin) == 0 {
		return baseWidth
	}

	margin := parseSpacing(style.Margin)
	return baseWidth - margin.Left - margin.Right
}

func (b *PDFBuilder) getContentArea(baseWidth float64, style *Style) (width float64, paddingLeft float64) {
	width = baseWidth
	paddingLeft = 0

	if style != nil {
		if len(style.Margin) > 0 {
			margin := parseSpacing(style.Margin)
			width -= margin.Left + margin.Right
			paddingLeft += margin.Left
		}

		if len(style.Padding) > 0 {
			padding := parseSpacing(style.Padding)
			width -= padding.Left + padding.Right
			paddingLeft += padding.Left
		}
	}

	return width, paddingLeft
}

// --- Utilitaires ---

func hexToRGB(hex string) (r, g, b int) {
	if hex == "" {
		return 0, 0, 0
	}
	h := strings.TrimPrefix(hex, "#")
	if len(h) == 3 {
		fmt.Sscanf(h, "%1x%1x%1x", &r, &g, &b)
		return r * 17, g * 17, b * 17
	}
	fmt.Sscanf(h, "%02x%02x%02x", &r, &g, &b)
	return
}

// --- Système de templating avec variables ---

// TemplateProcessor gère le remplacement des variables dans les templates
type TemplateProcessor struct {
	variables map[string]interface{}
}

// NewTemplateProcessor crée un nouveau processeur de template
func NewTemplateProcessor(variables map[string]interface{}) *TemplateProcessor {
	if variables == nil {
		variables = make(map[string]interface{})
	}
	return &TemplateProcessor{
		variables: variables,
	}
}

// ProcessTemplate traite un template JSON en remplaçant les variables et les boucles
func (tp *TemplateProcessor) ProcessTemplate(templateContent []byte) ([]byte, error) {
	content := string(templateContent)

	// 1. Traiter d'abord les boucles {{#array}}...{{/array}}
	content = tp.processLoops(content)

	// 2. Ensuite traiter les variables simples {{variable}}
	re := regexp.MustCompile(`\{\{\s*([^}]+)\s*\}\}`)
	processed := re.ReplaceAllStringFunc(content, func(match string) string {
		// Extraire le nom de la variable (enlever {{ et }})
		varName := strings.TrimSpace(strings.Trim(match, "{}"))

		// Skip les variables de boucle qui commencent par #
		if strings.HasPrefix(varName, "#") || strings.HasPrefix(varName, "/") {
			return match
		}

		// Gérer les champs imbriqués (ex: user.name)
		value := tp.getNestedValue(varName)

		// Convertir en string
		return tp.valueToString(value)
	})

	return []byte(processed), nil
}

// processLoops traite les boucles {{#array}}...{{/array}}
func (tp *TemplateProcessor) processLoops(content string) string {
	// Regex pour matcher {{#variable}}...{{/variable}}
	// Utilisation d'une approche différente car Go ne supporte pas les backreferences
	loopStartRe := regexp.MustCompile(`\{\{\s*#\s*([^}]+)\s*\}\}`)

	// Trouver toutes les boucles
	result := content

	for {
		// Trouver le début d'une boucle
		startMatch := loopStartRe.FindStringSubmatchIndex(result)
		if startMatch == nil {
			break
		}

		arrayName := result[startMatch[2]:startMatch[3]]

		// Trouver la fin correspondante
		endPattern := fmt.Sprintf(`\{\{\s*/\s*%s\s*\}\}`, regexp.QuoteMeta(arrayName))
		endRe := regexp.MustCompile(endPattern)
		endMatch := endRe.FindStringIndex(result[startMatch[1]:])

		if endMatch == nil {
			// Pas de fermeture trouvée, passer au suivant
			result = result[startMatch[1]:]
			continue
		}

		// Ajuster les positions
		endMatch[0] += startMatch[1]
		endMatch[1] += startMatch[1]

		// Extraire le contenu de la boucle
		loopContent := result[startMatch[1]:endMatch[0]]

		// Traiter la boucle
		processedLoop := tp.processLoop(arrayName, loopContent)

		// Remplacer dans le résultat
		result = result[:startMatch[0]] + processedLoop + result[endMatch[1]:]
	}

	return result
}

// processLoop traite une boucle individuelle
func (tp *TemplateProcessor) processLoop(arrayName, loopContent string) string {
	// Récupérer le tableau
	arrayValue := tp.getNestedValue(arrayName)

	switch arr := arrayValue.(type) {
	case []interface{}:
		var result strings.Builder

		for i, item := range arr {
			// Créer un processeur temporaire avec l'item courant + variables originales
			tempVars := make(map[string]interface{})

			// Copier les variables originales
			for k, v := range tp.variables {
				tempVars[k] = v
			}

			// Ajouter les variables de boucle
			switch itemMap := item.(type) {
			case map[string]interface{}:
				// Fusionner les propriétés de l'item
				for k, v := range itemMap {
					tempVars[k] = v
				}
			default:
				// Si l'item n'est pas un objet, l'assigner à "item"
				tempVars["item"] = item
			}

			// Ajouter l'index
			tempVars["index"] = i
			tempVars["index1"] = i + 1

			// Traiter le contenu de la boucle avec les nouvelles variables
			tempProcessor := NewTemplateProcessor(tempVars)

			// Traiter les variables dans le contenu de la boucle
			re := regexp.MustCompile(`\{\{\s*([^}#/]+)\s*\}\}`)
			processedLoopContent := re.ReplaceAllStringFunc(loopContent, func(varMatch string) string {
				varName := strings.TrimSpace(strings.Trim(varMatch, "{}"))
				value := tempProcessor.getNestedValue(varName)
				return tempProcessor.valueToString(value)
			})

			result.WriteString(processedLoopContent)

			// Ajouter une virgule entre les éléments si nécessaire
			if i < len(arr)-1 && strings.TrimSpace(loopContent) != "" {
				result.WriteString(",")
			}
		}

		return result.String()

	case []map[string]interface{}:
		var result strings.Builder

		for i, item := range arr {
			// Créer un processeur temporaire avec l'item courant + variables originales
			tempVars := make(map[string]interface{})

			// Copier les variables originales
			for k, v := range tp.variables {
				tempVars[k] = v
			}

			// Fusionner les propriétés de l'item
			for k, v := range item {
				tempVars[k] = v
			}

			// Ajouter l'index
			tempVars["index"] = i
			tempVars["index1"] = i + 1

			// Traiter le contenu de la boucle avec les nouvelles variables
			tempProcessor := NewTemplateProcessor(tempVars)

			// Traiter les variables dans le contenu de la boucle
			re := regexp.MustCompile(`\{\{\s*([^}#/]+)\s*\}\}`)
			processedLoopContent := re.ReplaceAllStringFunc(loopContent, func(varMatch string) string {
				varName := strings.TrimSpace(strings.Trim(varMatch, "{}"))
				value := tempProcessor.getNestedValue(varName)
				return tempProcessor.valueToString(value)
			})

			result.WriteString(processedLoopContent)

			// Ajouter une virgule entre les éléments si nécessaire
			if i < len(arr)-1 && strings.TrimSpace(loopContent) != "" {
				result.WriteString(",")
			}
		}

		return result.String()
	}

	// Si ce n'est pas un tableau, retourner vide
	return ""
}

// getNestedValue récupère une valeur potentiellement imbriquée
func (tp *TemplateProcessor) getNestedValue(path string) interface{} {
	parts := strings.Split(path, ".")

	var current interface{} = tp.variables

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, exists := v[part]; exists {
				current = val
			} else {
				return ""
			}
		default:
			return ""
		}
	}

	return current
}

// valueToString convertit une valeur en string pour l'insertion dans le JSON
func (tp *TemplateProcessor) valueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Échapper les caractères spéciaux JSON
		escaped := strings.ReplaceAll(v, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", "\\n")
		escaped = strings.ReplaceAll(escaped, "\r", "\\r")
		escaped = strings.ReplaceAll(escaped, "\t", "\\t")
		return escaped
	case int, int64, int32:
		return fmt.Sprintf("%d", v)
	case float64, float32:
		return fmt.Sprintf("%.2f", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return ""
	default:
		// Pour les autres types, convertir en string et échapper
		str := fmt.Sprintf("%v", v)
		escaped := strings.ReplaceAll(str, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", "\\n")
		escaped = strings.ReplaceAll(escaped, "\r", "\\r")
		escaped = strings.ReplaceAll(escaped, "\t", "\\t")
		return escaped
	}
}

// --- Fonctions utilitaires pour charger des templates ---

// LoadTemplateFromFile charge un template depuis un fichier
func LoadTemplateFromFile(templatePath string) ([]byte, error) {
	return os.ReadFile(templatePath)
}

// LoadTemplateFromReader charge un template depuis un Reader
func LoadTemplateFromReader(reader io.Reader) ([]byte, error) {
	return io.ReadAll(reader)
}

// ProcessTemplateFile traite un template depuis un fichier avec des variables
func ProcessTemplateFile(templatePath string, variables map[string]interface{}) (Template, error) {
	// Charger le template
	templateContent, err := LoadTemplateFromFile(templatePath)
	if err != nil {
		return Template{}, fmt.Errorf("failed to load template: %w", err)
	}

	// Traiter les variables
	processor := NewTemplateProcessor(variables)
	processedContent, err := processor.ProcessTemplate(templateContent)
	if err != nil {
		return Template{}, fmt.Errorf("failed to process template: %w", err)
	}

	// Parser le JSON final
	var template Template
	if err := json.Unmarshal(processedContent, &template); err != nil {
		return Template{}, fmt.Errorf("failed to parse processed template: %w", err)
	}

	return template, nil
}

// ProcessTemplateContent traite un contenu de template avec des variables
func ProcessTemplateContent(templateContent []byte, variables map[string]interface{}) (Template, error) {
	// Traiter les variables
	processor := NewTemplateProcessor(variables)
	processedContent, err := processor.ProcessTemplate(templateContent)
	if err != nil {
		return Template{}, fmt.Errorf("failed to process template: %w", err)
	}

	// Parser le JSON final
	var template Template
	if err := json.Unmarshal(processedContent, &template); err != nil {
		return Template{}, fmt.Errorf("failed to parse processed template: %w", err)
	}

	return template, nil
}

// --- Fonction principale pour générer un PDF depuis un template avec variables ---

// GeneratePDFFromFile génère un PDF depuis un fichier template avec des variables
func GeneratePDFFromFile(templatePath string, variables map[string]interface{}) ([]byte, error) {
	template, err := ProcessTemplateFile(templatePath, variables)
	if err != nil {
		return nil, err
	}

	return GeneratePDF(template)
}

// GeneratePDFFromContent génère un PDF depuis un contenu template avec des variables
func GeneratePDFFromContent(templateContent []byte, variables map[string]interface{}) ([]byte, error) {
	template, err := ProcessTemplateContent(templateContent, variables)
	if err != nil {
		return nil, err
	}

	return GeneratePDF(template)
}

// Fonction principale pour générer un PDF depuis un template JSON
func GeneratePDF(template Template) ([]byte, error) {
	builder := NewPDFBuilder(template)
	return builder.Build()
}
