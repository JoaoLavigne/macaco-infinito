package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"

	"macaco-infinito/internal/espera"
	"macaco-infinito/internal/generator"
	"macaco-infinito/internal/mailer"
	"macaco-infinito/internal/matcher"
	"macaco-infinito/internal/store"
)

//go:embed embed/index.html
var indexHTML string

var gerador = generator.NovoGerador(10)

type configuracao struct {
	SMTP struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"smtp"`
	DB struct {
		Path string `yaml:"path"`
	} `yaml:"db"`
	SiteURL    string `yaml:"site_url"`
	AlertEmail string `yaml:"alert_email"`
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

	enviador := mailer.Novo(mailer.Config{
		Host:    config.SMTP.Host,
		Port:    config.SMTP.Port,
		User:    config.SMTP.User,
		Pass:    config.SMTP.Pass,
		SiteURL: config.SiteURL,
	})

	go gerador.Iniciar(500 * time.Millisecond)

	detector := matcher.Novo(gerador.NovasLetras(), armazenamento, notificar(armazenamento, enviador))
	detector.Iniciar()

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
			registrarErro(armazenamento, "validacao_entrada", "corpo JSON inválido: "+err.Error(), "cadastro")
			responderJSON(w, http.StatusBadRequest, map[string]string{
				"erro": "Entrada inválida. Verifique a palavra e o e-mail.",
			})
			return
		}

		palavra := normalizarPalavra(pedido.Palavra)
		email := strings.ToLower(strings.TrimSpace(pedido.Email))

		if len(palavra) < 3 || len(palavra) > 15 {
			registrarErro(armazenamento, "validacao_entrada", "palavra fora do tamanho permitido", "cadastro")
			responderJSON(w, http.StatusBadRequest, map[string]string{
				"erro": "A palavra deve ter entre 3 e 15 letras.",
			})
			return
		}

		if !palavraAlfabetica(palavra) || !emailValido(email) {
			registrarErro(armazenamento, "validacao_entrada", "palavra ou e-mail inválido", "cadastro")
			responderJSON(w, http.StatusBadRequest, map[string]string{
				"erro": "Entrada inválida. Verifique a palavra e o e-mail.",
			})
			return
		}

		err := armazenamento.InserirCadastro(email, palavra, time.Now())
		if err == store.ErrEmailDuplicado {
			registrarErro(armazenamento, "validacao_entrada", "e-mail já cadastrado", "cadastro")
			responderJSON(w, http.StatusConflict, map[string]string{
				"erro": "e-mail já cadastrado",
			})
			return
		}
		if err != nil {
			log.Printf("Erro ao gravar cadastro: %v", err)
			registrarErro(armazenamento, "sql", err.Error(), "cadastro")
			responderJSON(w, http.StatusInternalServerError, map[string]string{
				"erro": "Erro interno. Tente novamente mais tarde.",
			})
			return
		}

		mensagem := fmt.Sprintf(
			"Palavra cadastrada. Será notificado quando o macaco digitar sua palavra. Boa sorte! O tempo médio de espera é de %s.",
			espera.Formatar(espera.Media(len(palavra))),
		)

		responderJSON(w, http.StatusOK, map[string]string{
			"mensagem": mensagem,
		})
	}
}

// notificador é a parte do mailer que o fluxo de notificação precisa.
type notificador interface {
	EnviarNotificacao(destinatario, palavra string, cadastradoEm, agora time.Time) error
}

// notificar envia o e-mail PRIMEIRO e só marca o cadastro como notificado
// se o SMTP confirmar o sucesso. Se o envio falhar, o cadastro continua
// pendente para o matcher tentar de novo.
func notificar(armazenamento *store.Store, enviador notificador) func(store.Cadastro) error {
	return func(c store.Cadastro) error {
		agora := time.Now()

		if err := enviador.EnviarNotificacao(c.Email, c.Palavra, c.Data, agora); err != nil {
			log.Printf("Erro ao enviar e-mail para %s: %v", c.Email, err)
			registrarErro(armazenamento, "smtp", err.Error(), "mail")
			return err
		}

		if err := armazenamento.AtualizarNotificado(c.ID, agora); err != nil {
			log.Printf("E-mail enviado, mas erro ao atualizar cadastro %d: %v", c.ID, err)
			registrarErro(armazenamento, "sql", err.Error(), "matcher")
			return err
		}

		log.Printf("E-mail enviado para %s (palavra %s)", c.Email, c.Palavra)
		return nil
	}
}

// registrarErro grava um diagnóstico na tabela logs sem interromper o fluxo
// principal. Se a gravação falhar, o erro é ignorado para evitar loop de falha.
func registrarErro(armazenamento *store.Store, tipo, mensagem, origem string) {
	if err := armazenamento.RegistrarLog(tipo, mensagem, origem); err != nil {
		log.Printf("Erro ao gravar log: %v", err)
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

func carregarConfiguracao() configuracao {
	var config configuracao
	config.DB.Path = "dados.db"
	config.SiteURL = "http://localhost:8080"

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
