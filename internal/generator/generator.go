package generator

import (
	"math/rand"
	"sync"
	"time"
)

// Gerador produz letras aleatórias de A a Z continuamente.
// Mantém as últimas N letras em memória para consulta pelo servidor HTTP.
type Gerador struct {
	mu       sync.Mutex
	letras   []rune // buffer circular das últimas letras
	tamanho  int    // quantas letras manter (padrão: 10)
	ativo    bool   // controla início/parada do loop
	contador int    // total de letras geradas desde o início
}

// NovoGerador cria um gerador com o tamanho de buffer especificado.
func NovoGerador(tamanho int) *Gerador {
	return &Gerador{
		letras:  make([]rune, 0, tamanho),
		tamanho: tamanho,
	}
}

// Iniciar começa a gerar letras continuamente em uma goroutine.
// Cada letra é gerada a cada intervalo especificado.
func (g *Gerador) Iniciar(intervalo time.Duration) {
	g.mu.Lock()
	g.ativo = true
	g.mu.Unlock()

	go func() {
		for {
			g.mu.Lock()
			if !g.ativo {
				g.mu.Unlock()
				return
			}
			// Gera uma letra aleatória de A a Z (65 a 90 na tabela ASCII)
			letra := rune('A' + rand.Intn(26))
			g.adicionarLetra(letra)
			g.mu.Unlock()
			time.Sleep(intervalo)
		}
	}()
}

// Parar interrompe o loop de geração.
func (g *Gerador) Parar() {
	g.mu.Lock()
	g.ativo = false
	g.mu.Unlock()
}

// adicionarLetra adiciona uma letra ao buffer, mantendo apenas as últimas 'tamanho'.
// Deve ser chamado com o mutex já travado.
func (g *Gerador) adicionarLetra(letra rune) {
	g.letras = append(g.letras, letra)
	if len(g.letras) > g.tamanho {
		g.letras = g.letras[len(g.letras)-g.tamanho:]
	}
	g.contador++
}

// LetrasRetornar retorna uma cópia das letras atuais para exibição na tela.
// Retorna uma cópia para evitar que o caller modifique o estado interno.
func (g *Gerador) LetrasRetornar() []rune {
	g.mu.Lock()
	defer g.mu.Unlock()
	result := make([]rune, len(g.letras))
	copy(result, g.letras)
	return result
}

// ContadorRetornar retorna o total de letras geradas desde o início.
// Útil para monitoramento (detectar se o gerador travou).
func (g *Gerador) ContadorRetornar() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.contador
}
