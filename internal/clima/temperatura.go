package clima

import "math"

// O Fahrenheit é arredondado para uma casa decimal e o Kelvin para duas. A
// conversão para Kelvin usa 273.15.
func paraFahrenheit(celsius float64) float64 {
	return math.Round((celsius*1.8+32)*10) / 10
}

func paraKelvin(celsius float64) float64 {
	return math.Round((celsius+273.15)*100) / 100
}
