---
doc: scope
status: approved
---

# Macaco Infinito

One line: Uma simulação do teorema do macaco infinito que gera letras aleatórias e avisa por e-mail quando uma palavra cadastrada por um usuário é formada.

## The Unique Kernel

A conexão pessoal entre a palavra escolhida por um usuário e o momento em que o acaso a forma, depois de um tempo de espera — a notificação por e-mail que torna a aleatoriedade significativa. Sem isso, é só um gerador aleatório de letras.

## Who It's For

Uma pessoa interessada em ciência e experimentos, que gosta de ver ideias abstratas ganharem vida por simulação. Alguém que cadastra uma palavra, esquece do macaco, e meses depois recebe a palavra formada por e-mail.

## The Core Loop

O usuário entra, digita um e-mail e uma palavra, e esquece da existência do macaco. As letras vão passando na tela (uma a cada 0,5s). Depois de um tempo variável, o usuário recebe um e-mail dizendo que sua palavra foi formada. Não há login, não há dashboard, não há nada para acompanhar em tempo real — apenas o cadastro e a surpresa futura.

## Inspiration & Identity

Cabeça de macaco de óculos escuros, não o corpo todo, sobre um fundo espacial em movimento (estrelas, planetas, galáxias) que remete ao infinito. Tom: lúdico, científico, misterioso e calmo.

## Why This Matters to the Learner

Quer praticar criar e manter um projeto no ar, com interesse na área DevOps/SRE. Quer algo que possa manter vivo e evoluir com operações reais.

## What "Working" Looks Like

Letras passando embaixo do macaco, uma a cada 0,5s. Usuário cadastra palavra + e-mail. Aparece um tempo médio de espera baseado na quantidade de letras, com um "boa sorte". Quando o macaco forma a palavra cadastrada, o usuário recebe a notificação por e-mail.

## The POC Boundary

O que entra:
- Gerador de letras aleatórias (uma a cada 0,5s)
- Cadastro de palavra + e-mail (palavra de 3 a 15 letras apenas)
- Notificação por e-mail quando a palavra do usuário é formada
- Armazenamento apenas de letras que formam palavra de usuário ou palavras de um dicionário popular (EN/PT/ES)

## Later

Vale a pena fazer, não agora:
- Interface visual com o macaco de óculos escuros e fundo espacial animado
- Cálculo do tempo médio de espera exibido na tela
- Dicionário de palavras populares em inglês, português e espanhol
- Armazenamento de letras que formam palavras do dicionário (não apenas de usuários)

## Explicitly Cut

Não entra neste POC, por motivo real:
- Armazenar todas as letras geradas (apenas as que derem match)
- Histórico de tentativas ou contador de letras geradas
- Conta de usuário com senha ou perfil
- Dashboard para acompanhar o progresso
- Mais de uma palavra por usuário
- Notificações por push ou SMS (apenas e-mail)