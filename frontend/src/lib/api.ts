import type {
  Account, Group, Income, IncomeSource, MonthSummary,
  PayAccountInput, Payment, YearSummary
} from './types'

// O Wails v2 injeta os bindings do backend em window.go.main.App.
declare global {
  interface Window {
    go?: {
      main: {
        App: Record<string, (...args: unknown[]) => Promise<unknown>>
      }
    }
  }
}

function call<T>(method: string, ...args: unknown[]): Promise<T> {
  const fn = window.go?.main?.App?.[method]
  if (!fn) {
    return Promise.reject(new Error(`Binding '${method}' não está disponível (rode dentro do Wails).`))
  }
  return fn(...args) as Promise<T>
}

export const api = {
  // Grupos
  getGroups: () => call<Group[]>('GetGroups'),
  createGroup: (name: string, icon: string, color: string) =>
    call<Group>('CreateGroup', name, icon, color),
  updateGroup: (id: number, name: string, icon: string, color: string, sortOrder: number) =>
    call<Group>('UpdateGroup', id, name, icon, color, sortOrder),
  deleteGroup: (id: number) => call<void>('DeleteGroup', id),

  // Contas
  getAccounts: () => call<Account[]>('GetAccounts'),
  createAccount: (name: string, amount: number, dueDay: number, groupId: number | null, notes: string, type: 'fixed' | 'percent', percent: number, sourceIds: number[]) =>
    call<Account>('CreateAccount', name, amount, dueDay, groupId, notes, type, percent, sourceIds),
  updateAccount: (id: number, name: string, amount: number, dueDay: number, groupId: number | null, active: boolean, notes: string, type: 'fixed' | 'percent', percent: number, sourceIds: number[]) =>
    call<Account>('UpdateAccount', id, name, amount, dueDay, groupId, active, notes, type, percent, sourceIds),
  deleteAccount: (id: number) => call<void>('DeleteAccount', id),
  getSuggestedPayment: (accountId: number, year: number, month: number) =>
    call<number>('GetSuggestedPayment', accountId, year, month),

  // Pagamentos
  payAccount: (input: PayAccountInput) => call<Payment>('PayAccount', input),
  unpayAccount: (accountId: number, year: number, month: number) =>
    call<void>('UnpayAccount', accountId, year, month),
  getPayments: (year: number, month: number) => call<Payment[]>('GetPayments', year, month),

  // Fontes de crédito
  getIncomeSources: () => call<IncomeSource[]>('GetIncomeSources'),
  createIncomeSource: (name: string, icon: string, color: string) =>
    call<IncomeSource>('CreateIncomeSource', name, icon, color),
  updateIncomeSource: (id: number, name: string, icon: string, color: string) =>
    call<IncomeSource>('UpdateIncomeSource', id, name, icon, color),
  deleteIncomeSource: (id: number) => call<void>('DeleteIncomeSource', id),

  // Entradas
  getIncomes: (year: number, month: number) => call<Income[]>('GetIncomes', year, month),
  createIncome: (sourceId: number, amount: number, date: string, description: string) =>
    call<Income>('CreateIncome', sourceId, amount, date, description),
  updateIncome: (id: number, sourceId: number, amount: number, date: string, description: string) =>
    call<Income>('UpdateIncome', id, sourceId, amount, date, description),
  deleteIncome: (id: number) => call<void>('DeleteIncome', id),

  // Relatórios
  getMonthSummary: (year: number, month: number) =>
    call<MonthSummary>('GetMonthSummary', year, month),
  getYearSummary: (year: number) => call<YearSummary>('GetYearSummary', year),
  getAvailableYears: () => call<number[]>('GetAvailableYears')
}

// Extrai uma mensagem amigável do erro retornado pelos bindings do Wails.
export function errMsg(e: unknown): string {
  if (e instanceof Error) return e.message
  if (typeof e === 'string') return e
  return String(e ?? 'Erro desconhecido')
}
