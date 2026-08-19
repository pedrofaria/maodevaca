<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, errMsg } from '../lib/api'
import { money } from '../lib/format'
import { COLOR_OPTIONS, ICON_OPTIONS } from '../lib/icons'
import type { Account, Group } from '../lib/types'

const loading = ref(true)
const error = ref('')
const accounts = ref<Account[]>([])
const groups = ref<Group[]>([])

// Modal de conta
const accountModalOpen = ref(false)
const editingAccount = ref<Account | null>(null)
const accountForm = ref({ name: '', amount: 0, dueDay: 1, groupId: null as number | null, notes: '', active: true })
const accountSaving = ref(false)
const accountErr = ref('')

// Modal de grupo
const groupModalOpen = ref(false)
const editingGroup = ref<Group | null>(null)
const groupForm = ref({ name: '', icon: 'i-lucide-folder', color: '#6366f1' })
const groupSaving = ref(false)
const groupErr = ref('')

const accountGroups = computed(() => accounts.value.filter(a => a.groupId !== null))
const ungrouped = computed(() => accounts.value.filter(a => a.groupId === null))

function accountsOfGroup(groupId: number): Account[] {
  return accountGroups.value.filter(a => a.groupId === groupId)
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [accs, grps] = await Promise.all([api.getAccounts(), api.getGroups()])
    accounts.value = accs ?? []
    groups.value = grps ?? []
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ---------- conta ----------
function openNewAccount(groupId: number | null = null) {
  editingAccount.value = null
  accountForm.value = { name: '', amount: 0, dueDay: 1, groupId, notes: '', active: true }
  accountErr.value = ''
  accountModalOpen.value = true
}

function openEditAccount(acc: Account) {
  editingAccount.value = acc
  accountForm.value = { name: acc.name, amount: acc.amount, dueDay: acc.dueDay, groupId: acc.groupId, notes: acc.notes, active: acc.active }
  accountErr.value = ''
  accountModalOpen.value = true
}

async function saveAccount() {
  accountErr.value = ''
  if (!accountForm.value.name.trim()) { accountErr.value = 'Informe o nome da conta.'; return }
  if (!(accountForm.value.amount > 0)) { accountErr.value = 'Informe um valor maior que zero.'; return }
  if (accountForm.value.dueDay < 1 || accountForm.value.dueDay > 31) { accountErr.value = 'Dia de vencimento deve estar entre 1 e 31.'; return }
  accountSaving.value = true
  try {
    const f = accountForm.value
    if (editingAccount.value) {
      await api.updateAccount(editingAccount.value.id, f.name, f.amount, f.dueDay, f.groupId, f.active, f.notes)
    } else {
      await api.createAccount(f.name, f.amount, f.dueDay, f.groupId, f.notes)
    }
    accountModalOpen.value = false
    await load()
  } catch (e) {
    accountErr.value = errMsg(e)
  } finally {
    accountSaving.value = false
  }
}

async function removeAccount(acc: Account) {
  if (!confirm(`Excluir a conta "${acc.name}"? Os pagamentos dela também serão removidos.`)) return
  try {
    await api.deleteAccount(acc.id)
    await load()
  } catch (e) {
    error.value = errMsg(e)
  }
}

// ---------- grupo ----------
function openNewGroup() {
  editingGroup.value = null
  groupForm.value = { name: '', icon: ICON_OPTIONS[0], color: COLOR_OPTIONS[0] }
  groupErr.value = ''
  groupModalOpen.value = true
}

function openEditGroup(g: Group) {
  editingGroup.value = g
  groupForm.value = { name: g.name, icon: g.icon || ICON_OPTIONS[0], color: g.color || COLOR_OPTIONS[0] }
  groupErr.value = ''
  groupModalOpen.value = true
}

async function saveGroup() {
  groupErr.value = ''
  if (!groupForm.value.name.trim()) { groupErr.value = 'Informe o nome do grupo.'; return }
  groupSaving.value = true
  try {
    const f = groupForm.value
    if (editingGroup.value) {
      await api.updateGroup(editingGroup.value.id, f.name, f.icon, f.color, editingGroup.value.sortOrder)
    } else {
      await api.createGroup(f.name, f.icon, f.color)
    }
    groupModalOpen.value = false
    await load()
  } catch (e) {
    groupErr.value = errMsg(e)
  } finally {
    groupSaving.value = false
  }
}

async function removeGroup(g: Group) {
  if (!confirm(`Excluir o grupo "${g.name}"? As contas dele não serão apagadas, apenas ficarão sem grupo.`)) return
  try {
    await api.deleteGroup(g.id)
    await load()
  } catch (e) {
    error.value = errMsg(e)
  }
}

const groupOptions = computed(() => groups.value.map(g => ({ label: g.name, value: g.id })))
</script>

<template>
  <div class="space-y-6 pb-6 lg:pb-8">
    <div class="sticky top-0 z-10 -mx-6 lg:-mx-8 px-6 lg:px-8 pt-6 lg:pt-8 pb-3 bg-neutral-50 dark:bg-neutral-950 border-b border-neutral-200/70 dark:border-neutral-800/70 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Contas e grupos</h1>
        <p class="text-sm text-neutral-500 dark:text-neutral-400">Cadastro das contas recorrentes e seus grupos</p>
      </div>
      <div class="flex items-center gap-2">
        <UButton icon="i-lucide-folder-plus" label="Novo grupo" color="neutral" variant="soft" @click="openNewGroup" />
        <UButton icon="i-lucide-plus" label="Nova conta" @click="openNewAccount()" />
      </div>
    </div>

    <UAlert v-if="error" color="error" :title="error" icon="i-lucide-alert-circle" />

    <div v-if="loading && !accounts.length" class="text-sm text-neutral-400 dark:text-neutral-500">Carregando…</div>

    <template v-else>
      <!-- Grupos -->
      <div v-for="g in groups" :key="g.id" class="space-y-2">
        <div class="flex items-center justify-between pt-2">
          <div class="flex items-center gap-2">
            <UIcon :name="g.icon || 'i-lucide-folder'" class="text-lg" :style="{ color: g.color }" />
            <h2 class="font-semibold text-neutral-800 dark:text-neutral-100">{{ g.name }}</h2>
            <UBadge :label="String(accountsOfGroup(g.id).length)" color="neutral" variant="subtle" />
            <UButton icon="i-lucide-plus" color="neutral" variant="ghost" size="xs" title="Nova conta neste grupo" @click="openNewAccount(g.id)" />
          </div>
          <div class="flex items-center gap-1">
            <UButton icon="i-lucide-pencil" color="neutral" variant="ghost" size="xs" @click="openEditGroup(g)" />
            <UButton icon="i-lucide-trash" color="neutral" variant="ghost" size="xs" @click="removeGroup(g)" />
          </div>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <div
            v-for="acc in accountsOfGroup(g.id)"
            :key="acc.id"
            class="flex items-center justify-between gap-3 rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 p-3"
            :class="acc.active ? '' : 'opacity-50'"
          >
            <div class="min-w-0">
              <p class="font-medium text-sm text-neutral-900 dark:text-neutral-100 truncate">{{ acc.name }}</p>
              <p class="text-xs text-neutral-400 dark:text-neutral-500">Vence dia {{ acc.dueDay }}<span v-if="!acc.active"> · inativa</span><span v-if="acc.notes"> · {{ acc.notes }}</span></p>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <span class="font-semibold text-sm text-neutral-900 dark:text-neutral-100">{{ money(acc.amount) }}</span>
              <UButton icon="i-lucide-pencil" color="neutral" variant="ghost" size="sm" @click="openEditAccount(acc)" />
              <UButton icon="i-lucide-trash" color="neutral" variant="ghost" size="sm" @click="removeAccount(acc)" />
            </div>
          </div>
          <div v-if="!accountsOfGroup(g.id).length" class="text-xs text-neutral-400 dark:text-neutral-500 py-1 pl-1 sm:col-span-2">
            Nenhuma conta neste grupo.
          </div>
        </div>
      </div>

      <!-- Sem grupo -->
      <div v-if="ungrouped.length || !groups.length" class="space-y-2">
        <div class="flex items-center justify-between pt-2">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-inbox" class="text-lg text-neutral-400 dark:text-neutral-500" />
            <h2 class="font-semibold text-neutral-800 dark:text-neutral-100">Sem grupo</h2>
            <UBadge :label="String(ungrouped.length)" color="neutral" variant="subtle" />
          </div>
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <div
            v-for="acc in ungrouped"
            :key="acc.id"
            class="flex items-center justify-between gap-3 rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 p-3"
            :class="acc.active ? '' : 'opacity-50'"
          >
            <div class="min-w-0">
              <p class="font-medium text-sm text-neutral-900 dark:text-neutral-100 truncate">{{ acc.name }}</p>
              <p class="text-xs text-neutral-400 dark:text-neutral-500">Vence dia {{ acc.dueDay }}<span v-if="!acc.active"> · inativa</span><span v-if="acc.notes"> · {{ acc.notes }}</span></p>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <span class="font-semibold text-sm text-neutral-900 dark:text-neutral-100">{{ money(acc.amount) }}</span>
              <UButton icon="i-lucide-pencil" color="neutral" variant="ghost" size="sm" @click="openEditAccount(acc)" />
              <UButton icon="i-lucide-trash" color="neutral" variant="ghost" size="sm" @click="removeAccount(acc)" />
            </div>
          </div>
          <div v-if="!ungrouped.length && !groups.length" class="text-sm text-neutral-400 dark:text-neutral-500 text-center py-8 sm:col-span-2">
            Nenhuma conta cadastrada ainda. Clique em <b>Nova conta</b> para começar.
          </div>
        </div>
      </div>
    </template>

    <!-- Modal: conta -->
    <UModal v-model:open="accountModalOpen" :title="editingAccount ? 'Editar conta' : 'Nova conta a pagar'" :ui="{ content: 'max-w-md' }">
      <template #body>
        <div class="space-y-4">
          <UAlert v-if="accountErr" color="error" :title="accountErr" icon="i-lucide-alert-circle" />
          <UFormField label="Nome da conta" required>
            <UInput v-model="accountForm.name" placeholder="Ex.: Energia, Internet, Aluguel" />
          </UFormField>
          <div class="grid grid-cols-2 gap-4">
            <UFormField label="Valor (R$)" required>
              <UInput v-model.number="accountForm.amount" type="number" min="0" step="0.01" placeholder="0,00" />
            </UFormField>
            <UFormField label="Vence no dia" required>
              <UInput v-model.number="accountForm.dueDay" type="number" min="1" max="31" />
            </UFormField>
          </div>
          <UFormField label="Grupo">
            <USelect v-model="accountForm.groupId" :items="groupOptions" placeholder="Sem grupo" />
          </UFormField>
          <UFormField label="Observações">
            <UInput v-model="accountForm.notes" placeholder="Opcional" />
          </UFormField>
          <UCheckbox v-model="accountForm.active" label="Conta ativa (aparece nos pagamentos)" />
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton label="Cancelar" color="neutral" variant="ghost" @click="accountModalOpen = false" />
          <UButton label="Salvar" :loading="accountSaving" @click="saveAccount" />
        </div>
      </template>
    </UModal>

    <!-- Modal: grupo -->
    <UModal v-model:open="groupModalOpen" :title="editingGroup ? 'Editar grupo' : 'Novo grupo'" :ui="{ content: 'max-w-md' }">
      <template #body>
        <div class="space-y-4">
          <UAlert v-if="groupErr" color="error" :title="groupErr" icon="i-lucide-alert-circle" />
          <UFormField label="Nome do grupo" required>
            <UInput v-model="groupForm.name" placeholder="Ex.: Casa, Assinaturas, Transporte" />
          </UFormField>
          <UFormField label="Ícone">
            <div class="flex flex-wrap gap-2">
              <button
                v-for="ic in ICON_OPTIONS"
                :key="ic"
                type="button"
                class="grid place-items-center size-9 rounded-lg border"
                :class="groupForm.icon === ic ? 'border-emerald-600 bg-emerald-50 text-emerald-700 ring-2 ring-emerald-600/30' : 'border-neutral-300 bg-neutral-200 text-neutral-700 dark:border-neutral-600 dark:bg-neutral-700 dark:text-neutral-100 hover:bg-neutral-300 dark:hover:bg-neutral-600'"
                @click="groupForm.icon = ic"
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
                :class="groupForm.color === c ? 'ring-2 ring-neutral-900 dark:ring-white ring-offset-2 scale-110' : ''"
                :style="{ backgroundColor: c }"
                @click="groupForm.color = c"
              />
            </div>
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton label="Cancelar" color="neutral" variant="ghost" @click="groupModalOpen = false" />
          <UButton label="Salvar" :loading="groupSaving" @click="saveGroup" />
        </div>
      </template>
    </UModal>
  </div>
</template>
