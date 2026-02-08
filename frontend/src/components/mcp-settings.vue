<script setup>
import { ref, onMounted } from "vue";
import {
  GetMCPEnabled,
  GetMCPStatus,
  ReloadMCPTools,
  GetMCPToolCount,
  GetBuiltinToolCount
} from "../../wailsjs/go/main/App";
import { NTag, NButton, NCard, NSwitch } from "naive-ui";
import { useMessage } from "naive-ui";
import { EventsEmit } from "../../wailsjs/runtime";

const message = useMessage();

const mcpEnabled = ref(false);
const mcpStatus = ref({});
const mcpToolCount = ref(0);
const builtinToolCount = ref(0);

onMounted(() => {
  loadData();
  EventsEmit("updateSettings", () => {
    loadData();
  });
});

async function loadData() {
  await loadMCPStatus();
  await loadToolCounts();
}

async function loadMCPStatus() {
  try {
    const enabled = await GetMCPEnabled();
    mcpEnabled.value = enabled;
    const status = await GetMCPStatus();
    mcpStatus.value = status || {};
  } catch (error) {
    console.error("Failed to load MCP status:", error);
  }
}

async function loadToolCounts() {
  try {
    mcpToolCount.value = await GetMCPToolCount();
    builtinToolCount.value = await GetBuiltinToolCount();
  } catch (error) {
    console.error("Failed to load tool counts:", error);
  }
}

async function reloadTools() {
  try {
    message.info("正在重载 MCP 工具...");
    const result = await ReloadMCPTools();
    message.success(result);
    await loadData();
  } catch (error) {
    message.error("重载失败: " + error.message);
  }
}
</script>

<template>
  <div class="mcp-settings">
    <NCard title="MCP 工具设置" hoverable>
      <template #header-extra>
        <NButton size="small" type="primary" @click="reloadTools">
          重载工具
        </NButton>
      </template>

      <div class="status-overview">
        <div class="status-item">
          <span class="label">MCP 启用:</span>
          <NSwitch v-model:value="mcpEnabled" size="small" />
        </div>
        <div class="status-item">
          <span class="label">内置工具:</span>
          <NTag type="info" round>{{ builtinToolCount }}</NTag>
        </div>
        <div class="status-item">
          <span class="label">MCP 工具:</span>
          <NTag type="default" round>{{ mcpToolCount }}</NTag>
        </div>
      </div>

      <div v-if="Object.keys(mcpStatus).length > 0" class="status-details">
        <div v-for="(state, name) in mcpStatus" :key="name" class="status-detail">
          <div class="server-name">{{ name }}</div>
          <div class="status-info">
            <span>状态: {{ state.status || 'unknown' }}</span>
            <span v-if="state.lastError" class="error-text">{{ state.lastError }}</span>
          </div>
        </div>
      </div>

      <div v-if="mcpEnabled && Object.keys(mcpStatus).length === 0" class="info-message">
        MCP 已启用但没有配置的服务器
      </div>
    </NCard>
  </div>
</template>

<style scoped>
.mcp-settings {
  padding: 16px;
}

.status-overview {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-top: 16px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-item .label {
  font-weight: 500;
}

.status-details {
  margin-top: 16px;
  max-height: 300px;
  overflow-y: auto;
}

.status-detail {
  padding: 12px;
  border-bottom: 1px solid var(--n-divider-color);
}

.server-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.status-info {
  font-size: 14px;
}

.error-text {
  color: var(--n-error-color);
  font-size: 12px;
  margin-top: 4px;
}

.info-message {
  margin-top: 16px;
  padding: 12px;
  background: var(--n-color-info-1);
  border-radius: 4px;
}
</style>
