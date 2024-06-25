package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/tidwall/gjson"
)

type Document struct {
	ID          string  `json:"id"`
	PhoneNumber string  `json:"phone_number"`
	IsVPC       bool    `json:"is_vpc"`
	Biometric   string  `json:"biometric"`
	IsChanged   bool    `json:"is_changed"`
	Value       float64 `json:"value"`
	Timestamp   string  `json:"timestamp"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Distance    float64 `json:"distance"`
	Status      string  `json:"status"`
	Name        string  `json:"name"`
	Age         int     `json:"age"`
	Score       float64 `json:"score"`
	Level       int     `json:"level"`
}

func fetchBatch(es *elasticsearch.Client, scrollID string, results chan<- []Document, progressCh chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	for scrollID != "" {
		req := esapi.ScrollRequest{
			ScrollID: scrollID,
			Scroll:   1 * time.Minute,
		}

		res, err := req.Do(context.Background(), es)
		if err != nil {
			log.Printf("Error getting response: %s", err)
			return
		}
		defer res.Body.Close()

		if res.IsError() {
			log.Printf("Error: %s", res.String())
			return
		}

		var documents []Document
		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("Error reading response body: %s", err)
			return
		}

		result := gjson.Parse(string(responseBody))

		for _, hit := range result.Get("hits.hits").Array() {
			documents = append(documents, Document{
				ID:          hit.Get("_id").String(),
				PhoneNumber: hit.Get("_source.phone_number").String(),
				IsVPC:       hit.Get("_source.is_vpc").Bool(),
				Biometric:   hit.Get("_source.biometric").String(),
				IsChanged:   hit.Get("_source.is_changed").Bool(),
				Value:       hit.Get("_source.value").Float(),
				Timestamp:   hit.Get("_source.timestamp").String(),
				Latitude:    hit.Get("_source.latitude").Float(),
				Longitude:   hit.Get("_source.longitude").Float(),
				Distance:    hit.Get("_source.distance").Float(),
				Status:      hit.Get("_source.status").String(),
				Name:        hit.Get("_source.name").String(),
				Age:         int(hit.Get("_source.age").Int()),
				Score:       hit.Get("_source.score").Float(),
				Level:       int(hit.Get("_source.level").Int()),
			})
		}

		newScrollID := result.Get("_scroll_id").String()
		results <- documents
		progressCh <- len(documents)

		if len(documents) == 0 {
			break
		}

		scrollID = newScrollID
	}
}

func writeCSV(documents []Document, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Error creating file: %s", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"ID", "PhoneNumber", "IsVPC", "Biometric", "IsChanged", "Value", "Timestamp", "Latitude", "Longitude", "Distance", "Status", "Name", "Age", "Score", "Level"})

	for _, doc := range documents {
		record := []string{
			doc.ID,
			doc.PhoneNumber,
			strconv.FormatBool(doc.IsVPC),
			doc.Biometric,
			strconv.FormatBool(doc.IsChanged),
			fmt.Sprintf("%f", doc.Value),
			doc.Timestamp,
			fmt.Sprintf("%f", doc.Latitude),
			fmt.Sprintf("%f", doc.Longitude),
			fmt.Sprintf("%f", doc.Distance),
			doc.Status,
			doc.Name,
			strconv.Itoa(doc.Age),
			fmt.Sprintf("%f", doc.Score),
			strconv.Itoa(doc.Level),
		}
		writer.Write(record)
	}

	return nil
}

func main() {
	start := time.Now()

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	})
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	indexName := "index-2026"
	batchSize := 1000
	totalDocuments := 1000000
	numWorkers := 10

	var wg sync.WaitGroup
	results := make(chan []Document, numWorkers)
	progressCh := make(chan int, numWorkers)

	allDocuments := []Document{}

	go func() {
		totalFetched := 0
		for progress := range progressCh {
			totalFetched += progress
			percent := float64(totalFetched) / float64(totalDocuments) * 100
			fmt.Printf("Progress: %.2f%% (%d/%d)\n", percent, totalFetched, totalDocuments)
		}
	}()

	query := `{
		"query": {
			"match_all": {}
		}
	}`
	batchSizePtr := batchSize
	req := esapi.SearchRequest{
		Index:  []string{indexName},
		Body:   strings.NewReader(query),
		Scroll: 1 * time.Minute,
		Size:   &batchSizePtr,
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %s", err)
	}
	log.Printf("Initial response body: %s", responseBody)

	initialScrollID := gjson.Get(string(responseBody), "_scroll_id").String()
	if initialScrollID == "" {
		log.Fatalf("Initial scroll ID is empty")
	}

	wg.Add(1)
	go fetchBatch(es, initialScrollID, results, progressCh, &wg)

	go func() {
		wg.Wait()
		close(results)
		close(progressCh)
	}()

	for batch := range results {
		allDocuments = append(allDocuments, batch...)
		log.Printf("Fetched %d documents, total documents so far: %d", len(batch), len(allDocuments))
	}

	err = writeCSV(allDocuments, "data.csv")
	if err != nil {
		log.Fatalf("Error writing to CSV: %s", err)
	}

	duration := time.Since(start)
	fmt.Printf("Data fetched and written to CSV in %v\n", duration)
}
