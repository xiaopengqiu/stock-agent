<script setup>
import {h, onBeforeMount, onBeforeUnmount, onMounted, reactive, ref} from "vue";
import {Add, ChatboxOutline} from "@vicons/ionicons5";
import {NButton, NEllipsis, NText, useMessage} from "naive-ui";
import {
  FollowFund,
  GetConfig,
  GetFollowedFund,
  GetfundList,
  GetVersionInfo,
  UnFollowFund
} from "../../wailsjs/go/main/App";
import vueDanmaku from 'vue3-danmaku'

const danmus = ref([])
const ws = ref(null)
const icon = ref(null)
const message = useMessage()

const data = reactive({
  modelName:"",
  chatId: "",
  question:"",
  name: "",
  code: "",
  fenshiURL:"",
  kURL:"",
  fullscreen: false,
  airesult: "",
  openAiEnable: false,
  loading: true,
  enableDanmu: false,
})

const followList=ref([])
const options=ref([])
const ticker=ref({})

onBeforeMount(()=>{
  GetConfig().then(result => {
    if (result.openAiEnable) {
      data.openAiEnable = true
    }
    if (result.enableDanmu) {
      data.enableDanmu = true
    }
  })
  GetFollowedFund().then(result => {
    followList.value = result
    console.log("followList",followList.value)
  })
})

onMounted(() => {
  GetVersionInfo().then((res) => {
    icon.value = res.icon;
  });
  // 创建 WebSocket 连接
  ws.value = new WebSocket('ws://8.134.249.145:16688/ws'); // 替换为你的 WebSocket 服务器地址
  //ws.value = new WebSocket('ws://localhost:16688/ws'); // 替换为你的 WebSocket 服务器地址

  ws.value.onopen = () => {
    console.log('WebSocket 连接已打开');
  };

  ws.value.onmessage = (event) => {
    if(data.enableDanmu){
      danmus.value.push(event.data);
    }
  };

  ws.value.onerror = (error) => {
    console.error('WebSocket 错误:', error);
  };

  ws.value.onclose = () => {
    console.log('WebSocket 连接已关闭');
  };

  ticker.value=setInterval(() => {
    GetFollowedFund().then(result => {
      followList.value = result
      console.log("followList",followList.value)
    })
  }, 1000*60)

})

onBeforeUnmount(() => {
  clearInterval(ticker.value)
  ws.value.close()
})



function SendDanmu(){
  ws.value.send(data.name)
}
function AddFund(){
  FollowFund(data.code).then(result=>{
    if(result){
      message.success("关注成功")
      GetFollowedFund().then(result => {
        followList.value = result
        console.log("followList",followList.value)
      })
    }
  })
}
function unFollow(code){
  UnFollowFund(code).then(result=>{
    if(result){
      message.success("取消关注成功")
      GetFollowedFund().then(result => {
        followList.value = result
        console.log("followList",followList.value)
      })
    }
  })
}

function getFundList(value){
  GetfundList(value).then(result=>{
    options.value=[]
    result.forEach(item=>{
      options.value.push({
        label: item.name,
        value: item.code,
      })
    })
  })
}
function onSelectFund(value){
  data.code=value
}
function formatterTitle(title){
  return () => h(NEllipsis,{
    style: {
      'font-size': '16px',
      'max-width': '180px',
    },
  },{default: () => title,}
  )
}
</script>

