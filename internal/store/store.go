package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ErrEmailDuplicado é retornado quando o e-mail já possui um cadastro.
var ErrEmailDuplicado = errors.New("e-mail já cadastrado")

// Store acessa o arquivo SQLite com os cadastros e os logs.
type Store struct {
	db *sql.DB
}

// Cadastro representa uma linha da tabela cadastros.
type Cadastro struct {
	ID           int64
	Email        string
	Palavra      string
	Data         time.Time
	Status       string
	NotificadoEm *time.Time
}

// Abrir conecta ao arquivo SQLite e garante que as tabelas existam.
func Abrir(caminho string) (*Store, error) {
	db, err := sql.Open("sqlite3", caminho)
	if err != nil {
		return nil, fmt.Errorf("abrir banco: %w", err)
	}

	// SQLite grava em um único arquivo. Uma conexão por vez evita
	// bloqueios de escrita concorrentes no POC.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("conectar ao banco: %w", err)
	}

	const esquema = `
CREATE TABLE IF NOT EXISTS cadastros (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	email         TEXT NOT NULL UNIQUE,
	palavra       TEXT NOT NULL,
	data          TEXT NOT NULL,
	status        TEXT NOT NULL,
	notificado_em TEXT
);

CREATE TABLE IF NOT EXISTS logs (
	id        INTEGER PRIMARY KEY AUTOINCREMENT,
	timestamp TEXT NOT NULL,
	tipo      TEXT NOT NULL,
	mensagem  TEXT NOT NULL,
	origem    TEXT NOT NULL
);`

	if _, err := db.Exec(esquema); err != nil {
		db.Close()
		return nil, fmt.Errorf("criar tabelas: %w", err)
	}

	return &Store{db: db}, nil
}

// InserirCadastro grava um novo cadastro com status pendente.
// Retorna ErrEmailDuplicado se o e-mail já existir.
func (s *Store) InserirCadastro(email, palavra string, agora time.Time) error {
	const consulta = `
INSERT INTO cadastros (email, palavra, data, status)
VALUES (?, ?, ?, 'pendente')
ON CONFLICT(email) DO NOTHING;`

	resultado, err := s.db.Exec(consulta, email, palavra, agora.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("inserir cadastro: %w", err)
	}

	linhas, err := resultado.RowsAffected()
	if err != nil {
		return fmt.Errorf("verificar inserção: %w", err)
	}
	if linhas == 0 {
		return ErrEmailDuplicado
	}

	return nil
}

// ListarPendentes devolve todos os cadastros com status pendente.
// O matcher usa esta lista para comparar as letras geradas com as palavras.
func (s *Store) ListarPendentes() ([]Cadastro, error) {
	const consulta = `
SELECT id, email, palavra, data, status, notificado_em
FROM cadastros
WHERE status = 'pendente'
ORDER BY id;`

	linhas, err := s.db.Query(consulta)
	if err != nil {
		return nil, fmt.Errorf("listar pendentes: %w", err)
	}
	defer linhas.Close()

	var cadastros []Cadastro
	for linhas.Next() {
		var c Cadastro
		var data string
		var notificadoEm sql.NullString

		if err := linhas.Scan(&c.ID, &c.Email, &c.Palavra, &data, &c.Status, &notificadoEm); err != nil {
			return nil, fmt.Errorf("ler cadastro: %w", err)
		}

		c.Data, err = time.Parse(time.RFC3339, data)
		if err != nil {
			return nil, fmt.Errorf("interpretar data do cadastro: %w", err)
		}

		if notificadoEm.Valid {
			quando, err := time.Parse(time.RFC3339, notificadoEm.String)
			if err != nil {
				return nil, fmt.Errorf("interpretar notificado_em: %w", err)
			}
			c.NotificadoEm = &quando
		}

		cadastros = append(cadastros, c)
	}

	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("listar pendentes: %w", err)
	}

	return cadastros, nil
}

// AtualizarNotificado marca um cadastro como notificado e grava o momento do envio.
func (s *Store) AtualizarNotificado(id int64, quando time.Time) error {
	const consulta = `
UPDATE cadastros
SET status = 'notificado', notificado_em = ?
WHERE id = ?;`

	resultado, err := s.db.Exec(consulta, quando.UTC().Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("atualizar cadastro: %w", err)
	}

	linhas, err := resultado.RowsAffected()
	if err != nil {
		return fmt.Errorf("verificar atualização: %w", err)
	}
	if linhas == 0 {
		return fmt.Errorf("cadastro %d não encontrado", id)
	}

	return nil
}

// RegistrarLog grava um erro de diagnóstico na tabela logs.
// Uma falha ao gravar o log não deve derrubar o chamador.
func (s *Store) RegistrarLog(tipo, mensagem, origem string) error {
	const consulta = `
INSERT INTO logs (timestamp, tipo, mensagem, origem)
VALUES (?, ?, ?, ?);`

	_, err := s.db.Exec(consulta, time.Now().UTC().Format(time.RFC3339), tipo, mensagem, origem)
	if err != nil {
		return fmt.Errorf("registrar log: %w", err)
	}

	return nil
}

// Fechar encerra a conexão com o banco.
func (s *Store) Fechar() error {
	return s.db.Close()
}
