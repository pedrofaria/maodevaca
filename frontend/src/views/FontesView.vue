<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, errMsg } from '../lib/api'
import { COLOR_OPTIONS, ICON_OPTIONS } from '../lib/icons'
import type { IncomeSource } from '../lib/types'

const loading = ref(true)
const error = ref('')
const sources = ref<IncomeSource[]>([])

// Modal de fonte
const sourceModalOpen = ref(false)
const editingSource = ref<IncomeSource | null>(null)
const sourceForm = ref({ name: '', icon: ICON_OPTIONS[2], color: COLOR_OPTIONS[0] })
const sourceSaving = ref(false)
const sourceErr = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    sources.value = (await api.getIncomeSources()) ?? []
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)

function openNewSource() {
  editingSource.value = null
  sourceForm.value = { name: '', icon: ICON_OPTIONS[2], color: COLOR_OPTIONS[0] }
  sourceErr.value = ''
  sourceModalOpen.value = true
}

function openEditSource(s: IncomeSource) {
  editingSource.value = s
  sourceForm.value = { name: s.name, icon: s.icon || ICON_OPTIONS[2], color: s.color || COLOR_OPTIONS[0] }
  sourceErr.value = ''
  sourceModalOpen.value = true
}

async function saveSource() {
  sourceErr.value = ''
  if (!sourceForm.value.name.trim()) { sourceErr.value = 'Informe o nome da fonte.'; return }
  sourceSaving.value = true
  try {
    const f = sourceForm.value
    if (editingSource.value) {
      await api.updateIncomeSource(editingSource.value.id, f.name, f.icon, f.color)
    } else {
      await api.createIncomeSource(f.name, f.icon, f.color)
    }
    sourceModalOpen.value = false
    await load()
  } catch (e) {
    sourceErr.value = errMsg(e)
  } finally {
    sourceSaving.value = false
  }
}

async function removeSource(s: IncomeSource) {
  if (!confirm(`Excluir a fonte "${s.name}"? As entradas dela também serão removidas.`)) return
  try {
    await api.deleteIncomeSource(s.id)
    await load()
  } catch (e) {
    error.value = errMsg(e)
  }
}
</script>

<template>
  <div class="space-y-6 pb-6 lg:pb-8">
    <div class="sticky top-0 z-10 -mx-6 lg:-mx-8 px-6 lg:px-8 pt-6 lg:pt-8 pb-3 bg-neutral-50 dark:bg-neutral-950 border-b border-neutral-200/70 dark:border-neutral-800/70 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Fontes</h1>
        <p class="text-sm text-neutral-500 dark:text-neutral-400">Fonte de crédito usadas ao registrar entradas</p>
      </div>
      <UButton icon="i-lucide-plus" label="Nova fonte" @click="openNewSource" />
    </div>

    <UAlert v-if="error" color="error" :title="error" icon="i-lucide-alert-circle" />

    <div v-if="loading && !sources.length" class="text-sm text-neutral-500 dark:text-neutral-400">Carregando…</div>

    <div v-else-if="sources.length" class="grid grid-cols-1 sm:grid-cols-2 gap-2">
      <div
        v-for="s in sources"
        :key="s.id"
        class="flex items-center justify-between gap-3 rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 p-3"
      >
        <div class="flex items-center gap-3 min-w-0">
          <div
            class="grid place-items-center size-10 rounded-xl shrink-0 text-white"
            :style="{ backgroundColor: s.color || '#10b981' }"
          >
            <UIcon :name="s.icon || 'i-lucide-briefcase'" class="text-lg" />
          </div>
          <div class="min-w-0">
            <p class="font-medium text-neutral-900 dark:text-neutral-100 truncate">{{ s.name }}</p>
            <p class="text-xs text-neutral-400 dark:text-neutral-500">Fonte de crédito</p>
          </div>
        </div>
        <div class="flex items-center gap-1 shrink-0">
          <UButton icon="i-lucide-pencil" color="neutral" variant="ghost" size="sm" @click="openEditSource(s)" />
          <UButton icon="i-lucide-trash" color="neutral" variant="ghost" size="sm" @click="removeSource(s)" />
        </div>
      </div>
    </div>

    <div v-else class="text-sm text-neutral-400 dark:text-neutral-500 text-center py-10">
      Nenhuma fonte cadastrada. Crie fontes como <i>Salário</i>, <i>Aluguel</i> ou <i>Rendimentos</i> para usar nas <b>Entradas</b>.
    </div>

    <!-- Modal: fonte -->
    <UModal v-model:open="sourceModalOpen" :title="editingSource ? 'Editar fonte' : 'Nova fonte de crédito'" :ui="{ content: 'max-w-md' }">
      <template #body>
        <div class="space-y-4">
          <UAlert v-if="sourceErr" color="error" :title="sourceErr" icon="i-lucide-alert-circle" />
          <UFormField label="Nome da fonte" required>
            <UInput v-model="sourceForm.name" placeholder="Ex.: Salário, Aluguel, Freelance" />
          </UFormField>
          <UFormField label="Ícone">
            <div class="flex flex-wrap gap-2">
              <button
                v-for="ic in ICON_OPTIONS"
                :key="ic"
                type="button"
                class="grid place-items-center size-9 rounded-lg border"
                :class="sourceForm.icon === ic ? 'border-emerald-600 bg-emerald-50 text-emerald-700 ring-2 ring-emerald-600/30' : 'border-neutral-300 bg-neutral-200 text-neutral-700 dark:border-neutral-600 dark:bg-neutral-700 dark:text-neutral-100 hover:bg-neutral-300 dark:hover:bg-neutral-600'"
                @click="sourceForm.icon = ic"
              >
                <UIcon :name="ic" />
              </button>
            </div>
          </UFormField>
          <UFormField label="Cor">
            <div class="flex flex-wrap gap-2">
              <button
                v-for="c in COLOR_OPTIONS"
                :key="c"
                type="button"
                class="size-8 rounded-full border-2 border-white shadow-sm"
                :class="sourceForm.color === c ? 'ring-2 ring-neutral-900 dark:ring-white ring-offset-2 scale-110' : ''"
                :style="{ backgroundColor: c }"
                @click="sourceForm.color = c"
              />
            </div>
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton label="Cancelar" color="neutral" variant="ghost" @click="sourceModalOpen = false" />
          <UButton label="Salvar" :loading="sourceSaving" @click="saveSource" />
        </div>
      </template>
    </UModal>
  </div>
</template>
