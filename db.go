package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// db é a conexão compartilhada com o SQLite.
var db *sql.DB

// initDB abre/cria o banco de dados no diretório de configuração do usuário
// e garante que o esquema exista.
func initDB() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("obtendo diretório de configuração: %w", err)
	}
	appDir := filepath.Join(configDir, "maodevaca")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return fmt.Errorf("criando diretório do app: %w", err)
	}
	path := filepath.Join(appDir, "maodevaca.db")

	db, err = sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("abrindo banco: %w", err)
	}
	// Uma única conexão garante que os PRAGMAs persistam.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec(`PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL;`); err != nil {
		return fmt.Errorf("configurando pragmas: %w", err)
	}

	if err := migrate(); err != nil {
		return fmt.Errorf("migrando esquema: %w", err)
	}
	if err := migrateAccountColumns(); err != nil {
		return fmt.Errorf("migrando colunas de contas: %w", err)
	}
	return nil
}

// migrateAccountColumns adiciona as colunas novas da tabela accounts em bancos
// já existentes (SQLite não suporta ADD COLUMN IF NOT EXISTS).
func migrateAccountColumns() error {
	cols := map[string]bool{}
	rows, err := db.Query(`PRAGMA table_info(accounts)`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		cols[name] = true
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if !cols["type"] {
		if _, err := db.Exec(`ALTER TABLE accounts ADD COLUMN type TEXT NOT NULL DEFAULT 'fixed'`); err != nil {
			return err
		}
	}
	if !cols["percent"] {
		if _, err := db.Exec(`ALTER TABLE accounts ADD COLUMN percent REAL NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	return nil
}

// migrate cria as tabelas caso ainda não existam.
func migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS groups (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    icon       TEXT NOT NULL DEFAULT '',
    color      TEXT NOT NULL DEFAULT '#64748b',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS accounts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id   INTEGER REFERENCES groups(id) ON DELETE SET NULL,
    name       TEXT NOT NULL,
    amount     REAL NOT NULL DEFAULT 0,
    due_day    INTEGER NOT NULL DEFAULT 1,
    active     INTEGER NOT NULL DEFAULT 1,
    notes      TEXT NOT NULL DEFAULT '',
    type       TEXT NOT NULL DEFAULT 'fixed',
    percent    REAL NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS account_sources (
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    source_id  INTEGER NOT NULL REFERENCES income_sources(id) ON DELETE CASCADE,
    PRIMARY KEY (account_id, source_id)
);

CREATE TABLE IF NOT EXISTS payments (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount     REAL NOT NULL,
    paid_on    TEXT NOT NULL,
    year       INTEGER NOT NULL,
    month      INTEGER NOT NULL,
    notes      TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(account_id, year, month)
);

CREATE TABLE IF NOT EXISTS income_sources (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    icon       TEXT NOT NULL DEFAULT '',
    color      TEXT NOT NULL DEFAULT '#10b981',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS incomes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id   INTEGER NOT NULL REFERENCES income_sources(id) ON DELETE CASCADE,
    amount      REAL NOT NULL,
    date        TEXT NOT NULL,
    year        INTEGER NOT NULL,
    month       INTEGER NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
`
	_, err := db.Exec(schema)
	return err
}
