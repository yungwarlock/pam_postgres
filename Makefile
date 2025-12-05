all: app
	@echo "Build complete"

.PHONY: all

dashboard/dist/index.html:
	cd dashboard && bun run build

app: dashboard/dist/index.html
	go build -o pam_postgres .

run: app
	./pam_postgres
.PHONY: run

clean:
	rm -f pam_postgres
	rm -rf dashboard/dist
.PHONY: clean