package davis

import (
	"math/rand"
	"time"
)

// Estructura para los datos del sensor
type SensorData struct {
	Timestamp   string  `json:"timestamp"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

// Crear una nueva instancia de rand.Rand con una fuente personalizada
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// Función para simular la lectura de datos de un sensor de temperatura y humedad
func ReadSensorData() (float64, float64) {
	temperature := 20.0 + rnd.Float64()*5.0
	humidity := 30.0 + rnd.Float64()*20.0
	return temperature, humidity
}

// Función para obtener los datos del sensor y formatearlos en una estructura
func GetSensorData() SensorData {
	temperature, humidity := ReadSensorData()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return SensorData{
		Timestamp:   timestamp,
		Temperature: temperature,
		Humidity:    humidity,
	}
}
