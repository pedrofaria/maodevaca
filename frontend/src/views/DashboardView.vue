<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, errMsg } from '../lib/api'
import { money, monthLabel } from '../lib/format'
import { nextMonth, prevMonth, viewMonth, viewYear } from '../lib/state'
import type { MonthSummary } from '../lib/types'

const router = useRouter()
const summary = ref<MonthSummary | null>(null)
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    summary.value = await api.getMonthSummary(viewYear.value, viewMonth.value)
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch([viewYear, viewMonth], load)
</script>

<template>
  <div class="space-y-6 pb-6 lg:pb-8">
    <div class="sticky top-0 z-10 -mx-6 lg:-mx-8 px-6 lg:px-8 pt-6 lg:pt-8 pb-3 bg-neutral-50 dark:bg-neutral-950 border-b border-neutral-200/70 dark:border-neutral-800/70 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Dashboard</h1>
        <p class="text-sm text-neutral-500 dark:text-neutral-400">Resumo de <span class="font-medium">{{ monthLabel(viewYear, viewMonth) }}</span></p>
      </div>
      <div class="flex items-center gap-1">
        <UButton icon="i-lucide-chevron-left" color="neutral" variant="ghost" square @click="prevMonth" />
        <span class="text-sm font-semibold w-44 text-center">{{ monthLabel(viewYear, viewMonth) }}</span>
        <UButton icon="i-lucide-chevron-right" color="neutral" variant="ghost" square @click="nextMonth" />
      </div>
    </div>

    <UAlert v-if="error" color="error" :title="error" icon="i-lucide-alert-circle" />

    <div v-if="loading && !summary" class="text-sm text-neutral-400 dark:text-neutral-500">Carregando…</div>

    <template v-else>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <UCard>
          <div class="flex items-center justify-between">
            <span class="text-sm text-neutral-500 dark:text-neutral-400">Entradas</span>
            <UIcon name="i-lucide-arrow-down-left" class="text-emerald-600" />
          </div>
          <p class="mt-2 text-2xl font-bold text-emerald-600">{{ money(summary?.incomesTotal) }}</p>
        </UCard>
        <UCard>
          <div class="flex items-center justify-between">
            <span class="text-sm text-neutral-500 dark:text-neutral-400">Saídas</span>
            <UIcon name="i-lucide-arrow-up-right" class="text-red-600" />
          </div>
          <p class="mt-2 text-2xl font-bold text-red-600">{{ money(summary?.expensesTotal) }}</p>
        </UCard>
        <UCard>
          <div class="flex items-center justify-between">
            <span class="text-sm text-neutral-500 dark:text-neutral-400">Saldo</span>
            <UIcon name="i-lucide-scale" class="text-neutral-500" />
          </div>
          <p class="mt-2 text-2xl font-bold" :class="(summary?.balance ?? 0) >= 0 ? 'text-neutral-900 dark:text-neutral-100' : 'text-red-600'">
            {{ money(summary?.balance) }}
          </p>
        </UCard>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <UCard>
          <template #header>
            <div class="font-semibold flex items-center gap-2">
              <UIcon name="i-lucide-arrow-down-left" class="text-emerald-600" /> Entradas do mês
            </div>
          </template>
          <div v-if="summary?.incomes?.length" class="divide-y divide-neutral-100 dark:divide-neutral-800">
            <div v-for="i in summary.incomes" :key="i.id" class="flex items-center justify-between py-2.5">
              <div>
                <p class="font-medium text-sm">{{ i.sourceName }}</p>
                <p class="text-xs text-neutral-400 dark:text-neutral-500">{{ i.date }}<span v-if="i.description"> · {{ i.description }}</span></p>
              </div>
              <span class="font-semibold text-emerald-600 text-sm">{{ money(i.amount) }}</span>
            </div>
          </div>
          <div v-else class="text-sm text-neutral-400 dark:text-neutral-500 text-center py-6">Nenhuma entrada este mês.</div>
          <template #footer>
            <UButton label="Adicionar entrada" icon="i-lucide-plus" color="emerald" variant="soft" size="sm" block @click="router.push('/transacoes')" />
          </template>
        </UCard>

        <UCard>
          <template #header>
            <div class="font-semibold flex items-center gap-2">
              <UIcon name="i-lucide-arrow-up-right" class="text-red-600" /> Pagamentos do mês
            </div>
          </template>
          <div v-if="summary?.payments?.length" class="divide-y divide-neutral-100 dark:divide-neutral-800">
            <div v-for="p in summary.payments" :key="p.id" class="flex items-center justify-between py-2.5">
              <div>
                <p class="font-medium text-sm">{{ p.accountName }}</p>
                <p class="text-xs text-neutral-400 dark:text-neutral-500">{{ p.paidOn }}</p>
              </div>
              <span class="font-semibold text-red-600 text-sm">{{ money(p.amount) }}</span>
            </div>
          </div>
          <div v-else class="text-sm text-neutral-400 dark:text-neutral-500 text-center py-6">Nenhum pagamento este mês.</div>
          <template #footer>
            <UButton label="Gerenciar contas" icon="i-lucide-wallet" color="primary" variant="soft" size="sm" block @click="router.push('/contas')" />
          </template>
        </UCard>
      </div>
    </template>
  </div>
</template>
