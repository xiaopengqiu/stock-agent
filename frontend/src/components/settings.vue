<script setup>
import {h, onBeforeUnmount, onMounted, ref} from "vue";
import {
  AddPrompt, DelPrompt,
  ExportConfig,
  GetConfig,
  GetPromptTemplates,
  SendDingDingMessageByType,
  UpdateConfig, CheckSponsorCode, ExportPrompts,
  GetToolConfig,
  SetToolConfig,
  ReloadTools,
} from "../../wailsjs/go/main/App";
import {NTag, NButton, NAlert, NCard, NSwitch, NTable, NModal, NForm, NFormItem, NInput, NSelect, NSpace, NPopconfirm, NDivider, useDialog, useMessage} from "naive-ui";
import {data, models} from "../../wailsjs/go/models";
import {EventsEmit} from "../../wailsjs/runtime";

const message = useMessage()

const formRef = ref(null)
const formValue = ref({
  ID: 1,
  tushareToken: '',
  dingPush: {
    enable: false,
    dingRobot: ''
  },
  localPush: {
    enable: true,
  },
  updateBasicInfoOnStart: false,
  refreshInterval: 1,
  openAI: {
    enable: false,
    aiConfigs: [], // AI配置列表
    prompt: "",
    questionTemplate: "{{stockName}}分析和总结",
    crawlTimeOut: 30,
    kDays: 30,
  },
  enableDanmu: false,
  browserPath: '',
  enableNews: false,
  darkTheme: true,
  enableFund: false,
  enablePushNews: false,
  enableOnlyPushRedNews: false,
  sponsorCode: "",
  httpProxy:"",
  httpProxyEnabled:false,
  enableThinking: false,
})

// 添加一个新的AI配置到列表
function addAiConfig() {
  formValue.value.openAI.aiConfigs.push(new data.AIConfig({
    name: '',
    baseUrl: 'https://api.deepseek.com',
    apiKey: '',
    modelName: 'deepseek-chat',
    temperature: 0.1,
    maxTokens: 32768,
    timeOut: 600,
  }));
}

// 从列表中移除一个AI配置
function removeAiConfig(index) {
  const originalCount = formValue.value.openAI.aiConfigs.length;
  // 使用filter创建新数组确保响应式更新
  formValue.value.openAI.aiConfigs = formValue.value.openAI.aiConfigs.filter((_, i) => i !== index);
}


const promptTemplates = ref([])
onMounted(() => {
  GetConfig().then(res => {
    formValue.value.ID = res.ID
    formValue.value.tushareToken = res.tushareToken
    formValue.value.dingPush = {
      enable: res.dingPushEnable,
      dingRobot: res.dingRobot
    }
    formValue.value.localPush = {
      enable: res.localPushEnable,
    }
    formValue.value.updateBasicInfoOnStart = res.updateBasicInfoOnStart
    formValue.value.refreshInterval = res.refreshInterval
    // 加载AI配置
    formValue.value.openAI = {
      enable: res.openAiEnable,
      aiConfigs: res.aiConfigs || [],
      prompt: res.prompt,
      questionTemplate: res.questionTemplate ? res.questionTemplate : '{{stockName}}分析和总结',
      crawlTimeOut: res.crawlTimeOut,
      kDays: res.kDays,
    }


    formValue.value.enableDanmu = res.enableDanmu
    formValue.value.browserPath = res.browserPath
    formValue.value.enableNews = res.enableNews
    formValue.value.darkTheme = res.darkTheme
    formValue.value.enableFund = res.enableFund
    formValue.value.enablePushNews = res.enablePushNews
    formValue.value.enableOnlyPushRedNews = res.enableOnlyPushRedNews
    formValue.value.sponsorCode = res.sponsorCode
    formValue.value.httpProxy=res.httpProxy;
    formValue.value.httpProxyEnabled=res.httpProxyEnabled;
    formValue.value.enableThinking=res.enableThinking || false;

  })

  GetPromptTemplates("", "").then(res => {
    promptTemplates.value = res
  })

  // 加载工具配置
  loadToolConfig()
})
onBeforeUnmount(() => {
  message.destroyAll()
})

