import { createMemoryHistory, createRouter } from 'vue-router'

import stockView from '../components/stock.vue'
import settingsView from '../components/settings.vue'
import about from "../components/about.vue";
import fundView from "../components/fund.vue";

const routes = [
    { path: '/', component: stockView,name: 'stock' },
    { path: '/fund', component: fundView,name: 'fund' },
    { path: '/settings', component: settingsView,name: 'settings' },
    { path: '/about', component: about,name: 'about' },
]

const router = createRouter({
    history: createMemoryHistory(),
    routes,
})

export default router