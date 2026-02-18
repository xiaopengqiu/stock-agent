<template>
  <n-spin :show="analyzing">
    <div style="height: calc(100vh - 200px)">
      <n-row :gutter="16">
        <!-- 左侧对话区域 -->
        <n-col :span="12" style="height: 100%">
          <n-card title="AI对话" style="height: 100%">
            <template #header-extra>
              <n-button size="small" @click="showHistory">
                <template #icon>
                  <n-icon><component :is="TimeOutline"/></n-icon>
                </template>
                历史记录
              </n-button>
            </template>
            <div style="height: 100%; display: flex; flex-direction: column;">
              <!-- 工具列表展示 -->
              <n-card v-if="toolsList.length > 0" size="small" style="margin-bottom: 10px;">
                <template #header>
                  <n-flex align="center">
                    <n-icon :size="16" style="margin-right: 5px">
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                        <path fill="currentColor" d="M19.14,12.94c0.04-0.3,0.06-0.61,0.06-0.94c0-0.32-0.02-0.64-0.07-0.94l0.01,0.05C19.14,12.94,19.14,12.94,19.14,12.94z M17.83,18.82c-0.13,0.18-0.28,0.35-0.45,0.5l-0.03-0.03C17.56,19.12,17.69,18.97,17.83,18.82z M11.81,14.56c-0.07-0.34-0.1-0.69-0.1-1.04c0-0.36,0.04-0.71,0.11-1.06C11.85,13.22,11.82,13.89,11.81,14.56z M14.46,15.38c-0.15-0.31-0.28-0.64-0.39-0.99l-0.05,0.02C14.07,14.73,14.25,15.05,14.46,15.38z M7.16,13.21c0.02-0.26,0.05-0.52,0.09-0.77C7.17,12.82,7.16,13.02,7.16,13.21z M5.52,15.12c0.09-0.26,0.2-0.51,0.31-0.76L5.74,14.41C5.63,14.64,5.56,14.89,5.52,15.12z M10.77,17.87c-0.05-0.28-0.09-0.57-0.12-0.86l-0.02,0.03C10.66,17.31,10.71,17.59,10.77,17.87z M13.41,18.4c-0.11-0.23-0.23-0.45-0.33-0.69l-0.04,0.01C13.16,17.95,13.28,18.16,13.41,18.4z M16.54,16.36l-0.03-0.04c-0.14,0.16-0.28,0.33-0.45,0.48c0.18-0.17,0.33-0.35,0.48-0.51V16.36z M16.16,17.52l-0.01-0.01c-0.15,0.16-0.31,0.31-0.49,0.45c0.19-0.15,0.36-0.31,0.52-0.47L16.16,17.52z M15.01,18.64c-0.14,0.14-0.29,0.28-0.46,0.4l0.02-0.02C14.75,18.89,14.88,18.77,15.01,18.64z M18.4,16.3l-0.01-0.01c-0.14,0.16-0.28,0.32-0.43,0.46c0.16-0.15,0.31-0.31,0.45-0.47L18.4,16.3z"/>
                        <path fill="currentColor" d="M12,2C6.48,2,2,6.48,2,12s4.48,10,10,10s10-4.48,10-10S17.52,2,12,2z M12,20c-4.41,0-8-3.59-8-8s3.59-8,8-8s8,3.59,8,8S16.41,20,12,20z M12,6c-3.31,0-6,2.69-6,6s2.69,6,6,6s6-2.69,6-6S15.31,6,12,6z"/>
                      </svg>
                    </n-icon>
                    <span>可用工具</span>
                  </n-flex>
                </template>
                <n-space>
                  <n-tag
                    v-for="tool in toolsList"
                    :key="tool"
                    :type="getToolStatus(tool)"
                    size="small"
                    :bordered="false"
                  >
                    {{ tool }}
                  </n-tag>
                </n-space>
              </n-card>

              <!-- 消息列表 -->
              <n-scrollbar style="flex: 1; margin-bottom: 10px;">
                <div v-for="(msg, index) in messages" :key="index" style="margin-bottom: 15px;">
                  <!-- 用户消息 -->
                  <div v-if="msg.role === 'user'" style="text-align: right;">
                    <n-card size="small" style="display: inline-block; max-width: 80%; background: #e6f7ff; border: none;">
                      {{ msg.content }}
                    </n-card>
                  </div>
                  <!-- AI消息 -->
                  <div v-else style="text-align: left;">
                    <n-card size="small" style="display: inline-block; max-width: 90%; background: #f5f5f5; border: none;">
                      <MdPreview :modelValue="msg.content" :theme="darkTheme ? 'dark' : 'light'"/>
                    </n-card>
                  </div>
                </div>
              </n-scrollbar>

              <!-- 输入框 -->
              <n-input
                v-model:value="inputText"
                type="textarea"
                placeholder="请输入您的选股需求，例如：推荐今日资金流入大的科技股"
                :autosize="{ minRows: 2, maxRows: 4 }"
                @keydown.enter.prevent="handleEnter"
                :disabled="analyzing"
              />
              <n-button
                type="primary"
                block
                :loading="analyzing"
                @click="sendMessage"
                style="margin-top: 10px;"
              >
                {{ analyzing ? '分析中...' : '开始分析' }}
              </n-button>
            </div>
          </n-card>
        </n-col>

        <!-- 右侧推荐结果区域 -->
        <n-col :span="12" style="height: 100%; display: flex; flex-direction: column;">
          <n-card title="推荐结果" style="height: 100%; display: flex; flex-direction: column;">
            <template #header-extra>
              <n-space>
                <n-button-group size="small">
                  <n-button
                    :type="viewMode === 'full' ? 'primary' : 'default'"
                    @click="viewMode = 'full'"
                  >
                    完整报告
                  </n-button>
                  <n-button
                    :type="viewMode === 'simple' ? 'primary' : 'default'"
                    @click="viewMode = 'simple'"
                  >
                    简洁列表
                  </n-button>
                </n-button-group>
                <n-dropdown :options="exportOptions" @select="exportReport">
                  <n-button size="small" type="tertiary">
                    <template #icon>
                      <n-icon><component :is="DownloadOutline"/></n-icon>
                    </template>
                  </n-button>
                </n-dropdown>
              </n-space>
            </template>

            <div style="flex: 1; overflow: hidden; min-height: 0; display: flex; flex-direction: column;">
              <!-- 完整报告视图 -->
              <div v-if="viewMode === 'full'" class="full-report-wrapper">

                  <MdPreview :modelValue="fullReport" :theme="darkTheme ? 'dark' : 'light'" class="md-preview-content"/>

              </div>

              <!-- 简洁列表视图 -->
              <div v-else class="simple-list-wrapper">
                <n-data-table
                  :columns="columns"
                  :data="simpleRecommendations"
                  :pagination="pagination"
                  size="small"
                  :scroll-x="1200"
                />
              </div>
            </div>
          </n-card>
        </n-col>
      </n-row>
    </div>
  </n-spin>

  <!-- 历史记录抽屉 -->
  <n-drawer v-model:show="historyVisible" width="50%" placement="right">
    <template #header>
      <n-text strong>历史推荐记录</n-text>
    </template>
    <n-spin :show="loadingHistory">
      <n-list v-if="historyList.length > 0">
        <n-list-item
          v-for="item in historyList"
          :key="item.ID"
          clickable
          @click="viewHistoryReport(item.ID)"
        >
          <n-thing>
            <template #header>
              <n-text>{{ item.QuerySummary }}</n-text>
            </template>
            <template #description>
              <n-space>
                <n-tag size="small" :type="item.Status === 'completed' ? 'success' : 'warning'">
                  {{ item.Status === 'completed' ? '已完成' : '处理中' }}
                </n-tag>
                <n-text type="info">
                  推荐数量: {{ item.RecommendCount }}
                </n-text>
                <n-text type="tertiary" depth="3">
                  {{ formatTime(item.CreatedAt) }}
                </n-text>
              </n-space>
            </template>
          </n-thing>
        </n-list-item>
      </n-list>
      <n-empty v-else description="暂无历史记录" />
    </n-spin>
  </n-drawer>
