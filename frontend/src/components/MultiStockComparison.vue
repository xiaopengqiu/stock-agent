<template>
  <n-card class="multi-stock-card" hoverable>
    <template #header>
      <div class="card-header">
        <div class="header-left">
          <n-icon :size="22" color="#2080f0">
            <BarChartIcon />
          </n-icon>
          <span class="card-title">多股票对比</span>
        </div>
        <n-space>
          <n-button-group size="small">
            <n-button 
              :type="compareMode === 'bar' ? 'primary' : 'default'"
              @click="compareMode = 'bar'"
            >
              柱状图
            </n-button>
            <n-button 
              :type="compareMode === 'line' ? 'primary' : 'default'"
              @click="compareMode = 'line'"
            >
              折线图
            </n-button>
            <n-button 
              :type="compareMode === 'heatmap' ? 'primary' : 'default'"
              @click="compareMode = 'heatmap'"
            >
              热力图
            </n-button>
          </n-button-group>
        </n-space>
      </div>
    </template>

    <div class="comparison-chart-wrapper">
      <v-chart 
        ref="chartRef"
        :option="chartOption" 
        :autoresize="true"
        style="height: 320px;" 
      />
    </div>

    <div class="comparison-metrics">
      <div class="metric-tabs">
        <div 
          v-for="metric in metrics" 
          :key="metric.key"
          class="metric-tab"
          :class="{ 'active': currentMetric === metric.key }"
          @click="currentMetric = metric.key"
        >
          {{ metric.label }}
        </div>
      </div>
    </div>

    <div class="stock-summary">
      <n-divider style="margin: 12px 0;" />
      <div class="summary-grid">
        <div 
          v-for="stock in stocks" 
          :key="stock.code"
          class="stock-summary-item"
          :class="{ 'highlighted': selectedStock?.code === stock.code }"
          @click="selectStock(stock)"
        >
          <div class="stock-name">{{ stock.name }}</div>
          <div class="stock-code">{{ stock.code }}</div>
          <div class="stock-metrics">
            <span class="metric-item" :class="{ 'positive': stock.return >= 0, 'negative': stock.return < 0 }">
              {{ stock.return >= 0 ? '+' : '' }}{{ stock.return.toFixed(1) }}%
            </span>
            <span class="metric-separator">|</span>
            <span class="metric-item">PE: {{ stock.pe.toFixed(1) }}</span>
            <span class="metric-separator">|</span>
            <span class="metric-item">ROE: {{ stock.roe.toFixed(1) }}%</span>
          </div>
        </div>
      </div>
    </div>

    <div class="selected-detail" v-if="selectedStock">
      <n-divider style="margin: 12px 0;" />
      <div class="selected-detail-content">
        <div class="detail-header">
          <n-tag :type="getStockTagType(selectedStock)" size="small">
            {{ selectedStock.recommend }}
          </n-tag>
          <span class="detail-name">{{ selectedStock.name }}</span>
          <span class="detail-code">{{ selectedStock.code }}</span>
        </div>
        <n-grid :x-gap="12" :y-gap="8" :cols="3">
          <n-grid-item>
            <n-statistic label="预期收益">
              <template #value>
                <span :class="{ 'text-positive': selectedStock.return >= 0, 'text-negative': selectedStock.return < 0 }">
                  {{ selectedStock.return >= 0 ? '+' : '' }}{{ selectedStock.return.toFixed(1) }}%
                </span>
              </template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="市盈率 PE">
              <template #value>{{ selectedStock.pe.toFixed(1) }}</template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="净资产收益率 ROE">
              <template #value>{{ selectedStock.roe.toFixed(1) }}%</template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="市净率 PB">
              <template #value>{{ selectedStock.pb.toFixed(1) }}</template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="风险等级">
              <template #value>
                <n-tag size="small" :type="getRiskTagType(selectedStock.risk)">
                  {{ selectedStock.risk < 30 ? '低' : selectedStock.risk < 50 ? '中' : '高' }}
                </n-tag>
              </template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="推荐置信度">
              <template #value>{{ selectedStock.confidence.toFixed(0) }}%</template>
            </n-statistic>
          </n-grid-item>
        </n-grid>
        <div class="detail-reason" v-if="selectedStock.reason">
          <n-alert type="info" :bordered="false" size="small">
            {{ selectedStock.reason }}
          </n-alert>
        </div>
      </div>
    </div>
  </n-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, LineChart, HeatmapChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  VisualMapComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import { NCard, NButton, NButtonGroup, NSpace, NIcon, NDivider, NTag, NGrid, NGridItem, NStatistic, NAlert } from 'naive-ui'
