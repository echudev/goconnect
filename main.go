package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/echudev/goconnect/drivers/davis"
	"github.com/echudev/goconnect/drivers/thermo"
)

type SensorData struct {
	Temperature float64
	Humidity    float64
	Pressure    float64
	Count       int
}

var mu sync.Mutex
var sensorDataMap = make(map[string]*SensorData)

func collectData(sensorFunc interface{}, interval time.Duration, stopChan <-chan struct{}, wg *sync.WaitGroup) {
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
			mu.Lock()
			if _, exists := sensorDataMap[timestamp]; !exists {
				sensorDataMap[timestamp] = &SensorData{}
			}
			data := sensorDataMap[timestamp]
			switch f := sensorFunc.(type) {
			case func() davis.SensorData:
				sensorData := f()
				data.Temperature += sensorData.Temperature
				data.Humidity += sensorData.Humidity
				data.Count++
			case func() thermo.CoData:
				coData := f()
				data.Pressure += coData.Co
				data.Count++
			}
			mu.Unlock()
		}
	}
}

func writeToCSV(filename string, stopChan <-chan struct{}, wg *sync.WaitGroup) {
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
			mu.Lock()
			data, exists := sensorDataMap[timestamp]
			if exists {
				avgTemp := data.Temperature / float64(data.Count)
				avgHum := data.Humidity / float64(data.Count)
				avgPress := data.Pressure / float64(data.Count)
				csvData := []string{
					timestamp,
					strconv.FormatFloat(avgTemp, 'f', 2, 64),
					strconv.FormatFloat(avgHum, 'f', 2, 64),
					strconv.FormatFloat(avgPress, 'f', 2, 64),
				}
				writeRowToCSV(filename, csvData)
				delete(sensorDataMap, timestamp) // Eliminar datos después de escribir en CSV
			}
			mu.Unlock()
		}
	}
}

func writeRowToCSV(filename string, data []string) {
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

func createDirectories(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directories:", err)
		}
	}
}

func main() {
	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	// Obtener la fecha actual y configurar la estructura de carpetas y archivo CSV
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	// Crear las carpetas necesarias
	path := filepath.Join("data", year, month)
	createDirectories(path)

	// Nombre del archivo CSV diario
	filename := filepath.Join(path, day+".csv")

	// Crear el archivo CSV y escribir el encabezado si no existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		file.Close()
		writeRowToCSV(filename, []string{"Timestamp", "Temperature (°C)", "Humidity (%)", "Pressure (hPa)"})
	}

	// Configurar y lanzar goroutines para cada sensor
	wg.Add(1)
	go collectData(davis.GetSensorData, 10*time.Second, stopChan, &wg)

	wg.Add(1)
	go collectData(thermo.GetCoData, 15*time.Second, stopChan, &wg)

	// Lanzar goroutine para escribir en el CSV cada minuto
	wg.Add(1)
	go writeToCSV(filename, stopChan, &wg)

	// Capturar señales de interrupción (Ctrl+C) para una terminación controlada
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Esperar a recibir una señal de interrupción
	<-sigChan
	close(stopChan)

	// Esperar a que todas las goroutines terminen
	wg.Wait()
	fmt.Println("Program terminated gracefully.")
}
