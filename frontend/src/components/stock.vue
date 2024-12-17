<script setup>
import {h, onBeforeMount, onBeforeUnmount, onMounted, reactive, ref} from 'vue'
import {Greet, Follow, UnFollow, GetFollowList, GetStockList, SetCostPriceAndVolume} from '../../wailsjs/go/main/App'
import {NButton, NFlex, NForm, NFormItem, NInput, NInputNumber, NText, useMessage, useModal} from 'naive-ui'

const message = useMessage()
const modal = useModal()

const stocks=ref([])
const results=ref({})
const ticker=ref({})
const stockList=ref([])
const followList=ref([])
const options=ref([])

const formModel = ref({
  costPrice: 0.000,
  volume: 0
})

const form = h(NForm, { labelPlacement: 'left',model: formModel}, [
  h(NFormItem, { label: '成本(元)',path:"costPrice" }, [h(NInputNumber, {path:"costPrice", onUpdateValue: updateCostPrice, clearable: true,precision:3, placeholder: '买入成本价(元)', })]),
  h(NFormItem, { label: '数量(股)',path:"volume" }, [h(NInputNumber, {path:"volume",onUpdateValue: updateVolume, clearable: true,precision:0, placeholder: '买入股数,1手=100股', })]),
]);


const data = reactive({
  name: "",
  code: "",
  resultText: "Please enter your name below 👇",
})


onBeforeMount(()=>{
  GetStockList("").then(result => {
    stockList.value = result
    options.value=result.map(item => {
      return {
        label: item.name+" "+item.ts_code,
        value: item.ts_code
      }
    })
  })
  GetFollowList().then(result => {
    followList.value = result
    for (const followedStock of result) {
      if (!stocks.value.includes(followedStock.StockCode)) {
        stocks.value.push(followedStock.StockCode)
      }
    }
    monitor()
    message.destroyAll
  })
})

onMounted(() => {
  message.loading("Loading...")
  console.log(`the component is now mounted.`)

    ticker.value=setInterval(() => {
      if(isTradingTime()){
        monitor()
      }
    }, 1000)

})

onBeforeUnmount(() => {
  console.log(`the component is now unmounted.`)
  clearInterval(ticker.value)
})


//判断是否是A股交易时间
function isTradingTime() {
  const now = new Date();
  const day = now.getDay(); // 获取星期几，0表示周日，1-6表示周一至周六
  if (day >= 1 && day <= 5) { // 周一至周五
    const hours = now.getHours();
    const minutes = now.getMinutes();
    const totalMinutes = hours * 60 + minutes;
    const startMorning = 9 * 60 + 15; // 上午9点15分换算成分钟数
    const endMorning = 11 * 60 + 30; // 上午11点30分换算成分钟数
    const startAfternoon = 13 * 60; // 下午13点换算成分钟数
    const endAfternoon = 15 * 60; // 下午15点换算成分钟数
    if ((totalMinutes >= startMorning && totalMinutes < endMorning) ||
        (totalMinutes >= startAfternoon && totalMinutes < endAfternoon)) {
      return true;
    }
  }
  return false;
}

function AddStock(){
  if (!stocks.value.includes(data.code)) {
      stocks.value.push(data.code)
      Follow(data.code).then(result => {
        message.success(result)
      })
  }
  monitor()
}



function removeMonitor(code,name) {
  console.log("removeMonitor",name,code)
  stocks.value.splice(stocks.value.indexOf(code),1)
  delete results.value[name]
  UnFollow(code).then(result => {
    message.success(result)
  })
}

function getStockList(){
  let result;
  result=stockList.value.filter(item => item.name.includes(data.name)||item.ts_code.includes(data.name))
  options.value=result.map(item => {
    return {
      label: item.name+" "+item.ts_code,
      value: item.ts_code
    }
  })
}

function monitor() {
  for (let code of stocks.value) {
   // console.log(code)
    Greet(code).then(result => {
      let s=(result["当前价格"]-result["昨日收盘价"])*100/result["昨日收盘价"]
      let roundedNum = s.toFixed(2);  // 将数字转换为保留两位小数的字符串形式
      result.s=roundedNum+"%"
      if (roundedNum>0) {
        result.type="error"
        result.color="#E88080"
      }else if (roundedNum<0) {
        result.type="success"
        result.color="#63E2B7"
      }else {
        result.type="default"
        result.color="#FFFFFF"
      }
      let res= followList.value.filter(item => item.StockCode===code)
      if (res.length>0) {
        result.costPrice=res[0].CostPrice
        result.volume=res[0].Volume
        result.profit=((result["当前价格"]-result.costPrice)*100/result.costPrice).toFixed(3)
      }
      results.value[result["股票名称"]]=result
    })
  }
}
function onSelect(item) {
  console.log(item)
  data.code=item.split(".")[1].toLowerCase()+item.split(".")[0]
}

