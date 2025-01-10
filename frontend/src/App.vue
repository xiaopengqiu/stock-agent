<script setup>
import stockInfo from './components/stock.vue'
import {
  Quit,
  WindowFullscreen,
  WindowHide,
  WindowIsFullscreen,
  WindowUnfullscreen
} from '../wailsjs/runtime'
import {h, ref} from "vue";
import { RouterLink } from 'vue-router'
import {darkTheme, NIcon} from 'naive-ui'
import {
  SettingsOutline,
  ReorderTwoOutline,
  ExpandOutline,
  RefreshOutline, PowerOutline, BarChartOutline,
} from '@vicons/ionicons5'

const content = ref('数据来源于网络,仅供参考\n投资有风险,入市需谨慎')
const isFullscreen = ref(false)
const activeKey = ref('stock')
const containerRef= ref({})
const menuOptions = ref([
  {
    label: () =>
        h(
            RouterLink,
            {
              to: {
                name: 'stock',
                params: {
                  id: 'zh-CN'
                }
              }
            },
            { default: () => '我的自选' }
        ),
    key: 'stock',
    icon: renderIcon(BarChartOutline)
  },
  {
    label: () =>
        h(
            RouterLink,
            {
              to: {
                name: 'settings',
                params: {
                  id: 'zh-CN'
                }
              }
            },
            { default: () => '设置' }
        ),
    key: 'settings',
    icon: renderIcon(SettingsOutline),
  },
  {
    label: ()=> h("a", {
      href: '#',
      onClick: toggleFullscreen
    }, { default: () => '全屏' }),
    key: 'full',
    icon: renderIcon(ExpandOutline),
  },
  {
    label: ()=> h("a", {
      href: '#',
      onClick: WindowHide,
    }, { default: () => '隐藏到托盘区' }),
    key: 'hide',
    icon: renderIcon(ReorderTwoOutline),

  },
  {
    label: ()=> h("a", {
      href: '#',
      onClick: Quit,
    }, { default: () => '退出程序' }),
    key: 'exit',
    icon: renderIcon(PowerOutline),
  },
])
function renderIcon(icon) {
  return () => h(NIcon, null, { default: () => h(icon) })
}
function toggleFullscreen(e) {
  console.log(e)
  WindowIsFullscreen().then((isFull) => {
    isFullscreen.value = isFull
    if (isFull) {
      WindowUnfullscreen()
      e.target.innerHTML = '全屏'
    } else {
      WindowFullscreen()
      e.target.innerHTML = '取消全屏'
    }
  })
}

</script>
<template>


  <n-config-provider :theme="darkTheme" ref="containerRef">
  <n-watermark
      :content="content"
      cross
      selectable
      :font-size="12"
      :line-height="12"
      :height="400"
      :width="200"
      :x-offset="50"
      :y-offset="50"
      :rotate="-15"
      style="height: 100%"
  >
  <n-flex justify="center">
    <n-message-provider >
      <n-modal-provider>
        <n-grid x-gap="12" :cols="1">
          <n-gi>
            <RouterView />
          </n-gi>
          <n-gi style="position: sticky;bottom:0;z-index: 9999;">
                  <n-card size="small">
                  <n-menu style="font-size: 18px;"
                          v-model:value="activeKey"
                          mode="horizontal"
                          :options="menuOptions"
                          responsive
                  />
                </n-card>
          </n-gi>
        </n-grid>
      </n-modal-provider>
    </n-message-provider>
  </n-flex>
  </n-watermark>
  </n-config-provider>
</template>
<style>

</style>
