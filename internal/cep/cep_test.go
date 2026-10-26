package cep

import "testing"

func TestValido(t *testing.T) {
	casos := []struct {
		valor  string
		espera bool
	}{
		{valor: "29902555", espera: true},
		{valor: "01001000", espera: true},
		{valor: "", espera: false},
		{valor: "1234567", espera: false},
		{valor: "123456789", espera: false},
		{valor: "0100100a", espera: false},
		{valor: "01001-000", espera: false},
		{valor: "２９９０２５５５", espera: false},
	}

	for _, caso := range casos {
		if obtido := Valido(caso.valor); obtido != caso.espera {
			t.Errorf("Valido(%q) = %v; queria %v", caso.valor, obtido, caso.espera)
		}
	}
}
