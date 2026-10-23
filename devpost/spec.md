---
doc: spec
status: approved
---

# Macaco Infinito — Technical Spec

## How This Works, In Plain Language

O sistema é um único programa (binário Go) que roda em uma máquina virtual na nuvem. Esse programa faz três coisas ao mesmo tempo:

1. Gera letras aleatórias continuamente, uma a cada 0,5 segundo, usando apenas letras de A a Z sem acento. As últimas 10 letras são mostradas na página web.
2. Recebe cadastros de visitantes que preenchem um e-mail e uma palavra de 3 a 15 letras. Cada cadastro é salvo em um arquivo de banco de dados SQLite (um único arquivo no disco).
3. Verifica se alguma palavra cadastrada foi formada pelas letras aleatórias. Quando acontece, envia um e-mail de notificação para o cadastro correspondente.

A página web é um HTML simples servido pelo próprio programa Go. O visitante não precisa instalar nada: basta acessar o endereço da VM no navegador.

Tudo roda como um único processo na VM. Se a VM reiniciar, o sistema se levanta automaticamente graças a uma configuração do systemd.

## The Core Journey Through the System

PRD ref: `prd.md > The Core Journey`.

1. O usuário acessa o endereço da VM no navegador → o servidor HTTP do binário Go devolve a página HTML embutida, mostrando as últimas 10 letras geradas e o formulário de cadastro.
2. O usuário preenche palavra (3–15 letras) e e-mail, e envia → a API do Go recebe, valida o tamanho da palavra, grava no SQLite com status `pendente`, e responde com a mensagem de confirmação incluindo o tempo médio de espera.
3. O usuário fecha o navegador e segue sua rotina → nada mais acontece na tela. O loop de geração de letras continua rodando em segundo plano no mesmo processo.
4. Quando o gerador forma exatamente a palavra cadastrada → o detector de match encontra a palavra no SQLite, muda o status para `notificado`, grava `notificado_em`, e dispara o envio de e-mail via SMTP.
5. O usuário recebe o e-mail com título "O macaco infinito digitou sua palavra" e link de volta para a página inicial.

PRD ref: `prd.md > Features and Behavior > Notificação por e-mail`.

## Stack

- Go 1.22 — linguagem e compilador. Gera um binário estático único, sem dependências de runtime na VM. https://go.dev/doc/
- mattn/go-sqlite3 — driver SQLite para Go. Permite ler/gravar o arquivo de banco local. https://github.com/mattn/go-sqlite3
- gopkg.in/yaml.v3 — parser YAML para ler o arquivo `config.yaml`. https://gopkg.in/yaml.v3
- net/http (biblioteca padrão Go) — servidor HTTP que serve a página e expõe a API de cadastro.
- net/smtp (biblioteca padrão Go) — envio de e-mails via SMTP.
- SQLite — banco de dados embutido em um único arquivo. Nenhum servidor separado. https://www.sqlite.org/
- systemd — gerenciador de processos do Linux na VM. Inicia o binário automaticamente no boot e reinicia se travar.
- Google Cloud Compute Engine e2-micro — VM que hospeda o binário 24/7. https://cloud.google.com/compute

Rationale: Go é escolhido porque compila para um único binário, fácil de deployar e monitorar — um processo para observar, alinhado ao foco SRE do aprendiz. SQLite elimina a necessidade de um servidor de banco separado. systemd garante que o serviço suba sozinho após reinício da VM.

## Where It Runs and How Someone Tries It

Runtime: Linux na VM Google Cloud e2-micro. O binário é compilado localmente e enviado para a VM.

Requisitos de ambiente: Go 1.22 instalado (para compilar), acesso SSH à VM GCP, conta de e-mail com senha de aplicativo configurada para SMTP.

Arquivo de configuração (config.yaml): criado manualmente na VM antes do primeiro uso, contendo credenciais SMTP e caminho do banco:

```yaml
smtp:
  host: "smtp.gmail.com"
  port: 587
  user: "seu-email@gmail.com"
  pass: "senha-de-aplicativo"
db:
  path: "/var/lib/macaco-infinito/dados.db"
alert_email: "seu-email-pessoal@gmail.com"
```

Como iniciar:
1. Compilar: go build -o macaco-infinito .
2. Enviar o binário para a VM: scp macaco-infinito usuario@vm-ip:/home/usuario/
3. Criar o diretório do banco: sudo mkdir -p /var/lib/macaco-infinito && sudo chown usuario:usuario /var/lib/macaco-infinito
4. Copiar config.yaml para ao lado do binário na VM.
5. Configurar systemd (arquivo /etc/systemd/system/macaco-infinito.service):

[Unit]
Description=Macaco Infinito
After=network.target

[Service]
Type=simple
User=usuario
WorkingDirectory=/home/usuario
ExecStart=/home/usuario/macaco-infinito
Restart=always

