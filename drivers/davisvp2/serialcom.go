package davisvp2

import (
	"math/rand"
)

// Función para simular la lectura de datos de una meteorológica davis vantage pro 2
func ReadSerialCOM() []float64 {
	temperature := 20.0 + rand.Float64()*5.0
	humidity := 30.0 + rand.Float64()*20.0
	pressure := 1000.0 + rand.Float64()*100.0
	uv := 0.0 + rand.Float64()*100.0
	rain := 0.0 + rand.Float64()*100.0

	return []float64{temperature, humidity, pressure, uv, rain}
}

// Función para obtener los datos del sensor y formatearlos en una estructura
func GetSerialCOM() map[string]float64 {

	values := ReadSerialCOM()

	return map[string]float64{
		"Temperature": values[0],
		"Humidity":    values[1],
		"Pressure":    values[2],
		"UV":          values[3],
		"Rain":        values[4],
	}
}
