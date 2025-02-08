import { createMemoryHistory, createRouter } from 'vue-router'

import stockView from '../components/stock.vue'
import settingsView from '../components/settings.vue'
import about from "../components/about.vue";

const routes = [
    { path: '/', component: stockView,name: 'stock' },
    { path: '/settings/:id', component: settingsView,name: 'settings' },
    { path: '/about', component: about,name: 'about' },
]

const router = createRouter({
    history: createMemoryHistory(),
    routes,
})

export default router