<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api, errMsg } from '../lib/api'
import { dateBR, money, todayISO } from '../lib/format'
import { nextMonth, prevMonth, viewMonth, viewYear } from '../lib/state'
import type { Account, Group, Income, IncomeSource, Payment } from '../lib/types'

const loading = ref(true)
const error = ref('')
const accounts = ref<Account[]>([])
const groups = ref<Group[]>([])
const payments = ref<Payment[]>([])
const incomes = ref<Income[]>([])
const sources = ref<IncomeSource[]>([])

// Modal de pagamento
const payModalOpen = ref(false)
const payingAccount = ref<Account | null>(null)
const payForm = ref({ amount: 0, paidOn: todayISO() })
const paySaving = ref(false)
const payErr = ref('')

// Modal de entrada
const incomeModalOpen = ref(false)
const incomeForm = ref({ sourceId: 0, amount: 0, date: todayISO(), description: '' })
const incomeSaving = ref(false)
const incomeErr = ref('')

// Estado de pagamentos do mês
const paidSet = computed(() => new Set(payments.value.map(p => p.accountId)))
const paidAmount = computed(() => new Map(payments.value.map(p => [p.accountId, p.amount] as const)))
const paidDate = computed(() => new Map(payments.value.map(p => [p.accountId, p.paidOn] as const)))

// Pagamento marcado com data no futuro: recebe destaque amarelo.
function isFuturePaid(accountId: number): boolean {
  const on = paidDate.value.get(accountId)
  return !!on && on > todayISO()
}

// Valor sugerido por conta percentual no mês em exibição (p/ exibir mesmo sem pagamento).
const suggestedAmount = ref<Map<number, number>>(new Map())

async function loadSuggested() {
  const m = new Map<number, number>()
  const percentAccs = accounts.value.filter(a => a.type === 'percent')
  await Promise.all(percentAccs.map(async (a) => {
    try {
      m.set(a.id, await api.getSuggestedPayment(a.id, viewYear.value, viewMonth.value))
    } catch { /* conta fica sem valor sugerido */ }
  }))
  suggestedAmount.value = m
}

const activeAccounts = computed(() => accounts.value.filter(a => a.active))
const inactiveAccounts = computed(() => accounts.value.filter(a => !a.active))

function groupActiveAccounts(gid: number): Account[] {
  return activeAccounts.value.filter(a => a.groupId === gid)
}
const activeGroupIds = computed(() => new Set(activeAccounts.value.filter(a => a.groupId !== null).map(a => a.groupId!)))
const displayGroups = computed(() => groups.value.filter(g => activeGroupIds.value.has(g.id)))
const ungroupedActive = computed(() => activeAccounts.value.filter(a => a.groupId === null))

function groupPaid(gid: number): number {
  let s = 0
  for (const a of groupActiveAccounts(gid)) s += paidAmount.value.get(a.id) ?? 0
  return s
}
const ungroupedPaid = computed(() => ungroupedActive.value.reduce((s, a) => s + (paidAmount.value.get(a.id) ?? 0), 0))
const totalPaid = computed(() => payments.value.reduce((s, p) => s + p.amount, 0))
const paidCount = computed(() => payments.value.length)

const monthTotal = computed(() => incomes.value.reduce((s, i) => s + i.amount, 0))
const sourceOptions = computed(() => sources.value.map(s => ({ label: s.name, value: s.id })))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [accs, grps, pays, inc, src] = await Promise.all([
      api.getAccounts(),
      api.getGroups(),
      api.getPayments(viewYear.value, viewMonth.value),
      api.getIncomes(viewYear.value, viewMonth.value),
      api.getIncomeSources()
    ])
    accounts.value = accs ?? []
    groups.value = grps ?? []
    payments.value = pays ?? []
    incomes.value = inc ?? []
    sources.value = src ?? []
    await loadSuggested()
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch([viewYear, viewMonth], load)

