<script setup lang="ts">
import {onBeforeMount, onUnmounted, ref} from 'vue'
import {HotTopic} from "../../wailsjs/go/main/App";
const list  = ref([])
const task =ref()

onBeforeMount(async () => {
  list.value = await HotTopic(10)
  setInterval(async ()=>{
    list.value = await HotTopic(10)
  }, 1000*10)
})
onUnmounted(()=>{
  clearInterval(task.value)
})
</script>

<template>
  <n-list bordered>
    <template #header>
      股吧热门
    </template>
    <n-list-item v-for="(item, index) in list" :key="index">
        <n-thing :title="item.nickname" :description="item.desc"  >
          <template v-if="item.squareImg" #avatar>
            <n-avatar :src="item.squareImg" :size="60">
            </n-avatar>
          </template>
          <template v-if="item.stock_list" #footer>
            <n-flex>
              <n-tag type="info" v-for="(v, i) in item.stock_list" :bordered="false">
                {{v.name}}
              </n-tag>
            </n-flex>
          </template>
        </n-thing>
    </n-list-item>
  </n-list>
</template>

<style scoped>

</style>