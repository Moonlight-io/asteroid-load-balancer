build:
	@go build -o bin/asteroid
run:
	@./bin/asteroid
install:
        @go get github.com/gorilla/mux