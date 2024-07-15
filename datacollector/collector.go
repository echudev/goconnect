package datacollector

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

type DataStruct struct {
	Data  map[string]float64
	Count int
}

var Mu sync.Mutex
var DataMap = make(map[string]*DataStruct)

func CollectData(driver func() map[string]float64, keys []string, interval time.Duration, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			fmt.Println("Stopping data collection...")
			return
		case <-ticker.C:
			timestamp := time.Now().Format("2006-01-02 15:04")
			Mu.Lock()
			if _, exists := DataMap[timestamp]; !exists {
				DataMap[timestamp] = &DataStruct{
					Data: make(map[string]float64),
				}
			}
			data := DataMap[timestamp]
			sensorData := driver()
			for _, key := range keys {
				data.Data[key] += sensorData[key]
			}
			fmt.Println(DataMap)
			data.Count++
			Mu.Unlock()
		}
	}
}

func WriteToCSV(filename string, keys []string, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			fmt.Println("Stopping CSV writing...")
			return
		case <-ticker.C:
			now := time.Now()
			timestamp := now.Add(-1 * time.Minute).Format("2006-01-02 15:04")
			Mu.Lock()
			data, exists := DataMap[timestamp]
			if exists {
				csvData := []string{timestamp}
				for _, key := range keys {
					avg := data.Data[key] / float64(data.Count)
					csvData = append(csvData, strconv.FormatFloat(avg, 'f', 2, 64))
				}
				WriteRowToCSV(filename, csvData)
				delete(DataMap, timestamp) // Eliminar datos después de escribir en CSV
			}
			Mu.Unlock()
		}
	}
}

func WriteRowToCSV(filename string, data []string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(data); err != nil {
		fmt.Println("Error writing to CSV:", err)
	}
}

func CreateDirectories(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directories:", err)
		}
	}
}
