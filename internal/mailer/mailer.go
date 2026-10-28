package mailer

import (
	"fmt"
	"math"
	"net/smtp"
	"strings"
	"time"

	"macaco-infinito/internal/espera"
)

// Config guarda os dados necessários para falar com o servidor SMTP.
type Config struct {
	Host    string
	Port    int
	User    string
	Pass    string
	SiteURL string
}

// Mailer envia as notificações por e-mail.
type Mailer struct {
	config Config
}

// Novo cria um mailer. Porta 587 e localhost:8080 são usados como padrão.
func Novo(config Config) *Mailer {
	if config.Port == 0 {
		config.Port = 587
	}
	if config.SiteURL == "" {
		config.SiteURL = "http://localhost:8080"
	}
	return &Mailer{config: config}
}

// EnviarNotificacao envia o e-mail avisando que a palavra foi formada.
// Retorna erro se o SMTP não estiver configurado ou recusar a mensagem.
func (m *Mailer) EnviarNotificacao(destinatario, palavra string, cadastradoEm, agora time.Time) error {
	if m.config.Host == "" {
		return fmt.Errorf("smtp não configurado")
	}

	remetente := m.config.User
	if remetente == "" {
		remetente = "macaco-infinito@localhost"
	}

	mensagem := montarMensagem(remetente, destinatario, m.corpo(palavra, cadastradoEm, agora))
	endereco := fmt.Sprintf("%s:%d", m.config.Host, m.config.Port)

	var auth smtp.Auth
	if m.config.User != "" {
		auth = smtp.PlainAuth("", m.config.User, m.config.Pass, m.config.Host)
	}

	return smtp.SendMail(endereco, auth, remetente, []string{destinatario}, mensagem)
}

// corpo monta o texto do e-mail comparando a espera real com a média.
func (m *Mailer) corpo(palavra string, cadastradoEm, agora time.Time) string {
	media := espera.Media(len(palavra))
	decorrido := agora.Sub(cadastradoEm).Seconds()
	if decorrido < 0 {
		decorrido = 0
	}

	comparacao := "abaixo"
	if decorrido >= media {
		comparacao = "acima"
	}
	diferenca := math.Abs(decorrido - media)

	return fmt.Sprintf(
		"Parabéns! O macaco infinito digitou sua palavra: %q.\n\n"+
			"Aposto que você nem lembrava mais. Você esperou %s, o que ficou %s %s da média de %s.\n\n"+
			"Acesse nosso site para cadastrar outra palavra: %s\n",
		palavra,
		espera.Formatar(decorrido),
		espera.Formatar(diferenca),
		comparacao,
		espera.Formatar(media),
		m.config.SiteURL,
	)
}

func montarMensagem(remetente, destinatario, corpo string) []byte {
	cabecalhos := []string{
		"From: Macaco Infinito <" + remetente + ">",
		"To: " + destinatario,
		"Subject: O macaco infinito digitou sua palavra",
		"MIME-Version: 1.0",
		`Content-Type: text/plain; charset="UTF-8"`,
		"Content-Transfer-Encoding: 8bit",
	}

	corpoCRLF := strings.ReplaceAll(corpo, "\n", "\r\n")
	return []byte(strings.Join(cabecalhos, "\r\n") + "\r\n\r\n" + corpoCRLF)
}
