package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/TcMits/aipstr"
)

func main() {
	result, _ := json.Marshal(aipstr.NewOrderByLexer())
	io.WriteString(os.Stdout, string(result))
}
