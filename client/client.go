package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	ctx := context.Background()
	cotacao, err := getCotacao(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer file.Close()
	file.WriteString(fmt.Sprintf("Dólar: %v", cotacao.Bid))
}

func getCotacao(ctx context.Context) (*Cotacao, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*300)
	defer cancel()
	url := "http://localhost:8080/cotacao"
	log.Printf("Obtendo cotação do dólar de %q\n", url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		return nil, err
	}
	return &cotacao, nil
}
