package main

import (
	"context"
	"fmt"
)

// App é a estrutura exposta ao frontend via bindings do Wails.
type App struct {
	ctx context.Context
}

// NewApp cria uma nova instância do App.
func NewApp() *App {
	return &App{}
}

// fmtE é um atalho para criar erros formatados.
func fmtE(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// startup é chamado quando o app inicia. O contexto é salvo
// para permitir chamadas ao runtime do Wails.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ---------- Grupos de contas ----------

// GetGroups retorna todos os grupos, ordenados, com a contagem de contas.
func (a *App) GetGroups() ([]Group, error) {
	rows, err := db.Query(`
		SELECT g.id, g.name, g.icon, g.color, g.sort_order,
		       (SELECT COUNT(*) FROM accounts acc WHERE acc.group_id = g.id) AS cnt
		FROM groups g
		ORDER BY g.sort_order ASC, g.name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []Group{}
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Icon, &g.Color, &g.SortOrder, &g.AccountCount); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// CreateGroup cria um novo grupo.
func (a *App) CreateGroup(name, icon, color string) (Group, error) {
	res, err := db.Exec(`INSERT INTO groups (name, icon, color) VALUES (?, ?, ?)`, name, icon, color)
	if err != nil {
		return Group{}, err
	}
	id, _ := res.LastInsertId()
	return Group{ID: id, Name: name, Icon: icon, Color: color}, nil
}

// UpdateGroup atualiza os dados de um grupo.
func (a *App) UpdateGroup(id int64, name, icon, color string, sortOrder int) (Group, error) {
	_, err := db.Exec(`UPDATE groups SET name=?, icon=?, color=?, sort_order=? WHERE id=?`,
		name, icon, color, sortOrder, id)
	if err != nil {
		return Group{}, err
	}
	return Group{ID: id, Name: name, Icon: icon, Color: color, SortOrder: sortOrder}, nil
}

// DeleteGroup remove um grupo. As contas do grupo ficam sem grupo (não são apagadas).
func (a *App) DeleteGroup(id int64) error {
	_, err := db.Exec(`DELETE FROM groups WHERE id=?`, id)
	return err
}

// ---------- Contas recorrentes ----------

// GetAccounts retorna todas as contas com o nome do grupo.
func (a *App) GetAccounts() ([]Account, error) {
	rows, err := db.Query(`
		SELECT acc.id, acc.group_id, COALESCE(g.name, ''), acc.name, acc.amount,
		       acc.due_day, acc.active, acc.notes, acc.created_at
		FROM accounts acc
		LEFT JOIN groups g ON g.id = acc.group_id
		ORDER BY acc.active DESC, COALESCE(g.sort_order, 999), acc.due_day ASC, acc.name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		var a Account
		var active int
		if err := rows.Scan(&a.ID, &a.GroupID, &a.GroupName, &a.Name, &a.Amount,
			&a.DueDay, &active, &a.Notes, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Active = active == 1
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

// CreateAccount cria uma nova conta recorrente.
func (a *App) CreateAccount(name string, amount float64, dueDay int, groupID *int64, notes string) (Account, error) {
	res, err := db.Exec(`INSERT INTO accounts (name, amount, due_day, group_id, notes) VALUES (?,?,?,?,?)`,
		name, amount, dueDay, groupID, notes)
	if err != nil {
		return Account{}, err
	}
	id, _ := res.LastInsertId()
	return Account{ID: id, Name: name, Amount: amount, DueDay: dueDay, GroupID: groupID, Notes: notes}, nil
}

// UpdateAccount atualiza uma conta recorrente.
func (a *App) UpdateAccount(id int64, name string, amount float64, dueDay int, groupID *int64, active bool, notes string) (Account, error) {
	act := 0
	if active {
		act = 1
	}
	_, err := db.Exec(`UPDATE accounts SET name=?, amount=?, due_day=?, group_id=?, active=?, notes=? WHERE id=?`,
		name, amount, dueDay, groupID, act, notes, id)
	if err != nil {
		return Account{}, err
	}
	return Account{ID: id, Name: name, Amount: amount, DueDay: dueDay, GroupID: groupID, Active: active, Notes: notes}, nil
}

// DeleteAccount remove uma conta e, em cascata, seus pagamentos.
func (a *App) DeleteAccount(id int64) error {
	_, err := db.Exec(`DELETE FROM accounts WHERE id=?`, id)
	return err
}

// ---------- Pagamentos ----------

// PayAccount marca uma conta como paga no mês (upsert: se já existe, atualiza).
func (a *App) PayAccount(in PayAccountInput) (Payment, error) {
	// Valida mês/ano básicos.
	if in.Year < 1970 || in.Year > 2100 || in.Month < 1 || in.Month > 12 {
		return Payment{}, fmt.Errorf("mês/ano inválidos")
	}
	if in.PaidOn == "" {
		return Payment{}, fmt.Errorf("data de pagamento obrigatória")
	}
	if in.Amount <= 0 {
		// Se o valor não foi informado, usa o valor cadastrado da conta.
		var amt float64
		if err := db.QueryRow(`SELECT amount FROM accounts WHERE id=?`, in.AccountID).Scan(&amt); err != nil {
			return Payment{}, fmt.Errorf("conta não encontrada")
		}
		in.Amount = amt
	}

	_, err := db.Exec(`
		INSERT INTO payments (account_id, amount, paid_on, year, month, notes)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(account_id, year, month) DO UPDATE SET
			amount = excluded.amount,
			paid_on = excluded.paid_on,
			notes   = excluded.notes`,
		in.AccountID, in.Amount, in.PaidOn, in.Year, in.Month, in.Notes)
	if err != nil {
		return Payment{}, err
	}
	return a.paymentForMonth(in.AccountID, in.Year, in.Month)
}

// UnpayAccount desmarca o pagamento de uma conta no mês.
func (a *App) UnpayAccount(accountID int64, year, month int) error {
	_, err := db.Exec(`DELETE FROM payments WHERE account_id=? AND year=? AND month=?`,
		accountID, year, month)
	return err
}

// GetPayments retorna todos os pagamentos de um mês, com nome da conta/grupo.
func (a *App) GetPayments(year, month int) ([]Payment, error) {
	rows, err := db.Query(`
		SELECT p.id, p.account_id, acc.name, COALESCE(g.name,''), p.amount,
		       p.paid_on, p.year, p.month, p.notes
		FROM payments p
		JOIN accounts acc ON acc.id = p.account_id
		LEFT JOIN groups g ON g.id = acc.group_id
		WHERE p.year=? AND p.month=?
		ORDER BY p.paid_on ASC, acc.name ASC`, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := []Payment{}
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.AccountID, &p.AccountName, &p.GroupName,
			&p.Amount, &p.PaidOn, &p.Year, &p.Month, &p.Notes); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

// paymentForMonth busca um único pagamento de conta para o mês.
func (a *App) paymentForMonth(accountID int64, year, month int) (Payment, error) {
	var p Payment
	err := db.QueryRow(`
		SELECT p.id, p.account_id, acc.name, COALESCE(g.name,''), p.amount,
		       p.paid_on, p.year, p.month, p.notes
		FROM payments p
		JOIN accounts acc ON acc.id = p.account_id
		LEFT JOIN groups g ON g.id = acc.group_id
		WHERE p.account_id=? AND p.year=? AND p.month=?`,
		accountID, year, month).
		Scan(&p.ID, &p.AccountID, &p.AccountName, &p.GroupName,
			&p.Amount, &p.PaidOn, &p.Year, &p.Month, &p.Notes)
	return p, err
}
