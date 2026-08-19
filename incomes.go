package main

import (
	"time"
)

// ---------- Fontes de crédito ----------

// GetIncomeSources retorna todas as fontes de crédito.
func (a *App) GetIncomeSources() ([]IncomeSource, error) {
	rows, err := db.Query(`SELECT id, name, icon, color, created_at FROM income_sources ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := []IncomeSource{}
	for rows.Next() {
		var s IncomeSource
		if err := rows.Scan(&s.ID, &s.Name, &s.Icon, &s.Color, &s.CreatedAt); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

// CreateIncomeSource cria uma nova fonte de crédito.
func (a *App) CreateIncomeSource(name, icon, color string) (IncomeSource, error) {
	res, err := db.Exec(`INSERT INTO income_sources (name, icon, color) VALUES (?, ?, ?)`, name, icon, color)
	if err != nil {
		return IncomeSource{}, err
	}
	id, _ := res.LastInsertId()
	return IncomeSource{ID: id, Name: name, Icon: icon, Color: color}, nil
}

// UpdateIncomeSource atualiza uma fonte de crédito.
func (a *App) UpdateIncomeSource(id int64, name, icon, color string) (IncomeSource, error) {
	_, err := db.Exec(`UPDATE income_sources SET name=?, icon=?, color=? WHERE id=?`, name, icon, color, id)
	if err != nil {
		return IncomeSource{}, err
	}
	return IncomeSource{ID: id, Name: name, Icon: icon, Color: color}, nil
}

// DeleteIncomeSource remove uma fonte e, em cascata, suas entradas.
func (a *App) DeleteIncomeSource(id int64) error {
	_, err := db.Exec(`DELETE FROM income_sources WHERE id=?`, id)
	return err
}

// ---------- Entradas (créditos) ----------

// GetIncomes retorna as entradas de um mês.
func (a *App) GetIncomes(year, month int) ([]Income, error) {
	rows, err := db.Query(`
		SELECT i.id, i.source_id, s.name, s.icon, s.color, i.amount,
		       i.date, i.year, i.month, i.description
		FROM incomes i
		JOIN income_sources s ON s.id = i.source_id
		WHERE i.year=? AND i.month=?
		ORDER BY i.date ASC, i.id ASC`, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	incomes := []Income{}
	for rows.Next() {
		var i Income
		if err := rows.Scan(&i.ID, &i.SourceID, &i.SourceName, &i.SourceIcon, &i.SourceColor,
			&i.Amount, &i.Date, &i.Year, &i.Month, &i.Description); err != nil {
			return nil, err
		}
		incomes = append(incomes, i)
	}
	return incomes, rows.Err()
}

// CreateIncome registra uma entrada de dinheiro.
func (a *App) CreateIncome(sourceID int64, amount float64, date, description string) (Income, error) {
	if amount <= 0 {
		return Income{}, fmtE("o valor deve ser maior que zero")
	}
	if date == "" {
		return Income{}, fmtE("a data é obrigatória")
	}
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return Income{}, fmtE("data inválida")
	}
	year, month := t.Year(), int(t.Month())

	res, err := db.Exec(`INSERT INTO incomes (source_id, amount, date, year, month, description) VALUES (?,?,?,?,?,?)`,
		sourceID, amount, date, year, month, description)
	if err != nil {
		return Income{}, err
	}
	id, _ := res.LastInsertId()
	return Income{ID: id, SourceID: sourceID, Amount: amount, Date: date, Year: year, Month: month, Description: description}, nil
}

// UpdateIncome atualiza uma entrada de dinheiro.
func (a *App) UpdateIncome(id, sourceID int64, amount float64, date, description string) (Income, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return Income{}, fmtE("data inválida")
	}
	year, month := t.Year(), int(t.Month())
	_, err = db.Exec(`UPDATE incomes SET source_id=?, amount=?, date=?, year=?, month=?, description=? WHERE id=?`,
		sourceID, amount, date, year, month, description, id)
	if err != nil {
		return Income{}, err
	}
	return Income{ID: id, SourceID: sourceID, Amount: amount, Date: date, Year: year, Month: month, Description: description}, nil
}

// DeleteIncome remove uma entrada de dinheiro.
func (a *App) DeleteIncome(id int64) error {
	_, err := db.Exec(`DELETE FROM incomes WHERE id=?`, id)
	return err
}