<template>
  <vue-danmaku v-model:danmus="danmus"  style="height:100px; width:100%;z-index: 9;position:absolute; top: 400px; pointer-events: none;" ></vue-danmaku>
  <n-flex justify="start" >
    <n-grid :x-gap="8" :cols="3"  :y-gap="8" >
      <n-gi v-for="info in  followList" style="margin-left: 2px" onmouseover="this.style.border='1px solid  #3498db' " onmouseout="this.style.border='0px'">
        <n-card :title="formatterTitle(info.name)">
          <template #header-extra>
            <n-tag size="small"  :bordered="false" type="info">{{info.code}}</n-tag>&nbsp;
            <n-tag size="small"  :bordered="false" type="success" @click="unFollow(info.code)"> 取消关注</n-tag>
          </template>
            <n-tag size="small" :type="info.netEstimatedUnit>0?'error':'success'" :bordered="false" v-if="info.netEstimatedUnit">估算净值：{{info.netEstimatedUnit}}({{info.netEstimatedUnitTime}})</n-tag>
          <n-divider vertical></n-divider>
          <n-tag size="small" :type="info.netUnitValue>0?'error':'success'" :bordered="false" v-if="info.netUnitValue">单位净值：{{info.netUnitValue}}({{info.netUnitValueDate}})</n-tag>
          <n-divider  v-if="info.netUnitValue"></n-divider>
            <n-flex justify="start">
            <n-tag size="small" :type="info.fundBasic.netGrowth1>0?'error':'success'" :bordered="false" v-if="info.fundBasic.netGrowth1">近一月：{{info.fundBasic.netGrowth1}}%</n-tag>
            <n-tag size="small" :type="info.fundBasic.netGrowth3>0?'error':'success'" :bordered="false" v-if="info.fundBasic.netGrowth3">近三月：{{info.fundBasic.netGrowth3}}%</n-tag>
            <n-tag size="small" :type="info.fundBasic.netGrowth6>0?'error':'success'" :bordered="false" v-if="info.fundBasic.netGrowth6">近六月：{{info.fundBasic.netGrowth6}}%</n-tag>
            <n-tag size="small" :type="info.fundBasic.netGrowth12>0?'error':'success'" :bordered="false" v-if="info.fundBasic.netGrowth12">近一年：{{info.fundBasic.netGrowth12}}%</n-tag>
            <n-tag size="small" :type="info.fundBasic.netGrowth36>0?'error':'success'" :bordered="false" v-if="info.fundBasic.netGrowth36">近三年：{{info.fundBasic.netGrowth36}}%</n-tag>
            <n-tag size="small" :type="info.fundBasic.netGrowth60>0?'error':'success'" :bordered="false" v-if="info.fundBasic.netGrowth60">近五年：{{info.fundBasic.netGrowth60}}%</n-tag>
            <n-tag size="small" :type="info.fundBasic.netGrowthYTD>0?'error':'success'" :bordered="false" v-if="info.fundBasic.netGrowthYTD" >今年来：{{info.fundBasic.netGrowthYTD}}%</n-tag>
            <n-tag size="small" :type="info.fundBasic.netGrowthAll>0?'error':'success'" :bordered="false" >成立来：{{info.fundBasic.netGrowthAll}}%</n-tag>
          </n-flex>
          <template #footer>
            <n-flex justify="start">
              <n-tag size="small"  :bordered="false" type="warning"> {{info.fundBasic.type}}</n-tag>
              <n-tag size="small"  :bordered="false" type="info"> {{info.fundBasic.company}}：{{info.fundBasic.manager}}</n-tag>
            </n-flex>
          </template>
<!--          <template #action>
            <n-flex justify="end">
            <n-button size="tiny" type="warning" @click="unFollow(info.code)">取消关注</n-button>
            </n-flex>
          </template>-->
        </n-card>
      </n-gi>
    </n-grid>
  </n-flex>
  <div style="position: fixed;bottom: 18px;right:0;z-index: 10;width: 480px">
    <n-input-group >
      <n-auto-complete  v-model:value="data.name"
                        :input-props="{
                                autocomplete: 'disabled',
                              }"
                        :options="options"
                        placeholder="基金名称/代码/弹幕"
                        clearable @update-value="getFundList" :on-select="onSelectFund"/>
      <n-button type="primary" @click="AddFund">
        <n-icon :component="Add"/> &nbsp;关注
      </n-button>
      <n-button type="error" @click="SendDanmu" v-if="data.enableDanmu">
        <n-icon :component="ChatboxOutline"/> &nbsp;发送弹幕
      </n-button>
    </n-input-group>
  </div>
</template>

<style scoped>

</style>