<script setup>
import {h, onMounted, ref} from 'vue'
import {
  AddPosition,
  UpdatePosition,
  DeletePosition,
  GetPositions,
  GetPositionByID,
  RefreshPositions,
  GetPositionSummary,
  AddPositionFromRecommendation,
  AnalyzePosition,
  GetPositionAnalysis,
  AnalyzeAllPositions,
  GetLatestPositionAnalyses,
  SyncFollowedToPositions,
  AnalyzePortfolio,
  GetLatestPortfolioAnalysis,
  Greet
} from "../../wailsjs/go/main/App";
import {NButton, NPopconfirm, useMessage, NModal, NCard, NStatistic, NGrid, NGi, NTabPane, NTabs, NSpin, NAlert, NFlex, NTag, NProgress, NIcon, NText} from "naive-ui";
import {SyncOutline} from "@vicons/ionicons5";

const message = useMessage()
const loading = ref(false)
const syncingFromFollowed = ref(false)
const positions = ref([])
const summary = ref({
  total_market_value: 0,
  total_profit_loss: 0,
  position_count: 0
})

// 分析弹窗相关
const showAnalysisModal = ref(false)
const analyzing = ref(false)
const currentPosition = ref(null)
const currentAnalysis = ref(null)

// 批量分析相关
const showBatchAnalysisModal = ref(false)
const batchAnalyzing = ref(false)
const batchProgress = ref(0)
const batchAnalyses = ref([])
const latestAnalysesMap = ref(new Map())

// 整体仓位分析相关
const showPortfolioAnalysisModal = ref(false)
const portfolioAnalyzing = ref(false)
const currentPortfolioAnalysis = ref(null)

// 添加持仓弹窗相关
const showAddModal = ref(false)
const addForm = ref({
  stockCode: '',
  stockName: '',
  quantity: 100,
  buyPrice: 0,
  buyDate: null,
  notes: ''
})
const addFormCurrentPrice = ref(0)
const addFormProfitLoss = ref(0)
const addFormProfitLossPct = ref(0)
const queryingStockPrice = ref(false)

// 编辑持仓弹窗相关
const showEditModal = ref(false)
const editingPositionId = ref(null)
const editForm = ref({
  stockCode: '',
  stockName: '',
  quantity: 100,
  buyPrice: 0,
  buyDate: null,
  notes: ''
})
const editFormCurrentPrice = ref(0)
const editFormProfitLoss = ref(0)
const editFormProfitLossPct = ref(0)
const editQueryingStockPrice = ref(false)

// 表格列定义
const columns = [
  {
    title: '股票',
    key: 'stock',
    width: 180,
    render: (row) => h('div', [
      h('div', { style: { fontWeight: 'bold' } }, row.stock_name),
      h('div', { style: { fontSize: '12px', color: '#999' } }, row.stock_code)
    ])
  },
  {
    title: '数量',
    key: 'quantity',
    width: 100,
    render: (row) => formatNumber(row.quantity, 0)
  },
  {
    title: '成本价',
    key: 'buyPrice',
    width: 100,
    render: (row) => '¥' + formatNumber(row.buy_price, 3)
  },
  {
    title: '现价',
    key: 'currentPrice',
    width: 100,
    render: (row) => '¥' + formatNumber(row.current_price)
  },
  {
    title: '盈亏',
    key: 'profitLoss',
    width: 120,
    render: (row) => h('span', {
      style: { color: row.profit_loss >= 0 ? '#18a058' : '#d03050' }
    }, '¥' + formatNumber(row.profit_loss))
  },
  {
    title: '盈亏%',
    key: 'profitLossPct',
    width: 110,
    render: (row) => h('span', {
      style: { color: row.profit_loss_pct >= 0 ? '#18a058' : '#d03050' }
    }, formatNumber(row.profit_loss_pct, 3) + '%')
  },
  {
    title: '市值',
    key: 'marketValue',
    width: 120,
    render: (row) => '¥' + formatNumber(row.market_value)
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    fixed: 'right',
    render: (row) => h('div', { style: { display: 'flex', gap: '8px' } }, [
      h(NButton, { size: 'small', type: 'primary', onClick: () => handleEdit(row) }, { default: () => '编辑' }),
      h(NButton, { size: 'small', type: 'info', onClick: () => handleAnalyze(row) }, { default: () => '分析' }),
      h(NPopconfirm, { onPositiveClick: () => handleDelete(row.id) }, {
        default: () => '确定删除该持仓？',
        trigger: () => h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' })
      })
    ])
  }
]