function saveConfig() {
  console.log('开始保存设置', formValue.value);
  // 构建配置时，包含aiConfigs列表
  let config = new data.SettingConfig({
    ID: formValue.value.ID,
    dingPushEnable: formValue.value.dingPush.enable,
    dingRobot: formValue.value.dingPush.dingRobot,
    localPushEnable: formValue.value.localPush.enable,
    updateBasicInfoOnStart: formValue.value.updateBasicInfoOnStart,
    refreshInterval: formValue.value.refreshInterval,
    openAiEnable: formValue.value.openAI.enable,
    aiConfigs: formValue.value.openAI.aiConfigs,
    // 序列化aiConfigs列表以传递给后端
    tushareToken: formValue.value.tushareToken,
    prompt: formValue.value.openAI.prompt,
    questionTemplate: formValue.value.openAI.questionTemplate,
    crawlTimeOut: formValue.value.openAI.crawlTimeOut,
    kDays: formValue.value.openAI.kDays,
    enableDanmu: formValue.value.enableDanmu,
    browserPath: formValue.value.browserPath,
    enableNews: formValue.value.enableNews,
    darkTheme: formValue.value.darkTheme,
    enableFund: formValue.value.enableFund,
    enablePushNews: formValue.value.enablePushNews,
    enableOnlyPushRedNews: formValue.value.enableOnlyPushRedNews,
    sponsorCode: formValue.value.sponsorCode,
    httpProxy:formValue.value.httpProxy,
    httpProxyEnabled:formValue.value.httpProxyEnabled,
    enableThinking: formValue.value.enableThinking,
  })

  if (config.sponsorCode) {
    CheckSponsorCode(config.sponsorCode).then(res => {
      if (res.code) {
        UpdateConfig(config).then(res => {
          message.success(res)
          EventsEmit("updateSettings", config);
        })
      } else {
        message.error(res.msg)
      }
    })
  } else {
    UpdateConfig(config).then(res => {
      message.success(res)
      EventsEmit("updateSettings", config);
    })
  }
}


function getHeight() {
  return document.documentElement.clientHeight
}

function sendTestNotice() {
  let markdown = "### go-stock test\n" + new Date()
  let msg = '{' +
      '     "msgtype": "markdown",' +
      '     "markdown": {' +
      '         "title":"go-stock' + new Date() + '",' +
      '         "text": "' + markdown + '"' +
      '     },' +
      '      "at": {' +
      '          "isAtAll": true' +
      '      }' +
      ' }'

  SendDingDingMessageByType(msg, "test-" + new Date().getTime(), 1).then(res => {
    message.info(res)
  })
}

function exportConfig() {
  ExportConfig().then(res => {
    message.info(res)
  })
}

function importConfig() {
  let input = document.createElement('input');
  input.type = 'file';
  input.accept = '.json';
  input.onchange = (e) => {
    let file = e.target.files[0];
    let reader = new FileReader();
    reader.onload = (e) => {
      let config = JSON.parse(e.target.result);
      formValue.value.ID = config.ID
      formValue.value.tushareToken = config.tushareToken
      formValue.value.dingPush = {
        enable: config.dingPushEnable,
        dingRobot: config.dingRobot
      }
      formValue.value.localPush = {
        enable: config.localPushEnable,
      }
      formValue.value.updateBasicInfoOnStart = config.updateBasicInfoOnStart
      formValue.value.refreshInterval = config.refreshInterval
      // 导入AI配置
      formValue.value.openAI = {
        enable: config.openAiEnable,
        aiConfigs: config.aiConfigs || [],
        prompt: config.prompt,
        questionTemplate: config.questionTemplate,
        crawlTimeOut: config.crawlTimeOut,
        kDays: config.kDays
      }
      formValue.value.enableDanmu = config.enableDanmu
      formValue.value.browserPath = config.browserPath
      formValue.value.enableNews = config.enableNews
      formValue.value.darkTheme = config.darkTheme
      formValue.value.enableFund = config.enableFund
      formValue.value.enablePushNews = config.enablePushNews
      formValue.value.enableOnlyPushRedNews = config.enableOnlyPushRedNews
      formValue.value.sponsorCode = config.sponsorCode
      formValue.value.httpProxy=config.httpProxy
      formValue.value.httpProxyEnabled=config.httpProxyEnabled
      formValue.value.enableThinking=config.enableThinking || false
    };
    reader.readAsText(file);
  };
  input.click();
}


