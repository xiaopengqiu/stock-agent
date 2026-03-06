<template>
  <n-card class="stock-radar-card" hoverable>
    <template #header>
      <div class="card-header">
        <div class="header-left">
          <n-icon :size="22" color="#2080f0">
            <TrendingUpIcon />
          </n-icon>
          <span class="card-title">综合评分雷达</span>
        </div>
        <n-tag type="info" size="small">
          总分: {{ totalScore }}
        </n-tag>
      </div>
    </template>

    <div class="radar-chart-wrapper">
      <v-chart 
        ref="chartRef"
        :option="radarOption" 
        :autoresize="true"
        style="height: 280px;" 
      />
    </div>

    <div class="score-summary">
      <div class="score-item" v-for="(score, key) in scores" :key="key">
        <div class="score-label">
          <span class="dot" :style="{ backgroundColor: getScoreColor(score) }"></span>
          {{ getScoreLabel(key) }}
        </div>
        <div class="score-bar">
          <div 
            class="score-fill" 
            :style="{ 
              width: `${score}%`,
              backgroundColor: getScoreColor(score)
            }"
          ></div>
        </div>
        <span class="score-value">{{ score }}</span>
      </div>
    </div>
  </n-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { RadarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import { NCard, NTag, NIcon } from 'naive-ui'
import { TrendingUp as TrendingUpIcon } from '@vicons/ionicons5'

use([
  CanvasRenderer,
  RadarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent
])

const props = defineProps({
  scores: {
    type: Object,
    default: () => ({
      technical: 75,
      fundamental: 80,
      capital: 65,
      news: 70,
      risk: 85
    })
  },
  stockName: {
    type: String,
    default: '综合评分'
  },
  darkTheme: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['click'])

const chartRef = ref(null)

// 总分计算
const totalScore = computed(() => {
  const scoreValues = Object.values(props.scores)
  if (scoreValues.length === 0) return 0
  const avg = scoreValues.reduce((a, b) => a + b, 0) / scoreValues.length
  return Math.round(avg)
})

// 评分标签映射
const scoreLabels = {
  technical: '技术面',
  fundamental: '基本面',
  capital: '资金面',
  news: '消息面',
  risk: '风控面'
}

function getScoreLabel(key) {
  return scoreLabels[key] || key
}

// 获取评分颜色
function getScoreColor(score) {
  if (score >= 80) return '#18a058'  // 优秀 - 绿色
  if (score >= 60) return '#f0a020'  // 良好 - 黄色
  if (score >= 40) return '#d4af37'  // 一般 - 金色
  return '#d03050'                       // 较差 - 红色
}

// 雷达图配置
const radarOption = computed(() => {
  const isDark = props.darkTheme
  const textColor = isDark ? '#ccc' : '#333'
  const bgColor = isDark ? '#1a1a1a' : '#fff'
  const gridColor = isDark ? '#333' : '#e8e8e8'

  const indicatorData = Object.entries(props.scores).map(([key, value]) => ({
    name: getScoreLabel(key),
    max: 100,
    color: getScoreColor(value)
  }))

  const values = Object.values(props.scores)

  return {
    backgroundColor: bgColor,
    tooltip: {
      trigger: 'item',
      backgroundColor: isDark ? '#2a2a2a' : '#fff',
      borderColor: isDark ? '#444' : '#ddd',
      textStyle: {
        color: textColor
      },
      formatter: (params) => {
        let result = `<div style="font-weight: 600; margin-bottom: 8px;">${params.name}</div>`
        const data = params.value
        const keys = Object.keys(props.scores)
        keys.forEach((key, index) => {
          const value = data[index]
          const color = getScoreColor(value)
          result += `<div style="margin: 4px 0;">
            <span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${color};margin-right:6px;"></span>
            ${getScoreLabel(key)}: <span style="font-weight:600;color:${color}">${value}</span>
          </div>`
        })
        return result
      }
    },
    radar: {
      indicator: indicatorData,
      shape: 'polygon',
      splitNumber: 4,
      axisName: {
        color: textColor,
        fontSize: 12,
        fontWeight: 500
      },
      splitLine: {
        lineStyle: {
          color: gridColor
        }
      },
      splitArea: {
        show: true,
        areaStyle: {
          color: [
            'rgba(32, 128, 240, 0.02)',
            'rgba(32, 128, 240, 0.04)',
            'rgba(32, 128, 240, 0.06)',
            'rgba(32, 128, 240, 0.08)'
          ]
        }
      },
      axisLine: {
        lineStyle: {
          color: gridColor
        }
      }
    },
    series: [
      {
        name: props.stockName,
        type: 'radar',
        data: [
          {
            value: values,
            name: props.stockName,
            symbol: 'circle',
            symbolSize: 8,
            lineStyle: {
              color: '#2080f0',
              width: 2
            },
            areaStyle: {
              color: {
                type: 'radial',
                x: 0.5,
                y: 0.5,
                r: 0.5,
                colorStops: [
                  { offset: 0, color: 'rgba(32, 128, 240, 0.3)' },
                  { offset: 1, color: 'rgba(32, 128, 240, 0.1)' }
                ]
              }
            },
            itemStyle: {
              color: '#2080f0',
              borderColor: '#fff',
              borderWidth: 2
            }
          }
        ]
      }
    ]
  }
})

onMounted(() => {
  // 初始化完成
})
</script>

<style scoped>
.stock-radar-card {
  transition: all 0.3s ease;
}

.stock-radar-card:hover {
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

.radar-chart-wrapper {
  margin: 8px 0 16px;
}

.score-summary {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.score-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.score-label {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 60px;
  font-size: 13px;
  color: #666;
  flex-shrink: 0;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.score-bar {
  flex: 1;
  height: 8px;
  background: #f0f0f0;
  border-radius: 4px;
  overflow: hidden;
}

.score-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.8s ease-out;
}

.score-value {
  width: 30px;
  text-align: right;
  font-size: 14px;
  font-weight: 600;
  font-family: 'Consolas', 'Monaco', monospace;
}

:deep(.dark) .card-title,
:deep(.dark) .score-label,
:deep(.dark) .score-value {
  color: #eee;
}

:deep(.dark) .score-bar {
  background: #2a2a2a;
}
</style>
