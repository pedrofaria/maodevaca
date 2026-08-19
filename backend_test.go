package main

import (
	"database/sql"
	"testing"
)

// setupTestDB aponta o package-level `db` para um SQLite em memória isolado.
func setupTestDB(t *testing.T) {
	t.Helper()
	var err error
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		t.Fatal(err)
	}
	if err := migrate(); err != nil {
		t.Fatal(err)
	}
}

func TestFinanceFlow(t *testing.T) {
	setupTestDB(t)
	a := NewApp()

	// Grupo + conta recorrente.
	g, err := a.CreateGroup("Casa", "i-lucide-house", "#6366f1")
	if err != nil {
		t.Fatal(err)
	}
	acc, err := a.CreateAccount("Aluguel", 1500.0, 5, &g.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if acc.Amount != 1500.0 {
		t.Fatalf("valor: %v", acc.Amount)
	}

	// Pagar a conta no mês.
	pay, err := a.PayAccount(PayAccountInput{AccountID: acc.ID, Year: 2026, Month: 8, Amount: 1500, PaidOn: "2026-08-05", Notes: ""})
	if err != nil {
		t.Fatal(err)
	}
	if pay.AccountName != "Aluguel" {
		t.Fatalf("nome do pagamento: %s", pay.AccountName)
	}

	// Fonte + entrada de crédito.
	s, err := a.CreateIncomeSource("Salário", "i-lucide-briefcase", "#10b981")
	if err != nil {
		t.Fatal(err)
	}
	inc, err := a.CreateIncome(s.ID, 5000, "2026-08-10", "Salário agosto")
	if err != nil {
		t.Fatal(err)
	}
	if inc.Year != 2026 || inc.Month != 8 {
		t.Fatalf("data da entrada: %d/%d", inc.Year, inc.Month)
	}

	// Resumo do mês.
	ms, err := a.GetMonthSummary(2026, 8)
	if err != nil {
		t.Fatal(err)
	}
	if ms.IncomesTotal != 5000 {
		t.Fatalf("entradas do mês: %v", ms.IncomesTotal)
	}
	if ms.ExpensesTotal != 1500 {
		t.Fatalf("saídas do mês: %v", ms.ExpensesTotal)
	}
	if ms.Balance != 3500 {
		t.Fatalf("saldo do mês: %v", ms.Balance)
	}
	if len(ms.Incomes) != 1 || len(ms.Payments) != 1 {
		t.Fatalf("listas: inc=%d pay=%d", len(ms.Incomes), len(ms.Payments))
	}

	// Resumo do ano (agosto = índice 7).
	ys, err := a.GetYearSummary(2026)
	if err != nil {
		t.Fatal(err)
	}
	if len(ys.Rows) != 12 {
		t.Fatalf("linhas do ano: %d", len(ys.Rows))
	}
	if ys.Rows[7].IncomesTotal != 5000 || ys.Rows[7].ExpensesTotal != 1500 || ys.Rows[7].Balance != 3500 {
		t.Fatalf("agosto no ano: %+v", ys.Rows[7])
	}
	if ys.IncomesTotal != 5000 || ys.ExpensesTotal != 1500 || ys.Balance != 3500 {
		t.Fatalf("totais anuais: %+v", ys)
	}

	// Anos disponíveis.
	years, err := a.GetAvailableYears()
	if err != nil {
		t.Fatal(err)
	}
	if len(years) != 1 || years[0] != 2026 {
		t.Fatalf("anos: %v", years)
	}

	// Desmarcar pagamento.
	if err := a.UnpayAccount(acc.ID, 2026, 8); err != nil {
		t.Fatal(err)
	}
	ms2, err := a.GetMonthSummary(2026, 8)
	if err != nil {
		t.Fatal(err)
	}
	if ms2.ExpensesTotal != 0 || ms2.Balance != 5000 {
		t.Fatalf("após desmarcar: %+v", ms2)
	}

	// Pagar novamente (upsert atualiza o mesmo registro).
	_, err = a.PayAccount(PayAccountInput{AccountID: acc.ID, Year: 2026, Month: 8, Amount: 1600, PaidOn: "2026-08-06", Notes: "com juros"})
	if err != nil {
		t.Fatal(err)
	}
	pays, err := a.GetPayments(2026, 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(pays) != 1 || pays[0].Amount != 1600 {
		t.Fatalf("upsert: %+v", pays)
	}

	// Deleção em cascata: apagar conta remove o pagamento.
	if err := a.DeleteAccount(acc.ID); err != nil {
		t.Fatal(err)
	}
	ms3, err := a.GetMonthSummary(2026, 8)
	if err != nil {
		t.Fatal(err)
	}
	if ms3.ExpensesTotal != 0 {
		t.Fatalf("cascata: %+v", ms3)
	}
}