window.onerror = function (event, source, lineno, colno, error) {
  EventsEmit("frontendError", {
    page: "settings.vue",
    message: event,
    source: source,
    lineno: lineno,
    colno: colno,
    error: error ? error.stack : null
  });
  return true;
};

const showManagePromptsModal = ref(false)
const promptTypeOptions = [
  {label: "模型系统Prompt", value: '模型系统Prompt'},
  {label: "模型用户Prompt", value: '模型用户Prompt'},]
const formPromptRef = ref(null)
const formPrompt = ref({
  ID: 0,
  Name: '',
  Content: '',
  Type: '',
})

function managePrompts() {
  formPrompt.value.ID = 0
  showManagePromptsModal.value = true
}

function savePrompt() {
  AddPrompt(formPrompt.value).then(res => {
    message.success(res)
    GetPromptTemplates("", "").then(res => {
      promptTemplates.value = res
    })
    showManagePromptsModal.value = false
  })
}

function exportPrompts() {
  ExportPrompts().then(res => {
    message.info(res)
  })
}

function importPrompt() {
  let input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = (e) => {
    let file = e.target.files[0]
    let reader  = new FileReader()
    reader.onload = (e) => {
      let prompts = JSON.parse(e.target.result)
      prompts.forEach((prompt, index) => {
        let data = {
          ID: prompt.ID,
          Name: prompt.Name,
          Content: prompt.Content,
          Type: prompt.Type
        }
        AddPrompt(data)
      })
    }
    reader.readAsText(file)
  }
  input.click()
}

function editPrompt(prompt) {
  formPrompt.value.ID = prompt.ID
  formPrompt.value.Name = prompt.name
  formPrompt.value.Content = prompt.content
  formPrompt.value.Type = prompt.type
  showManagePromptsModal.value = true
}

const dialog = useDialog()
function deletePrompt(Name, ID) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除 Prompt 模板「${Name}」吗？此操作无法撤销。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: () => {
      DelPrompt(ID).then(res => {
        message.success(res)
        GetPromptTemplates("", "").then(res => {
          promptTemplates.value = res
        })
      })
    }
  })
}

const userPromptExampleText = '例如：{{stockName}}[{{stockCode}}] 分析和总结'

// ========== 工具配置相关 ==========
// 工具配置数据
const toolConfig = ref({
  tools: [],
  version: "1.0",
});

// 加载状态
const toolLoading = ref(false);

// 添加/编辑工具弹窗
const showToolModal = ref(false);
const isEditing = ref(false);
const editingToolIndex = ref(-1);

// 工具表单
const toolForm = ref({
  name: "",
  type: "http",
  description: "",
  enabled: true,
  config: {
    url: "",
    method: "POST",
    headers: {},
  },
});

// 工具类型选项
const toolTypeOptions = [
  { label: "HTTP", value: "http" },
  { label: "内置", value: "builtin", disabled: true },
  { label: "MCP", value: "mcp", disabled: true },
];

// HTTP 方法选项
const methodOptions = [
  { label: "POST", value: "POST" },
  { label: "GET", value: "GET" },
];

