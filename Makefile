build:
	@go build -o bin/asteroid
run:
	@./bin/asteroid
deps:
	@go get ./... 
