import { computed, ref } from 'vue'

// Estado compartilhado de navegação por mês/ano entre as telas.
const now = new Date()
export const viewYear = ref(now.getFullYear())
export const viewMonth = ref(now.getMonth() + 1)

export const currentLabel = computed(() =>
  `${viewMonth.value}/${viewYear.value}`
)

export function prevMonth() {
  viewMonth.value -= 1
  if (viewMonth.value < 1) {
    viewMonth.value = 12
    viewYear.value -= 1
  }
}

export function nextMonth() {
  viewMonth.value += 1
  if (viewMonth.value > 12) {
    viewMonth.value = 1
    viewYear.value += 1
  }
}
