<script setup lang="ts">
import { ref, onMounted, h } from "vue";
import {
  GetToolConfig,
  SetToolConfig,
  ReloadTools,
  GetBuiltinToolCount,
} from "../../wailsjs/go/main/App";
import {
  NTag,
  NButton,
  NCard,
  NSwitch,
  NTable,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NSpace,
  NPopconfirm,
  useMessage,
  useDialog,
} from "naive-ui";
import type { DataTableColumns } from "naive-ui";

const message = useMessage();
const dialog = useDialog();

// 工具配置数据
const toolConfig = ref<ToolConfig>({
  tools: [],
  version: "1.0",
});

// 加载状态
const loading = ref(false);

// 添加/编辑工具弹窗
const showToolModal = ref(false);
const isEditing = ref(false);
const editingToolIndex = ref(-1);

// 工具表单
const toolForm = ref<ToolDefinition>({
  name: "",
  type: "http",
  description: "",
  enabled: true,
  config: {
    url: "",
    method: "POST",
    headers: {},
  },
});

// 工具类型选项
const toolTypeOptions = [
  { label: "HTTP", value: "http" },
  { label: "内置", value: "builtin", disabled: true },
  { label: "MCP", value: "mcp", disabled: true },
];

// HTTP 方法选项
const methodOptions = [
  { label: "POST", value: "POST" },
  { label: "GET", value: "GET" },
];

// 表格列定义
const columns: DataTableColumns<ToolDefinition> = [
  {
    title: "工具名称",
    key: "name",
    width: 180,
    render(row) {
      return h(
        "div",
        {
          style: { fontWeight: 500 },
        },
        row.name
      );
    },
  },
  {
    title: "类型",
    key: "type",
    width: 100,
    render(row) {
      const typeMap: Record<string, string> = {
        builtin: "内置",
        mcp: "MCP",
        http: "HTTP",
      };
      const typeColors: Record<string, any> = {
        builtin: "success",
        mcp: "warning",
        http: "info",
      };
      return h(NTag, {
        type: typeColors[row.type] || "default",
        size: "small",
        bordered: false,
      }, () => typeMap[row.type] || row.type);
    },
  },
  {
    title: "描述",
    key: "description",
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: "启用状态",
    key: "enabled",
    width: 100,
    align: "center",
    render(row, index) {
      return h(NSwitch, {
        value: row.enabled,
        size: "small",
        onUpdateValue: (value: boolean) => handleToggleEnabled(index, value),
      });
    },
  },
  {
    title: "操作",
    key: "actions",
    width: 150,
    align: "center",
    render(row, index) {
      // 内置工具不允许编辑/删除
      if (row.type === "builtin") {
        return h("span", { style: { color: "#999", fontSize: "12px" } }, "系统内置");
      }

      return h(NSpace, { size: "small", justify: "center" }, {
        default: () => [
          h(NButton, {
            size: "tiny",
            type: "primary",
            ghost: true,
            onClick: () => handleEditTool(index),
          }, () => "编辑"),
          h(NPopconfirm, {
            onPositiveClick: () => handleDeleteTool(index),
          }, {
            trigger: () => h(NButton, {
              size: "tiny",
              type: "error",
              ghost: true,
            }, () => "删除"),
            default: () => "确定要删除此工具吗？",
          }),
        ],
      });
    },
  },
];

// 类型定义
interface HttpToolConfig {
  url: string;
  method: string;
  headers: Record<string, string>;
}

interface ToolDefinition {
  name: string;
  type: "builtin" | "mcp" | "http";
  description: string;
  enabled: boolean;
  config?: HttpToolConfig;
}

interface ToolConfig {
  tools: ToolDefinition[];
  version: string;
}

// 加载工具配置
async function loadToolConfig() {
  loading.value = true;
  try {
    const configJson = await GetToolConfig();
    const config = JSON.parse(configJson) as ToolConfig;
    toolConfig.value = config;
  } catch (error) {
    message.error("加载工具配置失败: " + error);
  } finally {
    loading.value = false;
  }
}

// 保存工具配置
async function saveToolConfig() {
  try {
    const configJson = JSON.stringify(toolConfig.value, null, 2);
    const result = await SetToolConfig(configJson);
    message.success(result);
  } catch (error) {
    message.error("保存工具配置失败: " + error);
  }
}

// 启用/禁用工具
async function handleToggleEnabled(index: number, value: boolean) {
  toolConfig.value.tools[index].enabled = value;
  await saveToolConfig();
}

// 打开添加工具弹窗
function handleAddTool() {
  isEditing.value = false;
  editingToolIndex.value = -1;
  toolForm.value = {
    name: "",
    type: "http",
    description: "",
    enabled: true,
    config: {
      url: "",
      method: "POST",
      headers: {},
    },
  };
  showToolModal.value = true;
}

// 编辑工具
function handleEditTool(index: number) {
  const tool = toolConfig.value.tools[index];
  if (tool.type === "builtin") {
    message.warning("内置工具不能编辑");
    return;
  }
  isEditing.value = true;
  editingToolIndex.value = index;
  toolForm.value = JSON.parse(JSON.stringify(tool));
  showToolModal.value = true;
}

