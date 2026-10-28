---
doc: checklist
status: approved
---

# Build Checklist

Build mode: learn — cada slice é explicado em detalhe (funções, lógica, arquitetura, persistência) antes de avançar. O learner testa cada comportamento concluído.

## Slices

- [x] **1. Scaffold Go + dependências + config.yaml + main.go que compila e roda**
  Becomes usable: Um projeto Go que compila, lê o config.yaml e sobe um servidor HTTP vazio na porta 8080.
  Why now: Prova que a base do binário único funciona antes de adicionar comportamento.
  PRD ref: `prd.md > What We're Building`
  Spec ref: `spec.md > Stack`, `spec.md > File Structure`
  Build: Criar go.mod, instalar dependências (mattn/go-sqlite3, gopkg.in/yaml.v3), criar config.yaml.example, main.go com servidor HTTP vazio.
  Verify (mechanical): `go build -o macaco-infinito .` compila sem erros; `./macaco-infinito` inicia e responde na porta 8080.
  Learner check: Rode o binário localmente, acesse http://localhost:8080 e confirme que sobe sem erros.
  Commit: `Scaffold do projeto Go com servidor vazio`

- [x] **2. Servidor HTTP com página HTML (formulário, array de letras, botão do teorema)**
  Becomes usable: A página inicial renderiza no navegador com o formulário de cadastro, espaço para as letras e o botão "Conheça o Experimento do Macaco Infinito".
  Why now: Define a superfície de interação que todos os slices seguintes vão preencher.
  PRD ref: `prd.md > Screens and Layout`, `prd.md > Look and Feel`
  Spec ref: `spec.md > Components > Servidor HTTP e API`, `spec.md > Look and Feel`
  Build: HTML embutido via `//go:embed` com formulário (palavra, email, botão enviar), div para letras, botão do teorema com texto explicativo.
  Verify (mechanical): Acessar http://localhost:8080 e ver a página completa carregar.
  Learner check: Abra o navegador, confirme que o formulário e o botão do teorema aparecem como imaginado.
  Commit: `Página inicial com formulário e botão do teorema`

- [x] **3. Gerador de letras (1 letra a cada 0,5s, últimas 10 na tela)**
  Becomes usable: Letras aleatórias de A a Z aparecem na tela continuamente, atualizando a cada 0,5s.
  Why now: O núcleo do teorema do macaco infinito — sem letras, não há experimento.
  PRD ref: `prd.md > Features and Behavior > Geração de letras`
  Spec ref: `spec.md > Components > Gerador de Letras`, `spec.md > How This Works`
  Build: Goroutine que gera letra aleatória a cada 500ms, mantém slice das últimas 10, expõe via endpoint HTTP para o frontend atualizar.
  Verify (mechanical): Acesse a página e confirme que novas letras aparecem a cada 0,5s sem recarregar.
  Learner check: Observe as letras passando e confirme que são apenas A-Z sem acento/espaço.
  Commit: `Gerador de letras aleatórias contínuo`

- [ ] **4. Cadastro de palavra+email com validações + notificação de sucesso**
  Becomes usable: O usuário preenche palavra (3–15 letras) e email, envia, e recebe confirmação de sucesso com tempo médio de espera.
  Why now: É o ponto de entrada do usuário no experimento — sem cadastro, não há notificação futura.
  PRD ref: `prd.md > Features and Behavior > Cadastro de palavra e e-mail`
  Spec ref: `spec.md > Components > Servidor HTTP e API`, `spec.md > Data Model`
  Build: Endpoint POST /cadastro que normaliza palavra (remove espaços, uppercase), valida tamanho 3–15, rejeita não-alfabéticos, valida email (contém @), verifica UNIQUE constraint, usa prepared statements, grava no SQLite, retorna mensagem de sucesso com tempo médio pré-calculado.
  Verify (mechanical): Envie palavra válida e email; confirme mensagem de sucesso. Tente palavra com número; confirme rejeição genérica. Tente mesmo email duas vezes; confirme "email já cadastrado".
  Learner check: Preencha o formulário e confirme que a mensagem de sucesso aparece como esperado.
  Commit: `Cadastro de palavra e email com validações`

- [ ] **5. Armazenamento SQLite (inserir, listar pendentes, tabela de logs)**
  Becomes usable: Cadastros são salvos em arquivo SQLite no disco e persistem entre reinícios; erros são gravados na tabela de logs.
  Why now: Persistência é o primeiro lugar onde o modelo de dados pode estar errado — descobrir agora é barato.
  PRD ref: `prd.md > States and Boundaries > What persists between sessions`
  Spec ref: `spec.md > Data Model`, `spec.md > Components > Armazenamento`
  Build: Tabelas `cadastros` (id, email UNIQUE, palavra, data, status, notificado_em) e `logs` (id, timestamp, tipo, mensagem, origem). Funções para inserir, listar pendentes, atualizar status, gravar log.
  Verify (mechanical): Cadastre uma palavra, encerre o programa, rode novamente, confirme que o cadastro ainda está no banco.
  Learner check: Cadastre algo, pare o servidor, inicie novamente, confirme que seus dados estão lá.
  Commit: `Persistência em SQLite com tabelas cadastros e logs`