function search(code,name){
  setTimeout(() => {
    window.open("https://xueqiu.com/S/"+code)
  }, 500)
}
function updateCostPrice(v) {
  console.log(formModel.value.costPrice)
  formModel.value.costPrice=v
  console.log(formModel.value.costPrice)

}

function updateVolume(v) {
  console.log(formModel.value.volume)
  formModel.value.volume=v
  console.log(formModel.value.volume)
}

function setStock(code,name){

  let res=followList.value.filter(item => item.StockCode===code)
  console.log("res:",res)
   formModel.value.volume=res[0].Volume
   formModel.value.costPrice=res[0].CostPrice

  const m = modal.create({
    title: name,
    preset: 'card',
    style: {
      width: '400px'
    },
    content: () => form,
    footer: () =>
        h(NFlex, { justify: 'center' },[
          h(
              NButton,
              {size:'small', type: 'primary', onClick: () =>updateCostPriceAndVolume(m,code,formModel.value.costPrice,formModel.value.volume) },
              () => '保存'
          ),
            h(
            NButton,
            { size:'small', type: 'warning', onClick: () => m.destroy() },
            () => '关闭'
        ),
        ])


  })
}
function updateCostPriceAndVolume(m,code,price,volume){
  console.log(code,price,volume)
  SetCostPriceAndVolume(code,price,volume).then(result => {
    message.success(result)
    m.destroy()
    GetFollowList().then(result => {
      followList.value = result
      for (const followedStock of result) {
        if (!stocks.value.includes(followedStock.StockCode)) {
          stocks.value.push(followedStock.StockCode)
        }
      }
      monitor()
      message.destroyAll
    })
  })
}
</script>

<template>
    <n-grid :x-gap="8" :cols="3"  :y-gap="8">
      <n-gi v-for="result in results" >
         <n-card size="small" :data-code="result['股票代码']" :bordered="false" :title="result['股票名称']"  :content-style="'font-size: 18px;'" closable @close="removeMonitor(result['股票代码'],result['股票名称'])">
           <n-grid :cols="1" :y-gap="6">
             <n-gi>
               <n-text :type="result.type" >{{result["当前价格"]}}</n-text><n-text style="padding-left: 10px;" :type="result.type">{{ result.s}}</n-text>
             </n-gi>
           </n-grid>
           <n-grid :cols="2" :y-gap="4" :x-gap="4" :item-style="'font-size: 14px;'">
             <n-gi>
               <n-text :type="'info'">{{"最高 "+result["今日最高价"]}}</n-text>
             </n-gi>
             <n-gi>
               <n-text :type="'info'">{{"最低 "+result["今日最低价"]}}</n-text>
             </n-gi>
             <n-gi>
               <n-text :type="'info'">{{"昨收 "+result["昨日收盘价"]}}</n-text>
             </n-gi>
             <n-gi>
               <n-text :type="'info'">{{"今开 "+result["今日开盘价"]}}</n-text>
             </n-gi>
           </n-grid>
           <template #header-extra>
             <n-tag size="small" v-if="result.volume>0" :type="result.type">{{result.volume+"股"}}</n-tag>
           </template>
           <template #action>
             <n-flex justify="space-between">
               <n-tag size="small" v-if="result.costPrice>0" :type="result.type">{{"成本:"+result.costPrice+"  "+result.profit+"%"}}</n-tag>
               <n-button size="tiny" type="info" @click="setStock(result['股票代码'],result['股票名称'])"> 设置 </n-button>
                <n-button size="tiny" type="warning" @click="search(result['股票代码'],result['股票名称'])"> 详情 </n-button>
             </n-flex>
           </template>
         </n-card >
      </n-gi>
    </n-grid>
          <n-auto-complete v-model:value="data.name" type="text"
                           :input-props="{
                              autocomplete: 'disabled',
                            }"
                           :options="options"
                           placeholder="股票名称或者代码"
                           clearable class="input" @input="getStockList" :on-select="onSelect"/>
          <n-button type="info" @click="AddStock"> 添加 </n-button>
</template>

<style scoped>
.result {
  height: 20px;
  line-height: 20px;
  margin: 1.5rem auto;
}
.input-box {
  text-align: center;
}
.input {
  width: 200px;
  margin-right: 10px;
}

.light-green {
  height: 108px;
  background-color: rgba(0, 128, 0, 0.12);
}
.green {
  height: 108px;
  background-color: rgba(0, 128, 0, 0.24);
}
</style>
