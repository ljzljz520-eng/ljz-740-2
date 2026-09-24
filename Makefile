# Build/test helpers for the goocr bindings.
#
# A native OCR shared library implementing ocr.h is required for cgo builds.
# The reference stub in test/stub lets the whole pipeline run without a real
# model/backend.

GO        ?= go
CC         = gcc
STUB_DIR   = test/stub
UNAME_S   := $(shell uname -s)

ifeq ($(UNAME_S),Linux)
  STUB_LIB = $(STUB_DIR)/libocr.so
  STUB_RPATH_ENV = LD_LIBRARY_PATH=$(abspath $(STUB_DIR))
endif
ifeq ($(UNAME_S),Darwin)
  STUB_LIB = $(STUB_DIR)/libocr.dylib
  STUB_RPATH_ENV = DYLD_LIBRARY_PATH=$(abspath $(STUB_DIR))
endif

.PHONY: all build stub test test-stub fmt vet clean example

all: build

# Build the reference stub shared library.
stub: $(STUB_LIB)

$(STUB_DIR)/libocr.so: $(STUB_DIR)/ocr_stub.c ocr.h
	$(CC) -shared -fPIC -I. -DOCR_BUILDING_LIBRARY -o $@ $(STUB_DIR)/ocr_stub.c

$(STUB_DIR)/libocr.dylib: $(STUB_DIR)/ocr_stub.c ocr.h
	$(CC) -dynamiclib -I. -DOCR_BUILDING_LIBRARY -o $@ $(STUB_DIR)/ocr_stub.c

# Link against the stub and build everything.
build: stub
	CGO_LDFLAGS="-L$(abspath $(STUB_DIR))" $(GO) build ./...

test:
	$(GO) test ./...

# Run the full suite with the stub on the loader search path.
test-stub: stub
	CGO_LDFLAGS="-L$(abspath $(STUB_DIR))" \
	OCR_LIBRARY_PATH="$(abspath $(STUB_DIR))" \
	$(STUB_RPATH_ENV) $(GO) test -v ./...

example: stub
	@mkdir -p /tmp/goocr-models /tmp/goocr-images
	@printf 'pretend-png' > /tmp/goocr-images/sample.png
	CGO_LDFLAGS="-L$(abspath $(STUB_DIR))" $(GO) build -o $(STUB_DIR)/ocr ./cmd/ocr
	$(STUB_RPATH_ENV) $(STUB_DIR)/ocr -model /tmp/goocr-models \
		-image /tmp/goocr-images/sample.png -pretty

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

clean:
	rm -f $(STUB_DIR)/libocr.so $(STUB_DIR)/libocr.dylib $(STUB_DIR)/ocr
