package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

)

func StoreTokenUsage(userKey string, sessionID string, model string, promptTokens int, completionTokens int, totalTokens int) error {
	payload := map[string]interface{}{
		"user_key":          userKey,
		"sessionID":         sessionID,
		"model":             model,
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
		"total_tokens":      totalTokens,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error marshalling payload: %v", err)
	}
	log.Println("jsonPayload: ", string(jsonPayload))
	req, err := http.NewRequest("POST", Config.BaseUrl + "/api/tokens", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer " + userKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Error: status code %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}