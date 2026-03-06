<template>
  <n-card 
    class="tech-indicator-card" 
    hoverable
    :class="{ 'card-expanded': expanded }"
  >
    <template #header>
      <div class="indicator-header" @click="expanded = !expanded">
        <div class="header-left">
          <n-icon :size="20" :color="iconColor">
            <component :is="indicatorIcon" />
          </n-icon>
          <span class="indicator-name">{{ name }}</span>
        </div>
        <div class="header-right">
          <n-tag :type="statusType" size="small">
            {{ statusText }}
          </n-tag>
          <n-icon :size="16" class="expand-icon" :style="{ transform: expanded ? 'rotate(180deg)' : '' }">
            <ChevronDownIcon />
          </n-icon>
        </div>
      </div>
    </template>

    <div class="indicator-chart">
      <v-chart 
        ref="chartRef"
        :option="chartOption" 
        :autoresize="true"
        style="height: 120px;" 
      />
    </div>

    <div class="indicator-values">
      <div class="value-item" v-for="(value, key) in displayValues" :key="key">
        <span class="value-label">{{ key }}:</span>
        <span class="value-number" :class="{ 'positive': value > 0, 'negative': value < 0 }">
          {{ formatValue(value, key) }}
        </span>
      </div>
    </div>

    <div class="indicator-detail" v-if="expanded">
      <n-divider style="margin: 12px 0;" />
      <div class="detail-content">
        <div class="detail-section">
          <h4>当前状态</h4>
          <p>{{ analysisText }}</p>
        </div>
        <div class="detail-section">
          <h4>操作建议</h4>
          <n-alert :type="suggestionType" :bordered="false" size="small">
            {{ suggestionText }}
          </n-alert>
        </div>
      </div>
    </div>
  </n-card>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import { NCard, NTag, NIcon, NDivider, NAlert } from 'naive-ui'
import {
  TrendingUp as TrendingUpIcon,
  TrendingDown as TrendingDownIcon,
  BarChart as BarChartIcon,
  ChevronDown as ChevronDownIcon
} from '@vicons/ionicons5'

