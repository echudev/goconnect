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

	"github.com/echudev/goconnect/sensors"
)

// Función para escribir datos en un archivo CSV
func writeToCSV(filename string, data []string) {
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

// Función para crear las carpetas necesarias si no existen
func createDirectories(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directories:", err)
		}
	}
}

// Función principal que combina la lectura y almacenamiento de datos
func mainLoop(stopChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-stopChan:
			fmt.Println("Stopping mainLoop...")
			return
		case <-time.Tick(10 * time.Second):
			// Obtener la fecha actual
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
				writeToCSV(filename, []string{"Timestamp", "Temperature (°C)", "Humidity (%)", "Pressure (hPa)"})
			}

			// Leer datos del sensor
			sensorData := sensors.GetSensorData()
			pressData := sensors.GetPressData()

			// Almacenar los datos en el archivo CSV
			data := []string{sensorData.Timestamp, strconv.FormatFloat(sensorData.Temperature, 'f', 2, 64), strconv.FormatFloat(sensorData.Humidity, 'f', 2, 64), strconv.FormatFloat(pressData.Press, 'f', 2, 64)}
			writeToCSV(filename, data)

			// Mostrar los datos en la consola (opcional)
			fmt.Printf("Timestamp: %s, Temperature: %.2f °C, Humidity: %.2f %%, Pressure: %.2f hPa\n", sensorData.Timestamp, sensorData.Temperature, sensorData.Humidity, pressData.Press)
		}
	}
}

func main() {
	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	wg.Add(1)
	go mainLoop(stopChan, &wg)

	// Capturar señales de interrupción (Ctrl+C) para una terminación controlada
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Esperar a recibir una señal de interrupción
	<-sigChan
	close(stopChan)

	// Esperar a que mainLoop termine
	wg.Wait()
	fmt.Println("Program terminated gracefully.")
}
