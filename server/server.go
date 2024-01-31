package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"time"
)

type USDBRLQuotation struct {
	Usdbrl struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

type Quotation struct {
	ID  int `gorm:"primaryKey"`
	Bid string
}

var db *gorm.DB

func main() {
	fmt.Println("Server is starting...")

	db = getDatabase()

	http.HandleFunc("/quotation/usd-brl", usdBrlQuotationHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println(err)
	}
}

func getDatabase() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("quotation.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Quotation{})

	fmt.Println("Connected to the database")

	return db
}

func usdBrlQuotationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	quotation, err := getUsdBrlQuotation()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Internal server error"}`))
		return
	}
	_, err = saveToDatabase(quotation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Internal server error"}`))
		return
	}
	w.WriteHeader(http.StatusOK)

	bytesJson, err := json.Marshal(quotation.Usdbrl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Internal server error"}`))
		return
	}

	w.Write(bytesJson)
}

func getUsdBrlQuotation() (*USDBRLQuotation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, errors.New("context timeout")
	default:
		req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer res.Body.Close()

		responseBody, _ := io.ReadAll(res.Body)

		var usdBrlQuotation USDBRLQuotation
		err = json.Unmarshal(responseBody, &usdBrlQuotation)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		return &usdBrlQuotation, nil
	}
}

func saveToDatabase(usdBrlQuotation *USDBRLQuotation) (*Quotation, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("saveToDatabase context timeout")
		return nil, errors.New("saveToDatabase context timeout")
	default:
		quotation := Quotation{Bid: usdBrlQuotation.Usdbrl.Bid}
		db.Create(&quotation)
		return &quotation, nil
	}
}