// 批量分析表格列定义
const batchAnalysisColumns = [
  {
    title: '股票',
    key: 'stock',
    width: 180,
    render: (row) => h('div', [
      h('div', { style: { fontWeight: 'bold' } }, row.stock_name),
      h('div', { style: { fontSize: '12px', color: '#999' } }, row.stock_code)
    ])
  },
  {
    title: '总体建议',
    key: 'advice',
    width: 120,
    render: (row) => h(NTag, { type: getAdviceColor(row.overall_advice), size: 'small' }, { default: () => row.overall_advice })
  },
  {
    title: '置信度',
    key: 'confidence',
    width: 100,
    render: (row) => formatNumber(row.confidence * 100, 1) + '%'
  },
  {
    title: '建议补仓价',
    key: 'buyPrice',
    width: 120,
    render: (row) => row.suggested_buy_price ? '¥' + formatNumber(row.suggested_buy_price) : '-'
  },
  {
    title: '建议止盈价',
    key: 'sellPrice',
    width: 120,
    render: (row) => row.suggested_sell_price ? '¥' + formatNumber(row.suggested_sell_price) : '-'
  },
  {
    title: '止损价位',
    key: 'stopLoss',
    width: 120,
    render: (row) => h('span', { style: { color: '#d03050' } }, row.stop_loss_price ? '¥' + formatNumber(row.stop_loss_price) : '-')
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    fixed: 'right',
    render: (row) => h(NButton, { size: 'small', type: 'primary', onClick: () => viewAnalysisFromBatch(row.position_id) }, { default: () => '详情' })
  }
]

// 格式化数字
function formatNumber(num, decimals = 2) {
  if (num === null || num === undefined) return '-'
  return Number(num).toLocaleString('zh-CN', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  })
}

// 获取建议标签颜色
function getAdviceColor(advice) {
  const colors = {
    '持有': 'default',
    '加仓': 'success',
    '减仓': 'warning',
    '清仓': 'error'
  }
  return colors[advice] || 'default'
}

// 查询股票实时价格
async function queryStockPrice(stockCode, isEdit = false) {
  if (!stockCode || stockCode.trim() === '') {
    return
  }

  const querying = isEdit ? editQueryingStockPrice : queryingStockPrice
  const currentPrice = isEdit ? editFormCurrentPrice : addFormCurrentPrice
  const form = isEdit ? editForm : addForm

  querying.value = true
  try {
    const stockInfo = await Greet(stockCode.trim())
    if (stockInfo && stockInfo.Price) {
      const price = parseFloat(stockInfo.Price)
      if (!isNaN(price)) {
        currentPrice.value = price
        // 如果股票名称为空，自动填充
        if (stockInfo.Name && (!form.value.stockName || form.value.stockName === '')) {
          form.value.stockName = stockInfo.Name
        }
        // 重新计算盈亏
        calculateProfitLoss(isEdit)
      }
    }
  } catch (err) {
    console.error('查询股票价格失败:', err)
  } finally {
    querying.value = false
  }
}

// 处理成本价变化：自动查询股价并计算盈亏
async function handleBuyPriceChange(isEdit = false) {
  const form = isEdit ? editForm : addForm
  const currentPrice = isEdit ? editFormCurrentPrice : addFormCurrentPrice

  // 如果有股票代码，查询最新股价（如果还没有则必须查询）
  if (form.value.stockCode) {
    if (currentPrice.value <= 0) {
      // 还没有股价，必须查询
      await queryStockPrice(form.value.stockCode, isEdit)
    } else {
      // 已有股价，直接计算盈亏，但可以让用户手动点击刷新按钮更新
      calculateProfitLoss(isEdit)
    }
  } else {
    // 没有股票代码，只计算盈亏（显示为0）
    calculateProfitLoss(isEdit)
  }
}

