<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api, errMsg } from '../lib/api'
import { MONTHS, money } from '../lib/format'
import type { MonthSummary, YearSummary } from '../lib/types'

const tab = ref<'mes' | 'ano'>('mes')
const monthLoading = ref(false)
const yearLoading = ref(false)
const error = ref('')

const years = ref<number[]>([])
const year = ref(new Date().getFullYear())
const month = ref(new Date().getMonth() + 1)

const monthSummary = ref<MonthSummary | null>(null)
const yearSummary = ref<YearSummary | null>(null)

const yearOptions = computed(() => years.value.map(y => ({ label: String(y), value: y })))
const monthOptions = computed(() => MONTHS.map((m, i) => ({ label: m, value: i + 1 })))

function setTab(t: 'mes' | 'ano') {
  tab.value = t
  if (t === 'mes') loadMonth()
  else loadYear()
}

async function loadYears() {
  try {
    years.value = await api.getAvailableYears()
    if (!years.value.includes(year.value)) years.value = [year.value, ...years.value]
  } catch (e) {
    error.value = errMsg(e)
  }
}

async function loadMonth() {
  monthLoading.value = true
  error.value = ''
  try {
    monthSummary.value = await api.getMonthSummary(year.value, month.value)
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    monthLoading.value = false
  }
}

async function loadYear() {
  yearLoading.value = true
  error.value = ''
  try {
    yearSummary.value = await api.getYearSummary(year.value)
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    yearLoading.value = false
  }
}

// Recarrega a aba ativa quando o ano/mês muda.
watch([year, month], () => {
  if (tab.value === 'mes') loadMonth()
  else loadYear()
})

onMounted(async () => {
  await loadYears()
  await loadMonth()
})

// Barras do gráfico anual (altura em pixels relativas ao maior valor).
const CHART_H = 144 // px — deve casar com a altura fixa (h-36) do container.
const barMax = computed(() => {
  const rows = yearSummary.value?.rows ?? []
  let mx = 1
  for (const r of rows) mx = Math.max(mx, r.incomesTotal, r.expensesTotal)
  return mx
})

function barHeight(v: number): string {
  if (v <= 0) return '0px'
  return `${Math.max(3, Math.round((v / barMax.value) * CHART_H))}px`
}
</script>

