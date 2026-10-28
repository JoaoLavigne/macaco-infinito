package espera

import (
	"fmt"
	"math"
	"strings"
)

// Media devolve o tempo médio de espera, em segundos, para uma palavra
// de tamanho n: 26^n / 2 (duas letras por segundo).
func Media(tamanho int) float64 {
	return math.Pow(26, float64(tamanho)) / 2
}

// Formatar converte segundos em um texto legível (segundos, minutos,
// horas, dias, anos...).
func Formatar(segundos float64) string {
	const (
		minuto = 60.0
		hora   = 60 * minuto
		dia    = 24 * hora
		ano    = 365.25 * dia
	)

	switch {
	case segundos < minuto:
		return formatarNumero(segundos) + " segundos"
	case segundos < hora:
		return formatarNumero(segundos/minuto) + " minutos"
	case segundos < dia:
		return formatarNumero(segundos/hora) + " horas"
	case segundos < ano:
		return formatarNumero(segundos/dia) + " dias"
	}

	anos := segundos / ano
	switch {
	case anos < 1000:
		return formatarNumero(anos) + " anos"
	case anos < 1_000_000:
		return formatarNumero(anos/1_000) + " mil anos"
	case anos < 1_000_000_000:
		return formatarNumero(anos/1_000_000) + " milhões de anos"
	case anos < 1_000_000_000_000:
		return formatarNumero(anos/1_000_000_000) + " bilhões de anos"
	default:
		return formatarNumero(anos/1_000_000_000_000) + " trilhões de anos"
	}
}

func formatarNumero(valor float64) string {
	var texto string
	if valor < 10 {
		texto = fmt.Sprintf("%.1f", valor)
	} else {
		texto = fmt.Sprintf("%.0f", valor)
	}
	return strings.Replace(texto, ".", ",", 1)
}
