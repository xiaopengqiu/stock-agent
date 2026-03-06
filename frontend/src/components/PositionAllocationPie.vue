<template>
  <n-card class="position-pie-card" hoverable>
    <template #header>
      <div class="card-header">
        <div class="header-left">
          <n-icon :size="22" color="#d4af37">
            <PieChartIcon />
          </n-icon>
          <span class="card-title">仓位建议</span>
        </div>
        <n-tag type="warning" size="small">
          建议总仓位: {{ totalPosition }}%
        </n-tag>
      </div>
    </template>

    <div class="pie-wrapper">
      <div class="pie-chart-container">
        <v-chart 
          ref="chartRef"
          :option="pieOption" 
          :autoresize="true"
          style="height: 260px;" 
        />
      </div>
      <div class="pie-summary">
        <div class="summary-item">
          <span class="summary-label">建议仓位</span>
          <span class="summary-value" style="color: #d4af37;">{{ totalPosition }}%</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">股票数量</span>
          <span class="summary-value">{{ positions.length }}只</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">最高单仓</span>
          <span class="summary-value">{{ maxPosition }}%</span>
        </div>
      </div>
    </div>

    <div class="position-list">
      <div class="list-title">持仓明细</div>
      <div class="position-items">
        <div 
          v-for="(pos, index) in positions" 
          :key="pos.code"
          class="position-item"
          @click="selectPosition(pos)"
        >
          <div class="pos-left">
            <span class="pos-order">{{ index + 1 }}</span>
            <div class="pos-info">
              <div class="pos-name">{{ pos.name }}</div>
              <div class="pos-code">{{ pos.code }}</div>
            </div>
          </div>
          <div class="pos-right">
            <div class="pos-bar">
              <div 
                class="pos-bar-fill" 
                :style="{ 
                  width: `${pos.percent}%`,
                  backgroundColor: positionColors[index % positionColors.length]
                }"
              ></div>
            </div>
            <span class="pos-percent">{{ pos.percent }}%</span>
          </div>
        </div>
      </div>
    </div>

    <div class="market-temperature" v-if="showTemperature">
      <n-divider style="margin: 12px 0;" />
      <div class="temperature-container">
        <div class="temperature-label">市场热度</div>
        <div class="temperature-bar">
          <div class="temperature-scale">
            <span class="scale-item cold">寒冷 20%</span>
            <span class="scale-item cool">偏冷 40%</span>
            <span class="scale-item normal">正常 60%</span>
            <span class="scale-item warm">偏热 80%</span>
            <span class="scale-item hot">过热 100%</span>
          </div>
          <div class="temperature-indicator" :style="{ left: `${marketTemp}%` }">
            <div class="indicator-arrow"></div>
            <div class="indicator-value">{{ marketTemp }}%</div>
          </div>
        </div>
        <div class="temperature-status" :class="temperatureStatusClass">
          <n-icon size="16"><component :is="temperatureStatusIcon" /></n-icon>
          <span>{{ temperatureStatusText }}</span>
        </div>
      </div>
    </div>

    <div class="position-detail" v-if="selectedPosition">
      <n-divider style="margin: 12px 0;" />
      <div class="selected-detail">
        <div class="detail-header">
          <n-tag :type="getPositionType(selectedPosition)" size="small">
            {{ selectedPosition.recommend }}
          </n-tag>
          <span class="detail-name">{{ selectedPosition.name }}</span>
          <span class="detail-code">{{ selectedPosition.code }}</span>
        </div>
        <n-grid :x-gap="12" :y-gap="8" :cols="2">
          <n-grid-item>
            <n-statistic label="建议仓位">
              <template #value>
                <span style="color: #d4af37;">{{ selectedPosition.percent }}%</span>
              </template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="预期收益">
              <template #value>
                <span :class="{ 'text-positive': selectedPosition.return >= 0, 'text-negative': selectedPosition.return < 0 }">
                  {{ selectedPosition.return >= 0 ? '+' : '' }}{{ selectedPosition.return.toFixed(1) }}%
                </span>
              </template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="风险等级">
              <template #value>
                <n-tag size="small" :type="getRiskType(selectedPosition.risk)">
                  {{ selectedPosition.risk < 30 ? '低' : selectedPosition.risk < 50 ? '中' : '高' }}
                </n-tag>
              </template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="所属板块">
              <template #value>{{ selectedPosition.sector }}</template>
            </n-statistic>
          </n-grid-item>
        </n-grid>
        <div class="detail-reason" v-if="selectedPosition.reason">
          <n-alert type="info" :bordered="false" size="small">
            {{ selectedPosition.reason }}
          </n-alert>
        </div>
      </div>
    </div>
  </n-card>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import { NCard, NTag, NIcon, NDivider, NGrid, NGridItem, NStatistic, NAlert } from 'naive-ui'
