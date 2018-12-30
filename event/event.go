package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Event defines the JSON structure for emitting events
type Event struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type postBody struct {
	Body Event `json:"body"`
}

func post(payload []byte, url string) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "feedmonitor")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

// Send sends a JSON encoded Event a Dr.Gero host
func (e *Event) Send(host string) error {
	var payload postBody
	payload.Body = *e
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/event", host)
	err = post(body, url)
	if err != nil {
		return err
	}

	return nil
}