</template>

<script setup>
import {ref, onMounted, onBeforeUnmount, h, computed} from 'vue'
import {
  AIStockPickChat,
  GetStockPickReports,
  GetStockPickReport,
  GetStockPickRecommendations,
  FollowStockFromReport,
  ExportStockPickReport,
  CheckStockFollowed,
  GetStockPickStats
} from "../../wailsjs/go/main/App"
import {EventsOn, EventsOff} from "../../wailsjs/runtime"
import {MdPreview} from "md-editor-v3"
import {
  NCard,
  NRow,
  NCol,
  NSpin,
  NSpace,
  NTag,
  NScrollbar,
  NInput,
  NButton,
  NButtonGroup,
  NText,
  NIcon,
  NDataTable,
  NDrawer,
  NList,
  NListItem,
  NThing,
  NDropdown,
  NEmpty,
  useMessage,
  useNotification
} from "naive-ui"
import {TimeOutline, DownloadOutline} from '@vicons/ionicons5'

const message = useMessage()
const notify = useNotification()

// 状态管理
const analyzing = ref(false)
const inputText = ref('')
const messages = ref([
  {
    id: 1,
    role: 'assistant',
    content: '请告诉我您的选股需求，例如：\n\n- 推荐今日资金流入大的科技股\n- 寻找市盈率低于20且业绩增长的银行股\n- 推荐近期创新高的新能源龙头股',
    timestamp: Date.now()
  }
])
const toolsList = ref([])
const toolStatus = ref({})
const recommendations = ref([])
const viewMode = ref('full')
const fullReport = ref('')
const reportId = ref(null)
const darkTheme = ref(false)