// 计算盈亏
function calculateProfitLoss(isEdit = false) {
  const form = isEdit ? editForm : addForm
  const currentPrice = isEdit ? editFormCurrentPrice : addFormCurrentPrice
  const profitLoss = isEdit ? editFormProfitLoss : addFormProfitLoss
  const profitLossPct = isEdit ? editFormProfitLossPct : addFormProfitLossPct

  if (form.value.buyPrice > 0 && form.value.quantity > 0 && currentPrice.value > 0) {
    const marketValue = currentPrice.value * form.value.quantity
    const costValue = form.value.buyPrice * form.value.quantity
    profitLoss.value = marketValue - costValue
    if (costValue > 0) {
      profitLossPct.value = ((currentPrice.value - form.value.buyPrice) / form.value.buyPrice) * 100
    }
  } else {
    profitLoss.value = 0
    profitLossPct.value = 0
  }
}

// 从自选股同步持仓
async function syncFromFollowedStocks() {
  syncingFromFollowed.value = true
  try {
    const result = await SyncFollowedToPositions()
    message.success(result || '同步成功')
    await loadPositions()
  } catch (err) {
    message.error('同步失败: ' + err)
  } finally {
    syncingFromFollowed.value = false
  }
}

// 加载持仓数据
async function loadPositions() {
  loading.value = true
  try {
    positions.value = await GetPositions()
    await loadSummary()
  } catch (err) {
    message.error('加载持仓数据失败: ' + err)
  } finally {
    loading.value = false
  }
}

// 加载汇总数据
async function loadSummary() {
  try {
    summary.value = await GetPositionSummary()
  } catch (err) {
    console.error('加载汇总数据失败:', err)
  }
}

// 刷新持仓价格
async function refreshPositions() {
  loading.value = true
  try {
    const result = await RefreshPositions()
    message.success(result)
    await loadPositions()
  } catch (err) {
    message.error('刷新持仓数据失败: ' + err)
  } finally {
    loading.value = false
  }
}

// 添加持仓
async function handleAddPosition() {
  if (!addForm.value.stockCode || !addForm.value.stockName || !addForm.value.quantity || !addForm.value.buyPrice) {
    message.warning('请填写完整信息')
    return
  }

  try {
    const pos = {
      stock_code: addForm.value.stockCode,
      stock_name: addForm.value.stockName,
      quantity: addForm.value.quantity,
      buy_price: addForm.value.buyPrice,
      buy_date: addForm.value.buyDate ? new Date(addForm.value.buyDate).toISOString() : new Date().toISOString(),
      current_price: addFormCurrentPrice.value > 0 ? addFormCurrentPrice.value : addForm.value.buyPrice,
      notes: addForm.value.notes,
      is_active: true
    }
    await AddPosition(pos)
    message.success('持仓添加成功')
    showAddModal.value = false
    resetAddForm()
    await loadPositions()
  } catch (err) {
    message.error('添加持仓失败: ' + err)
  }
}

// 重置添加表单
function resetAddForm() {
  addForm.value = {
    stockCode: '',
    stockName: '',
    quantity: 100,
    buyPrice: 0,
    buyDate: null,
    notes: ''
  }
  addFormCurrentPrice.value = 0
  addFormProfitLoss.value = 0
  addFormProfitLossPct.value = 0
}

// 编辑持仓
async function handleEdit(row) {
  editingPositionId.value = row.id
  editForm.value = {
    stockCode: row.stock_code,
    stockName: row.stock_name,
    quantity: row.quantity,
    buyPrice: row.buy_price,
    buyDate: row.buy_date ? new Date(row.buy_date) : null,
    notes: row.notes || ''
  }
  // 初始化当前价格
  editFormCurrentPrice.value = row.current_price || 0
  // 计算盈亏
  calculateProfitLoss(true)
  showEditModal.value = true
  // 自动查询最新股价
  await queryStockPrice(row.stock_code, true)
}

// 保存编辑持仓
async function handleSaveEdit() {
  if (!editForm.value.stockCode || !editForm.value.stockName || !editForm.value.quantity || !editForm.value.buyPrice) {
    message.warning('请填写完整信息')
    return
  }

  try {
    const pos = {
      id: editingPositionId.value,
      stock_code: editForm.value.stockCode,
      stock_name: editForm.value.stockName,
      quantity: editForm.value.quantity,
      buy_price: editForm.value.buyPrice,
      buy_date: editForm.value.buyDate ? new Date(editForm.value.buyDate).toISOString() : null,
      notes: editForm.value.notes
    }
    await UpdatePosition(pos)
    message.success('持仓更新成功')
    showEditModal.value = false
    resetEditForm()
    await loadPositions()
  } catch (err) {
    message.error('更新持仓失败: ' + err)
  }
}