// 保存工具
async function handleSaveTool() {
  // 表单验证
  if (!toolForm.value.name.trim()) {
    message.error("工具名称不能为空");
    return;
  }
  if (!toolForm.value.description.trim()) {
    message.error("工具描述不能为空");
    return;
  }
  if (toolForm.value.type === "http") {
    if (!toolForm.value.config?.url.trim()) {
      message.error("HTTP URL 不能为空");
      return;
    }
  }

  // 检查名称是否重复（新增时）
  if (!isEditing.value) {
    const existingTool = toolConfig.value.tools.find(
      (t) => t.name === toolForm.value.name
    );
    if (existingTool) {
      message.error("工具名称已存在");
      return;
    }
  }

  // 保存
  if (isEditing.value && editingToolIndex.value >= 0) {
    toolConfig.value.tools[editingToolIndex.value] = JSON.parse(
      JSON.stringify(toolForm.value)
    );
  } else {
    toolConfig.value.tools.push(JSON.parse(JSON.stringify(toolForm.value)));
  }

  await saveToolConfig();
  showToolModal.value = false;
  message.success(isEditing.value ? "工具更新成功" : "工具添加成功");
}

// 删除工具
async function handleDeleteTool(index: number) {
  const tool = toolConfig.value.tools[index];
  if (tool.type === "builtin") {
    message.warning("内置工具不能删除");
    return;
  }
  toolConfig.value.tools.splice(index, 1);
  await saveToolConfig();
  message.success("工具删除成功");
}

// 重置配置
function handleResetConfig() {
  dialog.warning({
    title: "确认重置",
    content: "确定要重置工具配置为默认状态吗？这将删除所有自定义工具。",
    positiveText: "确认重置",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        // 保留内置工具，删除其他工具
        toolConfig.value.tools = toolConfig.value.tools.filter(
          (t) => t.type === "builtin"
        );
        await saveToolConfig();
        message.success("工具配置已重置");
      } catch (error) {
        message.error("重置失败: " + error);
      }
    },
  });
}

// 重新加载工具
async function handleReloadTools() {
  try {
    message.info("正在重新加载工具...");
    const result = await ReloadTools();
    message.success(result);
    await loadToolConfig();
  } catch (error) {
    message.error("重新加载失败: " + error);
  }
}

// 页面挂载时加载数据
onMounted(() => {
  loadToolConfig();
});
</script>

<template>
  <div class="tool-settings">
    <NCard title="工具配置" hoverable>
      <template #header-extra>
        <NSpace>
          <NButton size="small" type="info" @click="handleReloadTools">
            重新加载
          </NButton>
          <NButton size="small" type="warning" @click="handleResetConfig">
            重置配置
          </NButton>
          <NButton size="small" type="primary" @click="handleAddTool">
            添加工具
          </NButton>
        </NSpace>
      </template>

      <div class="tool-table-container">
        <n-data-table
          :columns="columns"
          :data="toolConfig.tools"
          :loading="loading"
          :pagination="false"
          size="small"
          striped
        />
      </div>

      <div v-if="toolConfig.tools.length === 0 && !loading" class="empty-state">
        <p>暂无工具配置</p>
        <NButton type="primary" @click="handleAddTool">添加第一个工具</NButton>
      </div>
    </NCard>

    <!-- 添加/编辑工具弹窗 -->
    <NModal
      v-model:show="showToolModal"
      :title="isEditing ? '编辑工具' : '添加工具'"
      preset="card"
      style="width: 600px"
      :mask-closable="false"
    >
      <NForm
        :model="toolForm"
        label-placement="left"
        label-width="100"
        require-mark-placement="right-hanging"
      >
        <NFormItem label="工具名称" required>
          <NInput
            v-model:value="toolForm.name"
            placeholder="请输入工具名称"
            :disabled="isEditing"
          />
        </NFormItem>

        <NFormItem label="工具类型" required>
          <NSelect
            v-model:value="toolForm.type"
            :options="toolTypeOptions"
            placeholder="请选择工具类型"
          />
        </NFormItem>

        <NFormItem label="工具描述" required>
          <NInput
            v-model:value="toolForm.description"
            type="textarea"
            placeholder="请输入工具描述"
            :rows="3"
          />
        </NFormItem>

        <!-- HTTP 类型配置 -->
        <template v-if="toolForm.type === 'http'">
          <NDivider title-placement="left">HTTP 配置</NDivider>

          <NFormItem label="请求地址" required>
            <NInput
              v-model:value="toolForm.config!.url"
              placeholder="请输入 HTTP URL，例如: http://localhost:8080/api/tool"
            />
          </NFormItem>

          <NFormItem label="请求方法">
            <NSelect
              v-model:value="toolForm.config!.method"
              :options="methodOptions"
            />
          </NFormItem>
        </template>

        <NFormItem label="启用状态">
          <NSwitch v-model:value="toolForm.enabled" />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showToolModal = false">取消</NButton>
          <NButton type="primary" @click="handleSaveTool">保存</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.tool-settings {
  padding: 16px;
}

.tool-table-container {
  margin-top: 16px;
}

.empty-state {
  text-align: center;
  padding: 48px;
  color: #999;
}

.empty-state p {
  margin-bottom: 16px;
}
</style>