// 历史记录相关
const historyVisible = ref(false)
const historyList = ref([])
const loadingHistory = ref(false)

// 导出选项
const exportOptions = [
  {
    label: '导出为Markdown',
    key: 'markdown'
  }
]

// 简洁列表数据
const simpleRecommendations = computed(() => {
  return recommendations.value.map((rec, index) => ({
    key: rec.stock_code,
    rank: index + 1,
    stock_code: rec.stock_code,
    stock_name: rec.stock_name,
    current_price: rec.current_price,
    price_change: rec.price_change,
    reason: rec.reason,
    target_change_percent: rec.target_change_percent,
    score: rec.score,
    is_followed: rec.is_followed
  }))
})

// 表格列配置
const columns = [
  { title: '排名', key: 'rank', width: 60, sorter: (a, b) => a.rank - b.rank },
  { title: '股票代码', key: 'stock_code', width: 100 },
  { title: '股票名称', key: 'stock_name', width: 120 },
  {
    title: '现价',
    key: 'current_price',
    width: 80,
    render: (row) => h(NText, { type: 'info' }, { default: () => row.current_price.toFixed(2) })
  },
  {
    title: '涨跌幅',
    key: 'price_change',
    width: 80,
    render: (row) => h(NText, { type: row.price_change >= 0 ? 'success' : 'error' }, { default: () => `${row.price_change >= 0 ? '+' : ''}${row.price_change.toFixed(2)}%` })
  },
  { title: '推荐理由', key: 'reason', ellipsis: { tooltip: true } },
  {
    title: '目标涨幅',
    key: 'target_change_percent',
    width: 90,
    sorter: (a, b) => a.target_change_percent - b.target_change_percent
  },
  {
    title: '评分',
    key: 'score',
    width: 80,
    sorter: (a, b) => a.score - b.score
  },
  {
    title: '操作',
    key: 'is_followed',
    width: 80,
    render: (row) => h(NButton, {
      size: 'tiny',
      type: row.is_followed ? 'default' : 'primary',
      disabled: row.is_followed,
      onClick: () => handleFollow(row.stock_code)
    }, { default: () => row.is_followed ? '已关注' : '关注' })
  }
]

// 分页配置
const pagination = ref({
  pageSize: 10
})

// 工具列表
const availableTools = [
  'SearchStockByIndicators',
  'GetStockKLine',
  'InteractiveAnswer',
  'GetStockResearchReport'
]

// 初始化
onMounted(() => {
  initEventListeners()
  toolsList.value = availableTools
})

onBeforeUnmount(() => {
  cleanupEventListeners()
})

// 事件监听
function initEventListeners() {
  EventsOn('ai-stock-pick-stream', handleStream)
  EventsOn('ai-stock-pick-start', handleStart)
  EventsOn('ai-stock-pick-tool', handleToolCall)
  EventsOn('ai-stock-pick-update', handleUpdate)
}

function cleanupEventListeners() {
  EventsOff('ai-stock-pick-stream')
  EventsOff('ai-stock-pick-start')
  EventsOff('ai-stock-pick-tool')
  EventsOff('ai-stock-pick-update')
}

// 处理回车发送
function handleEnter(e) {
  if (e.shiftKey) {
    // Shift+Enter 换行
    return
  }
  sendMessage()
}

