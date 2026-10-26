// Package clima concentra as conversões de temperatura e a orquestração
// entre o ViaCEP e a WeatherAPI.
package clima

// Clima é o corpo devolvido pelo serviço B.
type Clima struct {
	Cidade     string  `json:"city"`
	Celsius    float64 `json:"temp_C"`
	Fahrenheit float64 `json:"temp_F"`
	Kelvin     float64 `json:"temp_K"`
}
