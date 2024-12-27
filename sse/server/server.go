package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// DataPayload represents the JSON structure within the `data` field of each SSE event
type DataPayload struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type Event struct {
	ID        string      `json:"id,omitempty"`    // Optional ID
	EventType string      `json:"event,omitempty"` // Optional event type
	Data      DataPayload `json:"data"`            // Event data, serialized as JSON
	Retry     int         `json:"retry,omitempty"` // Optional retry interval in milliseconds
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Flushing ensures the response is sent immediately
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	// Send events in a loop
	for i := 0; i < 10; i++ {
		// Create a DataPayload for the current event
		data := DataPayload{
			Message:   fmt.Sprintf("Message #%d", i),
			Timestamp: time.Now().Format(time.RFC3339),
		}

		// Create an Event structure to hold the full event details
		event := Event{
			ID:        strconv.Itoa(i), // Optional ID as the counter
			EventType: "message",       // Custom event type, e.g., "message"
			Data:      data,            // Embed the data payload
			Retry:     5000,            // Optional retry interval (5 seconds)
		}

		// Serialize the DataPayload to JSON for the `data:` field
		jsonData, err := json.Marshal(event.Data)
		if err != nil {
			log.Printf("Error marshaling JSON data: %v\n", err)
			continue
		}
		// Send the SSE formatted event to the client
		fmt.Fprintf(w, "id: %s\n", event.ID)
		fmt.Fprintf(w, "event: %s\n", event.EventType)
		fmt.Fprintf(w, "retry: %d\n", event.Retry)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)

		// Flush to ensure the event is sent immediately
		flusher.Flush()

		// Simulate a delay between events
		time.Sleep(2 * time.Second)
	}
}

func main() {
	http.HandleFunc("/api/events/v1", sseHandler)
	log.Println("Starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
