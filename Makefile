# PDF Template System Makefile

# Variables
BINARY_NAME=pdf-template
WASM_NAME=pdf-template.wasm
TINY_WASM_NAME=pdf-template-tiny.wasm
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")

# Default target
.PHONY: all
all: build test

# Build targets

.PHONY: build
build: clean $(BINARY_NAME) $(WASM_NAME)  $(TINY_WASM_NAME)
	@echo "✅ WASM build complete!"

# Individual builds
$(BINARY_NAME): $(GO_FILES)
	@echo "📦 Building template system binary..."
	go build -o $(BINARY_NAME) main.go

$(WASM_NAME): $(GO_FILES)
	@echo "📦 Building WASM module..."
	CGO_ENABLED=0 GOOS=wasip1 GOARCH=wasm go build -ldflags="-s -w -X main.buildmode=production" -trimpath -tags=production -gcflags="-l=4" -o $(WASM_NAME) main.go
	@echo "� Optimizing WASM..."
	@wasm-opt $(WASM_NAME) -o $(WASM_NAME).opt --enable-bulk-memory -Oz 2>/dev/null && mv $(WASM_NAME).opt $(WASM_NAME) || echo "   ⚠️  wasm-opt not available, skipping optimization"
	@echo "📦 Compressing optimized WASM..."
	@gzip -k $(WASM_NAME) || echo "   ⚠️  gzip not available, skipping compression"
	@ls -lh $(WASM_NAME)*

$(TINY_WASM_NAME): $(GO_FILES)
	@echo "📦 Building WASM module with TinyGo..."
	tinygo build -target=wasip1 -opt=z -gc=leaking -scheduler=none -o $(TINY_WASM_NAME) main.go

# Test targets
.PHONY: test
test: build test-basic test-combined test-dynamic
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

# Performance tests
.PHONY: bench-report
bench-report: $(BINARY_NAME)
	@echo "📊 Generating PDF benchmark report..."
	@./generate_benchmark_report.sh

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
	@rm -f $(WASM_NAME)
	@rm -f $(WASM_NAME).gz
	@rm -f $(TINY_WASM_NAME)
	@rm -rf output/ *.pdf debug_*.json