use([
  CanvasRenderer,
  LineChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

const props = defineProps({
  name: {
    type: String,
    required: true
  },
  type: {
    type: String,
    default: 'MACD', // MACD, KDJ, RSI, BOLL
    validator: (v) => ['MACD', 'KDJ', 'RSI', 'BOLL'].includes(v)
  },
  data: {
    type: Array,
    default: () => []
  },
  currentValue: {
    type: Object,
    default: () => ({})
  },
  darkTheme: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['click'])

const expanded = ref(false)
const chartRef = ref(null)

// 图标映射
const indicatorIcon = computed(() => {
  const iconMap = {
    'MACD': BarChartIcon,
    'KDJ': TrendingUpIcon,
    'RSI': TrendingDownIcon,
    'BOLL': TrendingUpIcon
  }
  return iconMap[props.type] || BarChartIcon
})

// 图标颜色
const iconColor = computed(() => {
  const status = statusType.value
  const colorMap = {
    'success': '#18a058',
    'warning': '#f0a020',
    'error': '#d03050',
    'default': '#666'
  }
  return colorMap[status] || '#666'
})

// 状态类型
const statusType = computed(() => {
  // 根据指标类型和当前值判断状态
  if (!props.currentValue || Object.keys(props.currentValue).length === 0) {
    return 'default'
  }

  switch (props.type) {
    case 'MACD':
      if (props.currentValue.macd > 0 && props.currentValue.macd > props.currentValue.signal) {
        return 'success'
      } else if (props.currentValue.macd < 0 && props.currentValue.macd < props.currentValue.signal) {
        return 'error'
      }
      return 'warning'
    case 'KDJ':
      if (props.currentValue.k < 20 && props.currentValue.j < 20) {
        return 'success'
      } else if (props.currentValue.k > 80 && props.currentValue.j > 80) {
        return 'error'
      }
      return 'warning'
    case 'RSI':
      if (props.currentValue.rsi < 30) {
        return 'success'
      } else if (props.currentValue.rsi > 70) {
        return 'error'
      }
      return 'warning'
    case 'BOLL':
      const close = props.currentValue.close || 0
      const upper = props.currentValue.upper || 0
      const lower = props.currentValue.lower || 0
      if (close > upper) {
        return 'error'
      } else if (close < lower) {
        return 'success'
      }
      return 'warning'
    default:
      return 'default'
  }
})

// 状态文本
const statusText = computed(() => {
  const typeMap = {
    'success': '看多',
    'warning': '观望',
    'error': '看空',
    'default': '中性'
  }
  return typeMap[statusType.value] || '中性'
})

// 显示数值
const displayValues = computed(() => {
  if (!props.currentValue) return {}
  
  switch (props.type) {
    case 'MACD':
      return {
        'MACD': props.currentValue.macd || 0,
        'DIF': props.currentValue.dif || 0,
        'DEA': props.currentValue.signal || 0
      }
    case 'KDJ':
      return {
        'K值': props.currentValue.k || 0,
        'D值': props.currentValue.d || 0,
        'J值': props.currentValue.j || 0
      }
    case 'RSI':
      return {
        'RSI': props.currentValue.rsi || 0
      }
    case 'BOLL':
      return {
        '上轨': props.currentValue.upper || 0,
        '中轨': props.currentValue.mid || 0,
        '下轨': props.currentValue.lower || 0
      }
    default:
      return {}
  }
})

// 分析文本
const analysisText = computed(() => {
  switch (props.type) {
    case 'MACD':
      if (statusType.value === 'success') {
        return 'MACD金叉，红柱放大，多头力量较强，可考虑持有或加仓'
      } else if (statusType.value === 'error') {
        return 'MACD死叉，绿柱放大，空头力量较强，注意风险控制'
      }
      return 'MACD指标中性，等待明确信号'
    case 'KDJ':
      if (statusType.value === 'success') {
        return 'KDJ超卖区域，存在反弹可能，可关注'
      } else if (statusType.value === 'error') {
        return 'KDJ超买区域，存在回调风险，注意止盈'
      }
      return 'KDJ指标中性，等待明确信号'
    case 'RSI':
      if (statusType.value === 'success') {
        return 'RSI超卖，可能存在超跌反弹机会'
      } else if (statusType.value === 'error') {
        return 'RSI超买，需要警惕回调风险'
      }
      return 'RSI处于合理区间'
    case 'BOLL':
      if (statusType.value === 'success') {
        return '价格跌破下轨，存在超卖反弹可能'
      } else if (statusType.value === 'error') {
        return '价格突破上轨，注意回调风险'
      }
      return '价格在布林带内运行，趋势稳定'
    default:
      return '等待信号'
  }
})

// 建议类型
const suggestionType = computed(() => {
  return statusType.value
})

// 建议文本
const suggestionText = computed(() => {
  if (statusType.value === 'success') {
    return '可考虑逢低买入或持有，设置止损保护利润'
  } else if (statusType.value === 'error') {
    return '建议观望或适当减仓，等待更明确信号'
  }
  return '建议继续观察，等待明确信号后再操作'
})

// 格式化数值
function formatValue(value, key) {
  if (key === 'RSI' || key.includes('值')) {
    return value.toFixed(2)
  }
  if (value > 0) {
    return '+' + value.toFixed(4)
  }
  return value.toFixed(4)
}

// 图表配置
const chartOption = computed(() => {
  const isDark = props.darkTheme
  const textColor = isDark ? '#ccc' : '#333'
  const gridColor = isDark ? '#333' : '#eee'

  let baseOption = {
    grid: {
      top: 10,
      right: 10,
      bottom: 20,
      left: 40
    },
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark ? '#1a1a1a' : '#fff',
      borderColor: isDark ? '#333' : '#ddd',
      textStyle: {
        color: textColor
      }
    },
    xAxis: {
      type: 'category',
      data: props.data.map(d => d.date || d.time || ''),
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: {
        show: true,
        fontSize: 10,
        color: textColor,
        formatter: (value) => {
          if (value && value.length > 5) {
            return value.slice(-5)
          }
          return value
        }
      }
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: {
        lineStyle: {
          color: gridColor,
          type: 'dashed'
        }
      },
      axisLabel: {
        fontSize: 10,
        color: textColor
      }
    }
  }

  // 根据指标类型配置不同的图表
  switch (props.type) {
    case 'MACD':
      return {
        ...baseOption,
        legend: {
          show: true,
          top: 0,
          right: 10,
          textStyle: {
            fontSize: 10,
            color: textColor
          }
        },
        series: [
          {
            name: 'MACD',
            type: 'bar',
            data: props.data.map(d => d.macd || 0),
            itemStyle: {
              color: (params) => {
                return params.value >= 0 ? '#d03050' : '#18a058'
              }
            }
          },
          {
            name: 'DIF',
            type: 'line',
            data: props.data.map(d => d.dif || 0),
            smooth: true,
            symbol: 'none',
            lineStyle: {
              color: '#2080f0',
              width: 2
            }
          },
          {
            name: 'DEA',
            type: 'line',
            data: props.data.map(d => d.signal || 0),
            smooth: true,
            symbol: 'none',
            lineStyle: {
              color: '#f0a020',
              width: 2
            }
          }
        ]
      }

    case 'KDJ':
      return {
        ...baseOption,
        legend: {
          show: true,
          top: 0,
          right: 10,
          textStyle: {
            fontSize: 10,
            color: textColor
          }
        },
        series: [
          {
            name: 'K',
            type: 'line',
            data: props.data.map(d => d.k || 0),
            smooth: true,
            symbol: 'none',
            lineStyle: {
              color: '#2080f0',
              width: 2
            }
          },
          {
            name: 'D',
            type: 'line',
            data: props.data.map(d => d.d || 0),
            smooth: true,
            symbol: 'none',
            lineStyle: {
              color: '#f0a020',
              width: 2
            }
          },
          {
            name: 'J',
            type: 'line',
            data: props.data.map(d => d.j || 0),
            smooth: true,
            symbol: 'none',
            lineStyle: {
              color: '#d03050',
              width: 2
            }
          }
        ]
      }

    case 'RSI':
      return {
        ...baseOption,
        yAxis: {
          ...baseOption.yAxis,
          min: 0,
          max: 100,
          splitLine: {
            ...baseOption.yAxis.splitLine,
            show: true
          }
        },
        series: [
          {
            name: 'RSI',
            type: 'line',
            data: props.data.map(d => d.rsi || 0),
            smooth: true,
            symbol: 'circle',
            symbolSize: 4,
            lineStyle: {
              color: '#2080f0',
              width: 2
            },
            areaStyle: {
              color: {
                type: 'linear',
                x: 0,
                y: 0,
                x2: 0,
                y2: 1,
                colorStops: [
                  { offset: 0, color: 'rgba(32, 128, 240, 0.3)' },
                  { offset: 1, color: 'rgba(32, 128, 240, 0.05)' }
                ]
              }
            },
            markLine: {
              data: [
                { yAxis: 70, lineStyle: { color: '#d03050', type: 'dashed' }, label: { formatter: '超买', position: 'end' } },
                { yAxis: 30, lineStyle: { color: '#18a058', type: 'dashed' }, label: { formatter: '超卖', position: 'end' } }
              ]
            }
          }
        ]
      }

    case 'BOLL':
      return {
        ...baseOption,
        legend: {
          show: true,
          top: 0,
          right: 10,
          textStyle: {
            fontSize: 10,
            color: textColor
          }
        },
        series: [
          {
            name: '上轨',
            type: 'line',
            data: props.data.map(d => d.upper || 0),
            smooth: true,
            symbol: 'none',
            lineStyle: {
              color: '#d03050',
              width: 1,
              type: 'dashed'
            }
          },
          {
            name: '中轨',
            type: 'line',
            data: props.data.map(d => d.mid || 0),
            smooth: true,
            symbol: 'none',
            lineStyle: {
              color: '#f0a020',
              width: 1
            }
          },
          {
            name: '下轨',
            type: 'line',
            data: props.data.map(d => d.lower || 0),
            smooth: true,
            symbol: 'none',
            lineStyle: {
              color: '#18a058',
              width: 1,
              type: 'dashed'
            }
          }
        ]
      }

    default:
      return baseOption
  }
})

onMounted(() => {
  // 初始化图表
})
</script>

<style scoped>
.tech-indicator-card {
  transition: all 0.3s ease;
}

.tech-indicator-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-expanded {
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.15);
}

.indicator-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  user-select: none;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.indicator-name {
  font-size: 15px;
  font-weight: 600;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.expand-icon {
  transition: transform 0.3s ease;
  color: #999;
}

.indicator-chart {
  margin: 8px 0;
}

.indicator-values {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  padding-top: 8px;
}

.value-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.value-label {
  font-size: 13px;
  color: #666;
}

.value-number {
  font-size: 14px;
  font-weight: 600;
  font-family: 'Consolas', 'Monaco', monospace;
}

.value-number.positive {
  color: #d03050;
}

.value-number.negative {
  color: #18a058;
}

.detail-content {
  padding-top: 8px;
}

.detail-section {
  margin-bottom: 12px;
}

.detail-section h4 {
  font-size: 14px;
  margin-bottom: 8px;
  color: #333;
}

.detail-section p {
  font-size: 13px;
  line-height: 1.6;
  color: #666;
  margin: 0;
}

:deep(.dark) .indicator-name,
:deep(.dark) .value-label,
:deep(.dark) .detail-section h4 {
  color: #eee;
}

:deep(.dark) .value-number,
:deep(.dark) .detail-section p {
  color: #ccc;
}
</style>