- [ ] **6. Detector de match (buffer acumulativo, substring contígua, "ana" dentro de "banana")**
  Becomes usable: Quando as letras geradas formam uma palavra cadastrada (mesmo dentro de outra maior), o sistema detecta o match.
  Why now: É o coração do teorema — detectar que o acaso formou algo significativo.
  PRD ref: `prd.md > Features and Behavior > Notificação por e-mail`
  Spec ref: `spec.md > Components > Detector de Match`, `spec.md > Important Failure Modes` (match dividido entre letras)
  Build: Goroutine que mantém buffer das últimas 15 letras, a cada nova letra verifica se alguma palavra pendente aparece como substring contígua no buffer.
  Verify (mechanical): Cadastre "ana", deixe o gerador rodar até aparecer "ana" (mesmo dentro de outra sequência), confirme detecção.
  Learner check: Cadastre "sol" e deixe rodar até aparecer dentro de "solstício" — confirme que dispara.
  Commit: `Detector de match por substring contígua`

- [ ] **7. Envio de email via SMTP (antes de marcar como notificado)**
  Becomes usable: Quando o match é detectado, o usuário recebe o e-mail de notificação.
  Why now: A notificação por email é o kernel do projeto — sem ela, é só um gerador aleatório.
  PRD ref: `prd.md > Features and Behavior > Notificação por e-mail`
  Spec ref: `spec.md > Components > Enviador de E-mail`, `spec.md > Important Failure Modes` (crash entre match e envio)
  Build: Componente mailer que envia email via SMTP com título e corpo definidos no PRD. Tenta enviar PRIMEIRO, só marca como `notificado` no banco se o SMTP confirmar sucesso.
  Verify (mechanical): Cadastre uma palavra curta (ex: "gato"), aguarde ou force match, confirme recebimento do email.
  Learner check: Confirme que o email chega com o texto correto (título, corpo, link de volta).
  Commit: `Envio de notificação por email via SMTP`

- [ ] **8. Health checks `/health` + monitoramento (offset, email pendente, disco, throughput)**
  Becomes usable: O endpoint `/health` retorna status JSON e o sistema envia alertas por email quando algo falha.
  Why now: Você quer manter o projeto no ar com foco SRE — monitoramento é essencial.
  PRD ref: `prd.md > States and Boundaries > Error state`
  Spec ref: `spec.md > Monitoramento e Health Checks`
  Build: Endpoint GET /health retornando JSON com status dos componentes. Goroutine de monitoramento verifica: letras não avançaram por 120s, email pendente por 20min, disco abaixo de 20%, throughput abaixo de 100/min. Envia alerta para alert_email do config.yaml.
  Verify (mechanical): Acesse http://localhost:8080/health e confirme JSON retornado.
  Learner check: Confirme que o endpoint /health responde e entenda o que cada campo significa.
  Commit: `Health checks e monitoramento com alertas`

- [ ] **9. systemd + deploy na VM GCP com IP estático**
  Becomes usable: O sistema está rodando na VM do Google Cloud, acessível via IP estático, reiniciando automaticamente após reboot.
  Why now: O objetivo é manter o projeto no ar a longo prazo — deploy é o último passo do POC.
  PRD ref: `prd.md > What We're Building`
  Spec ref: `spec.md > Where It Runs and How Someone Tries It`
  Build: Compilar para Linux, enviar para VM, criar serviço systemd, reservar IP estático no GCP, configurar firewall para porta 8080.
  Verify (mechanical): Acesse o IP estático da VM no navegador e confirme que o sistema funciona.
  Learner check: Acesse a URL do deploy e confirme que o sistema está no ar.
  Commit: `Deploy na VM GCP com systemd e IP estático`

## Hands-on Checkpoints

- [ ] Early usable behavior explored — após o slice 4 (cadastro funcionando)
- [ ] Final kick-the-tires exploration and feedback completed — após o slice 9 (deploy)

## Final Review

- [ ] Final review complete — feedback resolved and learner confirms ready to ship

## Code Tour and App Map

- [ ] Learning activity complete — guided route through code
- [ ] `devpost/app-map.html` generated from finished code, checked, and shown

## Revisions

<!-- Leave this section empty when you write the checklist. `5-build` appends a bullet here when implementation contradicts an assumption in the plan. -->