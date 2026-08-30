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
	acc, err := a.CreateAccount("Aluguel", 1500.0, 5, &g.ID, "", "fixed", 0, nil)
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

// TestPercentAccount valida conta do tipo percentual: o valor sugerido é o
// percentual da soma das entradas das fontes vinculadas no mês.
func TestPercentAccount(t *testing.T) {
	setupTestDB(t)
	a := NewApp()

	// Duas fontes e entradas no mês.
	s1, err := a.CreateIncomeSource("Salário", "i-lucide-briefcase", "#10b981")
	if err != nil {
		t.Fatal(err)
	}
	s2, err := a.CreateIncomeSource("Aluguel", "i-lucide-home", "#6366f1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.CreateIncome(s1.ID, 8000, "2026-08-05", "Salário agosto"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.CreateIncome(s2.ID, 2000, "2026-08-10", "Aluguel recebido"); err != nil {
		t.Fatal(err)
	}

	// Conta percentual: 10% de Salário + Aluguel.
	acc, err := a.CreateAccount("Dízimo", 0, 5, nil, "", "percent", 10, []int64{s1.ID, s2.ID})
	if err != nil {
		t.Fatal(err)
	}
	if acc.Type != "percent" || acc.Percent != 10 || len(acc.SourceIDs) != 2 {
		t.Fatalf("conta percentual: %+v", acc)
	}

	// Valor sugerido = 10% de (8000 + 2000) = 1000.
	got, err := a.GetSuggestedPayment(acc.ID, 2026, 8)
	if err != nil {
		t.Fatal(err)
	}
	if got != 1000 {
		t.Fatalf("valor sugerido: %v (esperado 1000)", got)
	}

	// Sem entradas no mês, o valor sugerido é 0.
	got0, err := a.GetSuggestedPayment(acc.ID, 2026, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got0 != 0 {
		t.Fatalf("valor sugerido sem entradas: %v (esperado 0)", got0)
	}

	// GetAccounts carrega type/percent/fontes.
	accs, err := a.GetAccounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(accs) != 1 || accs[0].Type != "percent" || accs[0].Percent != 10 || len(accs[0].SourceIDs) != 2 {
		t.Fatalf("GetAccounts percentual: %+v", accs)
	}

	// Atualizar remove fontes que saíram (deixar só Salário) e muda o percentual.
	upd, err := a.UpdateAccount(acc.ID, "Dízimo", 0, 5, nil, true, "", "percent", 15, []int64{s1.ID})
	if err != nil {
		t.Fatal(err)
	}
	if upd.Percent != 15 || len(upd.SourceIDs) != 1 {
		t.Fatalf("update percentual: %+v", upd)
	}
	gotUpd, err := a.GetSuggestedPayment(acc.ID, 2026, 8)
	if err != nil {
		t.Fatal(err)
	}
	if gotUpd != 1200 { // 15% de 8000
		t.Fatalf("valor sugerido após update: %v (esperado 1200)", gotUpd)
	}
}