// 工具表格列定义
const toolColumns = [
  {
    title: "工具名称",
    key: "name",
    width: 180,
    render(row) {
      return h(
        "div",
        {
          style: { fontWeight: 500 },
        },
        row.name
      );
    },
  },
  {
    title: "类型",
    key: "type",
    width: 100,
    render(row) {
      const typeMap = {
        builtin: "内置",
        mcp: "MCP",
        http: "HTTP",
      };
      const typeColors = {
        builtin: "success",
        mcp: "warning",
        http: "info",
      };
      return h(NTag, {
        type: typeColors[row.type] || "default",
        size: "small",
        bordered: false,
      }, () => typeMap[row.type] || row.type);
    },
  },
  {
    title: "描述",
    key: "description",
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: "启用状态",
    key: "enabled",
    width: 100,
    align: "center",
    render(row, index) {
      return h(NSwitch, {
        value: row.enabled,
        size: "small",
        onUpdateValue: (value) => handleToggleEnabled(index, value),
      });
    },
  },
  {
    title: "操作",
    key: "actions",
    width: 150,
    align: "center",
    render(row, index) {
      // 内置工具不允许编辑/删除
      if (row.type === "builtin") {
        return h("span", { style: { color: "#999", fontSize: "12px" } }, "系统内置");
      }

      return h(NSpace, { size: "small", justify: "center" }, {
        default: () => [
          h(NButton, {
            size: "tiny",
            type: "primary",
            ghost: true,
            onClick: () => handleEditTool(index),
          }, () => "编辑"),
          h(NPopconfirm, {
            onPositiveClick: () => handleDeleteTool(index),
          }, {
            trigger: () => h(NButton, {
              size: "tiny",
              type: "error",
              ghost: true,
            }, () => "删除"),
            default: () => "确定要删除此工具吗？",
          }),
        ],
      });
    },
  },
];

// 加载工具配置
async function loadToolConfig() {
  toolLoading.value = true;
  try {
    const configJson = await GetToolConfig();
    const config = JSON.parse(configJson);
    toolConfig.value = config;
  } catch (error) {
    message.error("加载工具配置失败: " + error);
  } finally {
    toolLoading.value = false;
  }
}

// 保存工具配置
async function saveToolConfig() {
  try {
    const configJson = JSON.stringify(toolConfig.value, null, 2);
    const result = await SetToolConfig(configJson);
    message.success(result);
  } catch (error) {
    message.error("保存工具配置失败: " + error);
  }
}

// 启用/禁用工具
async function handleToggleEnabled(index, value) {
  toolConfig.value.tools[index].enabled = value;
  await saveToolConfig();
}

// 打开添加工具弹窗
function handleAddTool() {
  isEditing.value = false;
  editingToolIndex.value = -1;
  toolForm.value = {
    name: "",
    type: "http",
    description: "",
    enabled: true,
    config: {
      url: "",
      method: "POST",
      headers: {},
    },
  };
  showToolModal.value = true;
}

// 编辑工具
function handleEditTool(index) {
  const tool = toolConfig.value.tools[index];
  if (tool.type === "builtin") {
    message.warning("内置工具不能编辑");
    return;
  }
  isEditing.value = true;
  editingToolIndex.value = index;
  toolForm.value = JSON.parse(JSON.stringify(tool));
  showToolModal.value = true;
}

// 保存工具
async function handleSaveTool() {
  // 表单验证
  if (!toolForm.value.name.trim()) {
    message.error("工具名称不能为空");
    return;
  }
  if (!toolForm.value.description.trim()) {
    message.error("工具描述不能为空");
    return;
  }
  if (toolForm.value.type === "http") {
    if (!toolForm.value.config?.url.trim()) {
      message.error("HTTP URL 不能为空");
      return;
    }
  }

  // 检查名称是否重复（新增时）
  if (!isEditing.value) {
    const existingTool = toolConfig.value.tools.find(
      (t) => t.name === toolForm.value.name
    );
    if (existingTool) {
      message.error("工具名称已存在");
      return;
    }
  }

  // 保存
  if (isEditing.value && editingToolIndex.value >= 0) {
    toolConfig.value.tools[editingToolIndex.value] = JSON.parse(
      JSON.stringify(toolForm.value)
    );
  } else {
    toolConfig.value.tools.push(JSON.parse(JSON.stringify(toolForm.value)));
  }

  await saveToolConfig();
  showToolModal.value = false;
  message.success(isEditing.value ? "工具更新成功" : "工具添加成功");
}

// 删除工具
async function handleDeleteTool(index) {
  const tool = toolConfig.value.tools[index];
  if (tool.type === "builtin") {
    message.warning("内置工具不能删除");
    return;
  }
  toolConfig.value.tools.splice(index, 1);
  await saveToolConfig();
  message.success("工具删除成功");
}

