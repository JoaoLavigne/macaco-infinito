package matcher

import (
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

	detector := Novo(letras, armazenamento, func(c store.Cadastro) {
		mu.Lock()
		achados = append(achados, c.Palavra)
		mu.Unlock()
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
