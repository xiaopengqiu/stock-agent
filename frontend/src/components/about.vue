<script setup>
import { MdPreview } from 'md-editor-v3';
// preview.css相比style.css少了编辑器那部分样式
import 'md-editor-v3/lib/preview.css';
import {onMounted, ref} from 'vue';
import {GetVersionInfo} from "../../wailsjs/go/main/App";
const updateLog = ref('');
const versionInfo = ref('');
const icon = ref('https://raw.githubusercontent.com/ArvinLovegood/go-stock/master/build/appicon.png');
const alipay =ref('https://github.com/ArvinLovegood/go-stock/raw/master/build/screenshot/alipay.jpg')
const wxpay =ref('https://github.com/ArvinLovegood/go-stock/raw/master/build/screenshot/wxpay.jpg')
onMounted(() => {
  document.title = '关于软件';
  GetVersionInfo().then((res) => {
    updateLog.value = res.content;
    versionInfo.value = res.version;
    icon.value = res.icon;
    alipay.value=res.alipay;
    wxpay.value=res.wxpay;
  });
})
</script>

<template>
  <n-config-provider>
    <n-layout>
      <n-space vertical size="large">
        <!-- 软件描述 -->
        <n-card size="large">
          <n-divider title-placement="center">关于软件</n-divider>
          <n-space vertical >
            <n-image width="100" :src="icon" />
            <h1>go-stock <n-tag  size="small" round>{{versionInfo}}</n-tag></h1>
            <div style="justify-self: center;text-align: left" >
              <p>自选股行情实时监控，基于Wails和NaiveUI构建的AI赋能股票分析工具</p>
              <p>
              </p>
              <p>
                欢迎点赞GitHub：<a href="https://github.com/ArvinLovegood/go-stock" target="_blank">go-stock</a><n-divider vertical />
                <a href="https://github.com/ArvinLovegood/go-stock" target="_blank">GitHub</a><n-divider vertical />
                <a href="https://github.com/ArvinLovegood/go-stock/issues" target="_blank">Issues</a><n-divider vertical />
                <a href="https://github.com/ArvinLovegood/go-stock/releases" target="_blank">Releases</a><n-divider vertical />
              </p>
              <p v-if="updateLog">更新说明：{{updateLog}}</p>
            </div>
          </n-space>
          <n-divider title-placement="center">关于作者</n-divider>
          <n-space vertical>
<!--            <h1>关于作者</h1>-->
            <n-avatar width="100" src="https://avatars.githubusercontent.com/u/7401917?v=4" />
            <h2><a href="https://github.com/ArvinLovegood" target="_blank">@ArvinLovegood</a></h2>
            <p>一个热爱编程的小白，欢迎关注我的Github</p>
            <p>
              邮箱：<a href="mailto:sparkmemory@163.com">sparkmemory@163.com</a><n-divider vertical />
              QQ：  506808970<n-divider vertical />
              微信：ArvinLovegood</p><n-divider vertical />
              <p>开源不易，如果觉得好用，可以请作者喝杯咖啡。</p>
            <n-flex justify="center">
              <n-image width="200" :src="alipay" />
              <n-image width="200" :src="wxpay" />
            </n-flex>
          </n-space>
          <n-divider title-placement="center">鸣谢</n-divider>
          <div style="justify-self: center;text-align: left" >
            <p>
              感谢以下捐赠者：
              <n-gradient-text size="small" type="warning">*晨</n-gradient-text><n-divider vertical />
            </p>
            <p>
              感谢以下开发者：
              <a href="https://github.com/gnim2600" target="_blank">@gnim2600</a><n-divider vertical />
              <a href="https://github.com/XXXiaohuayanGGG" target="_blank">@XXXiaohuayanGGG</a><n-divider vertical />
              <a href="https://github.com/2lovecode" target="_blank">@2lovecode</a><n-divider vertical />
              <a href="https://github.com/JerryLookupU" target="_blank">@JerryLookupU</a><n-divider vertical />
            </p>
            <p>
              感谢以下开源项目：
              <a href="https://github.com/wailsapp/wails" target="_blank">Wails</a><n-divider vertical />
              <a href="https://github.com/vuejs" target="_blank">Vue</a><n-divider vertical />
              <a href="https://github.com/tusen-ai/naive-ui" target="_blank">NaiveUI</a><n-divider vertical />
            </p>
          </div>
        </n-card>
      </n-space>
    </n-layout>
  </n-config-provider>
</template>

<style scoped>
/* 可以在这里添加一些样式 */
h1, h2 {
  margin: 0;
  padding: 6px 0;
}

p {
  margin: 2px 0;
}

ul {
  list-style-type: disc;
  padding-left: 20px;
}

a {
  color: #18a058;
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
}
</style>
