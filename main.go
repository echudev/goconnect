package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/echudev/goconnect/datacollector"
	"github.com/echudev/goconnect/drivers/davisvp2"
	"github.com/echudev/goconnect/drivers/thermo48i"
)

func main() {
	fmt.Println("program started")
	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	// Obtener la fecha actual y configurar la estructura de carpetas y archivo CSV
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	// Crear las carpetas necesarias
	path := filepath.Join("data", year, month)
	datacollector.CreateDirectories(path)

	// Definir los sensores seleccionados por el usuario
	sensorsSelected := []struct {
		Driver       func() map[string]float64
		Keys         []string
		ScanInterval time.Duration
	}{
		{Driver: davisvp2.GetSerialCOM, Keys: []string{"Temperature", "Humidity", "Pressure"}, ScanInterval: 5 * time.Second},
		{Driver: thermo48i.GetModbusEthernet, Keys: []string{"Co"}, ScanInterval: 5 * time.Second},
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
		datacollector.WriteRowToCSV(filename, headers)
	}

	// Configurar y lanzar goroutines para cada sensor
	for _, sensor := range sensorsSelected {
		wg.Add(1)
		go datacollector.CollectData(sensor.Driver, sensor.Keys, sensor.ScanInterval, stopChan, &wg)
	}

	// Lanzar goroutine para escribir en el CSV cada minuto
	wg.Add(1)
	go datacollector.WriteToCSV(filename, headers[1:], stopChan, &wg) // Se pasa headers[1:] para excluir "Timestamp"

	// Capturar señales de interrupción (Ctrl+C) para una terminación controlada
	sigChan := make(chan os.Signal, 1)
	fmt.Println("Press Ctrl+C to stop program...")
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Esperar a recibir una señal de interrupción
	<-sigChan
	close(stopChan)

	// Esperar a que todas las goroutines terminen
	wg.Wait()
	fmt.Println("Program terminated gracefully.")
}