<template>
  <div class="space-y-6 pb-6 lg:pb-8">
    <div class="sticky top-0 z-10 -mx-6 lg:-mx-8 px-6 lg:px-8 pt-6 lg:pt-8 pb-3 bg-neutral-50 dark:bg-neutral-950 border-b border-neutral-200/70 dark:border-neutral-800/70 space-y-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Visualização</h1>
        <p class="text-sm text-neutral-500 dark:text-neutral-400">Entradas e saídas por mês e por ano</p>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex items-center gap-1 bg-neutral-100 dark:bg-neutral-800 rounded-lg p-1">
        <button
          type="button"
          class="px-4 py-1.5 rounded-md text-sm font-medium transition-colors"
          :class="tab === 'mes' ? 'bg-white dark:bg-neutral-900 shadow-sm text-neutral-900 dark:text-white' : 'text-neutral-500 dark:text-neutral-400 hover:text-neutral-700 dark:hover:text-neutral-200'"
          @click="setTab('mes')"
        >Por mês</button>
        <button
          type="button"
          class="px-4 py-1.5 rounded-md text-sm font-medium transition-colors"
          :class="tab === 'ano' ? 'bg-white dark:bg-neutral-900 shadow-sm text-neutral-900 dark:text-white' : 'text-neutral-500 dark:text-neutral-400 hover:text-neutral-700 dark:hover:text-neutral-200'"
          @click="setTab('ano')"
        >Por ano</button>
      </div>
      <div class="flex items-center gap-2">
        <USelect v-model="year" :items="yearOptions" class="w-28" />
        <USelect v-if="tab === 'mes'" v-model="month" :items="monthOptions" class="w-40" />
      </div>
      </div>
    </div>

    <UAlert v-if="error" color="error" :title="error" icon="i-lucide-alert-circle" />

    <div v-if="tab === 'mes'">
      <div v-if="monthLoading && !monthSummary" class="text-sm text-neutral-400 dark:text-neutral-500">Carregando…</div>
      <template v-else>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <UCard>
            <p class="text-sm text-neutral-500 dark:text-neutral-400">Entradas</p>
            <p class="mt-1 text-2xl font-bold text-emerald-600">{{ money(monthSummary?.incomesTotal) }}</p>
          </UCard>
          <UCard>
            <p class="text-sm text-neutral-500 dark:text-neutral-400">Saídas</p>
            <p class="mt-1 text-2xl font-bold text-red-600">{{ money(monthSummary?.expensesTotal) }}</p>
          </UCard>
          <UCard>
            <p class="text-sm text-neutral-500 dark:text-neutral-400">Saldo</p>
            <p class="mt-1 text-2xl font-bold" :class="(monthSummary?.balance ?? 0) >= 0 ? 'text-neutral-900 dark:text-neutral-100' : 'text-red-600'">
              {{ money(monthSummary?.balance) }}
            </p>
          </UCard>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mt-4">
          <UCard>
            <template #header><div class="font-semibold">Entradas ({{ MONTHS[month - 1] }})</div></template>
            <div v-if="monthSummary?.incomes?.length" class="divide-y divide-neutral-100 dark:divide-neutral-800">
              <div v-for="i in monthSummary.incomes" :key="i.id" class="flex items-center justify-between py-2">
                <div>
                  <p class="font-medium text-sm">{{ i.sourceName }}<span v-if="i.description"> · {{ i.description }}</span></p>
                  <p class="text-xs text-neutral-400 dark:text-neutral-500">{{ i.date }}</p>
                </div>
                <span class="font-semibold text-emerald-600 text-sm">{{ money(i.amount) }}</span>
              </div>
            </div>
            <div v-else class="text-sm text-neutral-400 dark:text-neutral-500 text-center py-4">Nenhuma entrada.</div>
          </UCard>
          <UCard>
            <template #header><div class="font-semibold">Saídas ({{ MONTHS[month - 1] }})</div></template>
            <div v-if="monthSummary?.payments?.length" class="divide-y divide-neutral-100 dark:divide-neutral-800">
              <div v-for="p in monthSummary.payments" :key="p.id" class="flex items-center justify-between py-2">
                <div>
                  <p class="font-medium text-sm">{{ p.accountName }}</p>
                  <p class="text-xs text-neutral-400 dark:text-neutral-500">{{ p.paidOn }}</p>
                </div>
                <span class="font-semibold text-red-600 text-sm">{{ money(p.amount) }}</span>
              </div>
            </div>
            <div v-else class="text-sm text-neutral-400 dark:text-neutral-500 text-center py-4">Nenhum pagamento.</div>
          </UCard>
        </div>
      </template>
    </div>

    <div v-if="tab === 'ano'">
      <div v-if="yearLoading && !yearSummary" class="text-sm text-neutral-400 dark:text-neutral-500">Carregando…</div>
      <template v-else>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <UCard>
            <p class="text-sm text-neutral-500 dark:text-neutral-400">Entradas anuais</p>
            <p class="mt-1 text-2xl font-bold text-emerald-600">{{ money(yearSummary?.incomesTotal) }}</p>
          </UCard>
          <UCard>
            <p class="text-sm text-neutral-500 dark:text-neutral-400">Saídas anuais</p>
            <p class="mt-1 text-2xl font-bold text-red-600">{{ money(yearSummary?.expensesTotal) }}</p>
          </UCard>
          <UCard>
            <p class="text-sm text-neutral-500 dark:text-neutral-400">Saldo anual</p>
            <p class="mt-1 text-2xl font-bold" :class="(yearSummary?.balance ?? 0) >= 0 ? 'text-neutral-900 dark:text-neutral-100' : 'text-red-600'">
              {{ money(yearSummary?.balance) }}
            </p>
          </UCard>
        </div>

        <!-- Gráfico de barras -->
        <UCard class="mt-4">
          <template #header>
            <div class="font-semibold flex items-center gap-4">
              <span>Comparativo mensal</span>
              <span class="flex items-center gap-1 text-xs font-normal text-neutral-500 dark:text-neutral-400"><span class="inline-block size-2.5 rounded bg-emerald-500"></span> Entradas</span>
              <span class="flex items-center gap-1 text-xs font-normal text-neutral-500 dark:text-neutral-400"><span class="inline-block size-2.5 rounded bg-red-500"></span> Saídas</span>
            </div>
          </template>
          <div class="flex items-end gap-2 h-48">
            <div v-for="r in yearSummary?.rows" :key="r.month" class="flex-1 flex flex-col items-center gap-1">
              <div class="flex items-end gap-1 w-full h-36" :title="`${r.label}: entradas ${money(r.incomesTotal)} · saídas ${money(r.expensesTotal)}`">
                <div class="flex-1 bg-emerald-500 rounded-t transition-all" :style="{ height: barHeight(r.incomesTotal) }"></div>
                <div class="flex-1 bg-red-500 rounded-t transition-all" :style="{ height: barHeight(r.expensesTotal) }"></div>
              </div>
              <span class="text-[10px] text-neutral-400 dark:text-neutral-500">{{ r.label.slice(0, 3) }}</span>
            </div>
          </div>
        </UCard>

        <!-- Tabela anual -->
        <UCard class="mt-4">
          <template #header><div class="font-semibold">Detalhamento de {{ year }}</div></template>
          <UTable
            :data="yearSummary?.rows ?? []"
            :columns="[
              { accessorKey: 'label', header: 'Mês' },
              { accessorKey: 'incomesTotal', header: 'Entradas', enableSorting: true },
              { accessorKey: 'expensesTotal', header: 'Saídas', enableSorting: true },
              { accessorKey: 'balance', header: 'Saldo', enableSorting: true }
            ]"
          >
            <template #incomesTotal-cell="{ getValue }">
              <span class="text-emerald-600 font-medium">{{ money(getValue()) }}</span>
            </template>
            <template #expensesTotal-cell="{ getValue }">
              <span class="text-red-600 font-medium">{{ money(getValue()) }}</span>
            </template>
            <template #balance-cell="{ getValue }">
              <span :class="getValue() >= 0 ? 'text-neutral-800 dark:text-neutral-100' : 'text-red-600'">{{ money(getValue()) }}</span>
            </template>
          </UTable>
        </UCard>
      </template>
    </div>
  </div>
</template>
