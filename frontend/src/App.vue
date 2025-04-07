<script setup>
import {
  EventsEmit,
  EventsOn,
  Quit,
  WindowFullscreen, WindowGetPosition,
  WindowHide,
  WindowSetPosition,
  WindowUnfullscreen
} from '../wailsjs/runtime'
import {h, onBeforeMount, onMounted, ref} from "vue";
import { RouterLink } from 'vue-router'
import {darkTheme, NGradientText, NIcon, NText,} from 'naive-ui'
import {
  SettingsOutline,
  ReorderTwoOutline,
  ExpandOutline,
  PowerOutline, LogoGithub, MoveOutline, WalletOutline, StarOutline, AlarmOutline, SparklesOutline,
} from '@vicons/ionicons5'
import {GetConfig} from "../wailsjs/go/main/App";
const enableNews= ref(false)
const contentStyle =  ref("")
const enableFund = ref(false)
const enableDarkTheme =  ref(null)
const content = ref('数据来源于网络,仅供参考;投资有风险,入市需谨慎')
const isFullscreen = ref(false)
const activeKey = ref('')
const containerRef= ref({})
const realtimeProfit= ref(0)
const telegraph= ref([])
const menuOptions = ref([
  {
    label: () =>
        h(
            RouterLink,
            {
              to: {
                name: 'stock',
                params: {
                },
              }
            },
            { default: () => '股票自选',}
        ),
    key: 'stock',
    icon: renderIcon(StarOutline),
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
                name: 'fund',
                params: {
                },
              }
            },
            { default: () => '基金自选',}
        ),
    show: enableFund.value,
    key: 'fund',
    icon: renderIcon(SparklesOutline),
    children:[
      {
        label: ()=> h(NText, {type:realtimeProfit.value>0?'error':'success'},{ default: () => '功能完善中！'}),
        key: 'realtimeProfit',
        show: realtimeProfit.value,
        icon: renderIcon(AlarmOutline),
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
                }
              }
            },
            { default: () => '设置' }
        ),
    key: 'settings',
    icon: renderIcon(SettingsOutline),
  },
  {
    label: () =>
        h(
            RouterLink,
            {
              to: {
                name: 'about',
                params: {
                }
              }
            },
            { default: () => '关于' }
        ),
    key: 'about',
    icon: renderIcon(LogoGithub),
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
  // {
  //   label: ()=> h("a", {
  //     href: 'javascript:void(0)',
  //     style: 'cursor: move;',
  //     onClick: toggleStartMoveWindow,
  //   }, { default: () => '移动' }),
  //   key: 'move',
  //   icon: renderIcon(MoveOutline),
  // },
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
// const drag = ref(false)
// const lastPos= ref({x:0,y:0})
// function toggleStartMoveWindow(e) {
//   drag.value=!drag.value
//   lastPos.value={x:e.clientX,y:e.clientY}
// }
// function dragstart(e) {
//   if (drag.value) {
//     let x=e.clientX-lastPos.value.x
//     let y=e.clientY-lastPos.value.y
//     WindowGetPosition().then((pos) => {
//       WindowSetPosition(pos.x+x,pos.y+y)
//     })
//   }
// }
// window.addEventListener('mousemove', dragstart)

EventsOn("realtime_profit",(data)=>{
  realtimeProfit.value=data
})
EventsOn("telegraph",(data)=>{
  telegraph.value=data
})

window.onerror = function (msg, source, lineno, colno, error) {
  // 将错误信息发送给后端
  EventsEmit("frontendError", {
    page: "App.vue",
    message: msg,
    source: source,
    lineno: lineno,
    colno: colno,
    error: error ? error.stack : null,
  });
  return true;
};

onBeforeMount(()=>{
  GetConfig().then((res)=>{
    console.log(res)
    enableFund.value=res.enableFund

    menuOptions.value.filter((item)=>{
      if(item.key==='fund'){
        item.show=res.enableFund
      }
    })

    if(res.darkTheme){
      enableDarkTheme.value=darkTheme
    }else{
      enableDarkTheme.value=null
    }
  })
})

onMounted(()=>{
  contentStyle.value="max-height: calc(90vh);overflow: hidden"
  GetConfig().then((res)=>{
    if(res.enableNews){
      enableNews.value=true
    }
    enableFund.value=res.enableFund
  })
})

</script>
<template>
  <n-config-provider  ref="containerRef" :theme="enableDarkTheme" >
    <n-message-provider >
      <n-notification-provider>
      <n-modal-provider>
        <n-dialog-provider>

        <n-watermark
      :content="content"
      cross
      selectable
      :font-size="16"
      :line-height="16"
      :width="500"
      :height="400"
      :x-offset="50"
      :y-offset="150"
      :rotate="-15"
  >
  <n-flex justify="center">
        <n-grid x-gap="12" :cols="1">
<!--
          <n-gi style="position: relative;top:1px;z-index: 19;width: 100%" v-if="telegraph.length>0">

          </n-gi>
-->

            <n-gi>
              <n-marquee :speed="100" style="position: relative;top:0;z-index: 19;width: 100%" v-if="(telegraph.length>0)&&(enableNews)">
                <n-tag type="warning" v-for="item in telegraph" style="margin-right: 10px">
                  {{item}}
                </n-tag>
                <!--              <n-text type="warning"> {{telegraph[0]}}</n-text>-->
              </n-marquee>
              <n-scrollbar :style="contentStyle">
              <RouterView />
              </n-scrollbar>
            </n-gi>


          <n-gi style="position: fixed;bottom:0;z-index: 9;width: 100%;">
                  <n-card size="small" style="--wails-draggable:drag">
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
        </n-dialog-provider>
      </n-modal-provider>
      </n-notification-provider>
    </n-message-provider>
  </n-config-provider>
</template>
<style>

</style>
