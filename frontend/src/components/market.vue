<script setup>
import {onBeforeMount, ref} from 'vue'
import {GetTelegraphList, GlobalStockIndexes} from "../../wailsjs/go/main/App";
import {EventsOn} from "../../wailsjs/runtime";
import NewsList from "./newsList.vue";
import KLineChart from "./KLineChart.vue";

const panelHeight = ref(window.innerHeight - 240)

const telegraphList = ref([])
const sinaNewsList = ref([])

const common = ref([])
const america = ref([])
const europe = ref([])
const asia = ref([])
const other = ref([])
const globalStockIndexes = ref(null)

function getIndex() {
  GlobalStockIndexes().then((res) => {
    globalStockIndexes.value = res
    common.value = res["common"]
    america.value = res["america"]
    europe.value = res["europe"]
    asia.value = res["asia"]
    other.value = res["other"]
    console.log(globalStockIndexes.value)
  })
}

onBeforeMount(() => {
  GetTelegraphList("财联社电报").then((res) => {
    telegraphList.value = res
  })
  GetTelegraphList("新浪财经").then((res) => {
    sinaNewsList.value = res
  })
  getIndex();

  setInterval(() => {
    getIndex()
  }, 3000)
})



EventsOn("newTelegraph", (data) => {
  for (let i = 0; i < data.length; i++) {
    telegraphList.value.pop()
  }
  telegraphList.value.unshift(...data)
})
EventsOn("newSinaNews", (data) => {
  for (let i = 0; i < data.length; i++) {
    sinaNewsList.value.pop()
  }
  sinaNewsList.value.unshift(...data)
})

//获取页面高度
window.onresize = () => {
  panelHeight.value = window.innerHeight - 240
}

function getAreaName(code){
  switch (code) {
    case "america":
      return "美国"
    case "europe":
      return "欧洲"
    case "asia":
      return "亚洲"
    case "common":
      return "常用"
    case "other":
      return "其他"
  }
}
</script>

<template>
  <n-card>
    <n-tabs type="line" animated>
      <n-tab-pane name="市场快讯" tab="市场快讯">
        <n-grid :cols="2" :y-gap="0">
          <n-gi>
            <news-list :newsList="telegraphList" :header-title="'财联社电报'"></news-list>
          </n-gi>
          <n-gi>
            <news-list :newsList="sinaNewsList" :header-title="'新浪财经'"></news-list>
          </n-gi>
        </n-grid>
      </n-tab-pane>
      <n-tab-pane name="全球股指" tab="全球股指">
        <n-tabs type="segment" animated>
          <n-tab-pane name="全球指数" tab="全球指数">
            <n-grid :cols="5" :y-gap="0">
              <n-gi v-for="(val, key) in globalStockIndexes" :key="key">
                <n-list bordered>
                  <template #header>
                    {{ getAreaName(key) }}
                  </template>
                  <n-list-item v-for="item in val" :key="item.code">
                    <n-grid :cols="3" :y-gap="0">
                      <n-gi>

                        <n-text :type="item.zdf>0?'error':'success'"><n-image :src="item.img"  width="20"/> &nbsp;{{ item.name }}</n-text>
                      </n-gi>
                      <n-gi>
                        <n-text :type="item.zdf>0?'error':'success'">{{ item.zxj }}</n-text>&nbsp;
                        <n-text :type="item.zdf>0?'error':'success'"><n-number-animation :precision="2" :from="0" :to="item.zdf" />%</n-text>

                      </n-gi>
                      <n-gi>
                        <n-text :type="item.state === 'open' ? 'success' : 'warning'">{{
                            item.state === 'open' ? '开市' : '休市'
                          }}
                        </n-text>
                      </n-gi>
                    </n-grid>
                  </n-list-item>
                </n-list>
              </n-gi>
            </n-grid>
          </n-tab-pane>
          <n-tab-pane name="上证指数" tab="上证指数">
            <k-line-chart code="sh000001" :chart-height="panelHeight" name="上证指数" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="深证成指" tab="深证成指">
            <k-line-chart code="sz399001" :chart-height="panelHeight" name="深证成指" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="创业板指" tab="创业板指">
            <k-line-chart code="sz399006" :chart-height="panelHeight" name="创业板指" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="恒生指数" tab="恒生指数">
            <k-line-chart code="hkHSI" :chart-height="panelHeight" name="恒生指数" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="纳斯达克" tab="纳斯达克">
            <k-line-chart code="us.IXIC" :chart-height="panelHeight" name="纳斯达克" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="道琼斯" tab="道琼斯">
            <k-line-chart code="us.DJI" :chart-height="panelHeight" name="道琼斯" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="标普500" tab="标普500">
            <k-line-chart code="us.INX" :chart-height="panelHeight" name="标普500" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
        </n-tabs>
      </n-tab-pane>
      <n-tab-pane name="指标行情" tab="指标行情">
        <n-tabs type="segment" animated>
          <n-tab-pane name="科创50" tab="科创50">
            <k-line-chart code="sh000688" :chart-height="panelHeight" name="科创50" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="沪深300" tab="沪深300">
            <k-line-chart code="sh000300" :chart-height="panelHeight" name="沪深300" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="上证50" tab="上证50">
            <k-line-chart code="sh000016" :chart-height="panelHeight" name="上证50" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="中证A500" tab="中证A500">
            <k-line-chart code="sh000510" :chart-height="panelHeight" name="中证A500" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="中证1000" tab="中证1000">
            <k-line-chart code="sh000852" :chart-height="panelHeight" name="中证1000" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="中证白酒" tab="中证白酒">
            <k-line-chart code="sz399997" :chart-height="panelHeight" name="中证白酒" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
          <n-tab-pane name="富时中国三倍做多" tab="富时中国三倍做多">
            <k-line-chart code="usYINN.AM" :chart-height="panelHeight" name="富时中国三倍做多" :k-days="20"
                          :dark-theme="true"></k-line-chart>
          </n-tab-pane>
        </n-tabs>
      </n-tab-pane>
    </n-tabs>
  </n-card>
</template>

<style scoped>

</style>