// 发送消息
async function sendMessage() {
  const query = inputText.value.trim()
  if (!query) {
    message.warning('请输入选股需求')
    return
  }

  // 添加用户消息
  messages.value.push({
    role: 'user',
    content: query,
    timestamp: Date.now()
  })

  // 清空输入框
  inputText.value = ''

  // 添加AI响应占位
  const aiMessage = {
    role: 'assistant',
    content: '',
    timestamp: Date.now()
  }
  messages.value.push(aiMessage)

  // 开始分析
  analyzing.value = true
  fullReport.value = '正在分析市场数据...'

  try {
    const result = await AIStockPickChat(query, 0)
    reportId.value = parseInt(result)
    message.success('荐股分析完成')
  } catch (error) {
    message.error('荐股分析失败: ' + error)
    analyzing.value = false
  }
}

// 处理流式响应
function handleStream(data) {
  if (data.content) {
    // 更新最后一条AI消息
    const lastMessage = messages.value[messages.value.length - 1]
    if (lastMessage && lastMessage.role === 'assistant') {
      lastMessage.content += data.content
    }
    fullReport.value += data.content
  }
}

// 处理开始事件
function handleStart(data) {
  message.info(data.message)
}

// 处理工具调用
function handleToolCall(data) {
  toolStatus.value[data.tool_name] = data.status
  message.info(`调用工具: ${data.tool_name}`)
}

// 处理更新事件
function handleUpdate(data) {
  if (data.recommendations) {
    recommendations.value = data.recommendations
  }

  // 更新市场分析
  if (data.market_analysis) {
    // 检查当前报告是否已包含市场分析，如果没有则添加
    if (!fullReport.value.includes('## 市场环境分析')) {
      fullReport.value += '\n\n## 市场环境分析\n\n' + data.market_analysis + '\n\n'
    }
  }

  // 更新筛选逻辑
  if (data.filter_logic) {
    if (!fullReport.value.includes('## 筛选逻辑')) {
      fullReport.value += '## 筛选逻辑\n\n' + data.filter_logic + '\n\n'
    }
  }

  // 如果AI分析完成，显示成功消息并停止加载状态
  if (data.status === 'completed') {
    analyzing.value = false

    // 更新推荐股票部分
    let recMarkdown = ''
    if (data.recommendations && data.recommendations.length > 0) {
      recMarkdown = '## 推荐股票\n\n'
      data.recommendations.forEach((rec, index) => {
        recMarkdown += `${index + 1}. [${rec.stock_code}] ${rec.stock_name} - ${rec.reason || ''}\n`
        if (rec.current_price) {
          recMarkdown += `   - 当前价格：${rec.current_price.toFixed(2)}\n`
        }
        if (rec.price_change) {
          recMarkdown += `   - 涨跌幅：${rec.price_change >= 0 ? '+' : ''}${rec.price_change.toFixed(2)}%\n`
        }
        if (rec.technical_analysis) {
          recMarkdown += `   - 技术面分析：${rec.technical_analysis}\n`
        }
        if (rec.fundamental_analysis) {
          recMarkdown += `   - 基本面分析：${rec.fundamental_analysis}\n`
        }
        if (rec.target_change_percent) {
          recMarkdown += `   - 目标涨幅：${rec.target_change_percent}%\n`
        }
        if (rec.risk_tips) {
          recMarkdown += `   - 风险提示：${rec.risk_tips}\n`
        }
        recMarkdown += '\n'
      })
    }

    // 添加推荐股票到完整报告（如果不存在）
    if (recMarkdown && !fullReport.value.includes('## 推荐股票')) {
      fullReport.value += recMarkdown + '\n---\n\n本报告由AI智能分析生成，仅供参考，不构成投资建议。股市有风险，投资需谨慎。'
    }

    // 显示成功消息
    notify.success({
      title: 'AI分析完成',
      content: `成功生成${data.candidates_count || 0}个推荐股票，扫描了${data.total_scanned || 0}只股票`,
      duration: 5000
    })

    message.success(`AI荐股分析完成！推荐${data.candidates_count || 0}只股票`)
  }
}

// 获取工具状态
function getToolStatus(toolName) {
  const status = toolStatus.value[toolName]
  if (status === 'success') return 'success'
  if (status === 'running') return 'warning'
  if (status === 'failed') return 'error'
  return 'default'
}

// 显示历史记录
async function showHistory() {
  historyVisible.value = true
  await loadHistory()
}

