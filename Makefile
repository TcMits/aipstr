all: create_lexer

create_lexer:
	go run ./cmd/filterlexer > filterlexer.json
	go run ./cmd/orderbylexer > orderbylexer.json
	participle gen lexer aipstr --name Filter < filterlexer.json | gofmt > ./filter_lexer.go
	participle gen lexer aipstr --name OrderBy < orderbylexer.json | gofmt > ./order_by_lexer.go
	rm filterlexer.json orderbylexer.json
