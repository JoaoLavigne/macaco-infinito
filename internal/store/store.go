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

// Store acessa o arquivo SQLite com os cadastros.
type Store struct {
	db *sql.DB
}

// Abrir conecta ao arquivo SQLite e garante que a tabela cadastros exista.
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
);`

	if _, err := db.Exec(esquema); err != nil {
		db.Close()
		return nil, fmt.Errorf("criar tabela cadastros: %w", err)
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

// Fechar encerra a conexão com o banco.
func (s *Store) Fechar() error {
	return s.db.Close()
}