// 加载历史记录
async function loadHistory() {
  loadingHistory.value = true
  try {
    const result = await GetStockPickReports(0, 20)

    console.log('GetStockPickReports 原始返回值:', result)
    console.log('返回值类型:', typeof result)

    // 新的返回格式: StockPickReportsResponse { items: [], total: number }
    let items = []

    if (result && result.items && Array.isArray(result.items)) {
      items = result.items
      console.log('从 result.items 获取，长度:', items.length)
      console.log('total:', result.total)
    } else {
      console.warn('GetStockPickReports 返回格式不正确:', result)
    }

    console.log('最终 items 数量:', items.length)

    if (items.length === 0) {
      message.info('暂无历史记录')
      historyList.value = []
      return
    }

    historyList.value = items
    message.success(`加载成功，共 ${items.length} 条记录`)
  } catch (error) {
    console.error('加载历史记录失败:', error)
    message.error('加载历史记录失败: ' + (error?.message || error || '未知错误'))
    historyList.value = []
  } finally {
    loadingHistory.value = false
  }
}

// 查看历史报告
async function viewHistoryReport(id) {
  try {
    console.log('加载历史报告:', id)
    const report = await GetStockPickReport(id)
    if (report) {
      reportId.value = report.id
      fullReport.value = formatReportToMarkdown(report)
      const recs = await GetStockPickRecommendations(id)
      recommendations.value = recs || []
      historyVisible.value = false
      message.success('报告加载成功')
    }
  } catch (error) {
    console.error('加载报告失败:', error)
    const errorMsg = error?.message || error?.toString() || error || '未知错误'

    // 判断是否是记录不存在的错误
    if (errorMsg.includes('record not found') || errorMsg.includes('记录不存在') || errorMsg.includes('已被删除')) {
      message.error('报告记录不存在或已被删除，请刷新历史记录列表')
      // 自动刷新历史记录列表
      setTimeout(async () => {
        await loadHistory()
      }, 1000)
    } else {
      message.error('加载报告失败: ' + errorMsg)
    }
  }
}

// 格式化报告为Markdown
function formatReportToMarkdown(report) {
  let markdown = ''
  markdown += `# AI荐股报告\n\n`
  markdown += `**生成时间：** ${formatTime(report.created_at)}\n\n`
  markdown += `**选股需求：** ${report.user_query}\n\n`

  if (report.market_analysis) {
    markdown += `## 市场环境分析\n\n${report.market_analysis}\n\n`
  }

  if (report.filter_logic) {
    markdown += `## 筛选逻辑\n\n${report.filter_logic}\n\n`
  }

  markdown += `## 推荐股票\n\n${report.recommendations}\n\n`

  markdown += `---\n\n`
  markdown += `*本报告由AI智能分析生成，仅供参考，不构成投资建议。股市有风险，投资需谨慎。*\n`

  return markdown
}

// 格式化时间
function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 导出报告
async function exportReport(format) {
  if (!reportId.value) {
    message.warning('当前没有可导出的报告')
    return
  }

  try {
    const result = await ExportStockPickReport(reportId.value, format)
    message.success(`已导出: ${result}`)
  } catch (error) {
    message.error('导出失败: ' + error)
  }
}

// 一键关注
async function handleFollow(stockCode) {
  try {
    const result = await FollowStockFromReport(reportId.value, stockCode)
    message.success(result)

    // 更新关注状态
    const index = recommendations.value.findIndex(r => r.stock_code === stockCode)
    if (index !== -1) {
      recommendations.value[index].is_followed = true
    }
  } catch (error) {
    message.error('关注失败: ' + error)
  }
}

// 检查股票是否已关注
async function checkFollowStatus() {
  for (const rec of recommendations.value) {
    try {
      const followed = await CheckStockFollowed(rec.stock_code)
      rec.is_followed = followed
    } catch (error) {
      console.error('检查关注状态失败:', error)
    }
  }
}

// 获取荐股统计
async function getStats() {
  try {
    const stats = await GetStockPickStats()
    console.log('荐股统计:', stats)
  } catch (error) {
    console.error('获取统计失败:', error)
  }
}
</script>

<style scoped>
.n-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.n-card > :deep(.n-card__content) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.full-report-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.full-report-wrapper :deep(.n-scrollbar) {
  flex: 1;
  min-height: 0;
}

.full-report-wrapper :deep(.n-scrollbar-container) {
  height: 100%;
  width: 100%;
  overflow: auto;
}

.md-preview-content {
  min-height: 100%;
  max-width: 100%;
  box-sizing: border-box;
  overflow-y: scroll;
  padding-right: 10px;
}

.simple-list-wrapper {
  flex: 1;
  overflow: auto;
  min-height: 0;
}
</style>
