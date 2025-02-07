<script setup>
import {computed, h, onBeforeMount, onBeforeUnmount, onMounted, reactive, ref} from 'vue'
import {
  Follow, GetConfig,
  GetFollowList,
  GetStockList,
  Greet, NewChat, NewChatStream,
  SendDingDingMessage, SendDingDingMessageByType,
  SetAlarmChangePercent,
  SetCostPriceAndVolume, SetStockSort,
  UnFollow
} from '../../wailsjs/go/main/App'
import {
  NAvatar,
  NButton,
  NFlex,
  NForm,
  NFormItem,
  NInputNumber,
  NText,
  useMessage,
  useModal,
  useNotification
} from 'naive-ui'
import {EventsOn, WindowFullscreen, WindowReload, WindowUnfullscreen} from '../../wailsjs/runtime'
import {Add, Search,StarOutline} from '@vicons/ionicons5'
import { MdPreview } from 'md-editor-v3';
// preview.css相比style.css少了编辑器那部分样式
import 'md-editor-v3/lib/preview.css';
const mdPreviewRef = ref(null)
const message = useMessage()
const modal = useModal()
const notify = useNotification()

const stocks=ref([])
const results=ref({})
const ticker=ref({})
const stockList=ref([])
const followList=ref([])
const options=ref([])
const modalShow = ref(false)
const modalShow2 = ref(false)
const modalShow3 = ref(false)
const modalShow4 = ref(false)
const addBTN = ref(true)
const formModel = ref({
  name: "",
  code: "",
  costPrice: 0.000,
  volume: 0,
  alarm: 0,
  alarmPrice:0,
  sort:999,
})

const data = reactive({
  name: "",
  code: "",
  fenshiURL:"",
  kURL:"",
  resultText: "Please enter your name below 👇",
  fullscreen: false,
  airesult: "",
  openAiEnable: false,
  loading: true,
})

const sortedResults = computed(() => {
  //console.log("computed",sortedResults.value)
  const sortedKeys =Object.keys(results.value).sort();
  const sortedObject = {};
  sortedKeys.forEach(key => {
    sortedObject[key] = results.value[key];
  });
  return sortedObject
});

