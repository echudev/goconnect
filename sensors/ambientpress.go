package sensors

import (
	"time"
)

// Estructura para los datos del sensor
type PressData struct {
	Timestamp string  `json:"timestamp"`
	Press     float64 `json:"press"`
}

// Función para simular la lectura de datos de un sensor de temperatura y humedad
func ReadPressData() float64 {
	Press := 1015 + rnd.Float64()*10.0
	return Press
}

// Función para obtener los datos del sensor y formatearlos en una estructura
func GetPressData() PressData {
	press := ReadPressData()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return PressData{
		Timestamp: timestamp,
		Press:     press,
	}
}
