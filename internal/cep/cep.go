// Package cep contém a validação de CEP usada pelos dois serviços.
package cep

// Valido informa se o valor tem exatamente oito dígitos.
func Valido(valor string) bool {
	if len(valor) != 8 {
		return false
	}

	for i := 0; i < len(valor); i++ {
		if valor[i] < '0' || valor[i] > '9' {
			return false
		}
	}

	return true
}
