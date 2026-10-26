package clima

import "math"

// O Fahrenheit é arredondado para uma casa decimal e o Kelvin para duas.
func paraFahrenheit(celsius float64) float64 {
	return math.Round((celsius*1.8+32)*10) / 10
}

func paraKelvin(celsius float64) float64 {
	return math.Round((celsius+273)*100) / 100
}
