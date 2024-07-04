package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

// Estructura para los datos del sensor
type SensorData struct {
	Timestamp   string  `json:"timestamp"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

// Función para simular la lectura de datos de un sensor de temperatura y humedad
func readSensorData() (float64, float64) {
	temperature := 20.0 + rand.Float64()*5.0
	humidity := 30.0 + rand.Float64()*20.0
	return temperature, humidity
}

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

// Función principal que combina la lectura y almacenamiento de datos
func mainLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// Crear un nuevo archivo CSV diario
		filename := time.Now().Format("2006-01-02") + ".csv"

		// Crear el archivo CSV y escribir el encabezado si no existe
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			file, err := os.Create(filename)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			file.Close()
			writeToCSV(filename, []string{"Timestamp", "Temperature (°C)", "Humidity (%)"})
		}

		// Leer datos del sensor cada 10 segundos
		for range time.Tick(10 * time.Second) {
			// Leer datos del sensor
			temperature, humidity := readSensorData()

			// Obtener la marca de tiempo actual
			timestamp := time.Now().Format("2006-01-02 15:04:05")

			// Almacenar los datos en el archivo CSV
			data := []string{timestamp, strconv.FormatFloat(temperature, 'f', 2, 64), strconv.FormatFloat(humidity, 'f', 2, 64)}
			writeToCSV(filename, data)

			// Mostrar los datos en la consola (opcional)
			fmt.Printf("Timestamp: %s, Temperature: %.2f °C, Humidity: %.2f %%\n", timestamp, temperature, humidity)
		}
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go mainLoop(&wg)

	// Esperar a que mainLoop termine
	wg.Wait()
}
