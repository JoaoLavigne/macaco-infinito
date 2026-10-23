---
doc: prd
status: approved
---

# Macaco Infinito — Product Requirements

One line: Uma simulação do teorema do macaco infinito que gera letras aleatórias e avisa por e-mail quando uma palavra cadastrada por um usuário é formada.

Source: `scope.md > The Unique Kernel`, `scope.md > Who It's For`, `scope.md > The Core Loop`, `scope.md > What "Working" Looks Like`, `scope.md > The POC Boundary`.

## The Core Journey

1. O usuário chega à página e vê o macaco de óculos escuros sobre um fundo espacial em movimento, com letras aleatórias passando na tela (uma a cada 0,5s, mostrando as 10 últimas letras; a nova apaga a primeira).
2. O usuário digita uma palavra no campo "insira uma palavra" e preenche seu e-mail no campo de notificação.
3. Ao enviar, aparece na tela: "Palavra cadastrada. Será notificado quando o macaco digitar sua palavra. Boa sorte! O tempo médio de espera é de X."
4. O usuário fecha a página e segue sua rotina. Nada mais acontece na tela.
5. Quando o gerador de letras formar exatamente a palavra cadastrada, o usuário recebe um e-mail com o título "O macaco infinito digitou sua palavra" e o corpo: "Parabéns! O macaco infinito digitou sua palavra: 'palavra'. Aposto que você nem lembrava mais. Foi X tempo acima da média ou Y tempo abaixo da média. Acesse nosso site para cadastrar outra palavra <link>."
6. O link leva de volta à página inicial, onde o usuário pode cadastrar outra palavra.

## Screens and Layout

Apenas uma superfície: a página inicial. Ela contém:
- Cabeça de macaco de óculos escuros (não o corpo todo)
- Fundo espacial em movimento (estrelas, planetas, galáxias)
- Array de letras: as 10 últimas letras geradas, atualizadas a cada 0,5s
- Campo de texto "insira uma palavra"
- Campo de e-mail para notificação
- Botão de envio
- Área de mensagem de confirmação após o cadastro
- Botão "Conheça o Experimento do Macaco Infinito" que, ao ser clicado, exibe o texto explicativo sobre o teorema do macaco infinito e o propósito do experimento:
  > "O teorema do macaco infinito diz que, se um macaco apertasse teclas ao acaso por um tempo infinito, ele acabaria digitando qualquer texto que você imaginar, inclusive as obras completas de Shakespeare.
  >
  > Por que isso é verdade? Qualquer sequência específica de letras tem uma chance pequena, mas não nula, de sair. Se você tenta infinitas vezes, uma chance pequena acaba acontecendo. É como jogar um dado: tirar seis é improvável numa jogada, mas se você joga sem parar, o seis aparece.
  >
  > O experimento tem o objetivo de mostrar na prática que quando falamos de infinito não existe "se vai acontecer" mas sim "quando".
  >
  > Cadastre o seu email e palavra e um dia a mágica vai acontecer."

Não há telas de login, dashboard, histórico ou acompanhamento em tempo real.

## Look and Feel

- Macaco de óculos escuros, apenas cabeça ou tronco
- Fundo espacial em movimento (estrelas, planetas, galáxias) remete ao infinito
- Tom: lúdico, científico, misterioso e calmo

Source: `scope.md > Inspiration & Identity`.

## Features and Behavior

### Geração de letras

O sistema gera letras aleatórias continuamente, uma a cada 0,5s. Na tela são mostradas apenas as 10 últimas letras; cada nova letra apaga a primeira para dar espaço.

- Acceptance criterion: As letras são visíveis na página pública sem necessidade de cadastro.
- Acceptance criterion: A cada 0,5s uma nova letra aparece e a mais antiga das 10 é removida.

### Cadastro de palavra e e-mail

O usuário insere uma palavra e um e-mail e envia.

- Acceptance criterion: Após o envio, aparece a mensagem "Palavra cadastrada. Será notificado quando o macaco digitar sua palavra. Boa sorte! O tempo médio de espera é de X."
- Acceptance criterion: Um mesmo e-mail só pode cadastrar uma palavra (evita ataques/spam).
- Acceptance criterion: Um mesmo e-mail só pode cadastrar uma palavra. Se o e-mail já estiver cadastrado, o cadastro é rejeitado com mensagem genérica "e-mail já cadastrado".
- Acceptance criterion: Após o cadastro bem-sucedido, aparece notificação de sucesso na tela.
- Acceptance criterion: A palavra deve ter entre 3 e 15 letras; palavras menores ou maiores são rejeitadas com mensagem clara na tela.
- Acceptance criterion: A entrada da palavra ignora espaços em branco e não diferencia maiúsculas de minúsculas (ex: " Casa " é tratado como "CASA").
- Acceptance criterion: Se a palavra contiver números ou caracteres não alfabéticos, o cadastro é rejeitado com mensagem genérica de erro e o usuário é solicitado a inserir novamente.
- Acceptance criterion: O e-mail é validado (verificação de formato simples: contém @ e domínio). E-mails inválidos são rejeitados com mensagem genérica de erro.
- Acceptance criterion: Ambos os campos são tratados com consultas parametrizadas (prepared statements) para prevenir SQL injection.
- Acceptance criterion: Erros de validação são armazenados em uma tabela de log e retornados ao usuário de forma genérica, sem expor detalhes internos do sistema.
- Acceptance criterion: O botão "Conheça o Experimento do Macaco Infinito" está visível na página principal e, ao ser clicado, exibe o texto explicativo sobre o teorema do macaco infinito e o propósito do experimento.

