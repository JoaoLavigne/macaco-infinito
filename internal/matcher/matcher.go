package matcher

import (
	"log"
	"strings"

	"macaco-infinito/internal/store"
)

// tamanhoBuffer é o número máximo de letras mantidas em memória.
// 15 cobre o maior tamanho de palavra permitido no cadastro.
const tamanhoBuffer = 15

// Matcher acompanha as letras geradas e detecta quando alguma palavra
// pendente aparece como substring contígua na sequência.
type Matcher struct {
	letras     <-chan rune
	store      *store.Store
	aoDetectar func(store.Cadastro)
	detectados map[int64]bool
}

// Novo cria um matcher que lê letras do canal informado, consulta os
// cadastros pendentes e chama aoDetectar para cada palavra encontrada.
func Novo(letras <-chan rune, armazenamento *store.Store, aoDetectar func(store.Cadastro)) *Matcher {
	return &Matcher{
		letras:     letras,
		store:      armazenamento,
		aoDetectar: aoDetectar,
		detectados: make(map[int64]bool),
	}
}

// Iniciar começa a observar as letras em uma goroutine.
func (m *Matcher) Iniciar() {
	go m.loop()
}

func (m *Matcher) loop() {
	buffer := make([]rune, 0, tamanhoBuffer)

	for letra := range m.letras {
		buffer = append(buffer, letra)
		if len(buffer) > tamanhoBuffer {
			buffer = buffer[len(buffer)-tamanhoBuffer:]
		}

		m.verificar(buffer)
	}
}

// verificar procura cada palavra pendente dentro do buffer atual.
// A busca é por substring contígua, então "ana" também é encontrada
// dentro de "banana".
func (m *Matcher) verificar(buffer []rune) {
	pendentes, err := m.store.ListarPendentes()
	if err != nil {
		log.Printf("matcher: erro ao listar pendentes: %v", err)
		return
	}

	texto := string(buffer)

	for _, cadastro := range pendentes {
		if m.detectados[cadastro.ID] {
			continue
		}

		palavra := strings.ToUpper(cadastro.Palavra)
		if len(palavra) > len(buffer) {
			continue
		}

		if strings.Contains(texto, palavra) {
			m.detectados[cadastro.ID] = true
			m.aoDetectar(cadastro)
		}
	}
}
