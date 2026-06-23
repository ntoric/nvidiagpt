package nvidia

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	APIKey string
	URL    string
	Model  string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model             string    `json:"model"`
	Messages          []Message `json:"messages"`
	MaxTokens         int       `json:"max_tokens"`
	Temperature       float64   `json:"temperature"`
	TopP              float64   `json:"top_p"`
	FrequencyPenalty  float64   `json:"frequency_penalty"`
	PresencePenalty   float64   `json:"presence_penalty"`
	Stream            bool      `json:"stream"`
}

type chatChoice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

func New(apiKey, url, model string) *Client {
	return &Client{
		APIKey: apiKey,
		URL:    url,
		Model:  model,
	}
}

func (c *Client) StreamChat(ctx context.Context, messages []Message, model string, onToken func(content string)) error {
	if model == "" {
		model = c.Model
	}
	payload := chatRequest{
		Model:             model,
		Messages:          messages,
		MaxTokens:         1024,
		Temperature:       0.7,
		TopP:              1.0,
		FrequencyPenalty:  0.0,
		PresencePenalty:   0.0,
		Stream:            true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nvidia api error (status %d): %s", resp.StatusCode, string(raw))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk chatResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content == "" {
				content = chunk.Choices[0].Message.Content
			}
			if content != "" {
				onToken(content)
			}
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	return scanner.Err()
}
