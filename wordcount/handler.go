package function

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	handler "github.com/openfaas/templates-sdk/go-http"
)

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {

	gatewayUrl, ok := os.LookupEnv("GATEWAY_URL")
	if !ok {
		log.Fatal("GATEWAY_URL environment variable not set")
	}
	log.Printf("gateway url: %s", gatewayUrl)

	val, ok := os.LookupEnv("NUM_WORKERS")
	if !ok {
		log.Fatal("NUM_WORKERS environment variable not set")
	}
	numWorkers, err := strconv.Atoi(val)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("number of workers: %v", numWorkers)

	log.Printf("splitting data into %v chunks...", numWorkers)
	chunks := SplitData(req.Body, numWorkers)

	// call mappers
	channel := make(chan []byte)
	for w := 0; w < numWorkers; w++ {
		log.Printf("calling function mapper%v...", w)
		go callFunction(gatewayUrl+"/function/wordcount-mapper"+strconv.Itoa(w),
			[]byte(chunks[w]),
			channel)
	}

	// receive map results and merge them
	log.Print("listening on receiving channel...")
	mergedMapResults := make(map[string][]int)
	for w := 0; w < numWorkers; w++ {
		mapResult := make(map[string]int)
		data := <-channel
		log.Print("got map result")
		if json.Unmarshal(data, &mapResult) != nil {
			log.Fatal("error while unmarshalling map data")
		}
		for word, count := range mapResult {
			mergedMapResults[word] = append(mergedMapResults[word], count)
		}
	}

	// split reducers' input
	log.Print("assigning words to reducers...")
	wordCounter := 0
	reduceInput := make([]map[string][]int, numWorkers)
	for i := range reduceInput {
		// init array of maps
		reduceInput[i] = make(map[string][]int)
	}
	for word, countList := range mergedMapResults {
		reduceInput[wordCounter%numWorkers][word] = countList
		wordCounter++
	}

	// call reducers
	for w := 0; w < numWorkers; w++ {
		input, err := json.MarshalIndent(reduceInput[w], "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("calling function reducer%v...", w)
		go callFunction(gatewayUrl+"/function/wordcount-reducer"+strconv.Itoa(w),
			input,
			channel)
	}

	// receive reduce results and merge them
	log.Print("listening on receiving channel...")
	mergedReduceResults := make(map[string]int)
	var reduceResult map[string]int
	for w := 0; w < numWorkers; w++ {
		data := <-channel
		log.Print("got reduce result")
		if json.Unmarshal(data, &reduceResult) != nil {
			log.Fatal("error while unmarshalling reduce data")
		}
		for word, count := range reduceResult {
			mergedReduceResults[word] = count
		}
	}

	// format final result
	result, err := json.MarshalIndent(mergedReduceResults, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// send result back
	log.Print("sending results back...")
	return handler.Response{
		Body:       result,
		StatusCode: http.StatusOK,
	}, err
}

func callFunction(url string, data []byte, c chan []byte) {

	response, err := http.Post(url, "", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}

	// defer body close
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(response.Body)

	// read all the response body
	result, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	c <- result
}

func SplitData(data []byte, n int) []string {

	reader := bytes.NewReader(data)
	dataSize := int64(len(data))
	//fmt.Println("data size:", dataSize, "bytes")
	chunkSize := dataSize / int64(n)
	//fmt.Println("chunk size:", chunkSize, "bytes")
	chunks := make([]string, n)

	buffer := make([]byte, chunkSize)
	for i := 0; i < n; i++ {
		if i == n-1 {
			chunkSize = dataSize
			buffer = make([]byte, chunkSize)
		}
		read, err := io.ReadAtLeast(reader, buffer, int(chunkSize))
		if err != nil {
			log.Fatal(err)
		}
		chunks[i] = string(buffer)
		//fmt.Printf("-------READ CHUNK %d (%d bytes)-------\n", i, read)
		//fmt.Println(chunks[i])
		if i != n-1 {
			pos := strings.LastIndexByte(chunks[i], ' ')
			chunks[i] = chunks[i][:pos]
			//fmt.Printf("-------CUT CHUNK %d (%d bytes)------\n", i, pos)
			//fmt.Println(chunks[i])
			// Position at the beginning of the cut word
			_, err := reader.Seek(-int64(read-pos), io.SeekCurrent)
			if err != nil {
				log.Fatal(err)
			}
			dataSize -= int64(pos)
		}
	}
	return chunks
}
