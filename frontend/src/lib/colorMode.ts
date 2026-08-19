import { computed, ref } from 'vue'

type Mode = 'light' | 'dark'
const STORAGE_KEY = 'maodevaca-color-mode'

function initialMode(): Mode {
  if (typeof window === 'undefined') return 'light'
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored === 'light' || stored === 'dark') return stored
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark'
  return 'light'
}

// Estado reativo compartilhado por toda a aplicação.
export const colorMode = ref<Mode>(initialMode())

function apply(): void {
  if (typeof document === 'undefined') return
  document.documentElement.classList.toggle('dark', colorMode.value === 'dark')
}

// Aplica o tema assim que o módulo é carregado.
apply()

// Booleano derivado do modo atual — verdadeiro quando o tema é escuro.
export const isDark = computed(() => colorMode.value === 'dark')

export function setColorMode(mode: Mode): void {
  colorMode.value = mode
  localStorage.setItem(STORAGE_KEY, mode)
  apply()
}

export function toggleColorMode(): void {
  setColorMode(colorMode.value === 'dark' ? 'light' : 'dark')
}

// Composable usado nos componentes.
export function useColorMode() {
  return {
    isDark,
    toggle: toggleColorMode
  }
}
