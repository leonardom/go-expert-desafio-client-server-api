package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	SERVER_PORT = "8080"
	DB_NAME     = "sqlite.db"
)

var db *sql.DB

type ExchangeRate struct {
	Code      string `json:"code"`
	CodeIn    string `json:"codein"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Bid       string `json:"bid"`
	Ask       string `json:"ask"`
	Timestamp string `json:"timestamp"`
}

type AwesomeApiResponse struct {
	USDBRL ExchangeRate
}

type ExchangeRateOutput struct {
	Bid string `json:"bid"`
}

func NewExchangeRateOutput(xr *ExchangeRate) *ExchangeRateOutput {
	return &ExchangeRateOutput{
		Bid: xr.Ask,
	}
}

func main() {
	log.Println("Creating database...")
	os.Remove(DB_NAME)
	file, err := os.Create(DB_NAME)
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("Database created!")

	db, err = sql.Open("sqlite3", "./"+DB_NAME)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()
	createTable(db)

	log.Printf("Server listening on port %v...\n", SERVER_PORT)
	http.HandleFunc("/cotacao", handleExchangeRate)
	http.ListenAndServe(":8080", nil)
}

func handleExchangeRate(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	xr, err := getExchangeRate(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	insertQuote(ctx, db, xr)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NewExchangeRateOutput(xr))
}

func getExchangeRate(ctx context.Context) (*ExchangeRate, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	log.Printf("Getting exchange rate from %q\n", url)
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
	var response AwesomeApiResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return &response.USDBRL, nil
}

func createTable(db *sql.DB) {
	sql := `CREATE TABLE exchange_rates (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"code" TEXT,
		"code_in" TEXT,
		"high" TEXT,
		"low" TEXT,
		"bid" TEXT,
		"ask" TEXT,
		"timestamp" TEXT
	  );`

	log.Println("Create exchange rates table...")
	statement, err := db.Prepare(sql)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
	log.Println("Table created!")
}

func insertQuote(ctx context.Context, db *sql.DB, xr *ExchangeRate) {
	log.Println("Inserting exchange rate record...")
	sql := `INSERT INTO exchange_rates(code, code_in, high, low, bid, ask, timestamp) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	statement, err := db.Prepare(sql)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer statement.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()
	_, err = statement.ExecContext(ctx, xr.Code, xr.CodeIn, xr.High, xr.Low, xr.Bid, xr.Ask, xr.Timestamp)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
