package clima

import "testing"

func TestConversaoDeTemperatura(t *testing.T) {
	casos := []struct {
		nome       string
		celsius    float64
		fahrenheit float64
		kelvin     float64
	}{
		{nome: "dia quente", celsius: 28.5, fahrenheit: 83.3, kelvin: 301.5},
		{nome: "ponto de congelamento", celsius: 0, fahrenheit: 32, kelvin: 273},
		{nome: "escalas iguais", celsius: -40, fahrenheit: -40, kelvin: 233},
		{nome: "linhares hoje", celsius: 23, fahrenheit: 73.4, kelvin: 296},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if obtido := paraFahrenheit(caso.celsius); obtido != caso.fahrenheit {
				t.Errorf("paraFahrenheit(%v) = %v; queria %v", caso.celsius, obtido, caso.fahrenheit)
			}

			if obtido := paraKelvin(caso.celsius); obtido != caso.kelvin {
				t.Errorf("paraKelvin(%v) = %v; queria %v", caso.celsius, obtido, caso.kelvin)
			}
		})
	}
}
