import { createMemoryHistory, createRouter } from 'vue-router'

import stockView from '../components/stock.vue'
import settingsView from '../components/settings.vue'

const routes = [
    { path: '/', component: stockView,name: 'stock' },
    { path: '/settings/:id', component: settingsView,name: 'settings' },
]

const router = createRouter({
    history: createMemoryHistory(),
    routes,
})

export default router