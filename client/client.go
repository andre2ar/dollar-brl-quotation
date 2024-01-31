package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type ServerQuotation struct {
	Bid string `json:"bid"`
}

func main() {
	quotation, err := getQuotation()
	if err != nil {
		log.Fatalln("Error getting quotation:", err)
	}

	saveToFile(quotation)

	fmt.Println("Quotation:", quotation.Bid)
}

func getQuotation() (*ServerQuotation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, errors.New("context timeout")
	default:
		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/quotation/usd-brl", nil)
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
		var quotation ServerQuotation
		err = json.Unmarshal(responseBody, &quotation)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		return &quotation, nil
	}
}

func saveToFile(quotation *ServerQuotation) {
	f, err := os.OpenFile("client/quotation.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	f.WriteString("Dollar: " + quotation.Bid + "\n")
}

func fileExists(filePath string) bool {
	_, error := os.Stat(filePath)

	return !errors.Is(error, os.ErrNotExist)
}
