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

	"github.com/echudev/goconnect/drivers/davisvp2"
	"github.com/echudev/goconnect/drivers/thermo48i"
)

type DataStruct struct {
	Data  map[string]float64
	Count int
}

var mu sync.Mutex
var dataMap = make(map[string]*DataStruct)

func collectData(driver func() map[string]float64, keys []string, interval time.Duration, stopChan <-chan struct{}, wg *sync.WaitGroup) {
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
			if _, exists := dataMap[timestamp]; !exists {
				dataMap[timestamp] = &DataStruct{
					Data: make(map[string]float64),
				}
			}
			data := dataMap[timestamp]
			sensorData := driver()
			for _, key := range keys {
				data.Data[key] += sensorData[key]
			}
			data.Count++
			mu.Unlock()
		}
	}
}

func writeToCSV(filename string, keys []string, stopChan <-chan struct{}, wg *sync.WaitGroup) {
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
			data, exists := dataMap[timestamp]
			if exists {
				csvData := []string{timestamp}
				for _, key := range keys {
					avg := data.Data[key] / float64(data.Count)
					csvData = append(csvData, strconv.FormatFloat(avg, 'f', 2, 64))
				}
				writeRowToCSV(filename, csvData)
				delete(dataMap, timestamp) // Eliminar datos después de escribir en CSV
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

	// Definir los sensores seleccionados por el usuario
	sensorsSelected := []struct {
		Driver   func() map[string]float64
		Keys     []string
		Interval time.Duration
	}{
		{Driver: davisvp2.GetSerialCOM, Keys: []string{"Temperature", "Humidity", "Pressure"}, Interval: 10 * time.Second},
		{Driver: thermo48i.GetModbusEthernet, Keys: []string{"Co"}, Interval: 10 * time.Second},
		// Agregar más sensores según sea necesario
	}

	// Generar las cabeceras del CSV dinámicamente
	headers := []string{"Timestamp"}
	for _, sensor := range sensorsSelected {
		headers = append(headers, sensor.Keys...)
	}

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
		writeRowToCSV(filename, headers)
	}

	// Configurar y lanzar goroutines para cada sensor
	for _, sensor := range sensorsSelected {
		wg.Add(1)
		go collectData(sensor.Driver, sensor.Keys, sensor.Interval, stopChan, &wg)
	}

	// Lanzar goroutine para escribir en el CSV cada minuto
	wg.Add(1)
	go writeToCSV(filename, headers[1:], stopChan, &wg) // Se pasa headers[1:] para excluir "Timestamp"

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
