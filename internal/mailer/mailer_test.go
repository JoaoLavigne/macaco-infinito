package mailer

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// servidorFake implementa o mínimo do protocolo SMTP para receber uma
// mensagem em teste, sem depender de credenciais reais.
func servidorFake(t *testing.T) (string, int, <-chan string) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ouvir: %v", err)
	}

	mensagens := make(chan string, 1)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		defer ln.Close()

		leitor := bufio.NewReader(conn)
		escritor := bufio.NewWriter(conn)
		responder := func(texto string) {
			escritor.WriteString(texto + "\r\n")
			escritor.Flush()
		}

		responder("220 localhost ESMTP")
		for {
			linha, err := leitor.ReadString('\n')
			if err != nil {
				return
			}
			comando := strings.ToUpper(strings.TrimRight(linha, "\r\n"))

			switch {
			case strings.HasPrefix(comando, "EHLO"), strings.HasPrefix(comando, "HELO"):
				responder("250-localhost")
				responder("250 HELP")
			case strings.HasPrefix(comando, "MAIL FROM"), strings.HasPrefix(comando, "RCPT TO"):
				responder("250 OK")
			case strings.HasPrefix(comando, "DATA"):
				responder("354 Fim dos dados com <CR><LF>.<CR><LF>")
				var corpo strings.Builder
				for {
					l, err := leitor.ReadString('\n')
					if err != nil {
						return
					}
					if l == ".\r\n" {
						break
					}
					corpo.WriteString(l)
				}
				responder("250 OK")
				mensagens <- corpo.String()
			case strings.HasPrefix(comando, "QUIT"):
				responder("221 Bye")
				return
			default:
				responder("250 OK")
			}
		}
	}()

	host, portaTexto, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("separar endereço: %v", err)
	}
	var porta int
	fmt.Sscanf(portaTexto, "%d", &porta)

	return host, porta, mensagens
}

func TestEnviarNotificacao(t *testing.T) {
	host, porta, mensagens := servidorFake(t)

	enviador := Novo(Config{Host: host, Port: porta, SiteURL: "http://exemplo.com"})
	cadastradoEm := time.Now().Add(-time.Hour)

	if err := enviador.EnviarNotificacao("eu@exemplo.com", "GATO", cadastradoEm, time.Now()); err != nil {
		t.Fatalf("enviar: %v", err)
	}

	select {
	case mensagem := <-mensagens:
		trechos := []string{
			"Subject: O macaco infinito digitou sua palavra",
			"GATO",
			"http://exemplo.com",
			"abaixo",
		}
		for _, trecho := range trechos {
			if !strings.Contains(mensagem, trecho) {
				t.Errorf("mensagem sem %q:\n%s", trecho, mensagem)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("servidor fake não recebeu a mensagem")
	}
}

func TestEnviarSemSMTPConfigurado(t *testing.T) {
	enviador := Novo(Config{})
	err := enviador.EnviarNotificacao("eu@exemplo.com", "GATO", time.Now(), time.Now())
	if err == nil {
		t.Fatal("esperava erro quando o SMTP não está configurado")
	}
}
