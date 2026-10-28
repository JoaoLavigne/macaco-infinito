package matcher

import (
	"log"
	"strings"
	"time"

	"macaco-infinito/internal/store"
)

// tamanhoBuffer é o número máximo de letras mantidas em memória.
// 15 cobre o maior tamanho de palavra permitido no cadastro.
const tamanhoBuffer = 15

// intervaloRetentativa define de quanto em quanto tempo o matcher
// tenta reenviar uma notificação que falhou.
const intervaloRetentativa = 30 * time.Second

// Matcher acompanha as letras geradas e detecta quando alguma palavra
// pendente aparece como substring contígua na sequência.
type Matcher struct {
	letras               <-chan rune
	store                *store.Store
	aoDetectar           func(store.Cadastro) error
	detectados           map[int64]bool
	falhas               map[int64]store.Cadastro
	ultimaRetentativa    time.Time
	intervaloRetentativa time.Duration
}

// Novo cria um matcher que lê letras do canal informado, consulta os
// cadastros pendentes e chama aoDetectar para cada palavra encontrada.
// A função aoDetectar devolve erro quando o envio falha; nesse caso o
// matcher tenta de novo mais tarde e só marca o cadastro como detectado
// após o sucesso.
func Novo(letras <-chan rune, armazenamento *store.Store, aoDetectar func(store.Cadastro) error) *Matcher {
	return &Matcher{
		letras:               letras,
		store:                armazenamento,
		aoDetectar:           aoDetectar,
		detectados:           make(map[int64]bool),
		falhas:               make(map[int64]store.Cadastro),
		intervaloRetentativa: intervaloRetentativa,
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
	m.retentarFalhas()

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
		if _, falhou := m.falhas[cadastro.ID]; falhou {
			continue
		}

		palavra := strings.ToUpper(cadastro.Palavra)
		if len(palavra) > len(buffer) {
			continue
		}

		if strings.Contains(texto, palavra) {
			m.enviar(cadastro)
		}
	}
}

// enviar chama o callback e guarda o cadastro para nova tentativa se falhar.
func (m *Matcher) enviar(cadastro store.Cadastro) {
	if err := m.aoDetectar(cadastro); err != nil {
		m.falhas[cadastro.ID] = cadastro
		m.ultimaRetentativa = time.Now()
		return
	}
	m.detectados[cadastro.ID] = true
}

// retentarFalhas tenta reenviar notificações que falharam, respeitando
// o intervalo mínimo entre tentativas.
func (m *Matcher) retentarFalhas() {
	if len(m.falhas) == 0 {
		return
	}
	if time.Since(m.ultimaRetentativa) < m.intervaloRetentativa {
		return
	}

	m.ultimaRetentativa = time.Now()
	for id, cadastro := range m.falhas {
		if err := m.aoDetectar(cadastro); err == nil {
			delete(m.falhas, id)
			m.detectados[id] = true
		}
	}
}
