package thermo48i

import (
	"math/rand"
)

// Función para simular la lectura de datos del analizador thermo 48i
func ReadModbusEthernet() []float64 {
	co := 0.5 + rand.Float64()*0.6
	return []float64{co}
}

// Función para obtener los datos del sensor y formatearlos en una estructura
func GetModbusEthernet() map[string]float64 {

	values := ReadModbusEthernet()

	return map[string]float64{
		"Co": values[0],
	}
}
