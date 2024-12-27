**/*_templ.go: **/*.templ
	templ generate

server: **/*_templ.go
	go run main.go server

watch:
	air -build.cmd='templ generate; go build -o tmp/main main.go' \
		-build.exclude_regex='_templ.go' \
		-build.include_ext='go,templ' \
		server

migration_create:
		go run github.com/amacneil/dbmate new '$(name)'