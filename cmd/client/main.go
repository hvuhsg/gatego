package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func sendRequest() *http.Response {
	// Sample data to send in JSON format
	data := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": 123,
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
	}

	// Create a new POST request with the JSON payload
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8004/?name=yoyo", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	// Set the appropriate Content-Type header for JSON
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Example", "Noot")

	// Send the POST request
	client := http.DefaultClient
	response, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request:", err)
	}

	return response
}
func main() {
	resp := sendRequest()
	defer resp.Body.Close() // Always defer closing the response body

	// Check the response status code
	if resp.StatusCode > 299 {
		log.Printf("Error: received status code %d", resp.StatusCode)
	}

	fmt.Println(resp)

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response body
	fmt.Println(string(data))
}
