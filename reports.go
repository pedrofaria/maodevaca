package main

import (
	"sort"
	"time"
)

var monthNames = []string{
	"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
	"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
}

// GetMonthSummary agrega entradas e saídas de um mês específico.
func (a *App) GetMonthSummary(year, month int) (MonthSummary, error) {
	summary := MonthSummary{Year: year, Month: month, Incomes: []Income{}, Payments: []Payment{}}

	// Entradas do mês.
	if err := db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) FROM incomes WHERE year=? AND month=?`, year, month).
		Scan(&summary.IncomesTotal); err != nil {
		return summary, err
	}

	// Saídas do mês (pagamentos).
	if err := db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) FROM payments WHERE year=? AND month=?`, year, month).
		Scan(&summary.ExpensesTotal); err != nil {
		return summary, err
	}

	summary.Balance = round2(summary.IncomesTotal - summary.ExpensesTotal)

	incomes, err := a.GetIncomes(year, month)
	if err != nil {
		return summary, err
	}
	summary.Incomes = incomes

	payments, err := a.GetPayments(year, month)
	if err != nil {
		return summary, err
	}
	summary.Payments = payments

	return summary, nil
}

// GetYearSummary agrega os 12 meses de um ano.
func (a *App) GetYearSummary(year int) (YearSummary, error) {
	summary := YearSummary{Year: year, Rows: []YearRow{}}

	rows, err := db.Query(`
		SELECT m AS month,
		       (SELECT COALESCE(SUM(amount),0) FROM incomes  WHERE year=? AND month=m),
		       (SELECT COALESCE(SUM(amount),0) FROM payments WHERE year=? AND month=m)
		FROM (
			SELECT 1 AS m UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5
			UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9 UNION SELECT 10
			UNION SELECT 11 UNION SELECT 12
		)`, year, year)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	for rows.Next() {
		var r YearRow
		if err := rows.Scan(&r.Month, &r.IncomesTotal, &r.ExpensesTotal); err != nil {
			return summary, err
		}
		r.Year = year
		r.Label = monthNames[r.Month-1]
		r.Balance = round2(r.IncomesTotal - r.ExpensesTotal)
		summary.IncomesTotal += r.IncomesTotal
		summary.ExpensesTotal += r.ExpensesTotal
		summary.Rows = append(summary.Rows, r)
	}
	if err := rows.Err(); err != nil {
		return summary, err
	}

	summary.IncomesTotal = round2(summary.IncomesTotal)
	summary.ExpensesTotal = round2(summary.ExpensesTotal)
	summary.Balance = round2(summary.IncomesTotal - summary.ExpensesTotal)
	return summary, nil
}

// GetAvailableYears retorna os anos que possuem lançamentos, do mais recente ao mais antigo.
func (a *App) GetAvailableYears() ([]int, error) {
	seen := map[int]bool{}
	rows, err := db.Query(`
		SELECT year FROM incomes
		UNION
		SELECT year FROM payments
		ORDER BY year DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var y int
		if err := rows.Scan(&y); err != nil {
			return nil, err
		}
		seen[y] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	years := make([]int, 0, len(seen))
	for y := range seen {
		years = append(years, y)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	if len(years) == 0 {
		years = append(years, time.Now().Year())
	}
	return years, nil
}

// round2 arredonda um valor monetário para 2 casas decimais.
func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
