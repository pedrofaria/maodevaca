<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useColorMode } from './lib/colorMode'

const route = useRoute()
const { isDark, toggle } = useColorMode()

type NavItem =
  | { type: 'link'; to: string; label: string; icon: string }
  | { type: 'divider' }

const navItems: NavItem[] = [
  { type: 'link', to: '/', label: 'Dashboard', icon: 'i-lucide-chart-bar' },
  { type: 'divider' },
  { type: 'link', to: '/contas', label: 'Contas e grupos', icon: 'i-lucide-folder-tree' },
  { type: 'link', to: '/fontes', label: 'Fontes', icon: 'i-lucide-tags' },
  { type: 'divider' },
  { type: 'link', to: '/transacoes', label: 'Transações', icon: 'i-lucide-arrow-left-right' }
]

function isActive(to: string): boolean {
  return route.path === to
}
</script>

<template>
  <UApp>
    <div class="flex h-screen">
      <aside class="w-64 shrink-0 border-r border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 flex flex-col">
        <div class="flex items-center gap-2.5 px-5 h-16 border-b border-neutral-100 dark:border-neutral-800">
          <div class="grid place-items-center size-9 rounded-xl bg-emerald-600 text-white">
            <UIcon name="i-lucide-piggy-bank" class="text-lg" />
          </div>
          <div class="leading-tight">
            <p class="text-base font-bold tracking-tight text-neutral-900 dark:text-white">Mão de Vaca</p>
            <p class="text-[11px] text-neutral-400 dark:text-neutral-500">Finanças pessoais</p>
          </div>
        </div>
        <nav class="flex-1 p-3 overflow-y-auto">
          <template v-for="(item, idx) in navItems" :key="item.type === 'divider' ? 'd' + idx : item.to">
            <div v-if="item.type === 'divider'" class="my-2 border-t border-neutral-200 dark:border-neutral-800" />
            <RouterLink
              v-else
              :to="item.to"
              class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors"
              :class="isActive(item.to)
                ? 'bg-emerald-600 text-white shadow-sm'
                : 'text-neutral-600 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800'"
            >
              <UIcon :name="item.icon" class="text-lg shrink-0" />
              {{ item.label }}
            </RouterLink>
          </template>
        </nav>
        <div class="p-3 border-t border-neutral-100 dark:border-neutral-800 space-y-1">
          <button
            type="button"
            class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-neutral-600 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors"
            @click="toggle"
          >
            <UIcon :name="isDark ? 'i-lucide-sun' : 'i-lucide-moon'" class="text-lg shrink-0" />
            {{ isDark ? 'Modo claro' : 'Modo escuro' }}
          </button>
          <div class="px-3 pt-1 text-xs text-neutral-400 dark:text-neutral-500">
            Seu dinheiro sob controle. 💪
          </div>
        </div>
      </aside>
      <main class="flex-1 overflow-y-auto px-6 lg:px-8">
        <RouterView />
      </main>
    </div>
  </UApp>
</template>
