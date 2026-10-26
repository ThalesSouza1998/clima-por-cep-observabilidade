package clima

import (
	"context"
	"errors"

	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/cep"
	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/viacep"
)

var (
	ErrCEPinvalido      = errors.New("cep inválido")
	ErrCEPNaoEncontrado = errors.New("cep não encontrado")
)

// BuscaCidade resolve o município a partir do CEP.
type BuscaCidade interface {
	Cidade(ctx context.Context, cep string) (string, error)
}

// Termometro informa a temperatura atual de uma cidade, em Celsius.
type Termometro interface {
	TemperaturaAtual(ctx context.Context, cidade string) (float64, error)
}

type Servico struct {
	cidades    BuscaCidade
	termometro Termometro
}

func NovoServico(cidades BuscaCidade, termometro Termometro) *Servico {
	return &Servico{cidades: cidades, termometro: termometro}
}

func (s *Servico) PorCEP(ctx context.Context, valor string) (Clima, error) {
	if !cep.Valido(valor) {
		return Clima{}, ErrCEPinvalido
	}

	cidade, err := s.cidades.Cidade(ctx, valor)
	if err != nil {
		if errors.Is(err, viacep.ErrCEPDesconhecido) {
			return Clima{}, ErrCEPNaoEncontrado
		}
		return Clima{}, err
	}

	celsius, err := s.termometro.TemperaturaAtual(ctx, cidade)
	if err != nil {
		return Clima{}, err
	}

	return Clima{
		Cidade:     cidade,
		Celsius:    celsius,
		Fahrenheit: paraFahrenheit(celsius),
		Kelvin:     paraKelvin(celsius),
	}, nil
}