import { BarChart as BarChartIcon } from '@vicons/ionicons5'

use([
  CanvasRenderer,
  BarChart,
  LineChart,
  HeatmapChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  VisualMapComponent
])

const props = defineProps({
  stocks: {
    type: Array,
    default: () => [
      { name: '贵州茅台', code: 'sh600519', return: 25, risk: 35, pe: 35, pb: 12, roe: 28, confidence: 85, recommend: '强烈推荐', sector: '白酒', reason: '业绩稳定，现金流优秀，品牌护城河深厚' },
      { name: '宁德时代', code: 'sz300750', return: 40, risk: 65, pe: 45, pb: 15, roe: 18, confidence: 75, recommend: '推荐', sector: '新能源', reason: '新能源电池龙头，技术领先全球' },
      { name: '腾讯控股', code: 'hk00700', return: 18, risk: 30, pe: 20, pb: 5, roe: 22, confidence: 90, recommend: '强烈推荐', sector: '互联网', reason: '社交生态稳固，游戏和云业务增长' },
      { name: '招商银行', code: 'sh600036', return: 12, risk: 20, pe: 10, pb: 1.5, roe: 16, confidence: 88, recommend: '推荐', sector: '银行', reason: '零售银行标杆，资产质量优良' },
      { name: '比亚迪', code: 'sz002594', return: 35, risk: 55, pe: 55, pb: 8, roe: 12, confidence: 78, recommend: '推荐', sector: '汽车', reason: '新能源汽车销量持续领先' }
    ]
  },
  darkTheme: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['stock-click'])

const chartRef = ref(null)
const compareMode = ref('bar')
const currentMetric = ref('return')
const selectedStock = ref(null)

const metrics = [
  { key: 'return', label: '预期收益' },
  { key: 'pe', label: '市盈率 PE' },
  { key: 'pb', label: '市净率 PB' },
  { key: 'roe', label: '净资产收益率 ROE' },
  { key: 'risk', label: '风险等级' },
  { key: 'confidence', label: '推荐置信度' }
]

const stockColors = ['#d03050', '#18a058', '#2080f0', '#f0a020', '#722ed1']

// 获取股票标签类型
function getStockTagType(stock) {
  if (stock.recommend === '强烈推荐') return 'success'
  if (stock.recommend === '推荐') return 'info'
  return 'default'
}

// 获取风险标签类型
function getRiskTagType(risk) {
  if (risk < 30) return 'success'
  if (risk < 50) return 'warning'
  return 'error'
}

// 选中股票
function selectStock(stock) {
  selectedStock.value = selectedStock.value?.code === stock.code ? null : stock
  if (selectedStock.value) {
    emit('stock-click', selectedStock.value)
  }
}

// 图表配置
const chartOption = computed(() => {
  const isDark = props.darkTheme
  const textColor = isDark ? '#ccc' : '#333'
  const bgColor = isDark ? '#1a1a1a' : '#fff'
  const gridColor = isDark ? '#333' : '#eee'

  const stockNames = props.stocks.map(s => s.name)
  const metricLabels = metrics.map(m => m.label)

  if (compareMode.value === 'bar') {
    const metric = currentMetric.value
    const values = props.stocks.map(s => s[metric])
    const label = metrics.find(m => m.key === metric)?.label || metric

    return {
      backgroundColor: bgColor,
      tooltip: {
        trigger: 'axis',
        backgroundColor: isDark ? '#2a2a2a' : '#fff',
        borderColor: isDark ? '#444' : '#ddd',
        textStyle: { color: textColor },
        formatter: (params) => {
          const stock = props.stocks[params[0].dataIndex]
          return `
            <div style="font-weight: 600; margin-bottom: 4px;">${stock.name}</div>
            <div>${label}: ${params[0].value.toFixed(metric === 'return' || metric === 'roe' || metric === 'risk' || metric === 'confidence' ? 1 : 1)}${metric === 'return' || metric === 'roe' || metric === 'risk' || metric === 'confidence' ? '%' : ''}</div>
          `
        }
      },
      grid: {
        left: '10%',
        right: '10%',
        top: 40,
        bottom: 60
      },
      xAxis: {
        type: 'category',
        data: stockNames,
        axisLabel: {
          color: textColor,
          fontSize: 11,
          interval: 0,
          rotate: 15
        },
        axisLine: { lineStyle: { color: gridColor } },
        axisTick: { lineStyle: { color: gridColor } }
      },
      yAxis: {
        type: 'value',
        name: label,
        nameLocation: 'middle',
        nameGap: 40,
        axisLabel: {
          color: textColor,
          formatter: '{value}' + (metric === 'return' || metric === 'roe' || metric === 'risk' || metric === 'confidence' ? '%' : '')
        },
        axisLine: { lineStyle: { color: gridColor } },
        axisTick: { lineStyle: { color: gridColor } },
        splitLine: { lineStyle: { color: gridColor, type: 'dashed' } }
      },
      series: [
        {
          name: label,
          type: 'bar',
          data: values.map((val, idx) => ({
            value: val,
            itemStyle: {
              color: stockColors[idx % stockColors.length]
            }
          })),
          barWidth: '60%',
          itemStyle: {
            borderRadius: [4, 4, 0, 0]
          },
          label: {
            show: true,
            position: 'top',
            color: textColor,
            fontSize: 11,
            formatter: '{c}' + (metric === 'return' || metric === 'roe' || metric === 'risk' || metric === 'confidence' ? '%' : '')
          }
        }
      ]
    }
  } else if (compareMode.value === 'line') {
    return {
      backgroundColor: bgColor,
      tooltip: {
        trigger: 'axis',
        backgroundColor: isDark ? '#2a2a2a' : '#fff',
        borderColor: isDark ? '#444' : '#ddd',
        textStyle: { color: textColor }
      },
      legend: {
        data: stockNames,
        top: 0,
        textStyle: { color: textColor }
      },
      grid: {
        left: '10%',
        right: '15%',
        top: 40,
        bottom: 40
      },
      xAxis: {
        type: 'category',
        data: metricLabels,
        axisLabel: { color: textColor, fontSize: 11 },
        axisLine: { lineStyle: { color: gridColor } },
        axisTick: { lineStyle: { color: gridColor } }
      },
      yAxis: {
        type: 'value',
        axisLabel: { color: textColor },
        axisLine: { lineStyle: { color: gridColor } },
        axisTick: { lineStyle: { color: gridColor } },
        splitLine: { lineStyle: { color: gridColor, type: 'dashed' } }
      },
      series: props.stocks.map((stock, idx) => ({
        name: stock.name,
        type: 'line',
        data: [stock.return, stock.pe, stock.pb, stock.roe, stock.risk, stock.confidence],
        symbol: 'circle',
        symbolSize: 8,
        lineStyle: {
          color: stockColors[idx % stockColors.length],
          width: 2
        },
        itemStyle: {
          color: stockColors[idx % stockColors.length]
        },
        smooth: true
      }))
    }
  } else {
    // 热力图模式
    const heatData = []
    props.stocks.forEach((stock, i) => {
      metrics.forEach((metric, j) => {
        heatData.push([j, i, stock[metric.key]])
      })
    })

    return {
      backgroundColor: bgColor,
      tooltip: {
        position: 'top',
        backgroundColor: isDark ? '#2a2a2a' : '#fff',
        borderColor: isDark ? '#444' : '#ddd',
        textStyle: { color: textColor },
        formatter: (params) => {
          const stock = props.stocks[params.value[1]]
          const metric = metrics[params.value[0]]
          return `
            <div style="font-weight: 600; margin-bottom: 4px;">${stock.name}</div>
            <div>${metric.label}: ${params.value[2].toFixed(1)}${metric.key === 'return' || metric.key === 'roe' || metric.key === 'risk' || metric.key === 'confidence' ? '%' : ''}</div>
          `
        }
      },
      grid: {
        left: '10%',
        right: '15%',
        top: 40,
        bottom: 60
      },
      xAxis: {
        type: 'category',
        data: metricLabels,
        splitArea: { show: true },
        axisLabel: {
          color: textColor,
          fontSize: 10,
          interval: 0,
          rotate: 45
        }
      },
      yAxis: {
        type: 'category',
        data: stockNames,
        splitArea: { show: true },
        axisLabel: { color: textColor, fontSize: 11 }
      },
      visualMap: {
        min: 0,
        max: 100,
        calculable: true,
        orient: 'vertical',
        right: '5%',
        top: 'center',
        textStyle: { color: textColor },
        inRange: {
          color: ['#50a3ba', '#eac736', '#d94e5d']
        }
      },
      series: [
        {
          name: '对比热力图',
          type: 'heatmap',
          data: heatData,
          label: {
            show: true,
            color: textColor,
            fontSize: 10,
            formatter: (params) => params.value[2].toFixed(0)
          },
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowColor: 'rgba(0, 0, 0, 0.5)'
            }
          }
        }
      ]
    }
  }
})

