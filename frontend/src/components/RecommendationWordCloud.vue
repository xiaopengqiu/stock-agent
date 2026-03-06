<template>
  <n-card class="word-cloud-card" hoverable>
    <template #header>
      <div class="card-header">
        <div class="header-left">
          <n-icon :size="22" color="#722ed1">
            <CloudOutline />
          </n-icon>
          <span class="card-title">推荐理由词云</span>
        </div>
        <n-tag type="info" size="small">
          {{ keywords.length }} 个关键词
        </n-tag>
      </div>
    </template>

    <div class="word-cloud-wrapper">
      <div 
        class="word-cloud-container"
        :style="cloudContainerStyle"
      >
        <span
          v-for="(keyword, index) in keywords"
          :key="keyword.text"
          class="word-item"
          :class="`word-size-${keyword.size}`"
          :style="getWordStyle(keyword, index)"
          @click="handleWordClick(keyword)"
        >
          {{ keyword.text }}
        </span>
      </div>
    </div>

    <div class="keywords-summary">
      <div class="summary-title">热门标签</div>
      <div class="tag-list">
        <n-tag
          v-for="tag in topTags"
          :key="tag.text"
          :type="getTagType(tag.category)"
          size="small"
          class="summary-tag"
          @click="handleWordClick(tag)"
        >
          {{ tag.text }}
        </n-tag>
      </div>
    </div>

    <div class="word-detail" v-if="selectedWord">
      <n-divider style="margin: 12px 0;" />
      <div class="selected-word-detail">
        <div class="word-header">
          <n-tag :type="getTagType(selectedWord.category)" size="small">
            {{ selectedWord.category }}
          </n-tag>
          <span class="word-text">{{ selectedWord.text }}</span>
          <span class="word-count">出现 {{ selectedWord.count }} 次</span>
        </div>
        <div class="word-stats">
          <div class="stat-item">
            <span class="stat-label">权重</span>
            <span class="stat-value">{{ selectedWord.weight.toFixed(1) }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">情感倾向</span>
            <span class="stat-value" :class="{ 'positive': selectedWord.sentiment > 0, 'negative': selectedWord.sentiment < 0 }">
              {{ selectedWord.sentiment > 0 ? '正面' : selectedWord.sentiment < 0 ? '负面' : '中性' }}
            </span>
          </div>
        </div>
        <div class="related-stocks" v-if="selectedWord.relatedStocks && selectedWord.relatedStocks.length > 0">
          <div class="related-title">相关股票</div>
          <div class="stock-list">
            <n-tag
              v-for="stock in selectedWord.relatedStocks"
              :key="stock.code"
              size="small"
              class="stock-tag"
              @click="$emit('stock-click', stock)"
            >
              {{ stock.name }}
            </n-tag>
          </div>
        </div>
      </div>
    </div>
  </n-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { NCard, NTag, NIcon, NDivider } from 'naive-ui'
import { CloudOutline } from '@vicons/ionicons5'

const props = defineProps({
  keywords: {
    type: Array,
    default: () => [
      { text: '业绩稳定', weight: 95, size: 5, count: 8, category: '基本面', sentiment: 1, relatedStocks: [{ name: '贵州茅台', code: 'sh600519' }] },
      { text: '新能源', weight: 90, size: 5, count: 7, category: '行业', sentiment: 1, relatedStocks: [{ name: '宁德时代', code: 'sz300750' }, { name: '比亚迪', code: 'sz002594' }] },
      { text: '现金流', weight: 85, size: 4, count: 6, category: '财务', sentiment: 1, relatedStocks: [{ name: '贵州茅台', code: 'sh600519' }] },
      { text: '技术领先', weight: 80, size: 4, count: 5, category: '技术', sentiment: 1, relatedStocks: [{ name: '宁德时代', code: 'sz300750' }] },
      { text: '估值合理', weight: 75, size: 4, count: 5, category: '估值', sentiment: 1, relatedStocks: [{ name: '腾讯控股', code: 'hk00700' }] },
      { text: '零售银行', weight: 70, size: 3, count: 4, category: '行业', sentiment: 0, relatedStocks: [{ name: '招商银行', code: 'sh600036' }] },
      { text: '国产替代', weight: 65, size: 3, count: 4, category: '主题', sentiment: 1, relatedStocks: [{ name: '中芯国际', code: 'sh688981' }] },
      { text: '品牌护城河', weight: 60, size: 3, count: 3, category: '竞争', sentiment: 1, relatedStocks: [{ name: '贵州茅台', code: 'sh600519' }] },
      { text: '销量领先', weight: 55, size: 2, count: 3, category: '市场', sentiment: 1, relatedStocks: [{ name: '比亚迪', code: 'sz002594' }] },
      { text: 'AI+安防', weight: 50, size: 2, count: 2, category: '主题', sentiment: 0, relatedStocks: [{ name: '海康威视', code: 'sz002415' }] },
      { text: 'CRO龙头', weight: 45, size: 2, count: 2, category: '行业', sentiment: 1, relatedStocks: [{ name: '药明康德', code: 'sh603259' }] },
      { text: '全球竞争力', weight: 40, size: 2, count: 2, category: '竞争', sentiment: 1, relatedStocks: [{ name: '药明康德', code: 'sh603259' }] }
    ]
  },
  darkTheme: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['word-click', 'stock-click'])

const selectedWord = ref(null)

// 热门标签（前6个）
const topTags = computed(() => {
  return props.keywords.slice(0, 6)
})

// 词云容器样式
const cloudContainerStyle = computed(() => {
  return {
    minHeight: '200px',
    position: 'relative'
  }
})

// 获取标签颜色类型
function getTagType(category) {
  const typeMap = {
    '基本面': 'success',
    '技术': 'info',
    '财务': 'warning',
    '估值': 'error',
    '行业': 'info',
    '主题': 'default',
    '竞争': 'warning',
    '市场': 'success'
  }
  return typeMap[category] || 'default'
}

// 获取词颜色
function getWordColor(keyword) {
  const colorMap = {
    '基本面': '#18a058',
    '技术': '#2080f0',
    '财务': '#f0a020',
    '估值': '#d03050',
    '行业': '#722ed1',
    '主题': '#13c2c2',
    '竞争': '#d4af37',
    '市场': '#eb2f96'
  }
  
  if (keyword.sentiment > 0) {
    return colorMap[keyword.category] || '#333'
  } else if (keyword.sentiment < 0) {
    return '#d03050'
  }
  return colorMap[keyword.category] || '#666'
}

// 获取词样式
function getWordStyle(keyword, index) {
  const colors = [
    '#d03050', '#18a058', '#2080f0', '#722ed1',
    '#f0a020', '#d4af37', '#13c2c2', '#eb2f96'
  ]
  
  const color = getWordColor(keyword)
  const rotation = (index % 3 - 1) * 10 // -10°, 0°, 10°
  const opacity = 0.6 + (keyword.weight / 100) * 0.4
  
  return {
    color: color,
    transform: `rotate(${rotation}deg)`,
    opacity: opacity,
    cursor: 'pointer',
    transition: 'all 0.3s ease'
  }
}

// 处理词点击
function handleWordClick(keyword) {
  selectedWord.value = keyword
  emit('word-click', keyword)
}

onMounted(() => {
  // 默认选中第一个词
  if (props.keywords.length > 0) {
    selectedWord.value = props.keywords[0]
  }
})
</script>

<style scoped>
.word-cloud-card {
  transition: all 0.3s ease;
}

.word-cloud-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(114, 46, 209, 0.15);
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

.word-cloud-wrapper {
  margin: 12px 0;
  padding: 20px;
  background: linear-gradient(135deg, 
    rgba(114, 46, 209, 0.02) 0%, 
    rgba(32, 128, 240, 0.02) 100%);
  border-radius: 12px;
  border: 1px dashed rgba(114, 46, 209, 0.2);
}

.word-cloud-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  justify-content: center;
  align-items: center;
  min-height: 180px;
}

.word-item {
  padding: 4px 12px;
  border-radius: 20px;
  font-weight: 500;
  user-select: none;
  position: relative;
}

.word-item:hover {
  transform: scale(1.1) !important;
  z-index: 10;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.word-size-5 { font-size: 28px; font-weight: 700; }
.word-size-4 { font-size: 24px; font-weight: 600; }
.word-size-3 { font-size: 20px; font-weight: 500; }
.word-size-2 { font-size: 16px; font-weight: 500; }

.keywords-summary {
  padding: 12px 0;
  border-top: 1px solid #eee;
}

.summary-title {
  font-size: 13px;
  color: #999;
  margin-bottom: 8px;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.summary-tag {
  cursor: pointer;
  transition: all 0.2s ease;
}

.summary-tag:hover {
  transform: scale(1.05);
}

.selected-word-detail {
  padding: 8px 0;
}

.word-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.word-text {
  font-size: 18px;
  font-weight: 600;
}

.word-count {
  font-size: 13px;
  color: #999;
}

.word-stats {
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

.stat-label {
  font-size: 13px;
  color: #666;
}

.stat-value {
  font-size: 14px;
  font-weight: 600;
  font-family: 'Consolas', 'Monaco', monospace;
}

.stat-value.positive {
  color: #18a058;
}

.stat-value.negative {
  color: #d03050;
}

.related-stocks {
  margin-top: 8px;
}

.related-title {
  font-size: 13px;
  color: #999;
  margin-bottom: 8px;
}

.stock-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.stock-tag {
  cursor: pointer;
  transition: all 0.2s ease;
}

.stock-tag:hover {
  transform: scale(1.05);
}

:deep(.dark) .card-title,
:deep(.dark) .word-text {
  color: #eee;
}

:deep(.dark) .word-count,
:deep(.dark) .summary-title,
:deep(.dark) .related-title,
:deep(.dark) .stat-label {
  color: #999;
}

:deep(.dark) .keywords-summary {
  border-top-color: #333;
}

:deep(.dark) .stat-item {
  background: #2a2a2a;
}

:deep(.dark) .word-cloud-wrapper {
  background: linear-gradient(135deg, 
    rgba(114, 46, 209, 0.05) 0%, 
    rgba(32, 128, 240, 0.05) 100%);
}
</style>