import { PieChart as PieChartIcon, ThermometerOutline, SnowOutline, SunnyOutline, FlameOutline } from '@vicons/ionicons5'

use([
  CanvasRenderer,
  PieChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent
])

const props = defineProps({
  positions: {
    type: Array,
    default: () => [
      { name: '贵州茅台', code: 'sh600519', percent: 25, return: 25, risk: 35, sector: '白酒', recommend: '核心配置', reason: '业绩稳定，现金流优秀，品牌护城河深厚' },
      { name: '腾讯控股', code: 'hk00700', percent: 20, return: 18, risk: 30, sector: '互联网', recommend: '核心配置', reason: '社交生态稳固，游戏和云业务增长' },
      { name: '宁德时代', code: 'sz300750', percent: 18, return: 40, risk: 65, sector: '新能源', recommend: '积极配置', reason: '新能源电池龙头，技术领先全球' },
      { name: '招商银行', code: 'sh600036', percent: 15, return: 12, risk: 20, sector: '银行', recommend: '稳健配置', reason: '零售银行标杆，资产质量优良' },
      { name: '比亚迪', code: 'sz002594', percent: 12, return: 35, risk: 55, sector: '汽车', recommend: '积极配置', reason: '新能源汽车销量持续领先' },
      { name: '药明康德', code: 'sh603259', percent: 10, return: 22, risk: 45, sector: '医药', recommend: '适度配置', reason: 'CRO行业龙头，全球竞争力强' }
    ]
  },
  marketTemperature: {
    type: Number,
    default: 65
  },
  showTemperature: {
    type: Boolean,
    default: true
  },
  darkTheme: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['position-click'])

const chartRef = ref(null)
const selectedPosition = ref(null)

const positionColors = ['#d03050', '#18a058', '#2080f0', '#f0a020', '#722ed1', '#d4af37']

// 总仓位
const totalPosition = computed(() => {
  return props.positions.reduce((sum, pos) => sum + pos.percent, 0)
})

// 最高单仓
const maxPosition = computed(() => {
  return Math.max(...props.positions.map(pos => pos.percent))
})

// 市场温度
const marketTemp = computed(() => {
  return props.marketTemperature
})

// 温度状态
const temperatureStatusClass = computed(() => {
  const temp = marketTemp.value
  if (temp < 30) return 'cold'
  if (temp < 50) return 'cool'
  if (temp < 70) return 'normal'
  if (temp < 85) return 'warm'
  return 'hot'
})

const temperatureStatusText = computed(() => {
  const temp = marketTemp.value
  if (temp < 30) return '市场寒冷，可加大仓位'
  if (temp < 50) return '市场偏冷，适合布局'
  if (temp < 70) return '市场正常，保持仓位'
  if (temp < 85) return '市场偏热，适度减仓'
  return '市场过热，注意风险'
})

const temperatureStatusIcon = computed(() => {
  const temp = marketTemp.value
  if (temp < 30) return SnowOutline
  if (temp < 50) return ThermometerOutline
  if (temp < 70) return SunnyOutline
  return FlameOutline
})

// 获取仓位类型
function getPositionType(pos) {
  if (pos.recommend === '核心配置') return 'success'
  if (pos.recommend === '积极配置') return 'info'
  if (pos.recommend === '适度配置') return 'warning'
  return 'default'
}

// 获取风险类型
function getRiskType(risk) {
  if (risk < 30) return 'success'
  if (risk < 50) return 'warning'
  return 'error'
}

// 选中仓位
function selectPosition(pos) {
  selectedPosition.value = selectedPosition.value?.code === pos.code ? null : pos
  if (selectedPosition.value) {
    emit('position-click', selectedPosition.value)
  }
}

// 饼图配置
const pieOption = computed(() => {
  const isDark = props.darkTheme
  const textColor = isDark ? '#ccc' : '#333'
  const bgColor = isDark ? '#1a1a1a' : '#fff'

  return {
    backgroundColor: bgColor,
    tooltip: {
      trigger: 'item',
      backgroundColor: isDark ? '#2a2a2a' : '#fff',
      borderColor: isDark ? '#444' : '#ddd',
      textStyle: { color: textColor },
      formatter: (params) => {
        return `
          <div style="font-weight: 600; margin-bottom: 4px;">${params.name}</div>
          <div>仓位: ${params.value}%</div>
          <div style="color: #999; font-size: 12px;">占比: ${params.percent.toFixed(1)}%</div>
        `
      }
    },
    legend: {
      show: true,
      orient: 'vertical',
      right: '5%',
      top: 'center',
      textStyle: {
        color: textColor,
        fontSize: 11
      }
    },
    series: [
      {
        name: '仓位分布',
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['35%', '50%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 8,
          borderColor: bgColor,
          borderWidth: 2
        },
        label: {
          show: false
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 14,
            fontWeight: 'bold'
          },
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        },
        data: props.positions.map((pos, idx) => ({
          value: pos.percent,
          name: pos.name,
          itemStyle: {
            color: positionColors[idx % positionColors.length]
          }
        }))
      }
    ]
  }
})

