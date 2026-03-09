import {createMemoryHistory, createRouter, createWebHashHistory, createWebHistory} from 'vue-router'
import stockView from '../components/stock.vue'
import settingsView from '../components/settings.vue'
import aboutView from "../components/about.vue"
import fundView from "../components/fund.vue"
import marketView from "../components/market.vue"
import agentChat from "../components/agent-chat.vue"
import mcpSettingsView from "../components/mcp-settings.vue"
import toolSettingsView from "../components/tool-settings.vue"
import aiStockPickView from "../components/ai-stock-pick.vue"
import positionAnalysisView from "../components/position-analysis.vue"

const routes = [
    { path: '/', component: stockView,name: 'stock'},
    { path: '/fund', component: fundView,name: 'fund' },
    { path: '/settings', component: settingsView,name: 'settings' },
    { path: '/about', component: aboutView,name: 'about' },
    { path: '/market', component: marketView,name: 'market' },
    { path: '/agent', component: agentChat,name: 'agent' },
    { path: '/mcp-settings', component: mcpSettingsView,name: 'mcp-settings' },
    { path: '/tool-settings', component: toolSettingsView,name: 'tool-settings' },
    { path: '/ai-stock-pick', component: aiStockPickView,name: 'ai-stock-pick' },
    { path: '/position-analysis', component: positionAnalysisView,name: 'position-analysis' },
]

const router = createRouter({
    //history: createWebHistory(),
    history: createWebHashHistory(),
    routes,
})

export default router
