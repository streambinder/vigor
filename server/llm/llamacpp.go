package llm

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type LlamaCpp struct {
	LLM
	uri string
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	tiers := os.Getenv("LLAMACPP_TIERS")
	if tiers == "" {
		return
	}

	for _, tier := range strings.Split(tiers, ",") {
		openLLMs = append(openLLMs, &LlamaCpp{uri: tier})
	}
}

func (llm *LlamaCpp) query(prompt string) ([]byte, error) {
	request, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/completion", llm.uri),
		bytes.NewBuffer([]byte(fmt.Sprintf(`{"prompt": "%s","n_predict": 128}`, prompt))),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create llama.cpp request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := (&http.Client{}).Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to send request to llama.cpp: %s", err)
	}
	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from llama.cpp: %s", err)
	}

	return bytes, err
}
