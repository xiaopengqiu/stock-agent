<script setup>

import {onMounted, ref, watch} from "vue";
import {GetConfig, SendDingDingMessageByType, UpdateConfig} from "../../wailsjs/go/main/App";
import {useMessage} from "naive-ui";
import {data} from "../../wailsjs/go/models";
const message = useMessage()

const formRef = ref(null)
const formValue = ref({
  ID:1,
  dingPush:{
    enable:false,
    dingRobot: ''
  },
  localPush:{
    enable:true,
  },
  updateBasicInfoOnStart:false,
  refreshInterval:1
})

onMounted(()=>{
  GetConfig().then(res=>{
    formValue.value.ID = res.ID
    formValue.value.dingPush = {
      enable:res.dingPushEnable,
      dingRobot:res.dingRobot
    }
    formValue.value.localPush = {
      enable:res.localPushEnable,
    }
    formValue.value.updateBasicInfoOnStart = res.updateBasicInfoOnStart
    formValue.value.refreshInterval = res.refreshInterval
    console.log(res)
  })
  //message.info("加载完成")
})


function saveConfig(){
  let config= new data.Settings({
    ID:formValue.value.ID,
    dingPushEnable:formValue.value.dingPush.enable,
    dingRobot:formValue.value.dingPush.dingRobot,
    localPushEnable:formValue.value.localPush.enable,
    updateBasicInfoOnStart:formValue.value.updateBasicInfoOnStart,
    refreshInterval:formValue.value.refreshInterval
  })

 //console.log("Settings",config)
  UpdateConfig(config).then(res=>{
    message.success(res)
  })
}


function getHeight() {
  return document.documentElement.clientHeight
}
function sendTestNotice(){
  let markdown="### go-stock test\n"+new Date()
  let msg='{' +
      '     "msgtype": "markdown",' +
      '     "markdown": {' +
      '         "title":"go-stock'+new Date()+'",' +
      '         "text": "'+markdown+'"' +
      '     },' +
      '      "at": {' +
      '          "isAtAll": true' +
      '      }' +
      ' }'

  SendDingDingMessageByType(msg, "test-"+new Date().getTime(),1).then(res=>{
    message.info(res)
  })
}
</script>

<template>
  <n-flex justify="left" style="margin-top: 12px;padding-left: 12px">
  <n-form ref="formRef"  :model="formValue" :label-placement="'left'" :label-align="'left'" style="width: 100%;height: 100%">
      <n-grid :cols="24" :x-gap="24" style="text-align: left">
        <n-gi :span="24">
          <n-text type="default" style="font-size: 25px;font-weight: bold">基础设置</n-text>
        </n-gi>
        <n-form-item-gi  :span="6" label="启动时更新A股/指数信息：" path="updateBasicInfoOnStart" >
          <n-switch v-model:value="formValue.updateBasicInfoOnStart" />
        </n-form-item-gi>
        <n-form-item-gi  :span="6" label="数据刷新间隔(重启生效)：" path="refreshInterval" >
          <n-input-number v-model:value="formValue.refreshInterval" placeholder="请输入数据刷新间隔(秒)">
            <template #suffix>
              秒
            </template>
          </n-input-number>
        </n-form-item-gi>
      </n-grid>

        <n-grid :cols="24" :x-gap="24" style="text-align: left">
          <n-gi :span="24">
            <n-text type="default" style="font-size: 25px;font-weight: bold">通知设置</n-text>
          </n-gi>
          <n-form-item-gi  :span="6" label="是否启用钉钉推送：" path="dingPush.enable" >
            <n-switch v-model:value="formValue.dingPush.enable" />
          </n-form-item-gi>
          <n-form-item-gi  :span="6" label="是否启用本地推送：" path="localPush.enable" >
            <n-switch v-model:value="formValue.localPush.enable" />
          </n-form-item-gi>
          <n-form-item-gi :span="24"  v-if="formValue.dingPush.enable" label="钉钉机器人接口地址：" path="dingPush.dingRobot" >
            <n-input  placeholder="请输入钉钉机器人接口地址"  v-model:value="formValue.dingPush.dingRobot"/>
            <n-button type="primary" @click="sendTestNotice">发送测试通知</n-button>
          </n-form-item-gi>
        </n-grid>
    <n-gi :span="24">
      <div style="display: flex; justify-content: center">
        <n-button  type="primary" @click="saveConfig">
          保存
        </n-button>
      </div>
    </n-gi>
  </n-form>
  </n-flex>
</template>

<style scoped>

</style>