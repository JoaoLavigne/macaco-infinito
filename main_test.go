package main

import (
	"errors"
	"testing"
	"time"

	"macaco-infinito/internal/store"
)

type notificadorFake struct {
	err      error
	chamadas int
}

func (f *notificadorFake) EnviarNotificacao(destinatario, palavra string, cadastradoEm, agora time.Time) error {
	f.chamadas++
	return f.err
}

func cadastroPendente(t *testing.T) (*store.Store, store.Cadastro) {
	t.Helper()

	armazenamento, err := store.Abrir(t.TempDir() + "/teste.db")
	if err != nil {
		t.Fatalf("abrir banco: %v", err)
	}
	t.Cleanup(func() { armazenamento.Fechar() })

	if err := armazenamento.InserirCadastro("eu@exemplo.com", "GATO", time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("inserir cadastro: %v", err)
	}

	pendentes, err := armazenamento.ListarPendentes()
	if err != nil || len(pendentes) != 1 {
		t.Fatalf("esperava 1 pendente, obteve %d (erro: %v)", len(pendentes), err)
	}

	return armazenamento, pendentes[0]
}

func TestNotificarMarcaSomenteAposSucesso(t *testing.T) {
	armazenamento, cadastro := cadastroPendente(t)
	enviador := &notificadorFake{}

	if err := notificar(armazenamento, enviador)(cadastro); err != nil {
		t.Fatalf("notificar: %v", err)
	}
	if enviador.chamadas != 1 {
		t.Fatalf("esperava 1 envio, houve %d", enviador.chamadas)
	}

	pendentes, err := armazenamento.ListarPendentes()
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	if len(pendentes) != 0 {
		t.Fatalf("cadastro deveria estar notificado, ainda pendente: %v", pendentes)
	}
}

func TestNotificarMantemPendenteSeEnvioFalhar(t *testing.T) {
	armazenamento, cadastro := cadastroPendente(t)
	enviador := &notificadorFake{err: errors.New("smtp fora do ar")}

	if err := notificar(armazenamento, enviador)(cadastro); err == nil {
		t.Fatal("esperava erro de envio")
	}

	pendentes, err := armazenamento.ListarPendentes()
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	if len(pendentes) != 1 {
		t.Fatalf("cadastro deveria continuar pendente, obteve %d", len(pendentes))
	}
}