onBeforeMount(()=>{
  GetStockList("").then(result => {
    stockList.value = result
    options.value=result.map(item => {
      return {
        label: item.name+" - "+item.ts_code,
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
    message.destroyAll()
  })
  GetConfig().then(result => {
    if (result.openAiEnable) {
      data.openAiEnable = true
    }
  })
})

onMounted(() => {
  message.loading("Loading...")
 // console.log(`the component is now mounted.`)

    ticker.value=setInterval(() => {
      if(isTradingTime()){
        //monitor()
        data.fenshiURL='http://image.sinajs.cn/newchart/min/n/'+data.code+'.gif'+"?t="+Date.now()
      }
    }, 3500)
})

onBeforeUnmount(() => {
 // console.log(`the component is now unmounted.`)
  clearInterval(ticker.value)
})

EventsOn("refresh",(data)=>{
  message.success(data)
})

EventsOn("showSearch",(data)=>{
  addBTN.value = data === 1;
})

EventsOn("stock_price",(data)=>{
  //console.log("stock_price",data['股票代码'])
  updateData(data)
})

EventsOn("refreshFollowList",(data)=>{

  WindowReload()
 // message.loading("refresh...")
 //  GetFollowList().then(result => {
 //    followList.value = result
 //    for (const followedStock of result) {
 //      if (!stocks.value.includes(followedStock.StockCode)) {
 //        stocks.value.push(followedStock.StockCode)
 //      }
 //    }
 //    monitor()
 //    message.destroyAll
 //  })
})

EventsOn("newChatStream",async (msg) => {
  //console.log("newChatStream:->",data.airesult)
  data.loading = false
  if (msg === "DONE") {
    message.info("AI分析完成！")
    message.destroyAll()
  } else {
    data.airesult = data.airesult + msg
  }
})

EventsOn("updateVersion",async (msg) => {
  const githubTimeStr = msg.published_at;
  // 创建一个 Date 对象
  const utcDate = new Date(githubTimeStr);

// 获取本地时间
  const date = new Date(utcDate.getTime() + utcDate.getTimezoneOffset() * 60 * 1000);

  const year = date.getFullYear();
// getMonth 返回值是 0 - 11，所以要加 1
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  const seconds = String(date.getSeconds()).padStart(2, '0');

  const formattedDate = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;

  console.log("GitHub UTC 时间:", utcDate);
  console.log("转换后的本地时间:", formattedDate);
  notify.info({
    avatar: () =>
        h(NAvatar, {
          size: 'small',
          round: false,
          src: 'https://github.com/ArvinLovegood/go-stock/raw/master/build/appicon.png'
        }),
    title: '发现新版本: ' + msg.tag_name,
    content: () => {
      //return h(MdPreview, {theme:'dark',modelValue:msg.commit?.message}, null)
      return h('div', {
        style: {
          'text-align': 'left',
          'font-size': '14px',
        }
      }, { default: () => msg.commit?.message })
    },
    duration: 0,
    meta: "发布时间:"+formattedDate,
    action: () => {
      return h(NButton, {
        type: 'primary',
        size: 'small',
        onClick: () => {
          window.open(msg.html_url)
        }
      }, { default: () => '查看' })
    }
  })
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
    monitor()
  }else{
    message.error("已经关注了")
  }
}



function removeMonitor(code,name,key) {
  console.log("removeMonitor",name,code,key)
  stocks.value.splice(stocks.value.indexOf(code),1)
  console.log("removeMonitor-key",key)
  console.log("removeMonitor-v",results.value[key])

  delete results.value[key]
  console.log("removeMonitor-v",results.value[key])

  UnFollow(code).then(result => {
    message.success(result)
  })
}

function getStockList(value){
 // console.log("getStockList",value)
  let result;
  result=stockList.value.filter(item => item.name.includes(value)||item.ts_code.includes(value))
  options.value=result.map(item => {
    return {
      label: item.name+" - "+item.ts_code,
      value: item.ts_code
    }
  })
  if(value&&value.indexOf("-")<=0){
    data.code=value
  }
}

async function updateData(result) {
  if(result["当前价格"]<=0){
    result["当前价格"]=result["卖一报价"]
  }

  if (result.changePercent>0) {
    result.type="error"
    result.color="#E88080"
  }else if (result.changePercent<0) {
    result.type="success"
    result.color="#63E2B7"
  }else {
    result.type="default"
    result.color="#FFFFFF"
  }

    if(result.profitAmount>0){
      result.profitType="error"
    }else if(result.profitAmount<0){
      result.profitType="success"
    }
    if(result["当前价格"]){
      if(result.alarmChangePercent>0&&Math.abs(result.changePercent)>=result.alarmChangePercent){
        SendMessage(result,1)
      }

      if(result.alarmPrice>0&&result["当前价格"]>=result.alarmPrice){
        SendMessage(result,2)
      }

      if(result.costPrice>0&&result["当前价格"]>=result.costPrice){
        SendMessage(result,3)
      }
    }

  result.key=GetSortKey(result.sort,result["股票代码"])
  results.value[GetSortKey(result.sort,result["股票代码"])]=result
}


async function monitor() {
  for (let code of stocks.value) {
   // console.log(code)
    Greet(code).then(result => {
      updateData(result)
    })
  }
}

//数字长度不够前面补0
function padZero(num, length) {
  return (Array(length).join('0') + num).slice(-length);
}

function GetSortKey(sort,code){
  return padZero(sort,6)+"_"+code
}

function onSelect(item) {
  //console.log("onSelect",item)

  if(item.indexOf("-")>0){
    item=item.split("-")[1].toLowerCase()
  }
  if(item.indexOf(".")>0){
    data.code=item.split(".")[1].toLowerCase()+item.split(".")[0]
  }

}

function search(code,name){
  setTimeout(() => {
    //window.open("https://xueqiu.com/S/"+code)
    //window.open("https://www.cls.cn/stock?code="+code)
    //window.open("https://quote.eastmoney.com/"+code+".html")
    //window.open("https://finance.sina.com.cn/realstock/company/"+code+"/nc.shtml")
    window.open("https://www.iwencai.com/unifiedwap/result?w="+code)
    //window.open("https://www.iwencai.com/chat/?question="+code)
  }, 500)
}
function setStock(code,name){
    let res=followList.value.filter(item => item.StockCode===code)
    //console.log("res:",res)
    formModel.value.name=name
    formModel.value.code=code
    formModel.value.volume=res[0].Volume
    formModel.value.costPrice=res[0].CostPrice
    formModel.value.alarm=res[0].AlarmChangePercent
    formModel.value.alarmPrice=res[0].AlarmPrice
    formModel.value.sort=res[0].Sort
    modalShow.value=true
}

function showFenshi(code,name){
  data.code=code
  data.name=name
  data.fenshiURL='http://image.sinajs.cn/newchart/min/n/'+data.code+'.gif'+"?t="+Date.now()
  modalShow2.value=true
}
function showK(code,name){
  data.code=code
  data.name=name
  data.kURL='http://image.sinajs.cn/newchart/daily/n/'+data.code+'.gif'+"?t="+Date.now()
  modalShow3.value=true
}


function updateCostPriceAndVolumeNew(code,price,volume,alarm,formModel){

  if(formModel.sort){
    SetStockSort(formModel.sort,code).then(result => {
      //message.success(result)
    })
  }

  if(alarm||formModel.alarmPrice){
    SetAlarmChangePercent(alarm,formModel.alarmPrice,code).then(result => {
      //message.success(result)
    })
  }
  SetCostPriceAndVolume(code,price,volume).then(result => {
    modalShow.value=false
    message.success(result)
    GetFollowList().then(result => {
      followList.value = result
      for (const followedStock of result) {
        if (!stocks.value.includes(followedStock.StockCode)) {
          stocks.value.push(followedStock.StockCode)
        }
      }
      monitor()
      message.destroyAll()
    })
  })
}

function fullscreen(){
  if(data.fullscreen){
    WindowUnfullscreen()
  }else{
    WindowFullscreen()
  }
  data.fullscreen=!data.fullscreen
}


//type 报警类型: 1 涨跌报警;2 股价报警 3 成本价报警
function SendMessage(result,type){
  let typeName=getTypeName(type)
  let img='http://image.sinajs.cn/newchart/min/n/'+result["股票代码"]+'.gif'+"?t="+Date.now()
  let markdown="### go-stock ["+typeName+"]\n\n"+
      "### "+result["股票名称"]+"("+result["股票代码"]+")\n" +
      "- 当前价格: "+result["当前价格"]+"  "+result.changePercent+"%\n" +
      "- 最高价: "+result["今日最高价"]+"  "+result.highRate+"\n" +
      "- 最低价: "+result["今日最低价"]+"  "+result.lowRate+"\n" +
      "- 昨收价: "+result["昨日收盘价"]+"\n" +
      "- 今开价: "+result["今日开盘价"]+"\n" +
      "- 成本价: "+result.costPrice+"  "+result.profit+"%  "+result.profitAmount+" ¥\n" +
      "- 成本数量: "+result.costVolume+"股\n" +
      "- 日期: "+result["日期"]+"  "+result["时间"]+"\n\n"+
      "![image]("+img+")\n"
  let title=result["股票名称"]+"("+result["股票代码"]+") "+result["当前价格"]+" "+result.changePercent

  let msg='{' +
      '     "msgtype": "markdown",' +
      '     "markdown": {' +
      '         "title":"['+typeName+"]"+title+'",' +
      '         "text": "'+markdown+'"' +
      '     },' +
      '      "at": {' +
      '          "isAtAll": true' +
      '      }' +
      ' }'
   // SendDingDingMessage(msg,result["股票代码"])
    SendDingDingMessageByType(msg,result["股票代码"],type)
}

function aiCheckStock(stock,stockCode){
  data.airesult=""
  data.name=stock
  data.code=stockCode
  data.loading=true
  modalShow4.value=true
  message.loading("ai检测中...",{
    duration: 0,
  })
  NewChatStream(stock,stockCode)
}

function getTypeName(type){
  switch (type)
  {
    case 1:
      return "涨跌报警"
    case 2:
      return "股价报警"
    case 3:
      return "成本价报警"
    default:
      return ""
  }
}

//获取高度
function getHeight() {
  return document.documentElement.clientHeight
}


</script>

<template>
  <n-grid :x-gap="8" :cols="3"  :y-gap="8" >
      <n-gi v-for="result in sortedResults" style="margin-left: 2px" onmouseover="this.style.border='1px solid  #3498db' " onmouseout="this.style.border='0px'">
         <n-card   :data-code="result['股票代码']" :bordered="false" :title="result['股票名称']"   :closable="false" @close="removeMonitor(result['股票代码'],result['股票名称'],result.key)">
           <n-grid :cols="1" :y-gap="6">
             <n-gi>
               <n-text :type="result.type" >
                 <n-number-animation :duration="1000" :precision="2" :from="result['上次当前价格']" :to="Number(result['当前价格'])" />
               </n-text>
               <n-text style="padding-left: 10px;" :type="result.type">
                 <n-number-animation :duration="1000" :precision="3" :from="0" :to="result.changePercent" />%
               </n-text>&nbsp;
               <n-text size="small" v-if="result.costVolume>0" :type="result.type">
                 <n-number-animation  :duration="1000" :precision="2" :from="0" :to="result.profitAmountToday" />
               </n-text>
             </n-gi>
           </n-grid>
             <n-grid :cols="2" :y-gap="4" :x-gap="4"  >
               <n-gi>
                 <n-text :type="'info'">{{"最高 "+result["今日最高价"]+" "+result.highRate }}%</n-text>
               </n-gi>
               <n-gi>
                 <n-text :type="'info'">{{"最低 "+result["今日最低价"]+" "+result.lowRate }}%</n-text>
               </n-gi>
               <n-gi>
                 <n-text :type="'info'">{{"昨收 "+result["昨日收盘价"]}}</n-text>
               </n-gi>
               <n-gi>
                 <n-text :type="'info'">{{"今开 "+result["今日开盘价"]}}</n-text>
               </n-gi>
             </n-grid>
           <template #header-extra>
             <n-button size="tiny" secondary type="primary" @click="removeMonitor(result['股票代码'],result['股票名称'],result.key)">
               取消关注
             </n-button>&nbsp;
             <n-button size="tiny" v-if="data.openAiEnable" secondary type="warning" @click="aiCheckStock(result['股票名称'],result['股票代码'])"> AI分析 </n-button>

           </template>
           <template #footer>
             <n-flex justify="center">
               <n-tag size="small" v-if="result.volume>0" :type="result.profitType">{{result.volume+"股"}}</n-tag>
              <n-tag size="small" v-if="result.costPrice>0" :type="result.profitType">{{"成本:"+result.costPrice+"*"+result.costVolume+" "+result.profit+"%"+" ( "+result.profitAmount+" ¥ )"}}</n-tag>
             </n-flex>
           </template>
           <template #action>
             <n-flex justify="space-between">
               <n-text :type="'info'">{{result["日期"]+" "+result["时间"]}}</n-text>
               <n-button size="tiny" type="info" @click="setStock(result['股票代码'],result['股票名称'])"> 成本 </n-button>
               <n-button size="tiny" type="success" @click="showFenshi(result['股票代码'],result['股票名称'])"> 分时 </n-button>
               <n-button size="tiny" type="error" @click="showK(result['股票代码'],result['股票名称'])"> 日K </n-button>
               <n-button size="tiny" type="warning" @click="search(result['股票代码'],result['股票名称'])"> 详情 </n-button>
             </n-flex>
           </template>
         </n-card >
      </n-gi>
    </n-grid>
  <div style="position: fixed;bottom: 18px;right:0;z-index: 10;width: 350px">
<!--    <n-card :bordered="false">-->
      <n-input-group>
<!--        <n-button  type="error" @click="addBTN=!addBTN" > <n-icon :component="Search"/>&nbsp;<n-text  v-if="addBTN">隐藏</n-text></n-button>-->
        <n-auto-complete v-model:value="data.name"  v-if="addBTN"
                         :input-props="{
                                autocomplete: 'disabled',
                              }"
                         :options="options"
                         placeholder="请输入股票/指数名称或者代码"
                         clearable @update-value="getStockList" :on-select="onSelect"/>
        <n-button type="primary" @click="AddStock"  v-if="addBTN">
          <n-icon :component="Add"/> &nbsp;关注该股票
        </n-button>
      </n-input-group>
<!--    </n-card>-->
  </div>
      <n-modal transform-origin="center" size="small" v-model:show="modalShow" :title="formModel.name" style="width: 400px" :preset="'card'">
            <n-form :model="formModel" :rules="{
              costPrice: { required: true, message: '请输入成本'},
              volume: { required: true, message: '请输入数量'},
              alarm:{required: true, message: '涨跌报警值'} ,
              alarmPrice: { required: true, message: '请输入报警价格'},
              sort: { required: true, message: '请输入排序值'},
            }" label-placement="left" label-width="80px">
              <n-form-item label="股票成本" path="costPrice">
                <n-input-number v-model:value="formModel.costPrice" min="0"  placeholder="请输入股票成本" >
                  <template #suffix>
                    ¥
                  </template>
                </n-input-number>
              </n-form-item>
              <n-form-item label="股票数量" path="volume">
                <n-input-number v-model:value="formModel.volume"  min="0" step="100" placeholder="请输入股票数量" >
                  <template #suffix>
                    股
                  </template>
                </n-input-number>
              </n-form-item>
              <n-form-item label="涨跌提醒" path="alarm">
              <n-input-number v-model:value="formModel.alarm"  min="0" placeholder="请输入涨跌报警值(%)" >
                <template #suffix>
                  %
                </template>
              </n-input-number>
              </n-form-item>
              <n-form-item label="股价提醒" path="alarmPrice">
                <n-input-number v-model:value="formModel.alarmPrice"  min="0" placeholder="请输入股价报警值(¥)" >
                  <template #suffix>
                    ¥
                  </template>
                </n-input-number>
              </n-form-item>
              <n-form-item label="股票排序" path="sort">
                <n-input-number v-model:value="formModel.sort"  min="0" placeholder="请输入股价排序值" >
                </n-input-number>
              </n-form-item>
            </n-form>
            <template #footer>
              <n-button type="primary" @click="updateCostPriceAndVolumeNew(formModel.code,formModel.costPrice,formModel.volume,formModel.alarm,formModel)">保存</n-button>
            </template>
      </n-modal>

  <n-modal v-model:show="modalShow2" :title="data.name" style="width: 600px" :preset="'card'">
    <n-image :src="data.fenshiURL" />
  </n-modal>
  <n-modal v-model:show="modalShow3" :title="data.name" style="width: 600px" :preset="'card'">
    <n-image :src="data.kURL" />
  </n-modal>

  <n-modal transform-origin="center" v-model:show="modalShow4"  preset="card" style="width: 800px;height: 480px" :title="'['+data.name+']AI分析结果'" >
    <n-spin size="small" :show="data.loading">
      <MdPreview  ref="mdPreviewRef" style="height: 380px" :modelValue="data.airesult" :theme="'dark'"/>
    </n-spin>
  </n-modal>
</template>

<style scoped>

</style>
