/**
* Author: suchiksagar
* Date : 4/4/24
**/
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
)

// RequestPayload represents the expected structure of the incoming request body
type RequestPayload struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Body     string `json:"body"`
}

// relayHandler reads the incoming request, extracts the downstream URL and method, and forwards the request
func relayHandler(w http.ResponseWriter, r *http.Request) {
	// Read and decode the request body to get the downstream URL and method
	var payload RequestPayload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Create a new request to the downstream service
	req, err := http.NewRequest(payload.Method, payload.Endpoint, bytes.NewBuffer([]byte(payload.Body)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers from the original request to the new request
	req.Header = r.Header

	// Send the request to the downstream service
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request to downstream service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy all headers from the downstream response to the original response writer
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set the status code from the downstream response
	w.WriteHeader(resp.StatusCode)

	// Stream the response body directly to the original response writer
	io.Copy(w, resp.Body)
}

func main() {
	// Define a command-line flag for the port, with a default value of "4488"
	var port string
	flag.StringVar(&port, "port", "4488", "the port to listen on")
	flag.Parse() // Parse the command-line flags

	// Set up the HTTP server
	http.HandleFunc("/relayserver", relayHandler)

	log.Printf("Server is running on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
