<template>
  <div class="report-viewer">
    <div class="report-header">
      <n-space justify="space-between" align="center">
        <n-tag type="info" size="large">
          <template #icon>
            <n-icon :size="18">
              <svg viewBox="0 0 24 24" fill="currentColor">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6zm-1 2l5 5h-5V4zM6 20V4h6v6a2 2 0 0 0 2 2h6v8H6z"/>
              </svg>
            </n-icon>
          </template>
          AI分析报告
        </n-tag>
        <n-text type="tertiary" depth="3" class="generate-time">
          {{ generateTime }}
        </n-text>
      </n-space>
    </div>
    
    <div class="report-content">
      <MdPreview 
        :modelValue="content" 
        :theme="darkTheme ? 'dark' : 'light'"
        :code-style="darkTheme ? 'atom-one-dark' : 'github'"
        class="custom-md-preview"
      />
    </div>
    
    <div class="report-footer">
      <n-divider />
      <n-alert type="warning" :bordered="false" size="small">
        <template #icon>
          <n-icon :size="18">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2L1 21h22L12 2zm0 3.83l7.53 13.17H4.47L12 5.83zM11 10v4h2v-4h-2zm0 6v2h2v-2h-2z"/>
            </svg>
          </n-icon>
        </template>
        本报告由AI智能分析生成，仅供参考，不构成投资建议。股市有风险，投资需谨慎。
      </n-alert>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { MdPreview } from 'md-editor-v3'
import { NTag, NText, NDivider, NAlert, NIcon, NSpace } from 'naive-ui'

const props = defineProps({
  content: {
    type: String,
    default: ''
  },
  generateTime: {
    type: String,
    default: ''
  },
  darkTheme: {
    type: Boolean,
    default: false
  }
})
</script>

<style scoped>
.report-viewer {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.report-header {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.generate-time {
  font-size: 14px;
}

.report-content {
  line-height: 1.8;
}

/* 自定义Markdown样式 */
.custom-md-preview {
  font-size: 15px;
}

.custom-md-preview :deep(h1) {
  font-size: 24px;
  font-weight: 600;
  margin: 24px 0 16px;
  padding-bottom: 8px;
  border-bottom: 2px solid #18a058;
  color: #333;
}

.custom-md-preview :deep(h2) {
  font-size: 20px;
  font-weight: 600;
  margin: 20px 0 12px;
  padding-left: 12px;
  border-left: 4px solid #2080f0;
  color: #333;
}

.custom-md-preview :deep(h3) {
  font-size: 18px;
  font-weight: 600;
  margin: 16px 0 10px;
  color: #333;
}

.custom-md-preview :deep(p) {
  margin: 12px 0;
  line-height: 1.8;
  color: #444;
}

.custom-md-preview :deep(ul),
.custom-md-preview :deep(ol) {
  padding-left: 24px;
  margin: 12px 0;
}

.custom-md-preview :deep(li) {
  margin: 6px 0;
  line-height: 1.6;
  color: #444;
}

.custom-md-preview :deep(blockquote) {
  border-left: 4px solid #d03050;
  padding-left: 16px;
  margin: 16px 0;
  background: #fff5f5;
  padding: 12px 16px;
  border-radius: 4px;
  color: #666;
}

.custom-md-preview :deep(strong) {
  color: #333;
  font-weight: 600;
}

.custom-md-preview :deep(code) {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 14px;
  color: #d03050;
}

.custom-md-preview :deep(pre) {
  background: #282c34;
  padding: 16px;
  border-radius: 6px;
  overflow-x: auto;
  margin: 16px 0;
}

.custom-md-preview :deep(pre code) {
  background: transparent;
  color: #abb2bf;
  padding: 0;
}

.custom-md-preview :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 16px 0;
}

.custom-md-preview :deep(th),
.custom-md-preview :deep(td) {
  border: 1px solid #e0e0e0;
  padding: 10px 14px;
  text-align: left;
}

.custom-md-preview :deep(th) {
  background: #f5f5f5;
  font-weight: 600;
  color: #333;
}

.custom-md-preview :deep(a) {
  color: #2080f0;
  text-decoration: none;
}

.custom-md-preview :deep(a:hover) {
  text-decoration: underline;
}

.custom-md-preview :deep(img) {
  max-width: 100%;
  border-radius: 6px;
  margin: 12px 0;
}

.report-footer {
  margin-top: 32px;
}

/* 深色主题适配 */
:deep(.dark) .report-viewer {
  color: #ccc;
}

:deep(.dark) .report-header {
  border-bottom-color: #333;
}

:deep(.dark) .custom-md-preview :deep(h1),
:deep(.dark) .custom-md-preview :deep(h2),
:deep(.dark) .custom-md-preview :deep(h3) {
  color: #eee;
}

:deep(.dark) .custom-md-preview :deep(p),
:deep(.dark) .custom-md-preview :deep(li) {
  color: #ccc;
}

:deep(.dark) .custom-md-preview :deep(blockquote) {
  background: #2d1f1f;
  color: #aaa;
}

:deep(.dark) .custom-md-preview :deep(th) {
  background: #2a2a2a;
  color: #eee;
}

:deep(.dark) .custom-md-preview :deep(th),
:deep(.dark) .custom-md-preview :deep(td) {
  border-color: #444;
}
</style>
