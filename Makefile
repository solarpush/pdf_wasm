# PDF Template System Makefile

# Variables
BINARY_NAME=pdf-template
WASM_NAME=pdf-template.wasm
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")

# Default target
.PHONY: all
all: build test

# Build targets
.PHONY: build
build: clean $(BINARY_NAME)
	@echo "✅ Build complete!"

.PHONY: build-wasm
build-wasm: clean-wasm $(WASM_NAME)
	@echo "✅ WASM build complete!"

# Individual builds
$(BINARY_NAME): $(GO_FILES)
	@echo "📦 Building template system binary..."
	go build -o $(BINARY_NAME) main.go

$(WASM_NAME): $(GO_FILES)
	@echo "📦 Building WASM module..."
	GOOS=js GOARCH=wasm go build -o $(WASM_NAME) main.go

# Test targets
.PHONY: test
test: build test-basic test-combined test-dynamic test-yaml
	@echo "✅ All tests passed successfully!"

.PHONY: test-basic
test-basic: $(BINARY_NAME)
	@echo "🧪 Testing basic template system..."
	@mkdir -p output
	@echo '{"page":{"format":"A4"},"elements":[{"type":"text","content":"Hello World","style":{"size":16,"align":"center"}}]}' | ./$(BINARY_NAME) > output/test_basic.pdf
	@echo "   ✅ Basic template test passed"

.PHONY: test-combined
test-combined: $(BINARY_NAME)
	@echo "🔀 Testing combined template + variables..."
	@mkdir -p output
	@cat test_simple.json | ./$(BINARY_NAME) > output/test_combined.pdf
	@cat test_with_loops.json | ./$(BINARY_NAME) > output/test_loops.pdf
	@echo "   ✅ Combined template + variables test passed"

.PHONY: test-dynamic
test-dynamic:
	@echo "🔄 Testing dynamic loops system..."
	@cd cmd/test_dynamic_loops && go run main.go > /dev/null 2>&1
	@echo "   ✅ Dynamic loops tests passed"

.PHONY: test-yaml
test-yaml:
	@echo "🎯 Testing YAML support..."
	@cd cmd/test_yaml && go run main.go
	@echo "   ✅ YAML support tests passed"

# Examples
.PHONY: examples
examples: test-dynamic
	@echo "📄 Examples generated in output/ directory"

# Development
.PHONY: fmt
fmt:
	@echo "🎨 Formatting Go code..."
	go fmt ./...

.PHONY: validate
validate:
	@echo "🔍 Validating JSON templates..."
	@python3 -m json.tool template_dynamic.json > /dev/null && echo "   ✅ template_dynamic.json is valid" || echo "   ❌ template_dynamic.json has errors"
	@python3 -m json.tool variables_dynamic.json > /dev/null && echo "   ✅ variables_dynamic.json is valid" || echo "   ❌ variables_dynamic.json has errors"

# Clean targets
.PHONY: clean
clean:
	@rm -f $(BINARY_NAME)

.PHONY: clean-wasm
clean-wasm:
	@rm -f $(WASM_NAME)

.PHONY: clean-all
clean-all: clean clean-wasm
	@rm -rf output/ *.pdf debug_*.json

# Documentation
.PHONY: docs
docs:
	@echo "📖 Documentation disponible:"
	@echo ""
	@echo "📋 README.md                - Guide principal avec boucles dynamiques"
	@echo "🎯 README_TEMPLATING.md     - Guide complet du système de templating"
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  make test-dynamic         - Tester les boucles dynamiques"
	@echo "  make examples             - Générer les exemples"
	@echo ""

# Help
.PHONY: help
help:
	@echo "🔧 PDF Template System - Available Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "   make build              - Build local binary"
	@echo "   make build-wasm         - Build WASM module for Node.js"
	@echo "   make all                - Build + test (default)"
	@echo ""
	@echo "Test Commands:"
	@echo "   make test               - Run all tests"
	@echo "   make test-basic         - Test basic templates"
	@echo "   make test-dynamic       - Test dynamic loops with {{#array}}"
	@echo "   make examples           - Generate examples"
	@echo ""
	@echo "Development:"
	@echo "   make validate           - Validate JSON templates"
	@echo "   make fmt                - Format Go code"
	@echo ""
	@echo "Documentation:"
	@echo "   make docs               - Show available documentation"
	@echo ""
	@echo "Maintenance:"
	@echo "   make clean-all          - Clean everything"
	@echo "   make help               - Show this help"