### Notificação por e-mail

Quando o gerador forma a palavra cadastrada, o sistema envia um e-mail.

- Acceptance criterion: O e-mail tem o título "O macaco infinito digitou sua palavra".
- Acceptance criterion: O corpo inclui a palavra cadastrada, uma comparação com o tempo médio (acima ou abaixo) e um link para a página inicial.
- Acceptance criterion: Nada acontece na tela no momento da notificação (o usuário pode não estar na página).
- Acceptance criterion: Se o servidor de e-mail estiver fora do ar, o e-mail atrasa mas não é perdido; as informações são guardadas até o serviço restabelecer.

## States and Boundaries

- **First use:** Usuário vê o macaco, as letras gerando e os campos de cadastro. Nada está cadastrado ainda.
- **Normal use:** Usuário cadastra palavra + e-mail, recebe confirmação com tempo médio, fecha a página.
- **Empty state:** Nenhuma palavra cadastrada; as letras continuam gerando normalmente.
- **Success state:** Palavra formada pelo gerador → e-mail enviado ao usuário.
- **Error state:** Servidor de e-mail indisponível → e-mail atrasado, não perdido; informações guardadas até restabelecer.
- **What persists between sessions:** A palavra cadastrada e o e-mail associado são mantidos até que o match ocorra.
- **What disappears:** Nada é mostrado ao usuário entre o cadastro e a notificação. As letras antigas são substituídas nas 10 últimas.
- **Boundaries:** Um e-mail = uma palavra. A palavra deve ter entre 3 e 15 letras. Sem login, sem dashboard, sem acompanhamento em tempo real, sem push/SMS.

## Product Decisions

- Mesma palavra pode ser cadastrada por vários usuários diferentes — não é um problema.
- Um e-mail só cadastra uma palavra — evita ataques e spam.
- O link do e-mail leva para a página inicial, não para uma página específica.
- O tempo médio de espera é exibido após o cadastro; o cálculo exato é definido no `4-spec`.
- Notificações são apenas por e-mail, não há push nem SMS.

## What We're Building

O proof of concept deve:
- Exibir o macaco de óculos escuros sobre fundo espacial animado
- Gerar letras aleatórias continuamente (1 a cada 0,5s), mostrando as 10 últimas na tela
- Permitir cadastro de uma palavra + e-mail por visitante
- Exibir confirmação com tempo médio de espera após o cadastro
- Enviar e-mail de notificação quando a palavra for formada, com link para a página inicial
- Manter cadastros pendentes mesmo se o servidor de e-mail falhar temporariamente
- Não armazenar letras que não formem palavra de usuário nem palavra do dicionário popular (desse último, apenas no escopo futuro)

## Deferred From the POC

- Interface visual completa com macaco e fundo espacial animado
- Cálculo e exibição automática do tempo médio de espera baseado no tamanho da palavra
- Dicionário de palavras populares em inglês, português e espanhol
- Armazenamento de letras que formam palavras do dicionário (não apenas de usuários)

## Possible Later Enhancements

- Dicionário EN/PT/ES para também notificar quando palavras populares são formadas
- Estatísticas públicas de quantas palavras foram formadas
- Permitir múltiplas palavras por e-mail
- Notificações por push ou SMS

## Non-Goals

- Armazenar todas as letras geradas (apenas as que derem match) — motivo: escopo enxuto, foco no kernel.
- Conta de usuário com senha ou perfil — motivo: o usuário cadastra e esquece; não precisa de conta.
- Dashboard para acompanhar progresso — motivo: a pessoa deve esquecer do experimento até receber o e-mail.
- Histórico de tentativas ou contador de letras geradas — motivo: não adiciona valor ao kernel.

## Open Questions

- Qual é o alfabeto exato usado (26 letras sem acento, ou inclui caracteres acentuados do português/espanhol)? Deve ser respondido antes do `4-spec`.
- Como o tempo médio de espera é calculado e exibido exatamente? Pode ser resolvido no `4-spec`.

