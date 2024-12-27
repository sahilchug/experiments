package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Event struct {
	ID    string `json:"id"`
	Event string `json:"event"`
	Data  string `json:"data"`
}

func connectSSE(url string) (<-chan Event, <-chan error) {
	events := make(chan Event)
	errors := make(chan error)

	go func() {
		defer close(events)
		defer close(errors)

		// Make the initial HTTP GET request to the SSE endpoint
		resp, err := http.Get(url)
		if err != nil {
			errors <- err
			return
		}

		defer resp.Body.Close()

		// Use a scanner to read lines from the response body
		scanner := bufio.NewScanner(resp.Body)
		var event Event

		for scanner.Scan() {
			line := scanner.Text()
			// Parse SSE fields: `data:`, `id:`, `event:`, etc.
			if strings.HasPrefix(line, "data: ") {
				event.Data += strings.TrimPrefix(line, "data: ") + "\n"
			} else if strings.HasPrefix(line, "id: ") {
				event.ID = strings.TrimPrefix(line, "id: ")
			} else if strings.HasPrefix(line, "event: ") {
				event.Event = strings.TrimPrefix(line, "event: ")
			} else if line == "" {
				// End of event block; send event to the channel
				events <- event
				event = Event{} // reset for the next event
			}
		}

		if err := scanner.Err(); err != nil {
			errors <- err
		}
	}()

	return events, errors
}

func main() {
	url := "http://127.0.0.1:8080/api/events/v1"
	events, errors := connectSSE(url)

	// Listen for events or errors in the main goroutine
	for {
		select {
		case event := <-events:
			fmt.Printf("Received event: ID=%s, Event=%s, Data=%s\n", event.ID, event.Event, event.Data)
		case err := <-errors:
			log.Printf("Error: %v\n", err)
			return
		}
	}
}
