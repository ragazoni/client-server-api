package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CurrencyData struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite3", "./cotacao.db")
	if err != nil {
		fmt.Println("Error opening database connection:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		data TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		bid DECIMAL(10,2)
	)`)
	if err != nil {
		fmt.Println("Error creating cotacoes table:", err)
		return
	}

	stmt, err := db.Prepare("INSERT INTO cotacoes(data, bid) VALUES(?,?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			log.Println(err)
			http.Error(w, "Erro ao buscar cotação do dólar", http.StatusInternalServerError)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			http.Error(w, "Erro ao buscar cotação do dólar", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "Erro ao buscar cotação do dólar", http.StatusInternalServerError)
			return
		}

		var currencyData CurrencyData
		err = json.Unmarshal(body, &currencyData)
		if err != nil {
			log.Println(err)
			http.Error(w, "Erro ao buscar cotação do dólar", http.StatusInternalServerError)
			return
		}

		ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancelDB()

		_, err = stmt.ExecContext(ctxDB, time.Now(), currencyData.USDBRL.Bid)
		if err != nil {
			log.Println(err)
			http.Error(w, "Erro ao salvar cotação no banco de dados", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(currencyData)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
