GO := go

GO_SRCS := *.go services/*/*.go
DASHBOARD_SRCS := dashboard/src/**/*.tsx dashboard/src/*.tsx

APP := pam_postgres

all: build/pam_postgres
	@echo "Build complete"
.PHONY: all

dashboard/dist/index.html: $(DASHBOARD_SRCS)
	cd dashboard && bun run build

build/$(APP): dashboard/dist/index.html $(GO_SRCS)
	$(GO) build -o build/$(APP) .

run: build/$(APP)
	./build/$(APP)
.PHONY: run

clean:
	rm -f build/$(APP)
	rm -rf dashboard/dist
.PHONY: clean