onMounted(() => {
  if (props.positions.length > 0) {
    selectedPosition.value = props.positions[0]
  }
})
</script>

<style scoped>
.position-pie-card {
  transition: all 0.3s ease;
}

.position-pie-card:hover {
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

.pie-wrapper {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 16px;
  margin: 12px 0;
}

.pie-chart-container {
  min-height: 220px;
}

.pie-summary {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 12px;
  padding-left: 16px;
  border-left: 1px solid #eee;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.summary-label {
  font-size: 13px;
  color: #999;
}

.summary-value {
  font-size: 18px;
  font-weight: 700;
  font-family: 'Consolas', 'Monaco', monospace;
}

.position-list {
  padding-top: 8px;
  border-top: 1px solid #eee;
}

.list-title {
  font-size: 13px;
  color: #999;
  margin-bottom: 8px;
}

.position-items {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.position-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.position-item:hover {
  background: #e9ecef;
  transform: translateX(4px);
}

.pos-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.pos-order {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #d4af37, #f0a020);
  color: white;
  font-size: 11px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pos-info {
  display: flex;
  flex-direction: column;
}

.pos-name {
  font-size: 14px;
  font-weight: 500;
}

.pos-code {
  font-size: 11px;
  color: #999;
  font-family: 'Consolas', 'Monaco', monospace;
}

.pos-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.pos-bar {
  width: 80px;
  height: 8px;
  background: #e0e0e0;
  border-radius: 4px;
  overflow: hidden;
}

.pos-bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.5s ease-out;
}

.pos-percent {
  width: 40px;
  text-align: right;
  font-size: 14px;
  font-weight: 600;
  font-family: 'Consolas', 'Monaco', monospace;
  color: #d4af37;
}

.market-temperature {
  padding-top: 4px;
}

.temperature-container {
  padding: 8px 0;
}

.temperature-label {
  font-size: 13px;
  color: #999;
  margin-bottom: 8px;
}

.temperature-bar {
  position: relative;
  height: 20px;
  background: linear-gradient(90deg, 
    #18a058 0%, 
    #18a058 20%,
    #2080f0 20%,
    #2080f0 40%,
    #f0a020 40%,
    #f0a020 60%,
    #f0a020 60%,
    #d4af37 80%,
    #d03050 80%,
    #d03050 100%
  );
  border-radius: 10px;
}

.temperature-scale {
  position: absolute;
  top: -24px;
  left: 0;
  right: 0;
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  color: #999;
}

.temperature-indicator {
  position: absolute;
  top: -30px;
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  transition: left 0.5s ease-out;
}

.indicator-arrow {
  width: 0;
  height: 0;
  border-left: 6px solid transparent;
  border-right: 6px solid transparent;
  border-top: 8px solid #333;
}

.indicator-value {
  font-size: 12px;
  font-weight: 600;
  background: #333;
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  margin-bottom: 2px;
}

.temperature-status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 20px;
  padding: 8px 16px;
  background: #f8f9fa;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
}

.temperature-status.cold {
  background: rgba(24, 160, 88, 0.1);
  color: #18a058;
}

.temperature-status.cool {
  background: rgba(32, 128, 240, 0.1);
  color: #2080f0;
}

.temperature-status.normal {
  background: rgba(240, 160, 32, 0.1);
  color: #f0a020;
}

.temperature-status.warm {
  background: rgba(212, 175, 55, 0.1);
  color: #d4af37;
}

.temperature-status.hot {
  background: rgba(208, 48, 80, 0.1);
  color: #d03050;
}

.position-detail {
  padding-top: 4px;
}

.selected-detail {
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
:deep(.dark) .pos-name,
:deep(.dark) .detail-name {
  color: #eee;
}

:deep(.dark) .pos-code,
:deep(.dark) .detail-code,
:deep(.dark) .list-title,
:deep(.dark) .temperature-label,
:deep(.dark) .temperature-scale {
  color: #999;
}

:deep(.dark) .pie-summary,
:deep(.dark) .position-list {
  border-left-color: #333;
  border-top-color: #333;
}

:deep(.dark) .position-item {
  background: #2a2a2a;
}

:deep(.dark) .position-item:hover {
  background: #333;
}

:deep(.dark) .pos-bar {
  background: #3a3a3a;
}

:deep(.dark) .temperature-status {
  background: #2a2a2a;
}
</style>
