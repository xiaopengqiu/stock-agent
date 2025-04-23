<script setup>
import {computed, h, onBeforeMount, onBeforeUnmount, onMounted, reactive, ref} from 'vue'
import {
  NFlex,
    NTimeline,
    NTimelineItem,
} from 'naive-ui'
import * as echarts from 'echarts';
import {GetTelegraphList} from "../../wailsjs/go/main/App";
import {EventsOn} from "../../wailsjs/runtime";

const telegraphList= ref([])
onBeforeMount(() => {
  GetTelegraphList().then((res) => {
    telegraphList.value = res
  })
})

EventsOn("newTelegraph", (data) => {
  for (let i = 0; i < data.length; i++) {
    telegraphList.value.pop()
  }
  telegraphList.value.unshift(...data)
})
</script>

<template>
  <n-grid :cols="2" :y-gap="6">
<!--    <n-gi>-->
<!--      <n-card title="上证指数">-->
<!--        卡片内容-->
<!--      </n-card>-->
<!--    </n-gi>-->
<!--    <n-gi>-->
<!--      <n-card title="深证成指">-->
<!--        卡片内容-->
<!--      </n-card>-->
<!--    </n-gi>-->
    <n-gi span="2">
      <n-flex justify="flex-start">
      <n-list bordered>
        <template #header>
          财联社电报
        </template>
        <n-list-item v-for="item in telegraphList"  >
          <n-space justify="start">
          <n-text   :bordered="false" :type="item.isRed?'error':'info'">  <n-tag size="small" :type="item.isRed?'error':'warning'"  :bordered="false"> {{item.time}}</n-tag>{{item.content}}</n-text>
          </n-space>
          <n-space v-if="item.subjects" style="margin-top: 2px">
            <n-tag :bordered="false" type="success" size="small" v-for="sub in item.subjects">
              {{sub}}
            </n-tag>
            <n-space v-if="item.stocks">
              <n-tag :bordered="false" type="warning" size="small" v-for="sub in item.stocks">
                {{sub}}
              </n-tag>
            </n-space>
            <n-tag v-if="item.url" :bordered="false" type="warning" size="small" >
              <a :href="item.url" target="_blank" ><n-text type="warning">查看原文</n-text></a>
            </n-tag>
          </n-space>
        </n-list-item>
      </n-list>
      </n-flex>
    </n-gi>
  </n-grid>
</template>

<style scoped>

</style>