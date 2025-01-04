**/*_templ.go: **/*.templ
	templ generate

server: **/*_templ.go
	go run main.go --database=database.db server --port=8081 

watch:
	air -build.cmd='templ generate; go build -o tmp/main main.go' \
		-build.exclude_regex='_templ.go' \
		-build.include_ext='go,templ' \
		server --port=8081 --database=database.db

migration_create:
		go run github.com/amacneil/dbmate new '$(name)'

check:
	gofmt -d -e -s .
	go vet ./...
	staticcheck ./...
	gosec -quiet ./...
	gocylo -over 10 -ignore '*_templ\.go' .
	golangci-lint run
	govulncheck
	go test -race ./...

home_software: **/*.go
	PKG_CONFIG_LIBDIR=/usr/lib/aarch64-linux-gnu/pkgconfig \
		CC="zig cc -target aarch64-linux-gnu -isystem /usr/include -L/usr/lib/aarch64-linux-gnu" \
		CXX="zig c++ -target aarch64-linux-gnu -isystem /usr/include -L/usr/lib/aarch64-linux-gnu" \
		CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build
