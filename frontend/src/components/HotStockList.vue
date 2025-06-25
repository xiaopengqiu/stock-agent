<script setup lang="ts">
import {onBeforeMount, onUnmounted, ref} from 'vue'
import {HotStock} from "../../wailsjs/go/main/App";
import KLineChart from "./KLineChart.vue";

const {marketType}=defineProps(
    {
      marketType: {
        type: String,
        default: '10'
      }
    }
)
const task =ref()

const list  = ref([])

onBeforeMount(async () => {
  list.value = await HotStock(marketType)
  task.value = setInterval(async () => {
    list.value = await HotStock(marketType)
  }, 5000)
})
onUnmounted(()=>{
  clearInterval(task.value)
})

function getMarketCode(item) {
  if (item.exchange	 === 'SZ') {
    return item.code.toLowerCase()
  }
  if (item.exchange	 === 'SH') {
    return item.code.toLowerCase()
  }
  if (item.exchange	 === 'HK') {
    return (item.exchange + item.code).toLowerCase()
  }
  return ("gb_"+item.code).toLowerCase()
}
</script>

<template>
  <n-table striped size="small">
    <n-thead>
      <n-tr>
        <n-th>股票名称</n-th>
        <n-th>涨跌幅</n-th>
        <n-th>当前价格</n-th>
        <n-th>热度</n-th>
        <n-th>热度变化</n-th>
        <n-th>排名变化</n-th>
      </n-tr>
    </n-thead>
    <n-tbody>
      <n-tr v-for="item in list" :key="item.code">
        <n-td><n-text type="info">
          <n-popover trigger="hover" placement="right">
            <template #trigger>
              <n-tag type="info"  :bordered="false">  {{item.name}} {{item.code}}</n-tag>
            </template>
            <k-line-chart style="width: 800px" :code="getMarketCode(item)" :chart-height="500" :name="item.name" :k-days="20" :dark-theme="true"></k-line-chart>
          </n-popover>
        </n-text></n-td>
        <n-td><n-text :type="item.percent>0?'error':'success'">{{item.percent}}%</n-text></n-td>
        <n-td><n-text type="info">{{item.current}}</n-text></n-td>
        <n-td><n-text type="info">{{item.value}}</n-text></n-td>
        <n-td><n-text type="info">{{item.increment}}</n-text></n-td>
        <n-td><n-text type="info">{{item.rank_change}}</n-text></n-td>
      </n-tr>
    </n-tbody>
  </n-table>
</template>

<style scoped>

</style>