[Install]
WantedBy=multi-user.target

6. Habilitar e iniciar: sudo systemctl enable macaco-infinito && sudo systemctl start macaco-infinito
7. Reservar um IP estático no GCP para que o link do e-mail nunca quebre após reinícios:
   gcloud compute addresses create macaco-infinito-ip --region=<sua-regiao>
   gcloud compute instances add-access-config <nome-da-vm> --access-config-name="ip-estatico" --address=macaco-infinito-ip --zone=<sua-zona>
8. Acessar: http://<ip-estatico>:8080 no navegador.

## Look and Feel

PRD ref: prd.md > Look and Feel; scope.md > Inspiration & Identity.

- Cabeça de macaco de óculos escuros (apenas cabeça/tronco)
- Fundo espacial em movimento (estrelas, planetas, galáxias)
- Tom: lúdico, científico, misterioso e calmo
- Fonte sem serifa, cores escuras de fundo com texto claro para contraste espacial

Como o frontend está embutido no binário Go, o HTML/CSS/JS são entregues pelo servidor HTTP interno. Para o POC, o visual pode ser uma página única com CSS inline; a animação de fundo pode ser feita com CSS puro ou um pequeno script JS inline.

## Components

### Gerador de Letras (internal/generator)

O que faz: gera letras aleatórias de A a Z continuamente, uma a cada 0,5 segundo. Mantém em memória as últimas 10 letras para exibição na tela.

O que fala com: expõe as letras atuais para o servidor HTTP e para o detector de match.

PRD ref: prd.md > Features and Behavior > Geração de letras.

### Servidor HTTP e API (internal/server)

