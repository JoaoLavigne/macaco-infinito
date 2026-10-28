package matcher

import (
	"errors"
	"sync"
	"testing"
	"time"

	"macaco-infinito/internal/store"
)

// TestDetectaSubstringContigua prova duas coisas:
// 1. uma palavra só dispara quando está completa (BANAN não basta);
// 2. "ANA" é encontrada dentro de "BANANA", mesmo cruzando letras.
func TestDetectaSubstringContigua(t *testing.T) {
	armazenamento, err := store.Abrir(t.TempDir() + "/teste.db")
	if err != nil {
		t.Fatalf("abrir banco: %v", err)
	}
	defer armazenamento.Fechar()

	if err := armazenamento.InserirCadastro("ana@exemplo.com", "ANA", time.Now()); err != nil {
		t.Fatalf("inserir ANA: %v", err)
	}
	if err := armazenamento.InserirCadastro("banana@exemplo.com", "BANANA", time.Now()); err != nil {
		t.Fatalf("inserir BANANA: %v", err)
	}

	letras := make(chan rune, 16)
	var mu sync.Mutex
	var achados []string

	detector := Novo(letras, armazenamento, func(c store.Cadastro) error {
		mu.Lock()
		achados = append(achados, c.Palavra)
		mu.Unlock()
		return nil
	})
	detector.Iniciar()
	defer close(letras)

	quantidade := func() int {
		mu.Lock()
		defer mu.Unlock()
		return len(achados)
	}

	for _, letra := range "BAN" {
		letras <- letra
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)
	if quantidade() != 0 {
		t.Fatalf("não deveria detectar antes da palavra completa, obteve %v", achados)
	}

	for _, letra := range "ANA" {
		letras <- letra
		time.Sleep(10 * time.Millisecond)
	}

	limite := time.After(2 * time.Second)
	for quantidade() != 2 {
		select {
		case <-limite:
			t.Fatalf("esperava ANA e BANANA, obteve %v", achados)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(achados) != 2 || achados[0] != "ANA" || achados[1] != "BANANA" {
		t.Fatalf("detecções inesperadas: %v", achados)
	}
}

// TestRetentaEnvioFalho garante que um match cujo envio falhou não é
// marcado como detectado e é tentado de novo mais tarde.
func TestRetentaEnvioFalho(t *testing.T) {
	armazenamento, err := store.Abrir(t.TempDir() + "/teste.db")
	if err != nil {
		t.Fatalf("abrir banco: %v", err)
	}
	defer armazenamento.Fechar()

	if err := armazenamento.InserirCadastro("ana@exemplo.com", "ANA", time.Now()); err != nil {
		t.Fatalf("inserir ANA: %v", err)
	}

	letras := make(chan rune, 32)
	var mu sync.Mutex
	tentativas := 0
	var notificados []string

	detector := Novo(letras, armazenamento, func(c store.Cadastro) error {
		mu.Lock()
		defer mu.Unlock()
		tentativas++
		if tentativas == 1 {
			return errors.New("smtp fora do ar")
		}
		notificados = append(notificados, c.Palavra)
		return nil
	})
	detector.intervaloRetentativa = 10 * time.Millisecond
	detector.Iniciar()
	defer close(letras)

	for _, letra := range "ANA" {
		letras <- letra
		time.Sleep(10 * time.Millisecond)
	}

	// Letras extras fazem o matcher acordar e tentar de novo.
	for i := 0; i < 10; i++ {
		letras <- 'X'
		time.Sleep(20 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if tentativas < 2 {
		t.Fatalf("esperava nova tentativa, houve %d", tentativas)
	}
	if len(notificados) != 1 || notificados[0] != "ANA" {
		t.Fatalf("esperava ANA notificada uma vez, obteve %v", notificados)
	}
}
