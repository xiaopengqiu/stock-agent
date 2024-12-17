<script setup>
import stockInfo from './components/stock.vue'
import {ref} from "vue";

const content = ref('数据来源于网络,仅供参考\n投资有风险,入市需谨慎')
</script>

<template>
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
      <stockInfo/>
    </n-message-provider>
  </n-flex>
  </n-watermark>
</template>
<style>

</style>
