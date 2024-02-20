package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Printf("erro create request: %+v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error sending request %+v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error decoding json", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("error decode json", err)

	}

	if value, ok := data["bid"].(string); ok {
		if err := os.WriteFile("cotacao.txt", []byte(fmt.Sprintf("Dólar: %s\n", value)), 0644); err != nil {
			fmt.Println("error writing to file contacao", err)
		}
		fmt.Println("Cotação salva com sucesso!")
	} else {
		fmt.Println("Campo 'bid' não encontrado no JSON")

	}
}
