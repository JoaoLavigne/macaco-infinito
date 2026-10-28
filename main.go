package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"macaco-infinito/internal/generator"
)

//go:embed embed/index.html
var indexHTML string

// gerador é a instância única do gerador de letras, compartilhada entre
// o loop de geração e o handler HTTP.
var gerador = generator.NovoGerador(10)

// main é o ponto de entrada do programa.
// No Go, a execução começa sempre nesta função.
//
// Responsabilidades do main neste slice:
//  1. Criar e iniciar o gerador de letras em uma goroutine.
//  2. Iniciar um servidor HTTP na porta 8080 com:
//     - GET / : serve o HTML embutido
//     - GET /letras : retorna as últimas 10 letras em JSON
//  3. Manter o programa rodando (bloquear a execução).
//
// Este arquivo NÃO contém lógica de negócio (letras, email, banco).
// Ele é apenas o "condutor de orquestra" que inicia os componentes.
//
// O HTML da página é embutido no binário via //go:embed, então não
// há arquivos separados para gerenciar na VM.
func main() {
	// Inicia o gerador de letras em uma goroutine paralela ao servidor HTTP.
	// O loop gera uma letra a cada 500ms sem bloquear o servidor.
	go gerador.Iniciar(500 * time.Millisecond)

	// HandleFunc registra um handler para a raiz ("/").
	// Qualquer requisição HTTP para a raiz será tratada pela função abaixo,
	// servindo o HTML embutido com Content-Type: text/html.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// w é o ResponseWriter: usamos para enviar a resposta ao navegador.
		// r é o Request: contém informações da requisição (URL, headers, corpo).
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(indexHTML))
	})

	// GET /letras retorna as últimas 10 letras geradas em formato JSON.
	// O frontend faz polling neste endpoint para atualizar a tela sem recarregar.
	http.HandleFunc("/letras", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		letras := gerador.LetrasRetornar()

		// Converte []rune para []string para serialização JSON limpa.
		letraStrings := make([]string, len(letras))
		for i, l := range letras {
			letraStrings[i] = string(l)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"letras": letraStrings,
		})
	})

	// Porta 8080 é padrão para desenvolvimento local.
	// ListenAndServe bloqueia a execução: o programa não encerra sozinho.
	// Se houver erro (ex: porta já em uso), log.Fatal encerra com mensagem.
	log.Println("Iniciando servidor na porta 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
