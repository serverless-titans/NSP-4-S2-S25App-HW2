package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// RequestPayload defines the input structure.
type requestPayload struct {
	Prompt string `json:"prompt"`
}

// ResponsePayload defines the output structure.
type responsePayload struct {
	Application string `json:"application"`
	Prompt      string `json:"prompt"`
	Response    string `json:"response"`
	Source      string `json:"source"`
}

type hfRequest struct {
	Inputs string `json:"inputs"`
}

type hfResponse struct {
	GeneratedText string `json:"generated_text"`
}

type routerChatRequest struct {
	Model    string              `json:"model"`
	Messages []routerChatMessage `json:"messages"`
}

type routerChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type routerChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type zenQuoteResponse struct {
	Quote  string `json:"q"`
	Author string `json:"a"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "ok",
		"environment": "staging",
	})
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,POST")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read body"})
		return
	}
	defer r.Body.Close()

	payload, err := parsePayload(string(body))
	if err != nil {
		log.Printf("Error parsing payload: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	answer, source, err := generateAnswer(r.Context(), payload.Prompt)
	if err != nil {
		log.Printf("Error generating answer: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responsePayload{
		Application: "NSP-4-S2-S25App - v1.1 - EC2 Staging",
		Prompt:      payload.Prompt,
		Response:    answer,
		Source:      source,
	})
}

func parsePayload(body string) (requestPayload, error) {
	var payload requestPayload
	if strings.TrimSpace(body) == "" {
		return payload, errors.New("request body is required")
	}

	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return payload, errors.New("request body must be valid JSON")
	}

	payload.Prompt = strings.TrimSpace(payload.Prompt)
	if payload.Prompt == "" {
		return payload, errors.New("prompt is required")
	}

	return payload, nil
}

func generateAnswer(ctx context.Context, prompt string) (string, string, error) {
	token := strings.TrimSpace(os.Getenv("HUGGINGFACE_API_TOKEN"))
	if token != "" {
		log.Printf("Attempting Hugging Face Inference Router...")
		answer, err := queryHuggingFace(ctx, "https://router.huggingface.co", prompt, token)
		if err == nil && strings.TrimSpace(answer) != "" {
			return answer, "huggingface-router", nil
		}
		log.Printf("Inference Router failed: %v", err)

		log.Printf("Attempting Hugging Face Standard Inference API (hf.co)...")
		answer, err = queryHuggingFace(ctx, "https://api-inference.hf.co", prompt, token)
		if err == nil && strings.TrimSpace(answer) != "" {
			return answer, "huggingface-standard", nil
		}
		log.Printf("Standard Inference API (hf.co) failed: %v", err)
	} else {
		log.Printf("HUGGINGFACE_API_TOKEN is not set, skipping HF.")
	}

	log.Printf("Attempting fallback to ZenQuotes API...")
	quote, author, err := fetchQuote(ctx)
	if err == nil {
		return fmt.Sprintf(
			"NSP-4-S2-S25App processed your prompt: %q. Public API context: %q - %s",
			prompt,
			quote,
			author,
		), "zenquotes", nil
	}
	log.Printf("Fallback API (ZenQuotes) also failed: %v", err)

	log.Printf("Attempting second fallback to Typefit API...")
	quote, err = fetchTypefitQuote(ctx)
	if err == nil {
		return fmt.Sprintf(
			"NSP-4-S2-S25App processed: %q. Secondary fallback context: %q",
			prompt,
			quote,
		), "typefit", nil
	}
	log.Printf("Second fallback API (Typefit) also failed: %v", err)

	return fmt.Sprintf(
		"NSP-4-S2-S25App processed: %s. (Note: All external LLM and quote APIs were unavailable)",
		prompt,
	), "local-fallback", nil
}

func queryHuggingFace(
	ctx context.Context,
	baseURL string,
	prompt string,
	token string,
) (string, error) {
	modelID := strings.TrimSpace(os.Getenv("HUGGINGFACE_MODEL_ID"))
	if modelID == "" {
		modelID = "nvidia/NVIDIA-Nemotron-3-Ultra-550B-A55B-NVFP4:together"
	}

	chatURL := fmt.Sprintf("%s/v1/chat/completions", baseURL)
	answer, err := tryChatCompletion(ctx, chatURL, modelID, prompt, token)
	if err == nil {
		return answer, nil
	}
	log.Printf("Chat Completion path failed for %s: %v", baseURL, err)

	taskURL := fmt.Sprintf("%s/models/%s", baseURL, modelID)
	return tryTaskInference(ctx, taskURL, prompt, token)
}

func tryChatCompletion(
	ctx context.Context,
	apiURL string,
	modelID string,
	prompt string,
	token string,
) (string, error) {
	body, err := json.Marshal(routerChatRequest{
		Model: modelID,
		Messages: []routerChatMessage{
			{
				Role:    "system",
				Content: "You are the backend for NSP-4-S2-S25App. Reply in one short sentence.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return performHFRequest(ctx, apiURL, body, token, true)
}

func tryTaskInference(
	ctx context.Context,
	apiURL string,
	prompt string,
	token string,
) (string, error) {
	body, err := json.Marshal(map[string]string{
		"inputs": fmt.Sprintf(
			"Instruct: You are the backend for NSP-4-S2-S25App. Reply in one short sentence to the following prompt.\nPrompt: %s\nOutput:",
			prompt,
		),
	})
	if err != nil {
		return "", err
	}

	return performHFRequest(ctx, apiURL, body, token, false)
}

func performHFRequest(
	ctx context.Context,
	apiURL string,
	body []byte,
	token string,
	isChat bool,
) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-wait-for-model", "true")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if isChat {
		var routerResponse routerChatResponse
		if err := json.Unmarshal(respBody, &routerResponse); err == nil &&
			len(routerResponse.Choices) > 0 {
			return strings.TrimSpace(routerResponse.Choices[0].Message.Content), nil
		}
	} else {
		var taskResp []hfResponse
		if err := json.Unmarshal(respBody, &taskResp); err == nil && len(taskResp) > 0 {
			return strings.TrimSpace(taskResp[0].GeneratedText), nil
		}
		var singleResp hfResponse
		if err := json.Unmarshal(respBody, &singleResp); err == nil {
			return strings.TrimSpace(singleResp.GeneratedText), nil
		}
	}

	return "", errors.New("unexpected response format")
}

func fetchTypefitQuote(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://type.fit/api/quotes", nil)
	if err != nil {
		return "", err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var quotes []struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&quotes); err != nil || len(quotes) == 0 {
		return "", errors.New("failed to parse typefit quotes")
	}

	return quotes[0].Text, nil
}

func fetchQuote(ctx context.Context) (string, string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://zenquotes.io/api/random",
		nil,
	)
	if err != nil {
		return "", "", err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("zenquotes API returned HTTP %d", resp.StatusCode)
	}

	var quotes []zenQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quotes); err != nil {
		return "", "", err
	}

	if len(quotes) == 0 {
		return "", "", errors.New("zenquotes returned an empty array")
	}

	return quotes[0].Quote, quotes[0].Author, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80" // Default port as requested
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", generateHandler) // Changed default route to generateHandler

	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
