// Tipos espelhando os modelos do backend Go.

export interface Group {
  id: number
  name: string
  icon: string
  color: string
  sortOrder: number
  accountCount: number
}

export interface Account {
  id: number
  groupId: number | null
  groupName: string
  name: string
  amount: number
  dueDay: number
  active: boolean
  notes: string
  type: 'fixed' | 'percent'
  percent: number
  sourceIds: number[]
  createdAt: string
}

export interface Payment {
  id: number
  accountId: number
  accountName: string
  groupName: string
  amount: number
  paidOn: string
  year: number
  month: number
  notes: string
}

export interface PayAccountInput {
  accountId: number
  year: number
  month: number
  amount: number
  paidOn: string
  notes: string
}

export interface IncomeSource {
  id: number
  name: string
  icon: string
  color: string
  createdAt: string
}

export interface Income {
  id: number
  sourceId: number
  sourceName: string
  sourceIcon: string
  sourceColor: string
  amount: number
  date: string
  year: number
  month: number
  description: string
}

export interface MonthSummary {
  year: number
  month: number
  incomesTotal: number
  expensesTotal: number
  balance: number
  incomes: Income[]
  payments: Payment[]
}

export interface YearRow {
  year: number
  month: number
  label: string
  incomesTotal: number
  expensesTotal: number
  balance: number
}

export interface YearSummary {
  year: number
  rows: YearRow[]
  incomesTotal: number
  expensesTotal: number
  balance: number
}