// 重置编辑表单
function resetEditForm() {
  editingPositionId.value = null
  editForm.value = {
    stockCode: '',
    stockName: '',
    quantity: 100,
    buyPrice: 0,
    buyDate: null,
    notes: ''
  }
  editFormCurrentPrice.value = 0
  editFormProfitLoss.value = 0
  editFormProfitLossPct.value = 0
}

// 分析持仓
async function handleAnalyze(row) {
  currentPosition.value = row
  currentAnalysis.value = null
  showAnalysisModal.value = true

  // 先尝试获取已有的分析结果
  try {
    const existingAnalysis = await GetPositionAnalysis(row.id)
    if (existingAnalysis) {
      currentAnalysis.value = existingAnalysis
      return
    }
  } catch (err) {
    console.log('暂无历史分析结果')
  }

  // 如果没有历史分析，自动进行新分析
  await doAnalyze()
}

// 执行分析
async function doAnalyze() {
  if (!currentPosition.value) return

  analyzing.value = true
  try {
    const analysis = await AnalyzePosition(currentPosition.value.id, 0)
    currentAnalysis.value = analysis
    message.success('分析完成')
  } catch (err) {
    message.error('分析失败: ' + err)
  } finally {
    analyzing.value = false
  }
}

// 删除持仓
async function handleDelete(id) {
  try {
    await DeletePosition(id)
    message.success('持仓删除成功')
    await loadPositions()
  } catch (err) {
    message.error('删除持仓失败: ' + err)
  }
}

// 批量分析持仓
async function handleBatchAnalyze() {
  if (positions.value.length === 0) {
    message.warning('暂无持仓可分析')
    return
  }

  showBatchAnalysisModal.value = true
  batchAnalyses.value = []
  batchProgress.value = 0

  // 先尝试获取已有的分析结果
  await loadLatestAnalyses()

  // 如果已有分析结果，直接显示
  if (batchAnalyses.value.length > 0) {
    return
  }

  // 否则执行新的批量分析
  await doBatchAnalyze()
}

// 加载最新分析结果
async function loadLatestAnalyses() {
  try {
    const analyses = await GetLatestPositionAnalyses()
    if (analyses && analyses.length > 0) {
      batchAnalyses.value = analyses
      latestAnalysesMap.value = new Map(analyses.map(a => [a.position_id, a]))
      return
    }
  } catch (err) {
    console.log('加载历史分析结果失败:', err)
  }
}

// 执行批量分析
async function doBatchAnalyze() {
  batchAnalyzing.value = true
  batchProgress.value = 0

  try {
    const analyses = await AnalyzeAllPositions(0)
    batchAnalyses.value = analyses
    latestAnalysesMap.value = new Map(analyses.map(a => [a.position_id, a]))
    batchProgress.value = 100
    message.success('批量分析完成')
  } catch (err) {
    message.error('批量分析失败: ' + err)
  } finally {
    batchAnalyzing.value = false
  }
}

// 从批量分析中查看单个持仓分析
function viewAnalysisFromBatch(positionId) {
  const pos = positions.value.find(p => p.id === positionId)
  if (!pos) return

  currentPosition.value = pos
  currentAnalysis.value = latestAnalysesMap.value.get(positionId) || null
  showBatchAnalysisModal.value = false
  showAnalysisModal.value = true
}

// 分析整体仓位
async function handlePortfolioAnalysis() {
  if (positions.value.length === 0) {
    message.warning('暂无持仓可分析')
    return
  }

  currentPortfolioAnalysis.value = null
  showPortfolioAnalysisModal.value = true

  // 先尝试获取已有的分析结果
  try {
    const existingAnalysis = await GetLatestPortfolioAnalysis()
    if (existingAnalysis) {
      currentPortfolioAnalysis.value = existingAnalysis
      return
    }
  } catch (err) {
    console.log('暂无历史整体分析结果')
  }

  // 如果没有历史分析，自动进行新分析
  await doPortfolioAnalysis()
}

