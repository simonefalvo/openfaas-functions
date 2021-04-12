package function

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	handler "github.com/openfaas/templates-sdk/go-http"
)

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {

	var errorResponse = handler.Response{
		Body:       nil,
		StatusCode: http.StatusInternalServerError,
	}

	chunk := string(req.Body)
	log.Printf("input received: %s", chunk)

	result := make(map[string]int)
	scanner := bufio.NewScanner(strings.NewReader(chunk))
	// Set the split function for the scanning operation.
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := scanner.Text()
		word = strings.Trim(word, ".,:;()[]{}!?'\"\"")
		wordLength := len(word)
		if wordLength != 0 {
			result[word] += 1
		}
	}
	if err := scanner.Err(); err != nil {
		return errorResponse, err
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errorResponse, err
	}

	return handler.Response{
		Body:       jsonResult,
		StatusCode: http.StatusOK,
	}, err
}
