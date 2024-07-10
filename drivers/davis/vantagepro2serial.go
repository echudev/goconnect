package davis

import (
	"math/rand"
)

// Estructura para los datos del sensor
type SensorData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

// Función para simular la lectura de datos de un sensor de temperatura y humedad
func ReadSensorData() (float64, float64) {
	temperature := 20.0 + rand.Float64()*5.0
	humidity := 30.0 + rand.Float64()*20.0
	return temperature, humidity
}

// Función para obtener los datos del sensor y formatearlos en una estructura
func GetSensorData() SensorData {
	temperature, humidity := ReadSensorData()
	return SensorData{
		Temperature: temperature,
		Humidity:    humidity,
	}
}