O que faz: serve a página HTML embutida (//go:embed), que inclui o formulário de cadastro, o array de letras e o botão "Conheça o Experimento do Macaco Infinito" com texto explicativo sobre o teorema. Valida entradas do usuário: normaliza a palavra (remove espaços, converte para maiúsculas), rejeita palavras com números ou caracteres não alfabéticos, valida formato simples do email (contém @ e domínio), verifica se o email já está cadastrado (rejeita com mensagem genérica "e-mail já cadastrado"), e usa consultas parametrizadas (prepared statements) para prevenir SQL injection. Expõe o endpoint de cadastro (POST /cadastro) que recebe palavra + e-mail, grava no banco com status `pendente` e retorna mensagem de confirmação de sucesso com tempo médio de espera. Erros de validação são gravados na tabela `logs` e retornados ao usuário de forma genérica.

O que fala com: grava no internal/store; lê letras atuais do internal/generator para exibição.

PRD ref: prd.md > Screens and Layout; prd.md > Features and Behavior > Cadastro de palavra e e-mail; prd.md > Features and Behavior > Texto explicativo do experimento (botão).

### Armazenamento (internal/store)

O que faz: acessa o arquivo SQLite. Operações: inserir cadastro, listar cadastros pendentes, atualizar status para notificado, gravar notificado_em.

O que fala com: arquivo SQLite local em /var/lib/macaco-infinito/dados.db.

PRD ref: prd.md > States and Boundaries > What persists between sessions.

### Detector de Match (internal/matcher)

O que faz: roda em loop (goroutine), compara as letras geradas com as palavras pendentes no banco usando busca por substring contígua. Quando encontra uma palavra formada (mesmo dentro de outra palavra maior, ex: "ana" dentro de "banana"), aciona o envio de e-mail para todos os cadastros que correspondem. Tanto quem espera por "ana" quanto quem espera por "banana" são notificados.

O que fala com: lê do internal/generator, consulta internal/store, chama internal/mailer.

PRD ref: prd.md > Features and Behavior > Notificação por e-mail.

### Enviador de E-mail (internal/mailer)

O que faz: envia e-mail via SMTP com título O macaco infinito digitou sua palavra, corpo com a palavra, comparação de tempo médio e link para a página inicial. Se o servidor SMTP estiver fora do ar, retenta posteriormente (o cadastro permanece com status pendente).

O que fala com: servidor SMTP externo.

PRD ref: prd.md > Features and Behavior > Notificação por e-mail; prd.md > States and Boundaries > Error state.

## Data Model

Tabela cadastros no SQLite:

| Campo | Tipo | Descrição |
|-------|------|-----------|
| id | INTEGER PRIMARY KEY AUTOINCREMENT | Identificador único |
| email | TEXT NOT NULL UNIQUE | E-mail do cadastro. Um e-mail só pode cadastrar uma palavra. |
| palavra | TEXT NOT NULL | Palavra cadastrada (3–15 letras, A-Z) |
| data | TEXT NOT NULL | Data/hora do cadastro (ISO 8601) |
| status | TEXT NOT NULL | pendente ou notificado |
| notificado_em | TEXT | Data/hora do envio do e-mail (NULL enquanto pendente) |

Tabela logs no SQLite:

| Campo | Tipo | Descrição |
|-------|------|-----------|
| id | INTEGER PRIMARY KEY AUTOINCREMENT | Identificador único |
| timestamp | TEXT NOT NULL | Data/hora do erro (ISO 8601) |
| tipo | TEXT NOT NULL | Tipo do erro (ex: validacao_entrada, sql, smtp, servidor) |
| mensagem | TEXT NOT NULL | Mensagem descritiva do erro para diagnóstico interno |
| origem | TEXT NOT NULL | Origem do erro (ex: cadastro, gerador, mail) |

O que persiste entre sessões: o arquivo SQLite no disco mantém todos os cadastros pendentes e os logs de erro. Se o programa reiniciar, ele retoma de onde parou. O e-mail é enviado assim que o match for detectado, mesmo que seja após um reinício.

Tempo médio de espera: mantido em memória (mapa Go) associando tamanho da palavra (3–15) ao tempo médio esperado em segundos, calculado pela fórmula 26^n / 2 segundos (2 letras por segundo). Exibido no momento do cadastro.

## File Structure

macaco-infinito/
├── main.go
├── config.yaml
├── internal/
│   ├── generator/
│   │   └── generator.go
│   ├── store/
│   │   └── store.go
│   ├── matcher/
│   │   └── matcher.go
│   ├── mailer/
│   │   └── mailer.go
│   └── server/
│       └── server.go
├── embed/
│   └── index.html
├── go.mod
└── go.sum
devpost/

O binário compilado é um único arquivo (macaco-infinito) que contém o HTML embutido. Não há dependências de runtime além do próprio binário e do arquivo config.yaml.

## External Services and Dependencies

### Google Cloud Compute Engine (e2-micro)

Chamada: nenhuma API direta do programa; apenas hospedagem da VM.
Documentação: https://cloud.google.com/compute
Chaves: acesso via SSH (chave SSH configurada no GCP).
Custo: camada gratuita (e2-micro) dentro dos limites do GCP Free Tier.

### SMTP (provedor comum, ex: Gmail)

Chamada: conexão SMTP na porta 587, autenticação LOGIN/PLAIN, envio de mensagens MIME.
Documentação: depende do provedor (ex: https://support.google.com/mail/answer/7126229 para senha de aplicativo Gmail).
Chaves: endereço de e-mail + senha de aplicativo (guardadas em config.yaml).
Rate limits: limites diários do provedor gratuito (ex: ~500/dia no Gmail gratuito), suficientes para o baixo volume de matches deste POC.
Custo: gratuito para uso pessoal limitado.

## Monitoramento e Health Checks

O sistema expõe um endpoint `/health` que retorna o status dos componentes internos em formato JSON. Um goroutine de monitoramento interno executa verificações periódicas e envia alertas para o e-mail pessoal do aprendiz (configurado em `config.yaml`):

1. **Offset de geração de letras**: se o contador de letras geradas não avançou por mais de 120 segundos, alerta de gerador travado.
2. **E-mail pendente**: se há um cadastro com status `pendente` há mais de 20 minutos sem e-mail enviado, alerta de falha de entrega.
3. **Espaço em disco**: se o disco da VM tem menos de 20% livre, alerta de capacidade crítica.
4. **Throughput de letras**: se a quantidade de caracteres gerados por minuto está abaixo de 100 (esperado: 120 por minuto, ou 2 por segundo), alerta de degradação iminente antes de uma queda.

Todos os alertas incluem timestamp e descrição do problema, e são enviados via SMTP para o e-mail configurado.

PRD ref: `prd.md > States and Boundaries > Error state`.

## Important Failure Modes

1. Crash entre match detectado e envio de e-mail: o sistema tenta enviar o e-mail via SMTP antes de marcar o cadastro como notificado no banco. Se o SMTP confirmar sucesso, então atualiza o status para notificado e grava notificado_em. Se crashar entre enviar o e-mail e atualizar o banco, o e-mail sera reenviado no restart (possível duplicata, mas nunca perdido). Se o SMTP falhar, o cadastro permanece como pendente e o matcher tenta novamente no proximo ciclo.
   PRD ref: prd.md > Features and Behavior > Notificacao por e-mail.

2. Match dividido entre letras consecutivas: o matcher mantem um buffer acumulativo das ultimas N letras geradas (N = tamanho da maior palavra cadastrada, maximo 15). A cada nova letra, ela e adicionada ao buffer e o matcher verifica se alguma palavra pendente aparece como substring contigua. Assim, ...an seguido de a resulta em ...ana no buffer e o match e detectado.
   PRD ref: prd.md > Features and Behavior > Geracao de letras.

3. Gerador de letras trava (panic ou erro nao tratado): o loop do gerador usa recuperacao de panic para nao derrubar o processo. Alem disso, o systemd esta configurado com Restart=always, reiniciando o binario automaticamente se o processo morrer. O SQLite persiste os cadastros pendentes, entao o sistema retoma de onde parou apos reinicio.
   PRD ref: prd.md > States and Boundaries > Normal use.

4. Servidor SMTP fora do ar: o e-mail atrasa, mas nao e perdido. O cadastro permanece com status pendente no SQLite. O detector de match tenta novamente no proximo ciclo. O usuario recebe o e-mail com atraso, mas nao perde a notificacao.
   PRD ref: prd.md > States and Boundaries > Error state.

5. VM reinicia: o systemd inicia o binario automaticamente. O SQLite retem os cadastros pendentes. Letras em memoria e o buffer do matcher sao perdidos (recomecam do zero), o que e aceitavel para o POC.
   PRD ref: prd.md > States and Boundaries > Normal use.

6. Palavra muito longa (15 letras): o tempo medio de espera e astronomico; o usuario pode nunca receber o e-mail em uma vida util razoavel. Isso e inerente ao teorema do macaco infinito e e comunicado na tela (tempo medio de espera exibido).
   PRD ref: prd.md > Features and Behavior > Cadastro de palavra e e-mail.

7. Disco da VM enche: o SQLite nao consegue gravar novos cadastros nem atualizar status. A API retorna erro generico ao usuario. O matcher nao consegue marcar como notificado e deixa como pendente para retry posterior. A gravacao na tabela logs tambem pode falhar; nesse caso, o erro e ignorado silenciosamente para evitar loop de falha. Mitigacao: monitorar espaco em disco da VM via health check (alerta abaixo de 20% livre).
   PRD ref: prd.md > States and Boundaries > Error state.

## What Was Simplified and Why

Interface visual embutida e simples em vez de frontend separado com macaco animado e fundo espacial: motivo: manter o POC enxuto; a visual completa fica para depois (scope.md > Later).

Cálculo de tempo médio pré-calculado em memória em vez de cálculo dinâmico por consulta: motivo: simplicidade; apenas 13 tamanhos possíveis (3–15), cabem em um mapa em memória.

Um único binário e uma única VM em vez de microsserviços, load balancer ou múltiplas instâncias: motivo: um processo para monitorar, alinhado ao objetivo de aprendizado SRE do aprendiz.

SQLite local em vez de banco hospedado (PostgreSQL, MySQL, etc.): motivo: zero infraestrutura de banco, um arquivo para backup.

Sem login, sem conta de usuário: motivo: o usuário cadastra e esquece; não precisa de autenticação (prd.md > Non-Goals).

## Decisions and Open Issues

### Decisions tomadas neste spec

Go 1.22 + binário único: escolhido pelo aprendiz para facilitar deploy e monitoramento. Tradeoff: não há separação de frontend/backend em tempo de execução.

SMTP de provedor comum (Gmail/Outlook): escolhido pelo aprendiz para evitar serviços adicionais. Tradeoff: limites de envio e possível entrega em spam, aceitáveis para baixo volume.

systemd para iniciar automaticamente: garante que o serviço suba após reinício da VM. Tradeoff: requer acesso root para configurar inicialmente.

config.yaml ao lado do binário: credenciais SMTP e caminho do banco. Tradeoff: senha em texto claro no disco da VM (aceitável para POC; .gitignore evita subir para o Git).

Prepared statements para todas as queries SQL: o componente `internal/server` usa consultas parametrizadas para gravar e ler do SQLite, prevenindo SQL injection nos campos email e palavra. Tradeoff: código ligeiramente mais verboso do que concatenação de strings, mas essencial para segurança.

Envio de e-mail antes de marcar como notificado: o matcher tenta enviar o e-mail via SMTP primeiro e, somente se o SMTP confirmar sucesso, atualiza o status para `notificado` no banco. Tradeoff: possível e-mail duplicado em caso de crash entre envio e atualização, mas nunca perde a notificação.

Buffer acumulativo de letras no matcher: o matcher mantém as últimas N letras geradas em memória (N = tamanho máximo da palavra cadastrada, 15) e verifica substring contígua a cada nova letra. Tradeoff: uso de memória pequeno e fixo, necessário para detectar matches que cruzam a fronteira entre letras geradas em momentos diferentes.

### Learner uncertainty

O aprendiz ainda não testou se a senha de aplicativo do Gmail/Outlook funciona com net/smtp do Go sem bibliotecas adicionais. Isso deve ser verificado nos primeiros testes de build na VM.

### Open Questions

Qual será o endereço IP público da VM, e será necessário reservar um IP estático para que o link do e-mail não quebre após reinícios? Deve ser verificado antes do deploy.

O net/smtp padrão do Go suporta STARTTLS corretamente com o provedor escolhido? Deve ser verificado durante o build.
