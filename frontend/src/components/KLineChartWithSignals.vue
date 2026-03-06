<template>
  <n-card class="kline-signals-card" hoverable>
    <template #header>
      <div class="card-header">
        <div class="header-left">
          <n-icon :size="22" color="#d03050">
            <TrendingUpIcon />
          </n-icon>
          <span class="card-title">{{ stockName }} K线图</span>
          <n-tag size="small" type="info" style="margin-left: 8px;">
            {{ stockCode }}
          </n-tag>
        </div>
        <div class="header-right">
          <n-space>
            <n-button-group size="small">
              <n-button 
                :type="timeRange === 30 ? 'primary' : 'default'"
                @click="timeRange = 30"
              >
                30天
              </n-button>
              <n-button 
                :type="timeRange === 60 ? 'primary' : 'default'"
                @click="timeRange = 60"
              >
                60天
              </n-button>
              <n-button 
                :type="timeRange === 120 ? 'primary' : 'default'"
                @click="timeRange = 120"
              >
                120天
              </n-button>
            </n-button-group>
          </n-space>
        </div>
      </div>
    </template>

    <div class="kline-chart-wrapper">
      <v-chart 
        ref="chartRef"
        :option="chartOption" 
        :autoresize="true"
        style="height: 400px;" 
      />
    </div>

    <div class="signals-legend">
      <div class="legend-item">
        <span class="legend-dot buy"></span>
        <span class="legend-text">AI建议买入</span>
      </div>
      <div class="legend-item">
        <span class="legend-dot sell"></span>
        <span class="legend-text">AI建议卖出</span>
      </div>
      <div class="legend-item">
        <span class="legend-dot hold"></span>
        <span class="legend-text">观望持有</span>
      </div>
    </div>
  </n-card>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { CandlestickChart, LineChart, BarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent,
  VisualMapComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import { NCard, NTag, NButton, NButtonGroup, NSpace, NIcon } from 'naive-ui'
import { TrendingUp as TrendingUpIcon } from '@vicons/ionicons5'

use([
  CanvasRenderer,
  CandlestickChart,
  LineChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent,
  VisualMapComponent
])

const props = defineProps({
  stockCode: {
    type: String,
    default: 'sh600519'
  },
  stockName: {
    type: String,
    default: '贵州茅台'
  },
  klineData: {
    type: Array,
    default: () => []
  },
  buySignals: {
    type: Array,
    default: () => []
  },
  sellSignals: {
    type: Array,
    default: () => []
  },
  holdSignals: {
    type: Array,
    default: () => []
  },
  darkTheme: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['signal-click', 'time-range-change'])

const chartRef = ref(null)
const timeRange = ref(60)

// 过滤数据根据时间范围
const filteredData = computed(() => {
  if (!props.klineData || props.klineData.length === 0) return []
  
  const data = [...props.klineData]
  const startIndex = Math.max(0, data.length - timeRange.value)
  return data.slice(startIndex)
})

// 过滤买卖信号
const filteredBuySignals = computed(() => {
  const dates = filteredData.value.map(d => d.date || d.time)
  return props.buySignals.filter(signal => dates.includes(signal.date || signal.time))
})

const filteredSellSignals = computed(() => {
  const dates = filteredData.value.map(d => d.date || d.time)
  return props.sellSignals.filter(signal => dates.includes(signal.date || signal.time))
})

const filteredHoldSignals = computed(() => {
  const dates = filteredData.value.map(d => d.date || d.time)
  return props.holdSignals.filter(signal => dates.includes(signal.date || signal.time))
})

// 图表配置
const chartOption = computed(() => {
  const isDark = props.darkTheme
  const textColor = isDark ? '#ccc' : '#333'
  const bgColor = isDark ? '#1a1a1a' : '#fff'
  const gridColor = isDark ? '#333' : '#eee'
  const upColor = '#d03050'  // 上涨红色
  const downColor = '#18a058' // 下跌绿色

  const dates = filteredData.value.map(d => d.date || d.time || '')
  const candleData = filteredData.value.map(d => [
    d.open || d.o || 0,
    d.close || d.c || 0,
    d.low || d.l || 0,
    d.high || d.h || 0
  ])
  const volumes = filteredData.value.map(d => d.volume || d.v || 0)

  // 准备买卖信号标记
  const buyMarkData = filteredBuySignals.value.map(signal => ({
    name: '买入信号',
    value: [signal.date || signal.time, signal.price || 0],
    itemStyle: { color: upColor }
  }))

  const sellMarkData = filteredSellSignals.value.map(signal => ({
    name: '卖出信号',
    value: [signal.date || signal.time, signal.price || 0],
    itemStyle: { color: downColor }
  }))

  const holdMarkData = filteredHoldSignals.value.map(signal => ({
    name: '观望信号',
    value: [signal.date || signal.time, signal.price || 0],
    itemStyle: { color: '#f0a020' }
  }))

  // 计算最小值和最大值用于信号显示
  const minPrice = Math.min(...filteredData.value.map(d => d.low || d.l || 999999))
  const maxPrice = Math.max(...filteredData.value.map(d => d.high || d.h || 0))
  const priceRange = maxPrice - minPrice

  return {
    backgroundColor: bgColor,
    animation: true,
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark ? '#2a2a2a' : 'rgba(255, 255, 255, 0.95)',
      borderColor: isDark ? '#444' : '#ddd',
      textStyle: {
        color: textColor
      },
      axisPointer: {
        type: 'cross',
        crossStyle: {
          color: '#999'
        }
      },
      formatter: (params) => {
        let result = ''
        const candleParam = params.find(p => p.seriesName === 'K线')
        if (candleParam) {
          const dataIndex = candleParam.dataIndex
          const data = filteredData.value[dataIndex]
          if (data) {
            result = `
              <div style="font-weight: 600; margin-bottom: 8px; border-bottom: 1px solid ${gridColor}; padding-bottom: 4px;">
                ${data.date || data.time || ''}
              </div>
              <div style="display: grid; grid-template-columns: auto auto; gap: 4px 12px;">
                <span>开盘:</span><span style="font-weight: 600;">${(data.open || data.o || 0).toFixed(2)}</span>
                <span>收盘:</span><span style="font-weight: 600; color: ${(data.close || data.c) >= (data.open || data.o) ? upColor : downColor};">${(data.close || data.c || 0).toFixed(2)}</span>
                <span>最高:</span><span style="font-weight: 600; color: ${upColor};">${(data.high || data.h || 0).toFixed(2)}</span>
                <span>最低:</span><span style="font-weight: 600; color: ${downColor};">${(data.low || data.l || 0).toFixed(2)}</span>
                <span>成交量:</span><span style="font-weight: 600;">${(data.volume || data.v || 0).toLocaleString()}</span>
              </div>
            `
          }
        }
        
        // 检查是否有信号
        const signalParams = params.filter(p => ['买入信号', '卖出信号', '观望信号'].includes(p.seriesName))
        if (signalParams.length > 0) {
          result += `<div style="margin-top: 8px; border-top: 1px solid ${gridColor}; padding-top: 4px;">`
          signalParams.forEach(sp => {
            const color = sp.seriesName === '买入信号' ? upColor : 
                         sp.seriesName === '卖出信号' ? downColor : '#f0a020'
            result += `<div style="color: ${color}; font-weight: 600;">${sp.seriesName}</div>`
          })
          result += '</div>'
        }
        
        return result || '暂无数据'
      }
    },
    legend: {
      data: ['K线', '成交量', '买入信号', '卖出信号', '观望信号'],
      top: 10,
      textStyle: {
        color: textColor
      }
    },
    grid: [
      {
        left: '10%',
        right: '8%',
        top: 60,
        height: '55%'
      },
      {
        left: '10%',
        right: '8%',
        top: '75%',
        height: '15%'
      }
    ],
    xAxis: [
      {
        type: 'category',
        data: dates,
        scale: true,
        boundaryGap: false,
        axisLine: { lineStyle: { color: gridColor } },
        axisLabel: { 
          color: textColor,
          fontSize: 10,
          formatter: (value) => {
            if (value && value.length > 5) {
              return value.slice(-5)
            }
            return value
          }
        },
        splitLine: { show: false },
        min: 'dataMin',
        max: 'dataMax'
      },
      {
        type: 'category',
        gridIndex: 1,
        data: dates,
        scale: true,
        boundaryGap: false,
        axisLine: { lineStyle: { color: gridColor } },
        axisLabel: { show: false },
        splitLine: { show: false },
        min: 'dataMin',
        max: 'dataMax'
      }
    ],
    yAxis: [
      {
        scale: true,
        splitArea: {
          show: true,
          areaStyle: {
            color: [
              'rgba(32, 128, 240, 0.01)',
              'rgba(32, 128, 240, 0.02)'
            ]
          }
        },
        axisLine: { lineStyle: { color: gridColor } },
        axisTick: { lineStyle: { color: gridColor } },
        axisLabel: { color: textColor, fontSize: 10 },
        splitLine: { lineStyle: { color: gridColor, type: 'dashed' } }
      },
      {
        scale: true,
        gridIndex: 1,
        splitNumber: 2,
        axisLine: { lineStyle: { color: gridColor } },
        axisTick: { lineStyle: { color: gridColor } },
        axisLabel: { show: false },
        splitLine: { lineStyle: { color: gridColor, type: 'dashed' } }
      }
    ],
    dataZoom: [
      {
        type: 'inside',
        xAxisIndex: [0, 1],
        start: 0,
        end: 100
      },
      {
        show: true,
        xAxisIndex: [0, 1],
        type: 'slider',
        bottom: 10,
        start: 0,
        end: 100,
        textStyle: {
          color: textColor
        },
        borderColor: gridColor,
        fillerColor: 'rgba(32, 128, 240, 0.15)',
        handleStyle: {
          color: '#2080f0'
        }
      }
    ],
    series: [
      {
        name: 'K线',
        type: 'candlestick',
        data: candleData,
        itemStyle: {
          color: upColor,
          color0: downColor,
          borderColor: upColor,
          borderColor0: downColor
        }
      },
      {
        name: '成交量',
        type: 'bar',
        xAxisIndex: 1,
        yAxisIndex: 1,
        data: volumes.map((vol, idx) => {
          const data = filteredData.value[idx]
          const isUp = (data?.close || data?.c || 0) >= (data?.open || data?.o || 0)
          return {
            value: vol,
            itemStyle: {
              color: isUp ? upColor : downColor,
              opacity: 0.7
            }
          }
        })
      },
      {
        name: '买入信号',
        type: 'scatter',
        data: buyMarkData.map(item => {
          const price = item.value[1]
          return {
            ...item,
            value: [item.value[0], price - priceRange * 0.05],
            symbol: 'pin',
            symbolSize: 30,
            symbolRotate: 180
          }
        }),
        itemStyle: {
          color: upColor
        }
      },
      {
        name: '卖出信号',
        type: 'scatter',
        data: sellMarkData.map(item => {
          const price = item.value[1]
          return {
            ...item,
            value: [item.value[0], price + priceRange * 0.05],
            symbol: 'pin',
            symbolSize: 30
          }
        }),
        itemStyle: {
          color: downColor
        }
      },
      {
        name: '观望信号',
        type: 'scatter',
        data: holdMarkData.map(item => {
          return {
            ...item,
            symbol: 'circle',
            symbolSize: 12
          }
        }),
        itemStyle: {
          color: '#f0a020'
        }
      }
    ]
  }
})

// 监听时间范围变化
watch(timeRange, (newVal) => {
  emit('time-range-change', newVal)
})

onMounted(() => {
  // 初始化完成
})
</script>

<style scoped>
.kline-signals-card {
  transition: all 0.3s ease;
}

.kline-signals-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  margin-left: 8px;
}

.kline-chart-wrapper {
  margin: 12px 0;
}

.signals-legend {
  display: flex;
  justify-content: center;
  gap: 24px;
  padding-top: 8px;
  border-top: 1px solid #eee;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.legend-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.legend-dot.buy {
  background: #d03050;
}

.legend-dot.sell {
  background: #18a058;
}

.legend-dot.hold {
  background: #f0a020;
}

.legend-text {
  font-size: 13px;
  color: #666;
}

:deep(.dark) .card-title,
:deep(.dark) .legend-text {
  color: #eee;
}

:deep(.dark) .signals-legend {
  border-top-color: #333;
}
</style>
