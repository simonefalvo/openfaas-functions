package function

import (
	"encoding/json"
	"log"
	"net/http"

	handler "github.com/openfaas/templates-sdk/go-http"
)

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {

	var errorResponse = handler.Response{
		Body:       nil,
		StatusCode: http.StatusInternalServerError,
	}

	var input map[string][]int
	err := json.Unmarshal(req.Body, &input)
	if err != nil {
		return errorResponse, err
	}
	log.Printf("input: %v", input)

	result := make(map[string]int)
	for word, countArray := range input {
		for _, partialCount := range countArray {
			result[word] += partialCount
		}
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
