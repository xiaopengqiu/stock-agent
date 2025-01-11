<script setup>

import {onMounted, ref, watch} from "vue";
import {GetConfig, SendDingDingMessageByType, UpdateConfig} from "../../wailsjs/go/main/App";
import {useMessage} from "naive-ui";
import {data} from "../../wailsjs/go/models";
const message = useMessage()

const formRef = ref(null)
const formValue = ref({
  ID:0,
  dingPush:{
    enable:false,
    dingRobot: ''
  },
  localPush:{
    enable:true,
  }
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
  })

  console.log("Settings",config)
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
    <n-card title="推送设置" style="height: 100%;">
        <n-form ref="formRef"  :model="formValue" :label-placement="'left'" :label-align="'left'">
          <n-grid :cols="24" :x-gap="24">
            <n-form-item-gi  :span="12" label="是否启用钉钉推送：" path="dingPush.enable" >
              <n-switch v-model:value="formValue.dingPush.enable" />
            </n-form-item-gi>
            <n-form-item-gi  :span="12" label="是否启用本地推送：" path="localPush.enable" >
              <n-switch v-model:value="formValue.localPush.enable" />
            </n-form-item-gi>
            <n-form-item-gi :span="24"  v-if="formValue.dingPush.enable" label="钉钉机器人接口地址：" path="dingPush.dingRobot" >
              <n-input  placeholder="请输入钉钉机器人接口地址"  v-model:value="formValue.dingPush.dingRobot"/>
              <n-button type="primary" @click="sendTestNotice">发送测试通知</n-button>
            </n-form-item-gi>
            <n-gi :span="24">
              <div style="display: flex; justify-content: flex-end">
                <n-button round type="primary" @click="saveConfig">
                  保存
                </n-button>
              </div>
            </n-gi>
          </n-grid>
        </n-form>
    </n-card>

</template>

<style scoped>

</style>