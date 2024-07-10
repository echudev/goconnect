package thermo

import (
	"math/rand"
)

// Estructura para los datos del sensor
type CoData struct {
	Timestamp string  `json:"timestamp"`
	Co        float64 `json:"co"`
}

// Función para simular la lectura de datos de un sensor de temperatura y humedad
func ReadCoData() float64 {
	co := 0.5 + rand.Float64()*0.6
	return co
}

// Función para obtener los datos del sensor y formatearlos en una estructura
func GetCoData(CoData) CoData {
	co := ReadCoData()
	return CoData{
		Co: co,
	}
}