onMounted(() => {
  if (props.stocks.length > 0) {
    selectedStock.value = props.stocks[0]
  }
})
</script>

<style scoped>
.multi-stock-card {
  transition: all 0.3s ease;
}

.multi-stock-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(32, 128, 240, 0.15);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
}

.comparison-chart-wrapper {
  margin: 12px 0;
}

.comparison-metrics {
  padding: 12px 0;
  border-top: 1px solid #eee;
}

.metric-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.metric-tab {
  padding: 6px 12px;
  font-size: 12px;
  border-radius: 20px;
  background: #f0f0f0;
  color: #666;
  cursor: pointer;
  transition: all 0.2s ease;
}

.metric-tab:hover {
  background: #e0e0e0;
}

.metric-tab.active {
  background: linear-gradient(135deg, #2080f0, #18a058);
  color: white;
}

.stock-summary {
  padding-top: 8px;
}

.summary-grid {
  display: grid;
  gap: 8px;
}

.stock-summary-item {
  padding: 10px 12px;
  background: #f8f9fa;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.stock-summary-item:hover {
  background: #e9ecef;
  transform: translateX(4px);
}

.stock-summary-item.highlighted {
  background: rgba(32, 128, 240, 0.1);
  border: 1px solid #2080f0;
}

.stock-name {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 4px;
}

.stock-code {
  font-size: 12px;
  color: #999;
  font-family: 'Consolas', 'Monaco', monospace;
  margin-bottom: 6px;
}

.stock-metrics {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}

.metric-item {
  font-weight: 500;
}

.metric-item.positive {
  color: #d03050;
}

.metric-item.negative {
  color: #18a058;
}

.metric-separator {
  color: #ddd;
}

.selected-detail {
  padding-top: 4px;
}

.selected-detail-content {
  padding: 8px 0;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.detail-name {
  font-size: 16px;
  font-weight: 600;
}

.detail-code {
  font-size: 13px;
  color: #999;
  font-family: 'Consolas', 'Monaco', monospace;
}

.detail-reason {
  margin-top: 12px;
}

.text-positive {
  color: #d03050;
}

.text-negative {
  color: #18a058;
}

:deep(.dark) .card-title,
:deep(.dark) .stock-name,
:deep(.dark) .detail-name {
  color: #eee;
}

:deep(.dark) .stock-code,
:deep(.dark) .detail-code,
:deep(.dark) .metric-separator {
  color: #999;
}

:deep(.dark) .comparison-metrics,
:deep(.dark) .stock-summary {
  border-top-color: #333;
}

:deep(.dark) .metric-tab {
  background: #2a2a2a;
  color: #ccc;
}

:deep(.dark) .metric-tab:hover {
  background: #333;
}

:deep(.dark) .stock-summary-item {
  background: #2a2a2a;
}

:deep(.dark) .stock-summary-item:hover {
  background: #333;
}

:deep(.dark) .stock-summary-item.highlighted {
  background: rgba(32, 128, 240, 0.2);
}
</style>