// ---------- pagamento ----------
async function openPay(acc: Account) {
  payingAccount.value = acc
  payForm.value = { amount: 0, paidOn: todayISO() }
  payErr.value = ''
  if (acc.type === 'percent') {
    // Valor sugerido = percentual das entradas das fontes vinculadas no mês.
    try {
      payForm.value.amount = await api.getSuggestedPayment(acc.id, viewYear.value, viewMonth.value)
    } catch (e) {
      payErr.value = errMsg(e)
    }
  } else {
    // Valor sugerido = valor cadastrado da conta.
    payForm.value.amount = acc.amount
  }
  payModalOpen.value = true
}

async function confirmPay() {
  if (!payingAccount.value) return
  payErr.value = ''
  if (!(payForm.value.amount > 0)) { payErr.value = 'Valor inválido.'; return }
  if (!payForm.value.paidOn) { payErr.value = 'Data de pagamento obrigatória.'; return }
  paySaving.value = true
  try {
    await api.payAccount({
      accountId: payingAccount.value.id,
      year: viewYear.value,
      month: viewMonth.value,
      amount: payForm.value.amount,
      paidOn: payForm.value.paidOn,
      notes: ''
    })
    payModalOpen.value = false
    await load()
  } catch (e) {
    payErr.value = errMsg(e)
  } finally {
    paySaving.value = false
  }
}

// Estado de confirmação de remoção de pagamento
const unpayConfirmOpen = ref(false)
const pendingUnpay = ref<Account | null>(null)
const unpaying = ref(false)
const unpayErr = ref('')

async function unpay(acc: Account) {
  pendingUnpay.value = acc
  unpayErr.value = ''
  unpayConfirmOpen.value = true
}

async function confirmUnpay() {
  if (!pendingUnpay.value) return
  unpayErr.value = ''
  unpaying.value = true
  try {
    await api.unpayAccount(pendingUnpay.value.id, viewYear.value, viewMonth.value)
    unpayConfirmOpen.value = false
    pendingUnpay.value = null
    await load()
  } catch (e) {
    unpayErr.value = errMsg(e)
  } finally {
    unpaying.value = false
  }
}

// ---------- entrada ----------
function openNewIncome() {
  incomeErr.value = ''
  if (!sources.value.length) {
    incomeErr.value = 'Você precisa criar uma fonte de crédito primeiro (menu Fontes).'
  }
  incomeForm.value = { sourceId: sources.value[0]?.id ?? 0, amount: 0, date: todayISO(), description: '' }
  incomeModalOpen.value = true
}

async function saveIncome() {
  incomeErr.value = ''
  if (!incomeForm.value.sourceId) { incomeErr.value = 'Selecione uma fonte de crédito.'; return }
  if (!(incomeForm.value.amount > 0)) { incomeErr.value = 'Informe um valor maior que zero.'; return }
  if (!incomeForm.value.date) { incomeErr.value = 'Informe a data.'; return }
  incomeSaving.value = true
  try {
    const f = incomeForm.value
    await api.createIncome(f.sourceId, f.amount, f.date, f.description)
    incomeModalOpen.value = false
    await load()
  } catch (e) {
    incomeErr.value = errMsg(e)
  } finally {
    incomeSaving.value = false
  }
}

// Estado de confirmação de exclusão de entrada
const confirmOpen = ref(false)
const pendingIncome = ref<Income | null>(null)
const deleting = ref(false)
const deleteErr = ref('')

async function removeIncome(i: Income) {
  pendingIncome.value = i
  deleteErr.value = ''
  confirmOpen.value = true
}