// 执行整体仓位分析
async function doPortfolioAnalysis() {
  portfolioAnalyzing.value = true
  try {
    const analysis = await AnalyzePortfolio(0)
    currentPortfolioAnalysis.value = analysis
    message.success('整体仓位分析完成')
  } catch (err) {
    message.error('整体仓位分析失败: ' + err)
  } finally {
    portfolioAnalyzing.value = false
  }
}

onMounted(() => {
  loadPositions()
})
</script>

<template>
  <div class="position-analysis">
    <n-spin :show="loading">
      <!-- 汇总卡片区域 -->
      <n-card title="持仓汇总" style="margin-bottom: 16px;">
        <n-grid :cols="4" :x-gap="16">
          <n-gi>
            <n-statistic label="总市值">
              <template #prefix>¥</template>
              {{ formatNumber(summary.total_market_value) }}
            </n-statistic>
          </n-gi>
          <n-gi>
            <n-statistic label="总盈亏">
              <template #prefix>¥</template>
              <template #value>
                <span :style="{ color: summary.total_profit_loss >= 0 ? '#18a058' : '#d03050' }">
                  {{ formatNumber(summary.total_profit_loss) }}
                </span>
              </template>
            </n-statistic>
          </n-gi>
          <n-gi>
            <n-statistic label="持仓数" :value="summary.position_count" />
          </n-gi>
          <n-gi>
            <n-flex justify="end" gap="8">
              <n-button type="default" :loading="syncingFromFollowed" @click="syncFromFollowedStocks">
                <template #icon>
                  <n-icon><sync-outline /></n-icon>
                </template>
                从自选股同步
              </n-button>
              <n-button type="primary" @click="showAddModal = true">
                添加持仓
              </n-button>
              <n-button type="info" @click="refreshPositions">
                刷新数据
              </n-button>
              <n-button type="warning" @click="handleBatchAnalyze">
                AI批量分析
              </n-button>
              <n-button type="success" @click="handlePortfolioAnalysis">
                整体仓位分析
              </n-button>
            </n-flex>
          </n-gi>
        </n-grid>
      </n-card>

      <!-- 持仓列表区域 -->
      <n-card title="持仓列表">
        <div v-if="positions.length === 0" style="text-align: center; padding: 40px;">
          <n-text type="info">暂无持仓数据，点击"添加持仓"开始</n-text>
        </div>
        <div v-else>
          <n-data-table
            :columns="columns"
            :data="positions"
            :pagination="false"
            :bordered="false"
            size="small"
          />
        </div>
      </n-card>
    </n-spin>

    <!-- 添加持仓弹窗 -->
    <n-modal v-model:show="showAddModal" preset="card" title="添加持仓" style="width: 500px;">
      <n-form :model="addForm" label-placement="left" label-width="100">
        <n-form-item label="股票代码">
          <n-input
            v-model:value="addForm.stockCode"
            placeholder="请输入股票代码，如：000001.SZ"
            @blur="queryStockPrice(addForm.stockCode, false)"
          />
        </n-form-item>
        <n-form-item label="股票名称">
          <n-input v-model:value="addForm.stockName" placeholder="请输入股票名称" />
        </n-form-item>
        <n-form-item label="持股数量">
          <n-input-number
            v-model:value="addForm.quantity"
            :min="1"
            placeholder="请输入持股数量"
            style="width: 100%;"
            @update:value="handleBuyPriceChange(false)"
          />
        </n-form-item>
        <n-form-item label="买入价格">
          <n-input-number
            v-model:value="addForm.buyPrice"
            :min="0"
            :precision="3"
            placeholder="请输入买入价格"
            style="width: 100%;"
            @update:value="handleBuyPriceChange(false)"
          />
        </n-form-item>
        <n-form-item label="当前股价">
          <n-input-number
            :value="addFormCurrentPrice"
            :precision="3"
            placeholder="自动查询"
            style="width: 100%;"
            readonly
            :disabled="true"
          />
          <n-button
            v-if="addForm.stockCode"
            size="small"
            type="info"
            :loading="queryingStockPrice"
            style="margin-top: 8px;"
            @click="queryStockPrice(addForm.stockCode, false)"
          >
            {{ queryingStockPrice ? '查询中...' : '刷新股价' }}
          </n-button>
        </n-form-item>
        <n-form-item label="预估盈亏" v-if="addFormCurrentPrice > 0 && addForm.buyPrice > 0 && addForm.quantity > 0">
          <div>
            <span :style="{ color: addFormProfitLoss >= 0 ? '#18a058' : '#d03050', fontSize: '18px', fontWeight: 'bold' }">
              ¥{{ formatNumber(addFormProfitLoss) }}
            </span>
            <span :style="{ color: addFormProfitLossPct >= 0 ? '#18a058' : '#d03050', marginLeft: '12px' }">
              ({{ formatNumber(addFormProfitLossPct, 3) }}%)
            </span>
          </div>
        </n-form-item>
        <n-form-item label="买入日期">
          <n-date-picker v-model:value="addForm.buyDate" type="date" placeholder="请选择买入日期" style="width: 100%;" />
        </n-form-item>
        <n-form-item label="备注">
          <n-input v-model:value="addForm.notes" type="textarea" placeholder="请输入备注" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-flex justify="end">
          <n-button @click="showAddModal = false" style="margin-right: 8px;">取消</n-button>
          <n-button type="primary" @click="handleAddPosition">确认添加</n-button>
        </n-flex>
      </template>
    </n-modal>

    <!-- 编辑持仓弹窗 -->
    <n-modal v-model:show="showEditModal" preset="card" title="编辑持仓" style="width: 500px;">
      <n-form :model="editForm" label-placement="left" label-width="100">
        <n-form-item label="股票代码">
          <n-input
            v-model:value="editForm.stockCode"
            placeholder="请输入股票代码，如：000001.SZ"
            @blur="queryStockPrice(editForm.stockCode, true)"
          />
        </n-form-item>
        <n-form-item label="股票名称">
          <n-input v-model:value="editForm.stockName" placeholder="请输入股票名称" />
        </n-form-item>
        <n-form-item label="持股数量">
          <n-input-number
            v-model:value="editForm.quantity"
            :min="1"
            placeholder="请输入持股数量"
            style="width: 100%;"
            @update:value="handleBuyPriceChange(true)"
          />
        </n-form-item>
        <n-form-item label="买入价格">
          <n-input-number
            v-model:value="editForm.buyPrice"
            :min="0"
            :precision="3"
            placeholder="请输入买入价格"
            style="width: 100%;"
            @update:value="handleBuyPriceChange(true)"
          />
        </n-form-item>
        <n-form-item label="当前股价">
          <n-input-number
            :value="editFormCurrentPrice"
            :precision="3"
            placeholder="自动查询"
            style="width: 100%;"
            readonly
            :disabled="true"
          />
          <n-button
            v-if="editForm.stockCode"
            size="small"
            type="info"
            :loading="editQueryingStockPrice"
            style="margin-top: 8px;"
            @click="queryStockPrice(editForm.stockCode, true)"
          >
            {{ editQueryingStockPrice ? '查询中...' : '刷新股价' }}
          </n-button>
        </n-form-item>
        <n-form-item label="预估盈亏" v-if="editFormCurrentPrice > 0 && editForm.buyPrice > 0 && editForm.quantity > 0">
          <div>
            <span :style="{ color: editFormProfitLoss >= 0 ? '#18a058' : '#d03050', fontSize: '18px', fontWeight: 'bold' }">
              ¥{{ formatNumber(editFormProfitLoss) }}
            </span>
            <span :style="{ color: editFormProfitLossPct >= 0 ? '#18a058' : '#d03050', marginLeft: '12px' }">
              ({{ formatNumber(editFormProfitLossPct, 3) }}%)
            </span>
          </div>
        </n-form-item>
        <n-form-item label="买入日期">
          <n-date-picker v-model:value="editForm.buyDate" type="date" placeholder="请选择买入日期" style="width: 100%;" />
        </n-form-item>
        <n-form-item label="备注">
          <n-input v-model:value="editForm.notes" type="textarea" placeholder="请输入备注" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-flex justify="end">
          <n-button @click="showEditModal = false" style="margin-right: 8px;">取消</n-button>
          <n-button type="primary" @click="handleSaveEdit">保存修改</n-button>
        </n-flex>
      </template>
    </n-modal>

    <!-- 持仓分析详情弹窗 -->
    <n-modal v-model:show="showAnalysisModal" preset="card" title="持仓分析" style="width: 800px;">
      <n-spin :show="analyzing">
        <div v-if="currentPosition" style="margin-bottom: 16px;">
          <n-alert type="info">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>{{ currentPosition.stock_name }} ({{ currentPosition.stock_code }})</span>
                <n-tag v-if="currentAnalysis" :type="getAdviceColor(currentAnalysis.overall_advice)">
                  {{ currentAnalysis.overall_advice }}
                </n-tag>
              </div>
            </template>
            <n-grid :cols="3" :x-gap="16">
              <n-gi>成本价: ¥{{ formatNumber(currentPosition.buy_price, 3) }}</n-gi>
              <n-gi>现价: ¥{{ formatNumber(currentPosition.current_price) }}</n-gi>
              <n-gi>
                盈亏:
                <span :style="{ color: currentPosition.profit_loss >= 0 ? '#18a058' : '#d03050' }">
                  ¥{{ formatNumber(currentPosition.profit_loss) }} ({{ formatNumber(currentPosition.profit_loss_pct, 3) }}%)
                </span>
              </n-gi>
            </n-grid>
          </n-alert>
        </div>

        <div v-if="!currentAnalysis && !analyzing" style="text-align: center; padding: 40px;">
          <n-button type="primary" @click="doAnalyze">开始分析</n-button>
        </div>

        <div v-if="currentAnalysis">
          <n-tabs type="line">
            <n-tab-pane name="summary" tab="分析摘要">
              <n-card style="margin-top: 16px;">
                <n-grid :cols="2" :x-gap="16" :y-gap="16">
                  <n-gi>
                    <n-statistic label="总体建议">
                      <n-tag :type="getAdviceColor(currentAnalysis.overall_advice)" size="large">
                        {{ currentAnalysis.overall_advice }}
                      </n-tag>
                    </n-statistic>
                  </n-gi>
                  <n-gi>
                    <n-statistic label="置信度" :value="currentAnalysis.confidence * 100" suffix="%" />
                  </n-gi>
                  <n-gi>
                    <n-statistic label="建议补仓价">
                      <template #value>
                        <span v-if="currentAnalysis.suggested_buy_price">¥{{ formatNumber(currentAnalysis.suggested_buy_price) }}</span>
                        <span v-else>-</span>
                      </template>
                    </n-statistic>
                  </n-gi>
                  <n-gi>
                    <n-statistic label="建议止盈价">
                      <template #value>
                        <span v-if="currentAnalysis.suggested_sell_price">¥{{ formatNumber(currentAnalysis.suggested_sell_price) }}</span>
                        <span v-else>-</span>
                      </template>
                    </n-statistic>
                  </n-gi>
                  <n-gi :span="2">
                    <n-statistic label="止损价位">
                      <template #value>
                        <span style="color: #d03050;">¥{{ formatNumber(currentAnalysis.stop_loss_price) }}</span>
                      </template>
                    </n-statistic>
                  </n-gi>
                </n-grid>
              </n-card>
            </n-tab-pane>
            <n-tab-pane name="technical" tab="技术面分析">
              <n-card style="margin-top: 16px;">
                <p style="white-space: pre-wrap; line-height: 1.8;">{{ currentAnalysis.technical_analysis }}</p>
              </n-card>
            </n-tab-pane>
            <n-tab-pane name="fundamental" tab="基本面分析">
              <n-card style="margin-top: 16px;">
                <p style="white-space: pre-wrap; line-height: 1.8;">{{ currentAnalysis.fundamental_analysis }}</p>
              </n-card>
            </n-tab-pane>
            <n-tab-pane name="risk" tab="风险分析">
              <n-card style="margin-top: 16px;">
                <p style="white-space: pre-wrap; line-height: 1.8;">{{ currentAnalysis.risk_analysis }}</p>
              </n-card>
            </n-tab-pane>
          </n-tabs>
        </div>
      </n-spin>
      <template #footer>
        <n-flex justify="end" gap="8">
          <n-button v-if="currentAnalysis" type="info" @click="doAnalyze">重新分析</n-button>
          <n-button @click="showAnalysisModal = false">关闭</n-button>
        </n-flex>
      </template>
    </n-modal>

    <!-- 批量分析弹窗 -->
    <n-modal v-model:show="showBatchAnalysisModal" preset="card" title="批量持仓分析" style="width: 900px;">
      <n-spin :show="batchAnalyzing">
        <div v-if="batchAnalyzing" style="padding: 20px 0;">
          <n-progress type="line" :percentage="batchProgress" :show-indicator="true" status="active" />
          <div style="text-align: center; margin-top: 16px; color: #666;">
            正在分析持仓，请稍候...
          </div>
        </div>

        <div v-else-if="batchAnalyses.length === 0" style="text-align: center; padding: 40px;">
          <n-button type="primary" @click="doBatchAnalyze">开始批量分析</n-button>
        </div>

        <div v-else>
          <n-card style="margin-bottom: 16px;">
            <n-grid :cols="4" :x-gap="16">
              <n-gi>
                <n-statistic label="持仓总数" :value="batchAnalyses.length" />
              </n-gi>
              <n-gi>
                <n-statistic label="建议加仓" :value="batchAnalyses.filter(a => a.overall_advice === '加仓').length" />
              </n-gi>
              <n-gi>
                <n-statistic label="建议持有" :value="batchAnalyses.filter(a => a.overall_advice === '持有').length" />
              </n-gi>
              <n-gi>
                <n-statistic label="建议减仓/清仓" :value="batchAnalyses.filter(a => a.overall_advice === '减仓' || a.overall_advice === '清仓').length" />
              </n-gi>
            </n-grid>
          </n-card>

          <n-data-table
            :columns="batchAnalysisColumns"
            :data="batchAnalyses"
            :pagination="{ pageSize: 10 }"
            :bordered="false"
            size="small"
          />
        </div>
      </n-spin>
      <template #footer>
        <n-flex justify="end" gap="8">
          <n-button v-if="!batchAnalyzing && batchAnalyses.length > 0" type="info" @click="doBatchAnalyze">重新分析</n-button>
          <n-button @click="showBatchAnalysisModal = false">关闭</n-button>
        </n-flex>
      </template>
    </n-modal>

    <!-- 整体仓位分析弹窗 -->
    <n-modal v-model:show="showPortfolioAnalysisModal" preset="card" title="整体仓位分析" style="width: 900px;">
      <n-spin :show="portfolioAnalyzing">
        <div v-if="!currentPortfolioAnalysis && !portfolioAnalyzing" style="text-align: center; padding: 40px;">
          <n-button type="primary" @click="doPortfolioAnalysis">开始分析</n-button>
        </div>

        <div v-if="currentPortfolioAnalysis">
          <n-tabs type="line">
            <n-tab-pane name="overall" tab="整体评估">
              <n-card style="margin-top: 16px;">
                <p style="white-space: pre-wrap; line-height: 1.8;">{{ currentPortfolioAnalysis.overall_assessment }}</p>
              </n-card>
            </n-tab-pane>
            <n-tab-pane name="allocation" tab="仓位分布分析">
              <n-card style="margin-top: 16px;">
                <p style="white-space: pre-wrap; line-height: 1.8;">{{ currentPortfolioAnalysis.allocation_analysis }}</p>
              </n-card>
            </n-tab-pane>
            <n-tab-pane name="risk" tab="风险评估">
              <n-card style="margin-top: 16px;">
                <p style="white-space: pre-wrap; line-height: 1.8;">{{ currentPortfolioAnalysis.risk_assessment }}</p>
              </n-card>
            </n-tab-pane>
            <n-tab-pane name="suggestions" tab="调整建议">
              <n-card style="margin-top: 16px;">
                <p style="white-space: pre-wrap; line-height: 1.8;">{{ currentPortfolioAnalysis.adjustment_suggestions }}</p>
              </n-card>
            </n-tab-pane>
          </n-tabs>
        </div>
      </n-spin>
      <template #footer>
        <n-flex justify="end" gap="8">
          <n-button v-if="currentPortfolioAnalysis" type="info" @click="doPortfolioAnalysis">重新分析</n-button>
          <n-button @click="showPortfolioAnalysisModal = false">关闭</n-button>
        </n-flex>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.position-analysis {
  padding: 0;
}
</style>
