package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/TcMits/aipstr"
)

func main() {
	result, _ := json.Marshal(aipstr.NewFilterLexer())
	io.WriteString(os.Stdout, string(result))
}
