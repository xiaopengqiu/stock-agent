<script setup>
const { headerTitle,newsList } = defineProps({
  headerTitle: {
    type: String,
    default: '市场资讯'
  },
  newsList: {
    type: Array,
    default: () => []
  }
})
</script>

<template>
  <n-list bordered>
    <template #header>
      {{ headerTitle }}
    </template>
    <n-list-item v-for="item in newsList">
      <n-space justify="start">
        <n-text justify="start" :bordered="false" :type="item.isRed?'error':'info'">
          <n-tag size="small" :type="item.isRed?'error':'warning'" :bordered="false"> {{ item.time }}</n-tag>
          {{ item.content }}
        </n-text>
      </n-space>
      <n-space v-if="item.subjects" style="margin-top: 2px">
        <n-tag :bordered="false" type="success" size="small" v-for="sub in item.subjects">
          {{ sub }}
        </n-tag>
        <n-space v-if="item.stocks">
          <n-tag :bordered="false" type="warning" size="small" v-for="sub in item.stocks">
            {{ sub }}
          </n-tag>
        </n-space>
        <n-tag v-if="item.url" :bordered="false" type="warning" size="small">
          <a :href="item.url" target="_blank">
            <n-text type="warning">查看原文</n-text>
          </a>
        </n-tag>
      </n-space>
    </n-list-item>
  </n-list>
</template>
<style scoped>

</style>