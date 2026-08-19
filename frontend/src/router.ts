import type { RouteRecordRaw } from 'vue-router'

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'dashboard',
    component: () => import('./views/DashboardView.vue'),
    meta: { title: 'Dashboard' }
  },
  {
    path: '/contas',
    name: 'accounts',
    component: () => import('./views/ContasView.vue'),
    meta: { title: 'Contas e grupos' }
  },
  {
    path: '/transacoes',
    name: 'transactions',
    component: () => import('./views/TransacoesView.vue'),
    meta: { title: 'Transações' }
  },
  {
    path: '/fontes',
    name: 'sources',
    component: () => import('./views/FontesView.vue'),
    meta: { title: 'Fontes' }
  },
  {
    path: '/relatorios',
    name: 'reports',
    component: () => import('./views/ReportsView.vue'),
    meta: { title: 'Visualização' }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]