// 重置配置
function handleResetConfig() {
  dialog.warning({
    title: "确认重置",
    content: "确定要重置工具配置为默认状态吗？这将删除所有自定义工具。",
    positiveText: "确认重置",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        // 保留内置工具，删除其他工具
        toolConfig.value.tools = toolConfig.value.tools.filter(
          (t) => t.type === "builtin"
        );
        await saveToolConfig();
        message.success("工具配置已重置");
      } catch (error) {
        message.error("重置失败: " + error);
      }
    },
  });
}

// 重新加载工具
async function handleReloadTools() {
  try {
    message.info("正在重新加载工具...");
    const result = await ReloadTools();
    message.success(result);
    await loadToolConfig();
  } catch (error) {
    message.error("重新加载失败: " + error);
  }
}
</script>

<template>
  <n-flex justify="left" style="text-align: left; --wails-draggable:no-drag">
    <n-form ref="formRef" :label-placement="'left'" :label-align="'left'">
      <n-space vertical size="large">
        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => '基础设置')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left">
            <n-form-item-gi :span="10" label="Tushare Token：" path="tushareToken">
              <n-input type="text" placeholder="Tushare api token" v-model:value="formValue.tushareToken" clearable/>
            </n-form-item-gi>
            <n-form-item-gi :span="4" label="启动时更新基础信息：" path="updateBasicInfoOnStart">
              <n-switch v-model:value="formValue.updateBasicInfoOnStart"/>
            </n-form-item-gi>
            <n-form-item-gi :span="4" label="数据刷新间隔：" path="refreshInterval">
              <n-input-number v-model:value="formValue.refreshInterval" placeholder="请输入数据刷新间隔(秒)">
                <template #suffix>秒</template>
              </n-input-number>
            </n-form-item-gi>
            <n-form-item-gi :span="6" label="暗黑主题：" path="darkTheme">
              <n-switch v-model:value="formValue.darkTheme"/>
            </n-form-item-gi>
            <n-form-item-gi :span="10" label="浏览器安装路径：" path="browserPath">
              <n-input type="text" placeholder="浏览器安装路径" v-model:value="formValue.browserPath" clearable/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="指数基金：" path="enableFund">
              <n-switch v-model:value="formValue.enableFund"/>
            </n-form-item-gi>
            <n-form-item-gi :span="11" label="赞助码：" path="sponsorCode">
              <n-input-group>
                <n-input :show-count="true" placeholder="赞助码" v-model:value="formValue.sponsorCode"/>
                <n-button type="success" secondary strong
                          @click="CheckSponsorCode(formValue.sponsorCode).then((res) => {message.warning(res.msg)})">验证
                </n-button>
              </n-input-group>
            </n-form-item-gi>
          </n-grid>
        </n-card>

        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => '通知设置')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left">
            <n-form-item-gi :span="3" label="钉钉推送：" path="dingPush.enable">
              <n-switch v-model:value="formValue.dingPush.enable"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="本地推送：" path="localPush.enable">
              <n-switch v-model:value="formValue.localPush.enable"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="弹幕功能：" path="enableDanmu">
              <n-switch v-model:value="formValue.enableDanmu"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="显示滚动快讯：" path="enableNews">
              <n-switch v-model:value="formValue.enableNews"/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="市场资讯提醒：" path="enablePushNews">
              <n-switch v-model:value="formValue.enablePushNews"/>
            </n-form-item-gi>
            <n-form-item-gi v-if="formValue.enablePushNews" :span="4" label="只提醒红字或关注个股的新闻：" path="enableOnlyPushRedNews">
              <n-switch v-model:value="formValue.enableOnlyPushRedNews"/>
            </n-form-item-gi>

            <n-form-item-gi :span="22" v-if="formValue.dingPush.enable" label="钉钉机器人接口地址："
                            path="dingPush.dingRobot">
              <n-input placeholder="请输入钉钉机器人接口地址" v-model:value="formValue.dingPush.dingRobot"/>
              <n-button type="primary" @click="sendTestNotice">发送测试通知</n-button>
            </n-form-item-gi>
          </n-grid>
        </n-card>

        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => 'AI设置')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left;">
            <n-form-item-gi :span="24" label="AI诊股：">
              <n-switch v-model:value="formValue.openAI.enable"/>
            </n-form-item-gi>

            <n-form-item-gi :span="6" v-if="formValue.openAI.enable" label="Crawler Timeout(秒)"
                            title="资讯采集超时时间(秒)" path="openAI.crawlTimeOut">
              <n-input-number min="30" step="1" v-model:value="formValue.openAI.crawlTimeOut"/>
            </n-form-item-gi>
            <n-form-item-gi :span="4" v-if="formValue.openAI.enable" title="天数越多消耗tokens越多"
                            label="日K线数据(天)" path="openAI.kDays">
              <n-input-number min="30" step="1" max="365" v-model:value="formValue.openAI.kDays"/>
            </n-form-item-gi>
            <n-form-item-gi :span="2" label="http代理" path="httpProxyEnabled">
              <n-switch v-model:value="formValue.httpProxyEnabled"/>
            </n-form-item-gi>
            <n-form-item-gi :span="10" v-if="formValue.httpProxyEnabled" title="http代理地址"
                            label="http代理地址" path="httpProxy">
              <n-input type="text" placeholder="http代理地址" v-model:value="formValue.httpProxy" clearable/>
            </n-form-item-gi>
            <n-form-item-gi :span="3" label="深度思考" path="enableThinking">
              <n-switch v-model:value="formValue.enableThinking"/>
            </n-form-item-gi>

            <n-gi :span="24" v-if="formValue.openAI.enable">
              <n-divider title-placement="left">AI模型服务配置</n-divider>
            </n-gi>
            <n-gi :span="24" v-if="formValue.openAI.enable">
              <n-space vertical>
                <n-card v-for="(aiConfig, index) in formValue.openAI.aiConfigs" :key="index" :bordered="true"
                        size="small">
                  <template #header>
                    <n-flex justify="space-between" align="center">
                      <n-text depth="3">AI 配置 #{{ index + 1 }}</n-text>
                      <n-button type="error" size="tiny" ghost @click="removeAiConfig(index)">删除</n-button>
                    </n-flex>
                  </template>
                  <n-grid :cols="24" :x-gap="24">
                    <n-form-item-gi :span="24" hidden label="配置ID" :path="`openAI.aiConfigs[${index}].ID`">
                      <n-input type="text" placeholder="配置ID" v-model:value="aiConfig.ID" clearable/>
                    </n-form-item-gi>
                    <n-form-item-gi :span="12" label="配置名称" :path="`openAI.aiConfigs[${index}].name`">
                      <n-input type="text" placeholder="配置名称" v-model:value="aiConfig.name" clearable/>
                    </n-form-item-gi>
                    <n-form-item-gi :span="12" label="接口地址" :path="`openAI.aiConfigs[${index}].baseUrl`">
                      <n-input type="text" placeholder="AI接口地址" v-model:value="aiConfig.baseUrl" clearable/>
                    </n-form-item-gi>
                    <n-form-item-gi :span="12" label="令牌(apiKey)" :path="`openAI.aiConfigs[${index}].apiKey`">
                      <n-input type="password" placeholder="apiKey" v-model:value="aiConfig.apiKey" clearable
                               show-password-on="click"/>
                    </n-form-item-gi>
                    <n-form-item-gi :span="8" label="模型名称" :path="`openAI.aiConfigs[${index}].modelName`">
                      <n-input type="text" placeholder="AI模型名称" v-model:value="aiConfig.modelName" clearable/>
                    </n-form-item-gi>
                    <n-form-item-gi :span="5" label="Temperature" :path="`openAI.aiConfigs[${index}].temperature`">
                      <n-input-number placeholder="temperature" v-model:value="aiConfig.temperature" :step="0.1"/>
                    </n-form-item-gi>
                    <n-form-item-gi :span="5" label="MaxTokens" :path="`openAI.aiConfigs[${index}].maxTokens`">
                      <n-input-number placeholder="maxTokens" v-model:value="aiConfig.maxTokens"/>
                    </n-form-item-gi>
                    <n-form-item-gi :span="5" label="Timeout(秒)" :path="`openAI.aiConfigs[${index}].timeOut`">
                      <n-input-number min="60" step="1" placeholder="超时(秒)" v-model:value="aiConfig.timeOut"/>
                    </n-form-item-gi>
                  </n-grid>
                </n-card>
                <n-button type="primary" dashed @click=" addAiConfig" style="width: 100%;">+ 添加AI配置</n-button>
              </n-space>
            </n-gi>

            <n-gi :span="24">
              <n-divider/>
            </n-gi>

            <n-gi :span="24">
              <n-space vertical>
                <n-space justify="center">
                  <n-button type="primary" strong @click="saveConfig">保存设置</n-button>
                  <n-button type="info" @click="exportConfig">导出配置</n-button>
                  <n-button type="error" @click="importConfig">导入配置</n-button>
                </n-space>
              </n-space>
            </n-gi>

          </n-grid>
        </n-card>

        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => 'Prompt模板设置')" size="small">
          <n-grid :cols="24" :x-gap="24" style="text-align: left;">
            <n-gi :span="24" v-if="formValue.openAI.enable">
              <n-divider title-placement="left">Prompt 内容设置</n-divider>
            </n-gi>
            <n-form-item-gi :span="12" v-if="formValue.openAI.enable" label="系统 Prompt 示例" label-placement="left">
              <n-alert type="info" :show-icon="true" bordered class="prompt-alert equal-height">
                <div class="prompt-content">
                  请输入系统 Prompt，用于定义模型的基础行为和身份。
                </div>
              </n-alert>
            </n-form-item-gi>
            <n-form-item-gi :span="12" v-if="formValue.openAI.enable" label="用户 Prompt 示例" label-placement="left">
              <n-alert type="info" :show-icon="true" bordered class="prompt-alert equal-height">
                <div class="prompt-content">
                  {{ userPromptExampleText }}
                </div>
              </n-alert>
            </n-form-item-gi>

            <n-gi :span="24" v-if="promptTemplates.length > 0">
              <n-form-item-gi :span="24" label="模型系统 Prompt 模板">
                <n-tag size="medium" secondary v-if="formValue.openAI.enable" v-for="prompt in promptTemplates.filter(p => p.type === '模型系统Prompt')" closable
                       @close="deletePrompt(prompt.name, prompt.ID)" @click="editPrompt(prompt)" :title="prompt.content" :type="'success'" :bordered="false">
                  {{ prompt.name }}
                </n-tag>
              </n-form-item-gi>
            </n-gi>

            <n-gi :span="24" v-if="promptTemplates.length > 0">
              <n-form-item-gi :span="24" label="模型用户 Prompt 模板">
                <n-tag size="medium" secondary v-if="formValue.openAI.enable" v-for="prompt in promptTemplates.filter(p => p.type === '模型用户Prompt')" closable
                       @close="deletePrompt(prompt.name, prompt.ID)" @click="editPrompt(prompt)" :title="prompt.content" :type="'info'" :bordered="false">
                  {{ prompt.name }}
                </n-tag>
              </n-form-item-gi>
            </n-gi>


            <n-gi :span="24">
              <n-space vertical>
                <n-space justify="center">
                  <n-button type="primary" dashed @click="managePrompts" style="width: 100%;">添加提示词模板</n-button>
                  <n-button type="info" dashed @click="exportPrompts" style="width: 100%;">导出提示词模板</n-button>
                  <n-button type="info" dashed @click="importPrompt" style="width: 100%;">导入提示词模板</n-button>
                </n-space>

              </n-space>
            </n-gi>
          </n-grid>
        </n-card>

        <n-card :title="() => h(NTag, { type: 'primary', bordered: false }, () => '工具配置')" size="small">
          <template #header-extra>
            <NSpace size="small">
              <NButton size="small" type="info" @click="handleReloadTools">
                重新加载
              </NButton>
              <NButton size="small" type="warning" @click="handleResetConfig">
                重置配置
              </NButton>
              <NButton size="small" type="primary" @click="handleAddTool">
                添加工具
              </NButton>
            </NSpace>
          </template>

          <div class="tool-table-container">
            <n-data-table
              :columns="toolColumns"
              :data="toolConfig.tools"
              :loading="toolLoading"
              :pagination="false"
              size="small"
              striped
            />
          </div>

          <div v-if="toolConfig.tools.length === 0 && !toolLoading" class="empty-state">
            <p>暂无工具配置</p>
            <NButton type="primary" @click="handleAddTool">添加第一个工具</NButton>
          </div>
        </n-card>

      </n-space>
    </n-form>
  </n-flex>

  <n-modal v-model:show="showManagePromptsModal" closable :mask-closable="false">
    <n-card style="width: 800px; height: 600px; text-align: left" :bordered="false"
            :title="(formPrompt.ID > 0 ? '修改' : '添加') + '提示词'" size="huge" role="dialog" aria-modal="true">
      <n-form ref="formPromptRef" :label-placement="'left'" :label-align="'left'">
        <n-form-item label="名称">
          <n-input v-model:value="formPrompt.Name" placeholder="请输入提示词名称"/>
        </n-form-item>
        <n-form-item label="类型">
          <n-select v-model:value="formPrompt.Type" :options="promptTypeOptions" placeholder="请选择提示词类型"/>
        </n-form-item>
        <n-form-item label="内容">
          <n-input v-model:value="formPrompt.Content" type="textarea" :show-count="true" placeholder="请输入prompt"
                   :autosize="{ minRows: 12, maxRows: 12, }"/>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-flex justify="end">
          <n-button type="primary" @click="savePrompt">保存</n-button>
          <n-button type="warning" @click="showManagePromptsModal = false">取消</n-button>
        </n-flex>
      </template>
    </n-card>
  </n-modal>

  <!-- 工具配置弹窗 -->
  <NModal
    v-model:show="showToolModal"
    :title="isEditing ? '编辑工具' : '添加工具'"
    preset="card"
    style="width: 600px"
    :mask-closable="false"
  >
    <NForm
      :model="toolForm"
      label-placement="left"
      label-width="100"
      require-mark-placement="right-hanging"
    >
      <NFormItem label="工具名称" required>
        <NInput
          v-model:value="toolForm.name"
          placeholder="请输入工具名称"
          :disabled="isEditing"
        />
      </NFormItem>

      <NFormItem label="工具类型" required>
        <NSelect
          v-model:value="toolForm.type"
          :options="toolTypeOptions"
          placeholder="请选择工具类型"
        />
      </NFormItem>

      <NFormItem label="工具描述" required>
        <NInput
          v-model:value="toolForm.description"
          type="textarea"
          placeholder="请输入工具描述"
          :rows="3"
        />
      </NFormItem>

      <!-- HTTP 类型配置 -->
      <template v-if="toolForm.type === 'http'">
        <NDivider title-placement="left">HTTP 配置</NDivider>

        <NFormItem label="请求地址" required>
          <NInput
            v-model:value="toolForm.config.url"
            placeholder="请输入 HTTP URL，例如: http://localhost:8080/api/tool"
          />
        </NFormItem>

        <NFormItem label="请求方法">
          <NSelect
            v-model:value="toolForm.config.method"
            :options="methodOptions"
          />
        </NFormItem>
      </template>

      <NFormItem label="启用状态">
        <NSwitch v-model:value="toolForm.enabled" />
      </NFormItem>
    </NForm>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="showToolModal = false">取消</NButton>
        <NButton type="primary" @click="handleSaveTool">保存</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped>
.cardHeaderClass {
  font-size: 16px;
  font-weight: bold;
  color: red;
}
.prompt-alert {
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  background-color: #f9fafb;
  padding: 12px 16px;
  transition: all 0.2s ease-in-out;
}

.prompt-alert:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.prompt-content {
  font-size: 14px;
  color: #555;
  line-height: 1.6;
  white-space: pre-line;
}

.tool-table-container {
  margin-top: 16px;
}

.empty-state {
  text-align: center;
  padding: 48px;
  color: #999;
}

.empty-state p {
  margin-bottom: 16px;
}
</style>