async function confirmRemoveIncome() {
  if (!pendingIncome.value) return
  deleteErr.value = ''
  deleting.value = true
  try {
    await api.deleteIncome(pendingIncome.value.id)
    confirmOpen.value = false
    pendingIncome.value = null
    await load()
  } catch (e) {
    deleteErr.value = errMsg(e)
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div class="space-y-6 pb-6 lg:pb-8">
    <div class="sticky top-0 z-10 -mx-6 lg:-mx-8 px-6 lg:px-8 pt-6 lg:pt-8 pb-3 bg-neutral-50 dark:bg-neutral-950 border-b border-neutral-200/70 dark:border-neutral-800/70 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Transações</h1>
        <p class="text-sm text-neutral-500 dark:text-neutral-400">Pagamentos e entradas · <span class="font-medium">{{ viewMonth }}/{{ viewYear }}</span></p>
      </div>
      <div class="flex items-center gap-2">
        <div class="flex items-center gap-1 mr-1">
          <UButton icon="i-lucide-chevron-left" color="neutral" variant="ghost" square @click="prevMonth" />
          <span class="text-sm font-semibold w-24 text-center">{{ viewMonth }}/{{ viewYear }}</span>
          <UButton icon="i-lucide-chevron-right" color="neutral" variant="ghost" square @click="nextMonth" />
        </div>
      </div>
    </div>

    <UAlert v-if="error" color="error" :title="error" icon="i-lucide-alert-circle" />

    <div v-if="loading && !accounts.length && !incomes.length" class="text-sm text-neutral-400 dark:text-neutral-500">Carregando…</div>

    <template v-else>
      <!-- ============ PAGAMENTOS ============ -->
      <section class="space-y-3">
        <div class="rounded-lg bg-red-50 dark:bg-red-950/60 border border-red-200 dark:border-red-800 px-4 py-3 grid grid-cols-[1fr_auto_1fr] items-center gap-2">
          <span class="text-sm font-medium text-red-800 dark:text-red-300">Pagamentos de {{ viewMonth }}/{{ viewYear }}</span>
          <span class="text-sm text-red-600 dark:text-red-400">{{ paidCount }} de {{ activeAccounts.length }} {{ activeAccounts.length === 1 ? 'paga' : 'pagas' }}</span>
          <span class="text-xl font-bold text-red-700 dark:text-red-400 justify-self-end">{{ money(totalPaid) }}</span>
        </div>

        <!-- Grupos -->
        <div v-for="g in displayGroups" :key="g.id" class="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <div class="flex items-center justify-between pt-1 sm:col-span-2">
            <div class="flex items-center gap-2">
              <UIcon :name="g.icon || 'i-lucide-folder'" class="text-lg" :style="{ color: g.color }" />
              <h3 class="font-semibold text-neutral-800 dark:text-neutral-100">{{ g.name }}</h3>
              <UBadge :label="String(groupActiveAccounts(g.id).length)" color="neutral" variant="subtle" />
            </div>
            <span class="text-xs text-neutral-500 dark:text-neutral-400">Pago: <b class="text-neutral-900 dark:text-neutral-100">{{ money(groupPaid(g.id)) }}</b></span>
          </div>
          <div
            v-for="acc in groupActiveAccounts(g.id)"
            :key="acc.id"
            class="flex items-center justify-between gap-3 rounded-lg border p-3"
            :class="isFuturePaid(acc.id) ? 'bg-amber-50 border-amber-200 dark:bg-yellow-500/15 dark:border-yellow-500/40' : 'bg-white border-neutral-200 dark:bg-neutral-900 dark:border-neutral-800'"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div
                class="grid place-items-center size-9 rounded-lg shrink-0"
                :class="paidSet.has(acc.id) ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-950/60' : 'bg-neutral-100 text-neutral-500 dark:bg-neutral-800 dark:text-neutral-400'"
              >
                <UIcon :name="paidSet.has(acc.id) ? 'i-lucide-check' : 'i-lucide-clock'" class="text-lg" />
              </div>
              <div class="min-w-0">
                <p class="font-medium text-sm text-neutral-900 dark:text-neutral-100 truncate">{{ acc.name }}</p>
                <p class="text-xs text-neutral-400 dark:text-neutral-500">Vence dia {{ acc.dueDay }}</p>
              </div>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <template v-if="paidSet.has(acc.id)">
                <span class="font-semibold text-sm text-red-600 dark:text-red-400">{{ money(paidAmount.get(acc.id)) }}</span>
                <UButton
                  :label="'Pago em ' + dateBR(paidDate.get(acc.id) ?? '')"
                  icon="i-lucide-check"
                  color="success"
                  variant="soft"
                  size="sm"
                  @click="unpay(acc)"
                />
              </template>
              <span v-if="!paidSet.has(acc.id) && acc.type === 'percent'" class="font-semibold text-sm text-neutral-500 dark:text-neutral-400">{{ money(suggestedAmount.get(acc.id) ?? 0) }}</span>
              <UButton
                v-if="!paidSet.has(acc.id)"
                icon="i-lucide-circle-check"
                label="Pagar"
                color="primary"
                size="sm"
                @click="openPay(acc)"
              />
            </div>
          </div>
          <div v-if="!groupActiveAccounts(g.id).length" class="text-xs text-neutral-400 dark:text-neutral-500 py-1 pl-1 sm:col-span-2">
            Nenhuma conta neste grupo.
          </div>
        </div>

        <!-- Sem grupo -->
        <div v-if="ungroupedActive.length" class="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <div class="flex items-center justify-between pt-1 sm:col-span-2">
            <div class="flex items-center gap-2">
              <UIcon name="i-lucide-inbox" class="text-lg text-neutral-400 dark:text-neutral-500" />
              <h3 class="font-semibold text-neutral-800 dark:text-neutral-100">Sem grupo</h3>
              <UBadge :label="String(ungroupedActive.length)" color="neutral" variant="subtle" />
            </div>
            <span class="text-xs text-neutral-500 dark:text-neutral-400">Pago: <b class="text-neutral-900 dark:text-neutral-100">{{ money(ungroupedPaid) }}</b></span>
          </div>
          <div
            v-for="acc in ungroupedActive"
            :key="acc.id"
            class="flex items-center justify-between gap-3 rounded-lg border p-3"
            :class="isFuturePaid(acc.id) ? 'bg-amber-50 border-amber-200 dark:bg-yellow-500/15 dark:border-yellow-500/40' : 'bg-white border-neutral-200 dark:bg-neutral-900 dark:border-neutral-800'"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div
                class="grid place-items-center size-9 rounded-lg shrink-0"
                :class="paidSet.has(acc.id) ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-950/60' : 'bg-neutral-100 text-neutral-500 dark:bg-neutral-800 dark:text-neutral-400'"
              >
                <UIcon :name="paidSet.has(acc.id) ? 'i-lucide-check' : 'i-lucide-clock'" class="text-lg" />
              </div>
              <div class="min-w-0">
                <p class="font-medium text-sm text-neutral-900 dark:text-neutral-100 truncate">{{ acc.name }}</p>
                <p class="text-xs text-neutral-400 dark:text-neutral-500">Vence dia {{ acc.dueDay }}</p>
              </div>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <template v-if="paidSet.has(acc.id)">
                <span class="font-semibold text-sm text-red-600 dark:text-red-400">{{ money(paidAmount.get(acc.id)) }}</span>
                <UButton
                  :label="'Pago em ' + dateBR(paidDate.get(acc.id) ?? '')"
                  icon="i-lucide-check"
                  color="success"
                  variant="soft"
                  size="sm"
                  @click="unpay(acc)"
                />
              </template>
              <span v-if="!paidSet.has(acc.id) && acc.type === 'percent'" class="font-semibold text-sm text-neutral-500 dark:text-neutral-400">{{ money(suggestedAmount.get(acc.id) ?? 0) }}</span>
              <UButton
                v-if="!paidSet.has(acc.id)"
                icon="i-lucide-circle-check"
                label="Pagar"
                color="primary"
                size="sm"
                @click="openPay(acc)"
              />
            </div>
          </div>
        </div>

        <!-- Contas inativas -->
        <div v-if="inactiveAccounts.length" class="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-2">
          <h3 class="font-semibold text-neutral-400 dark:text-neutral-500 sm:col-span-2">Contas inativas</h3>
          <div
            v-for="acc in inactiveAccounts"
            :key="acc.id"
            class="flex items-center justify-between gap-3 rounded-lg border p-3 opacity-60"
            :class="isFuturePaid(acc.id) ? 'bg-amber-50 border-amber-200 dark:bg-yellow-500/15 dark:border-yellow-500/40' : 'bg-neutral-50 border-neutral-200 dark:bg-neutral-800/50 dark:border-neutral-800'"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div class="grid place-items-center size-9 rounded-lg bg-neutral-100 dark:bg-neutral-800 text-neutral-400 dark:text-neutral-500 shrink-0">
                <UIcon name="i-lucide-power" class="text-lg" />
              </div>
              <div class="min-w-0">
                <p class="font-medium text-sm text-neutral-900 dark:text-neutral-100 truncate">{{ acc.name }}</p>
                <p class="text-xs text-neutral-400 dark:text-neutral-500">Vence dia {{ acc.dueDay }}</p>
              </div>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <template v-if="paidSet.has(acc.id)">
                <span class="font-semibold text-sm text-neutral-400 dark:text-neutral-500">{{ money(paidAmount.get(acc.id)) }}</span>
                <UButton icon="i-lucide-check" label="Pago" color="success" variant="soft" size="sm" @click="unpay(acc)" />
              </template>
              <span v-else class="font-semibold text-sm text-neutral-400 dark:text-neutral-500">{{ acc.type === 'percent' ? money(suggestedAmount.get(acc.id) ?? 0) : money(acc.amount) }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- ============ ENTRADAS ============ -->
      <section class="space-y-3 pt-2">
        <div class="rounded-lg bg-emerald-50 dark:bg-emerald-950/60 border border-emerald-200 dark:border-emerald-800 px-4 py-3 grid grid-cols-[1fr_auto_1fr] items-center gap-2">
          <span class="text-sm font-medium text-emerald-800 dark:text-emerald-300">Entradas de {{ viewMonth }}/{{ viewYear }}</span>
          <UButton icon="i-lucide-plus" label="Nova entrada" @click="openNewIncome" />
          <span class="text-xl font-bold text-emerald-700 dark:text-emerald-400 justify-self-end">{{ money(monthTotal) }}</span>
        </div>

        <div v-if="incomes.length" class="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <div
            v-for="i in incomes"
            :key="i.id"
            class="flex items-center justify-between gap-3 rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 p-3"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div
                class="grid place-items-center size-10 rounded-xl shrink-0 text-white"
                :style="{ backgroundColor: i.sourceColor || '#10b981' }"
              >
                <UIcon :name="i.sourceIcon || 'i-lucide-briefcase'" class="text-lg" />
              </div>
              <div class="min-w-0">
                <p class="font-medium text-neutral-900 dark:text-neutral-100 truncate">{{ i.sourceName }}</p>
                <p class="text-xs text-neutral-400 dark:text-neutral-500">{{ dateBR(i.date) }}<span v-if="i.description"> · {{ i.description }}</span></p>
              </div>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <span class="font-semibold text-emerald-600 dark:text-emerald-400">{{ money(i.amount) }}</span>
              <UButton icon="i-lucide-trash" color="neutral" variant="ghost" size="sm" @click="removeIncome(i)" />
            </div>
          </div>
        </div>
        <div v-else class="text-sm text-neutral-400 dark:text-neutral-500 text-center py-4">Nenhuma entrada lançada este mês.</div>
      </section>
    </template>

    <!-- Modal: pagar -->
    <UModal v-model:open="payModalOpen" :title="'Pagar: ' + (payingAccount?.name ?? '')" :ui="{ content: 'max-w-sm' }">
      <template #body>
        <div class="space-y-4">
          <UAlert v-if="payErr" color="error" :title="payErr" icon="i-lucide-alert-circle" />
          <div class="grid grid-cols-2 gap-4">
            <UFormField label="Valor pago (R$)" required>
              <UInput v-model.number="payForm.amount" type="number" min="0" step="0.01" />
            </UFormField>
            <UFormField label="Data" required>
              <UInput v-model="payForm.paidOn" type="date" />
            </UFormField>
          </div>
          <p class="text-xs text-neutral-400 dark:text-neutral-500">
            Registrará o pagamento de <b>{{ payingAccount?.name }}</b> no mês {{ viewMonth }}/{{ viewYear }}.
          </p>
          <p v-if="payingAccount?.type === 'percent'" class="text-xs text-neutral-500 dark:text-neutral-400 rounded-lg bg-neutral-100 dark:bg-neutral-800 p-2">
            Conta <b>percentual</b>: valor sugerido de {{ payingAccount.percent }}% da soma das entradas das fontes vinculadas. Ajuste se necessário.
          </p>
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton label="Cancelar" color="neutral" variant="ghost" @click="payModalOpen = false" />
          <UButton label="Confirmar pagamento" color="success" :loading="paySaving" @click="confirmPay" />
        </div>
      </template>
    </UModal>

    <!-- Modal: entrada -->
    <UModal v-model:open="incomeModalOpen" title="Nova entrada de dinheiro" :ui="{ content: 'max-w-md' }">
      <template #body>
        <div class="space-y-4">
          <UAlert v-if="incomeErr" color="error" :title="incomeErr" icon="i-lucide-alert-circle" />
          <UFormField label="Fonte de crédito" required>
            <USelect v-model="incomeForm.sourceId" :items="sourceOptions" placeholder="Selecione uma fonte" />
          </UFormField>
          <div class="grid grid-cols-2 gap-4">
            <UFormField label="Valor (R$)" required>
              <UInput v-model.number="incomeForm.amount" type="number" min="0" step="0.01" placeholder="0,00" />
            </UFormField>
            <UFormField label="Data" required>
              <UInput v-model="incomeForm.date" type="date" />
            </UFormField>
          </div>
          <UFormField label="Descrição">
            <UInput v-model="incomeForm.description" placeholder="Ex.: Salário de julho (opcional)" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton label="Cancelar" color="neutral" variant="ghost" @click="incomeModalOpen = false" />
          <UButton label="Registrar" :loading="incomeSaving" @click="saveIncome" />
        </div>
      </template>
    </UModal>
    <!-- Modal: confirmação de remoção de pagamento -->
    <UModal v-model:open="unpayConfirmOpen" title="Remover pagamento" :ui="{ content: 'max-w-sm' }">
      <template #body>
        <div class="space-y-4">
          <div class="flex items-start gap-3">
            <div class="grid place-items-center size-10 rounded-xl shrink-0 bg-red-50 text-red-600 dark:bg-red-950/60">
              <UIcon name="i-lucide-alert-triangle" class="text-lg" />
            </div>
            <div class="text-sm text-neutral-600 dark:text-neutral-300">
              <p>
                Tem certeza que deseja remover o pagamento de
                <b class="text-neutral-900 dark:text-neutral-100">{{ pendingUnpay?.name }}</b>
                em {{ viewMonth }}/{{ viewYear }}?
              </p>
              <p class="text-xs text-neutral-400 dark:text-neutral-500 mt-1">A conta voltará a aparecer como pendente.</p>
            </div>
          </div>
          <UAlert v-if="unpayErr" color="error" :title="unpayErr" icon="i-lucide-alert-circle" />
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton label="Cancelar" color="neutral" variant="ghost" :disabled="unpaying" @click="unpayConfirmOpen = false" />
          <UButton label="Remover" color="error" :loading="unpaying" @click="confirmUnpay" />
        </div>
      </template>
    </UModal>

    <!-- Modal: confirmação de exclusão de entrada -->
    <UModal v-model:open="confirmOpen" title="Excluir entrada" :ui="{ content: 'max-w-sm' }">
      <template #body>
        <div class="space-y-4">
          <div class="flex items-start gap-3">
            <div class="grid place-items-center size-10 rounded-xl shrink-0 bg-red-50 text-red-600 dark:bg-red-950/60">
              <UIcon name="i-lucide-alert-triangle" class="text-lg" />
            </div>
            <div class="text-sm text-neutral-600 dark:text-neutral-300">
              <p>
                Tem certeza que deseja excluir a entrada de
                <b class="text-neutral-900 dark:text-neutral-100">{{ money(pendingIncome?.amount ?? 0) }}</b>
                da fonte <b class="text-neutral-900 dark:text-neutral-100">{{ pendingIncome?.sourceName }}</b>?
              </p>
              <p class="text-xs text-neutral-400 dark:text-neutral-500 mt-1">Esta ação não pode ser desfeita.</p>
            </div>
          </div>
          <UAlert v-if="deleteErr" color="error" :title="deleteErr" icon="i-lucide-alert-circle" />
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton label="Cancelar" color="neutral" variant="ghost" :disabled="deleting" @click="confirmOpen = false" />
          <UButton label="Excluir" color="error" :loading="deleting" @click="confirmRemoveIncome" />
        </div>
      </template>
    </UModal>
  </div>
</template>
