<script setup>
import stockInfo from './components/stock.vue'
import {
  EventsOn,
  Quit,
  WindowFullscreen, WindowGetPosition,
  WindowHide,
  WindowIsFullscreen, WindowSetPosition,
  WindowUnfullscreen
} from '../wailsjs/runtime'
import {h, ref} from "vue";
import { RouterLink } from 'vue-router'
import {darkTheme, NIcon, NText,} from 'naive-ui'
import {
  SettingsOutline,
  ReorderTwoOutline,
  ExpandOutline,
  RefreshOutline, PowerOutline, BarChartOutline, MoveOutline, WalletOutline,
} from '@vicons/ionicons5'

const content = ref('数据来源于网络,仅供参考\n投资有风险,入市需谨慎')
const isFullscreen = ref(false)
const activeKey = ref('stock')
const containerRef= ref({})
const realtimeProfit= ref("")
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
                },
              }
            },
            { default: () => '我的自选',}
        ),
    key: 'stock',
    icon: renderIcon(BarChartOutline),
    children:[
      {
        label: ()=> h(NText, {type:realtimeProfit.value>0?'error':'success'},{ default: () => '当日盈亏 '+realtimeProfit.value+"¥"}),
        key: 'realtimeProfit',
        show: realtimeProfit.value,
        icon: renderIcon(WalletOutline),
      },
    ]
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
      onClick: toggleFullscreen,
      title: '全屏 Ctrl+F 退出全屏 Esc',
    }, { default: () => isFullscreen.value?'取消全屏':'全屏' }),
    key: 'full',
    icon: renderIcon(ExpandOutline),
  },
  {
    label: ()=> h("a", {
      href: '#',
      onClick: WindowHide,
      title: '隐藏到托盘区 Ctrl+H',
    }, { default: () => '隐藏到托盘区' }),
    key: 'hide',
    icon: renderIcon(ReorderTwoOutline),
  },
  {
    label: ()=> h("a", {
      href: 'javascript:void(0)',
      style: 'cursor: move;',
      onClick: toggleStartMoveWindow,
    }, { default: () => '移动' }),
    key: 'move',
    icon: renderIcon(MoveOutline),
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
  //console.log(e)
    if (isFullscreen.value) {
      WindowUnfullscreen()
      //e.target.innerHTML = '全屏'
    } else {
      WindowFullscreen()
     // e.target.innerHTML = '取消全屏'
    }
  isFullscreen.value=!isFullscreen.value
}
const drag = ref(false)
const lastPos= ref({x:0,y:0})
function toggleStartMoveWindow(e) {
  drag.value=!drag.value
  lastPos.value={x:e.clientX,y:e.clientY}
}
function dragstart(e) {
  if (drag.value) {
    let x=e.clientX-lastPos.value.x
    let y=e.clientY-lastPos.value.y
    WindowGetPosition().then((pos) => {
      WindowSetPosition(pos.x+x,pos.y+y)
    })
  }
}
window.addEventListener('mousemove', dragstart)

EventsOn("realtime_profit",(data)=>{
  realtimeProfit.value=data
})
</script>
<template>


  <n-config-provider :theme="darkTheme" ref="containerRef">
    <n-message-provider >
      <n-notification-provider>
      <n-modal-provider>
  <n-watermark
      :content="content"
      cross
      selectable
      :font-size="18"
      :line-height="18"
      :height="400"
      :width="600"
      :x-offset="50"
      :y-offset="50"
      :rotate="-15"
      style="height: 100%"
  >
  <n-flex justify="center">

        <n-grid x-gap="12" :cols="1">
          <n-gi style="padding-bottom: 70px">
            <RouterView />
          </n-gi>
          <n-gi style="position: fixed;bottom:0;z-index: 9;width: 100%">
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

  </n-flex>
  </n-watermark>
      </n-modal-provider>
      </n-notification-provider>
    </n-message-provider>
  </n-config-provider>
</template>
<style>

</style>
