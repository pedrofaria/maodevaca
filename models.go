package main

// Group representa um grupo de contas recorrentes (ex.: "Casa", "Assinaturas").
type Group struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	Color        string `json:"color"`
	SortOrder    int    `json:"sortOrder"`
	AccountCount int    `json:"accountCount"`
}

// Account representa uma conta recorrente (despesa fixa mensal).
// Type pode ser "fixed" (valor fixo em Amount) ou "percent"
// (valor = Percent% da soma das entradas das fontes em SourceIDs).
type Account struct {
	ID        int64   `json:"id"`
	GroupID   *int64  `json:"groupId"`
	GroupName string  `json:"groupName"`
	Name      string  `json:"name"`
	Amount    float64 `json:"amount"`
	DueDay    int     `json:"dueDay"`
	Active    bool    `json:"active"`
	Notes     string  `json:"notes"`
	Type      string  `json:"type"`
	Percent   float64 `json:"percent"`
	SourceIDs []int64 `json:"sourceIds"`
	CreatedAt string  `json:"createdAt"`
}

// Payment representa o pagamento de uma conta em um determinado mês.
type Payment struct {
	ID          int64   `json:"id"`
	AccountID   int64   `json:"accountId"`
	AccountName string  `json:"accountName"`
	GroupName   string  `json:"groupName"`
	Amount      float64 `json:"amount"`
	PaidOn      string  `json:"paidOn"`
	Year        int     `json:"year"`
	Month       int     `json:"month"`
	Notes       string  `json:"notes"`
}

// PayAccountInput são os dados para marcar/pagar uma conta no mês.
type PayAccountInput struct {
	AccountID int64   `json:"accountId"`
	Year      int     `json:"year"`
	Month     int     `json:"month"`
	Amount    float64 `json:"amount"`
	PaidOn    string  `json:"paidOn"`
	Notes     string  `json:"notes"`
}

// IncomeSource representa uma fonte de crédito/entrada (ex.: "Salário", "Aluguel").
type IncomeSource struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	CreatedAt string `json:"createdAt"`
}

// Income representa uma entrada de dinheiro (crédito).
type Income struct {
	ID          int64   `json:"id"`
	SourceID    int64   `json:"sourceId"`
	SourceName  string  `json:"sourceName"`
	SourceIcon  string  `json:"sourceIcon"`
	SourceColor string  `json:"sourceColor"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Year        int     `json:"year"`
	Month       int     `json:"month"`
	Description string  `json:"description"`
}

// MonthSummary agrega entradas e saídas de um mês específico.
type MonthSummary struct {
	Year          int       `json:"year"`
	Month         int       `json:"month"`
	IncomesTotal  float64   `json:"incomesTotal"`
	ExpensesTotal float64   `json:"expensesTotal"`
	Balance       float64   `json:"balance"`
	Incomes       []Income  `json:"incomes"`
	Payments      []Payment `json:"payments"`
}

// YearRow representa o resumo de um único mês dentro de um ano.
type YearRow struct {
	Year          int     `json:"year"`
	Month         int     `json:"month"`
	Label         string  `json:"label"`
	IncomesTotal  float64 `json:"incomesTotal"`
	ExpensesTotal float64 `json:"expensesTotal"`
	Balance       float64 `json:"balance"`
}

// YearSummary agrega os 12 meses de um ano.
type YearSummary struct {
	Year          int       `json:"year"`
	Rows          []YearRow `json:"rows"`
	IncomesTotal  float64   `json:"incomesTotal"`
	ExpensesTotal float64   `json:"expensesTotal"`
	Balance       float64   `json:"balance"`
}
