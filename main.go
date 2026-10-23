package main

import (
	"fmt"
	"log"
	"net/http"
)

// main é o ponto de entrada do programa.
// No Go, a execução começa sempre nesta função.
//
// Responsabilidades do main neste slice:
//   1. Ler o arquivo de configuração (config.yaml) - se existir.
//   2. Iniciar um servidor HTTP simples na porta 8080.
//   3. Manter o programa rodando (bloquear a execução).
//
// Este arquivo NÃO contém lógica de negócio (letras, email, banco).
// Ele é apenas o "condutor de orquestra" que inicia os componentes.
func main() {
	// Inicia o servidor HTTP.
	// HandleFunc registra um handler para a raiz ("/").
	// Qualquer requisição HTTP para a raiz será tratada pela função anônima abaixo.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// w é o ResponseWriter: usamos para enviar a resposta ao navegador.
		// r é o Request: contém informações da requisição (URL, headers, corpo).
		fmt.Fprintf(w, "Macaco Infinito - online\n")
	})

	// Porta 8080 é padrão para desenvolvimento local.
	// ListenAndServe bloqueia a execução: o programa não encerra sozinho.
	// Se houver erro (ex: porta já em uso), log.Fatal encerra com mensagem.
	log.Println("Iniciando servidor na porta 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}