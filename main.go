package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"

	"macaco-infinito/internal/generator"
	"macaco-infinito/internal/store"
)

//go:embed embed/index.html
var indexHTML string

var gerador = generator.NovoGerador(10)

var temposEspera = calcularTemposEspera()

type configuracao struct {
	DB struct {
		Path string `yaml:"path"`
	} `yaml:"db"`
}

type pedidoCadastro struct {
	Palavra string `json:"palavra"`
	Email   string `json:"email"`
}

func main() {
	config := carregarConfiguracao()

	armazenamento, err := store.Abrir(config.DB.Path)
	if err != nil {
		log.Fatalf("Erro ao abrir banco de dados: %v", err)
	}
	defer armazenamento.Fechar()

	go gerador.Iniciar(500 * time.Millisecond)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(indexHTML))
	})

	http.HandleFunc("/letras", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		letras := gerador.LetrasRetornar()
		letraStrings := make([]string, len(letras))
		for i, l := range letras {
			letraStrings[i] = string(l)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"letras": letraStrings,
		})
	})

	http.HandleFunc("/cadastro", cadastrar(armazenamento))

	log.Println("Iniciando servidor na porta 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func cadastrar(armazenamento *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			responderJSON(w, http.StatusMethodNotAllowed, map[string]string{
				"erro": "Método não permitido.",
			})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var pedido pedidoCadastro
		if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
			responderJSON(w, http.StatusBadRequest, map[string]string{
				"erro": "Entrada inválida. Verifique a palavra e o e-mail.",
			})
			return
		}

		palavra := normalizarPalavra(pedido.Palavra)
		email := strings.ToLower(strings.TrimSpace(pedido.Email))

		if len(palavra) < 3 || len(palavra) > 15 {
			responderJSON(w, http.StatusBadRequest, map[string]string{
				"erro": "A palavra deve ter entre 3 e 15 letras.",
			})
			return
		}

		if !palavraAlfabetica(palavra) || !emailValido(email) {
			responderJSON(w, http.StatusBadRequest, map[string]string{
				"erro": "Entrada inválida. Verifique a palavra e o e-mail.",
			})
			return
		}

		err := armazenamento.InserirCadastro(email, palavra, time.Now())
		if err == store.ErrEmailDuplicado {
			responderJSON(w, http.StatusConflict, map[string]string{
				"erro": "e-mail já cadastrado",
			})
			return
		}
		if err != nil {
			log.Printf("Erro ao gravar cadastro: %v", err)
			responderJSON(w, http.StatusInternalServerError, map[string]string{
				"erro": "Erro interno. Tente novamente mais tarde.",
			})
			return
		}

		mensagem := fmt.Sprintf(
			"Palavra cadastrada. Será notificado quando o macaco digitar sua palavra. Boa sorte! O tempo médio de espera é de %s.",
			temposEspera[len(palavra)],
		)

		responderJSON(w, http.StatusOK, map[string]string{
			"mensagem": mensagem,
		})
	}
}

func responderJSON(w http.ResponseWriter, status int, corpo map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(corpo)
}

func normalizarPalavra(palavra string) string {
	semEspacos := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, palavra)
	return strings.ToUpper(semEspacos)
}

func palavraAlfabetica(palavra string) bool {
	for _, r := range palavra {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func emailValido(email string) bool {
	if strings.Count(email, "@") != 1 {
		return false
	}

	partes := strings.SplitN(email, "@", 2)
	if partes[0] == "" || partes[1] == "" {
		return false
	}
	if !strings.Contains(partes[1], ".") {
		return false
	}

	endereco, err := mail.ParseAddress(email)
	return err == nil && endereco.Address == email
}

func calcularTemposEspera() map[int]string {
	tempos := make(map[int]string, 13)
	for tamanho := 3; tamanho <= 15; tamanho++ {
		segundos := math.Pow(26, float64(tamanho)) / 2
		tempos[tamanho] = formatarDuracao(segundos)
	}
	return tempos
}

func formatarDuracao(segundos float64) string {
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

func carregarConfiguracao() configuracao {
	var config configuracao
	config.DB.Path = "dados.db"

	dados, err := os.ReadFile("config.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return config
		}
		log.Fatalf("Erro ao ler config.yaml: %v", err)
	}

	if err := yaml.Unmarshal(dados, &config); err != nil {
		log.Fatalf("Erro ao interpretar config.yaml: %v", err)
	}

	if config.DB.Path == "" {
		config.DB.Path = "dados.db"
	}

	return config
}
