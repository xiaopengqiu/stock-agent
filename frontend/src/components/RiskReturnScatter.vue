<template>
  <n-card class="risk-return-card" hoverable>
    <template #header>
      <div class="card-header">
        <div class="header-left">
          <n-icon :size="22" color="#d4af37">
            <TrendingUpIcon />
          </n-icon>
          <span class="card-title">风险收益散点图</span>
        </div>
        <n-tag type="warning" size="small">
          找到 {{ stocks.length }} 只股票
        </n-tag>
      </div>
    </template>

    <div class="scatter-chart-wrapper">
      <!-- 象限标识 -->
      <div class="quadrant-guide">
        <div class="quadrant q1">高收益低风险✨</div>
        <div class="quadrant q2">高收益高风险⚡</div>
        <div class="quadrant q3">低收益低风险🛡️</div>
        <div class="quadrant q4">低收益高风险⚠️</div>
      </div>
      
      <v-chart 
        ref="chartRef"
        :option="chartOption" 
        :autoresize="true"
        style="height: 380px;" 
        @click="handleChartClick"
      />
    </div>

    <div class="stock-summary">
      <div class="summary-stat">
        <span class="stat-label">黄金股票</span>
        <span class="stat-value" style="color: #18a058;">{{ goldenStocks.length }}</span>
      </div>
      <div class="summary-stat">
        <span class="stat-label">激进选择</span>
        <span class="stat-value" style="color: #2080f0;">{{ aggressiveStocks.length }}</span>
      </div>
      <div class="summary-stat">
        <span class="stat-label">稳健选择</span>
        <span class="stat-value" style="color: #f0a020;">{{ safeStocks.length }}</span>
      </div>
      <div class="summary-stat">
        <span class="stat-label">谨慎观望</span>
        <span class="stat-value" style="color: #d03050;">{{ riskyStocks.length }}</span>
      </div>
    </div>

    <div class="stock-list" v-if="selectedStock">
      <n-divider style="margin: 12px 0;" />
      <div class="selected-stock-detail">
        <div class="stock-header">
          <n-tag :type="getStockTagType(selectedStock)" size="small">
            {{ getStockQuadrant(selectedStock) }}
          </n-tag>
          <span class="stock-name">{{ selectedStock.name }}</span>
          <span class="stock-code">{{ selectedStock.code }}</span>
        </div>
        <div class="stock-stats">
          <div class="stat-item">
            <span class="stat-item-label">预期收益</span>
            <span class="stat-item-value" :class="{ 'positive': selectedStock.return >= 0, 'negative': selectedStock.return < 0 }">
              {{ selectedStock.return >= 0 ? '+' : '' }}{{ selectedStock.return.toFixed(1) }}%
            </span>
          </div>
          <div class="stat-item">
            <span class="stat-item-label">风险等级</span>
            <span class="stat-item-value">{{ selectedStock.risk.toFixed(1) }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-item-label">推荐置信度</span>
            <span class="stat-item-value">{{ selectedStock.confidence.toFixed(0) }}%</span>
          </div>
          <div class="stat-item" v-if="selectedStock.sector">
            <span class="stat-item-label">所属板块</span>
            <span class="stat-item-value">{{ selectedStock.sector }}</span>
          </div>
        </div>
        <div class="stock-reason" v-if="selectedStock.reason">
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
import { ScatterChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  VisualMapComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import { NCard, NTag, NIcon, NDivider, NAlert } from 'naive-ui'
import { TrendingUp as TrendingUpIcon } from '@vicons/ionicons5'

use([
  CanvasRenderer,
  ScatterChart,
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
      { name: '贵州茅台', code: 'sh600519', return: 25, risk: 35, confidence: 85, sector: '白酒', reason: '业绩稳定，现金流优秀' },
      { name: '宁德时代', code: 'sz300750', return: 40, risk: 65, confidence: 75, sector: '新能源', reason: '新能源龙头，技术领先' },
      { name: '腾讯控股', code: 'hk00700', return: 18, risk: 30, confidence: 90, sector: '互联网', reason: '社交生态稳固，估值合理' },
      { name: '招商银行', code: 'sh600036', return: 12, risk: 20, confidence: 88, sector: '银行', reason: '零售银行标杆，资产质量好' },
      { name: '比亚迪', code: 'sz002594', return: 35, risk: 55, confidence: 78, sector: '汽车', reason: '新能源汽车销量领先' },
      { name: '药明康德', code: 'sh603259', return: 22, risk: 45, confidence: 72, sector: '医药', reason: 'CRO龙头，全球竞争力' },
      { name: '海康威视', code: 'sz002415', return: 15, risk: 40, confidence: 68, sector: '安防', reason: '安防龙头，AI+安防' },
      { name: '中芯国际', code: 'sh688981', return: 28, risk: 70, confidence: 65, sector: '芯片', reason: '国产替代，自主可控' }
    ]
  },
  darkTheme: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['stock-click'])

const chartRef = ref(null)
const selectedStock = ref(null)

// 按象限分类股票
const goldenStocks = computed(() => {
  return props.stocks.filter(s => s.return >= 20 && s.risk <= 40)
})

const aggressiveStocks = computed(() => {
  return props.stocks.filter(s => s.return >= 20 && s.risk > 40)
})

const safeStocks = computed(() => {
  return props.stocks.filter(s => s.return < 20 && s.risk <= 40)
})

const riskyStocks = computed(() => {
  return props.stocks.filter(s => s.return < 20 && s.risk > 40)
})

// 获取象限标签
function getStockQuadrant(stock) {
  if (stock.return >= 20 && stock.risk <= 40) return '黄金股票'
  if (stock.return >= 20 && stock.risk > 40) return '激进选择'
  if (stock.return < 20 && stock.risk <= 40) return '稳健选择'
  return '谨慎观望'
}

// 获取标签类型
function getStockTagType(stock) {
  if (stock.return >= 20 && stock.risk <= 40) return 'success'
  if (stock.return >= 20 && stock.risk > 40) return 'info'
  if (stock.return < 20 && stock.risk <= 40) return 'warning'
  return 'error'
}

// 获取板块颜色
function getSectorColor(sector) {
  const colorMap = {
    '白酒': '#d03050',
    '新能源': '#18a058',
    '互联网': '#2080f0',
    '银行': '#f0a020',
    '汽车': '#d4af37',
    '医药': '#722ed1',
    '安防': '#13c2c2',
    '芯片': '#eb2f96'
  }
  return colorMap[sector] || '#666'
}

// 图表配置
const chartOption = computed(() => {
  const isDark = props.darkTheme
  const textColor = isDark ? '#ccc' : '#333'
  const bgColor = isDark ? '#1a1a1a' : '#fff'
  const gridColor = isDark ? '#333' : '#eee'

  return {
    backgroundColor: bgColor,
    tooltip: {
      trigger: 'item',
      backgroundColor: isDark ? '#2a2a2a' : 'rgba(255, 255, 255, 0.95)',
      borderColor: isDark ? '#444' : '#ddd',
      textStyle: {
        color: textColor
      },
      formatter: (params) => {
        const stock = props.stocks[params.dataIndex]
        if (!stock) return ''
        
        const quadrant = getStockQuadrant(stock)
        const tagType = getStockTagType(stock)
        const typeColor = {
          'success': '#18a058',
          'info': '#2080f0',
          'warning': '#f0a020',
          'error': '#d03050'
        }
        
        return `
          <div style="padding: 4px 0;">
            <div style="font-weight: 600; font-size: 14px; margin-bottom: 8px; border-bottom: 1px solid ${gridColor}; padding-bottom: 4px;">
              ${stock.name} <span style="color: #999; font-weight: 400;">${stock.code}</span>
            </div>
            <div style="display: inline-block; padding: 2px 8px; border-radius: 4px; background: ${typeColor[tagType]}20; color: ${typeColor[tagType]}; font-size: 12px; margin-bottom: 8px;">
              ${quadrant}
            </div>
            <div style="display: grid; grid-template-columns: auto auto; gap: 4px 16px; font-size: 13px;">
              <span>预期收益:</span><span style="font-weight: 600; color: ${stock.return >= 0 ? '#d03050' : '#18a058'}">${stock.return >= 0 ? '+' : ''}${stock.return.toFixed(1)}%</span>
              <span>风险等级:</span><span style="font-weight: 600;">${stock.risk.toFixed(1)}</span>
              <span>推荐置信度:</span><span style="font-weight: 600;">${stock.confidence.toFixed(0)}%</span>
              ${stock.sector ? `<span>所属板块:</span><span style="font-weight: 600; color: ${getSectorColor(stock.sector)};">${stock.sector}</span>` : ''}
            </div>
            ${stock.reason ? `
              <div style="margin-top: 8px; padding-top: 8px; border-top: 1px solid ${gridColor};">
                <div style="color: #999; font-size: 12px; margin-bottom: 4px;">推荐理由</div>
                <div style="font-size: 13px; color: ${textColor};">${stock.reason}</div>
              </div>
            ` : ''}
          </div>
        `
      }
    },
    legend: {
      show: true,
      top: 10,
      right: 10,
      textStyle: {
        color: textColor
      }
    },
    grid: {
      left: '10%',
      right: '15%',
      top: 60,
      bottom: 60
    },
    xAxis: {
      type: 'value',
      name: '预期收益 (%)',
      nameLocation: 'middle',
      nameGap: 35,
      min: 0,
      max: 50,
      splitLine: {
        lineStyle: {
          color: gridColor,
          type: 'dashed'
        }
      },
      axisLine: {
        lineStyle: {
          color: gridColor
        }
      },
      axisLabel: {
        color: textColor,
        formatter: '{value}%'
      },
      axisPointer: {
        lineStyle: {
          color: '#2080f0',
          type: 'dashed'
        }
      }
    },
    yAxis: {
      type: 'value',
      name: '风险等级',
      nameLocation: 'middle',
      nameGap: 45,
      min: 0,
      max: 80,
      splitLine: {
        lineStyle: {
          color: gridColor,
          type: 'dashed'
        }
      },
      axisLine: {
        lineStyle: {
          color: gridColor
        }
      },
      axisLabel: {
        color: textColor
      },
      axisPointer: {
        lineStyle: {
          color: '#2080f0',
          type: 'dashed'
        }
      }
    },
    visualMap: [
      {
        show: true,
        right: 10,
        top: 'center',
        dimension: 2,
        min: 50,
        max: 100,
        text: ['置信度高', '置信度低'],
        textStyle: {
          color: textColor
        },
        inRange: {
          symbolSize: [15, 40],
          color: [
            '#d03050',
            '#eb2f96',
            '#722ed1',
            '#2080f0',
            '#13c2c2',
            '#18a058'
          ]
        }
      }
    ],
    series: [
      {
        name: '推荐股票',
        type: 'scatter',
        data: props.stocks.map(stock => ([
          stock.return,
          stock.risk,
          stock.confidence,
          stock.sector
        ])),
        symbolSize: (data) => {
          const confidence = data[2]
          return 15 + (confidence - 50) * 0.5
        },
        itemStyle: {
          opacity: 0.8
        },
        markLine: {
          silent: true,
          lineStyle: {
            color: '#999',
            type: 'dashed'
          },
          label: {
            show: false
          },
          data: [
            { xAxis: 20 },
            { yAxis: 40 }
          ]
        },
        markArea: {
          silent: true,
          data: [
            [
              { name: '黄金区域', xAxis: 20, yAxis: 0, itemStyle: { color: 'rgba(24, 160, 88, 0.05)' } },
              { xAxis: 50, yAxis: 40 }
            ]
          ]
        }
      }
    ]
  }
})

// 处理图表点击
function handleChartClick(params) {
  if (params.componentType === 'series') {
    const stock = props.stocks[params.dataIndex]
    if (stock) {
      selectedStock.value = stock
      emit('stock-click', stock)
    }
  }
}

onMounted(() => {
  // 默认选中第一只黄金股票
  if (goldenStocks.value.length > 0) {
    selectedStock.value = goldenStocks.value[0]
  } else if (props.stocks.length > 0) {
    selectedStock.value = props.stocks[0]
  }
})
</script>

<style scoped>
.risk-return-card {
  transition: all 0.3s ease;
}

.risk-return-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(212, 175, 55, 0.15);
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

.scatter-chart-wrapper {
  position: relative;
  margin: 8px 0 16px;
}

.quadrant-guide {
  position: absolute;
  top: 60px;
  right: 10px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  z-index: 10;
  pointer-events: none;
}

.quadrant {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  white-space: nowrap;
}

.quadrant.q1 {
  background: rgba(24, 160, 88, 0.1);
  color: #18a058;
}

.quadrant.q2 {
  background: rgba(32, 128, 240, 0.1);
  color: #2080f0;
}

.quadrant.q3 {
  background: rgba(240, 160, 32, 0.1);
  color: #f0a020;
}

.quadrant.q4 {
  background: rgba(208, 48, 80, 0.1);
  color: #d03050;
}

.stock-summary {
  display: flex;
  justify-content: space-around;
  padding: 12px 0;
  border-top: 1px solid #eee;
}

.summary-stat {
  text-align: center;
}

.stat-label {
  display: block;
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
}

.stat-value {
  display: block;
  font-size: 18px;
  font-weight: 700;
  font-family: 'Consolas', 'Monaco', monospace;
}

.selected-stock-detail {
  padding: 8px 0;
}

.stock-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.stock-name {
  font-size: 16px;
  font-weight: 600;
}

.stock-code {
  font-size: 13px;
  color: #999;
  font-family: 'Consolas', 'Monaco', monospace;
}

.stock-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 12px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 6px;
}

.stat-item-label {
  font-size: 13px;
  color: #666;
}

.stat-item-value {
  font-size: 14px;
  font-weight: 600;
  font-family: 'Consolas', 'Monaco', monospace;
}

.stat-item-value.positive {
  color: #d03050;
}

.stat-item-value.negative {
  color: #18a058;
}

.stock-reason {
  margin-top: 8px;
}

:deep(.dark) .card-title,
:deep(.dark) .stock-name,
:deep(.dark) .stat-item-label {
  color: #eee;
}

:deep(.dark) .stock-code,
:deep(.dark) .stat-label {
  color: #999;
}

:deep(.dark) .stock-summary {
  border-top-color: #333;
}

:deep(.dark) .stat-item {
  background: #2a2a2a;
}
</style>
