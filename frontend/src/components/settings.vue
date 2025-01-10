<script setup>

import {ref} from "vue";

const formRef = ref(null)
const formValue = ref({
  dingPush:{
    enable:false,
    dingRobot: 'https://oapi.dingtalk.com/robot/send?access_token=0237527b404598f37ae5d83ef36e936860c7ba5d3892cd43f64c4159d3ed7cb1'
  },
  localPush:{
    enable:true,
  }

})
function getHeight() {
  return document.documentElement.clientHeight
}
</script>

<template>
  <n-flex justify="start">
    <n-card title="推送设置">
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
            </n-form-item-gi>
          </n-grid>
        </n-form>
    </n-card>
  </n-flex>

</template>

<style scoped>

</style>