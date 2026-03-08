<template>
    <div style="height: calc(100vh - 200px); display: flex; flex-direction: column;">
      <!-- 顶部分析状态提示 -->
      <div v-if="analyzing" style="margin-bottom: 12px;">
        <n-alert type="info" :bordered="false">
          <template #icon>
            <n-icon><component :is="SparklesOutline"/></n-icon>
          </template>
          <n-space>
            <span>AI正在分析市场数据，请稍候...</span>
            <div class="typing-dots" style="margin-top: 4px;">
              <span></span><span></span><span></span>
            </div>
          </n-space>
        </n-alert>
      </div>
      <n-row :gutter="16" style="flex: 1; min-height: 0; overflow: hidden;">
        <!-- 左侧对话区域 -->
        <n-col :span="12" class="chat-column">
          <n-card title="AI对话" class="chat-card">
            <template #header-extra>
              <n-button size="small" @click="showHistory">
                <template #icon>
                  <n-icon><component :is="TimeOutline"/></n-icon>
                </template>
                历史记录
              </n-button>
            </template>

            <!-- 对话区域主容器 - 使用固定高度布局 -->
            <div class="chat-container">
              <!-- 工具列表展示 -->
              <div class="tools-section" v-if="toolsList.length > 0">
                <n-card size="small">
                  <template #header>
                    <n-flex align="center">
                      <n-icon :size="16" style="margin-right: 5px">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                          <path fill="currentColor" d="M19.14,12.94c0.04-0.3,0.06-0.61,0.06-0.94c0-0.32-0.02-0.64-0.07-0.94l0.01,0.05C19.14,12.94,19.14,12.94,19.14,12.94z M17.83,18.82c-0.13,0.18-0.28,0.35-0.45,0.5l-0.03-0.03C17.56,19.12,17.69,18.97,17.83,18.82z M11.81,14.56c-0.07-0.34-0.1-0.69-0.1-1.04c0-0.36,0.04-0.71,0.11-1.06C11.85,13.22,11.82,13.89,11.81,14.56z M14.46,15.38c-0.15-0.31-0.28-0.64-0.39-0.99l-0.05,0.02C14.07,14.73,14.25,15.05,14.46,15.38z M7.16,13.21c0.02-0.26,0.05-0.52,0.09-0.77C7.17,12.82,7.16,13.02,7.16,13.21z M5.52,15.12c0.09-0.26,0.2-0.51,0.31-0.76L5.74,14.41C5.63,14.64,5.56,14.89,5.52,15.12z M10.77,17.87c-0.05-0.28-0.09-0.57-0.12-0.86l-0.02,0.03C10.66,17.31,10.71,17.59,10.77,17.87z M13.41,18.4c-0.11-0.23-0.23-0.45-0.33-0.69l-0.04,0.01C13.16,17.95,13.28,18.16,13.41,18.4z M16.54,16.36l-0.03-0.04c-0.14,0.16-0.28,0.33-0.45,0.48c0.18-0.17,0.33-0.35,0.48-0.51V16.36z M16.16,17.52l-0.01-0.01c-0.15,0.16-0.31,0.31-0.49,0.45c0.19-0.15,0.36-0.31,0.52-0.47L16.16,17.52z M15.01,18.64c-0.14,0.14-0.29,0.28-0.46,0.4l0.02-0.02C14.75,18.89,14.88,18.77,15.01,18.64z M18.4,16.3l-0.01-0.01c-0.14,0.16-0.28,0.32-0.43,0.46c0.16-0.15,0.31-0.31,0.45-0.47L18.4,16.3z"/>
                          <path fill="currentColor" d="M12,2C6.48,2,2,6.48,2,12s4.48,10,10,10s10-4.48,10-10S17.52,2,12,2z M12,20c-4.41,0-8-3.59-8-8s3.59-8,8-8s8,3.59,8,8S16.41,20,12,20z M12,6c-3.31,0-6,2.69-6,6s2.69,6,6,6s6-2.69,6-6S15.31,6,12,6z"/>
                        </svg>
                      </n-icon>
                      <span>可用工具</span>
                    </n-flex>
                  </template>
                  <n-space>
                    <n-tag
                      v-for="tool in toolsList"
                      :key="tool"
                      :type="getToolStatus(tool)"
                      size="small"
                      :bordered="false"
                    >
                      {{ tool }}
                    </n-tag>
                  </n-space>
                </n-card>
              </div>

              <!-- 消息列表区域 - 可滚动 -->
              <div class="messages-section">
                <div ref="messagesContainer" class="messages-scroll-container">
                  <div class="messages-content">
                    <div v-for="(msg, index) in messages" :key="index" class="message-item">
                      <!-- 用户消息 -->
                      <div v-if="msg.role === 'user'" class="message-wrapper user-message">
                        <div class="message-content-wrapper">
                          <div class="message-bubble user-bubble">
                            <div class="message-text">{{ msg.content }}</div>
                          </div>
                          <div class="message-avatar user-avatar">
                            <n-avatar size="small" color="#2080f0">
                              <template #icon>
                                <n-icon><component :is="UserOutline"/></n-icon>
                              </template>
                            </n-avatar>
                          </div>
                        </div>
                      </div>
                      <!-- AI消息 -->
                      <div v-else class="message-wrapper ai-message">
                        <div class="message-content-wrapper">
                          <div class="message-avatar ai-avatar">
                            <n-avatar size="small" color="#18a058">
                              <template #icon>
                                <n-icon><component :is="SparklesOutline"/></n-icon>
                              </template>
                            </n-avatar>
                          </div>
                          <div class="message-bubble ai-bubble">
                            <MdPreview :modelValue="msg.content" :theme="darkTheme ? 'dark' : 'light'"/>
                          </div>
                        </div>
                      </div>
                    </div>
                    <!-- 输入指示器 -->
                    <div v-if="analyzing && !firstTokenReceived" class="message-wrapper ai-message">
                      <div class="message-content-wrapper">
                        <div class="message-avatar ai-avatar">
                          <n-avatar size="small" color="#18a058">
                            <template #icon>
                              <n-icon><component :is="SparklesOutline"/></n-icon>
                            </template>
                          </n-avatar>
                        </div>
                        <div class="message-bubble ai-bubble typing-indicator">
                          <div class="typing-dots">
                            <span></span><span></span><span></span>
                          </div>
                        </div>
                      </div>
                    </div>
                    <!-- 底部占位，确保最后一条消息不被遮挡 -->
                    <div style="height: 8px;"></div>
                  </div>
                </div>
              </div>

              <!-- 底部输入区域 - 固定 -->
              <div class="input-section">
                <!-- 输入区域工具栏 -->
                <div class="input-toolbar" v-if="!analyzing">
                  <n-button-group size="tiny">
                    <n-button quaternary @click="clearInput" :disabled="!inputText">
                      <template #icon>
                        <n-icon><component :is="RefreshOutline"/></n-icon>
                      </template>
                      清空输入
                    </n-button>
                    <n-button quaternary @click="clearChat" :disabled="messages.length <= 1">
                      <template #icon>
                        <n-icon><component :is="TrashOutline"/></n-icon>
                      </template>
                      清空对话
                    </n-button>
                    <n-button quaternary @click="copyLastMessage" :disabled="messages.length < 2">
                      <template #icon>
                        <n-icon><component :is="CopyOutline"/></n-icon>
                      </template>
                      复制
                    </n-button>
                  </n-button-group>
                </div>

                <!-- 输入框 -->
                <div class="input-area">
                  <n-input
                    v-model:value="inputText"
                    type="textarea"
                    placeholder="请输入您的选股需求，例如：推荐今日资金流入大的科技股"
                    :autosize="{ minRows: 2, maxRows: 6 }"
                    @keydown.enter.prevent="handleEnter"
                    :disabled="analyzing"
                    class="chat-input"
                  />
                </div>
                <n-button
                  type="primary"
                  block
                  :loading="analyzing"
                  @click="sendMessage"
                  class="send-button"
                >
                  <template #icon>
                    <n-icon v-if="!analyzing"><component :is="SendOutline"/></n-icon>
                  </template>
                  {{ analyzing ? '分析中...' : '开始分析' }}
                </n-button>
              </div>
            </div>
          </n-card>
        </n-col>

        <!-- 右侧推荐结果区域 -->
        <n-col :span="12" class="chat-column">
          <n-card title="推荐结果" class="chat-card">
            <template #header-extra>
              <n-space>
                <n-button-group size="small">
                  <n-button
                    :type="viewMode === 'full' ? 'primary' : 'default'"
                    @click="viewMode = 'full'"
                  >
                    完整报告
                  </n-button>
                  <n-button
                    :type="viewMode === 'card' ? 'primary' : 'default'"
                    @click="viewMode = 'card'"
                  >
                    卡片视图
                  </n-button>
                  <n-button
                    :type="viewMode === 'simple' ? 'primary' : 'default'"
                    @click="viewMode = 'simple'"
                  >
                    简洁列表
                  </n-button>
                </n-button-group>
                <n-dropdown :options="exportOptions" @select="exportReport">
                  <n-button size="small" type="tertiary">
                    <template #icon>
                      <n-icon><component :is="DownloadOutline"/></n-icon>
                    </template>
                  </n-button>
                </n-dropdown>
              </n-space>
            </template>

            <div class="result-content-wrapper">
              <!-- 完整报告视图 -->
              <div v-if="viewMode === 'full'" class="full-report-wrapper">
                <n-scrollbar style="height: calc(100vh - 280px);">
                  <div style="padding: 16px; padding-bottom: 60px;">
                    <MdPreview :modelValue="fullReport" :theme="darkTheme ? 'dark' : 'light'" class="md-preview-content"/>
                  </div>
                </n-scrollbar>
              </div>

              <!-- 卡片视图 - 优化版 -->
              <div v-else-if="viewMode === 'card'" class="card-view-wrapper">
                <n-scrollbar style="height: calc(100vh - 280px);">
                  <div style="padding: 16px; padding-bottom: 60px;">
                    
                    <!-- 推荐股票卡片（放在最前面） -->
                    <n-divider title-placement="left" v-if="cardRecommendations.length > 0">
                      <span style="font-size: 14px; font-weight: 600;">📋 推荐股票</span>
                    </n-divider>
                    
                    <div class="card-grid" v-if="cardRecommendations.length > 0">
                    <n-card
                      v-for="(item, index) in cardRecommendations"
                      :key="item.stockCode"
                      class="stock-card"
                      size="small"
                      :segmented="{ content: true }"
                    >
                      <!-- 卡片头部 - 股票名称和代码 -->
                      <template #header>
                        <div class="card-header">
                          <div class="stock-title">
                            <span class="stock-name">{{ item.stockName }}</span>
                            <span class="stock-code">{{ item.stockCode }}</span>
                            <n-tag :type="getTradeSuggestionType(item.tradeSuggestion)" size="tiny" class="suggestion-tag">
                              {{ item.tradeSuggestion }}
                            </n-tag>
                          </div>
                          <div class="stock-price">
                            <span class="current-price" :class="{ 'price-up': item.priceChange >= 0, 'price-down': item.priceChange < 0 }">
                              {{ item.currentPrice.toFixed(2) }}
                            </span>
                            <span class="price-change" :class="{ 'price-up': item.priceChange >= 0, 'price-down': item.priceChange < 0 }">
                              {{ item.priceChange >= 0 ? '+' : '' }}{{ item.priceChange.toFixed(2) }}%
                            </span>
                          </div>
                        </div>
                      </template>

                      <!-- 买卖点信息区域 -->
                      <div class="trading-points">
                        <!-- 买入区域 -->
                        <div class="trading-point buy-point">
                          <div class="point-label">
                            <n-icon size="14" color="#18a058">
                              <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M12 4l-1.41 1.41L16.17 11H4v2h12.17l-5.58 5.59L12 20l8-8z"/>
                              </svg>
                            </n-icon>
                            <span class="label-text">买入区间</span>
                          </div>
                          <div class="price-value buy-price">{{ item.buyPriceRange }}</div>
                        </div>

                        <!-- 目标价 -->
                        <div class="trading-point target-point">
                          <div class="point-label">
                            <n-icon size="14" color="#2080f0">
                              <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-2-5.5l6-4.5-6-4.5z"/>
                              </svg>
                            </n-icon>
                            <span class="label-text">目标价</span>
                          </div>
                          <div class="price-value target-price">{{ item.targetPrice.toFixed(2) }}</div>
                          <div class="expected-return" :class="{ 'positive': item.expectedReturn > 0, 'negative': item.expectedReturn < 0 }">
                            {{ item.expectedReturn >= 0 ? '+' : '' }}{{ item.expectedReturn.toFixed(1) }}%
                          </div>
                        </div>

                        <!-- 止损价 -->
                        <div class="trading-point stop-loss-point">
                          <div class="point-label">
                            <n-icon size="14" color="#d03050">
                              <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
                              </svg>
                            </n-icon>
                            <span class="label-text">止损价</span>
                          </div>
                          <div class="price-value stop-loss-price">{{ item.stopLossPrice.toFixed(2) }}</div>
                          <div class="stop-loss-rate" v-if="item.stopLossRate">
                            止损 {{ item.stopLossRate }}%
                          </div>
                        </div>
                      </div>

                      <!-- 板块和推荐理由 -->
                      <div class="stock-info">
                        <div class="sector-concept" v-if="item.sectorConcept">
                          <n-tag size="tiny" type="info" :bordered="false">
                            {{ item.sectorConcept }}
                          </n-tag>
                        </div>
                        <div class="reason" v-if="item.reason">
                          <n-text depth="2" class="reason-text">
                            {{ item.reason }}
                          </n-text>
                        </div>
                      </div>

                      <!-- 技术面分析 -->
                      <div class="analysis-section" v-if="item.technicalAnalysis">
                        <n-divider style="margin: 12px 0;">
                          <span style="font-size: 12px; color: #666;">📊 技术面分析</span>
                        </n-divider>
                        <n-text depth="2" class="analysis-text">
                          {{ item.technicalAnalysis }}
                        </n-text>
                      </div>

                      <!-- 基本面分析 -->
                      <div class="analysis-section" v-if="item.fundamentalAnalysis">
                        <n-divider style="margin: 12px 0;">
                          <span style="font-size: 12px; color: #666;">📈 基本面分析</span>
                        </n-divider>
                        <n-text depth="2" class="analysis-text">
                          {{ item.fundamentalAnalysis }}
                        </n-text>
                      </div>

                      <!-- 风险提示 -->
                      <div class="risk-warning" v-if="item.riskTips">
                        <n-alert type="warning" size="small" :bordered="false">
                          <template #icon>
                            <n-icon>
                              <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M12 2L1 21h22L12 2zm0 3.83l7.53 13.17H4.47L12 5.83z"/>
                              </svg>
                            </n-icon>
                          </template>
                          {{ item.riskTips }}
                        </n-alert>
                      </div>

                      <!-- 备注 -->
                      <div class="remarks" v-if="item.remarks">
                        <n-text depth="3" class="remarks-text">
                          {{ item.remarks }}
                        </n-text>
                      </div>

                      <!-- 推荐时间 -->
                      <div class="recommend-time">
                        <n-text depth="3" class="time-text">
                          推荐时间: {{ item.recommendedAt }}
                        </n-text>
                      </div>

                      <!-- 卡片底部操作栏 -->
                      <div class="card-footer">
                        <n-space justify="space-between" align="center">
                          <n-tag :type="item.isFollowed ? 'success' : 'default'" size="small">
                            {{ item.isFollowed ? '已关注' : '未关注' }}
                          </n-tag>
                          <n-space>
                            <n-button
                              size="small"
                              type="info"
                              @click="handleAddToPosition(item)"
                            >
                              加入持仓
                            </n-button>
                            <n-button
                              size="small"
                              :type="item.isFollowed ? 'default' : 'primary'"
                              @click="handleFollow(item.stockCode)"
                            >
                              {{ item.isFollowed ? '取消关注' : '关注' }}
                            </n-button>
                          </n-space>
                        </n-space>
                      </div>
                    </n-card>
                    </div>
                    
                    <!-- 数据分析抽屉区域 -->
                    <n-divider title-placement="left" style="margin-top: 32px;" v-if="cardRecommendations.length > 0">
                      <span style="font-size: 14px; font-weight: 600;">📊 数据分析</span>
                    </n-divider>
                    
                    <n-collapse v-if="cardRecommendations.length > 0" style="margin-top: 16px;" :default-expanded-names="['technical']">
                      <!-- 技术面分析 -->
                      <n-collapse-item title="📊 技术面分析" name="technical">
                        <div style="padding: 16px 0;">
                          <n-tabs type="line" animated v-model:value="activeTechTab">
                            <n-tab-pane v-for="item in cardRecommendations.slice(0, 5)" :key="item.stockCode" :name="item.stockCode" :tab="item.stockName">
                              <div style="padding: 8px 0;">
                                <!-- K线图 -->
                                <n-card title="K线走势" size="small" style="margin-bottom: 16px;">
                                  <KLineChart
                                    :code="item.stockCode"
                                    :name="item.stockName"
                                    :dark-theme="darkTheme"
                                    :chart-height="350"
                                    :k-days="90"
                                  />
                                </n-card>

                                <!-- 技术指标 -->
                                <n-grid :x-gap="16" :y-gap="16" :cols="2" responsive="screen">
                                  <n-grid-item span="1">
                                    <n-card size="small" title="技术指标">
                                      <div style="margin-bottom: 8px;">
                                        <n-tag size="small" type="info" style="margin-right: 8px; margin-bottom: 4px;">
                                          {{ item.trend === 'up' ? '📈 上升趋势' : item.trend === 'down' ? '📉 下降趋势' : '➡️ 震荡' }}
                                        </n-tag>
                                        <n-tag size="small" v-if="item.rsi" style="margin-right: 8px; margin-bottom: 4px;" :type="getRsiType(item.rsi)">
                                          RSI: {{ item.rsi.toFixed(1) }}
                                        </n-tag>
                                        <n-tag size="small" v-if="item.macd" style="margin-right: 8px; margin-bottom: 4px;">
                                          MACD: {{ item.macd }}
                                        </n-tag>
                                        <n-tag size="small" v-if="item.kdj" style="margin-right: 8px; margin-bottom: 4px;">
                                          KDJ: {{ item.kdj }}
                                        </n-tag>
                                      </div>
                                      <n-text v-if="item.technicalAnalysis" type="info" depth="3">
                                        {{ item.technicalAnalysis }}
                                      </n-text>
                                    </n-card>
                                  </n-grid-item>

                                  <n-grid-item span="1">
                                    <n-card size="small" title="价格目标">
                                      <n-space vertical style="width: 100%;">
                                        <div style="display: flex; justify-content: space-between; align-items: center;">
                                          <n-text depth="2">当前价格</n-text>
                                          <n-text strong style="font-size: 16px;">{{ item.currentPrice?.toFixed(2) }}</n-text>
                                        </div>
                                        <n-divider style="margin: 8px 0;" />
                                        <div style="display: flex; justify-content: space-between; align-items: center;">
                                          <n-text depth="2">目标价</n-text>
                                          <n-text strong type="success">{{ item.targetPrice?.toFixed(2) }}</n-text>
                                        </div>
                                        <div style="display: flex; justify-content: space-between; align-items: center;">
                                          <n-text depth="2">预期涨幅</n-text>
                                          <n-tag :type="item.targetChangePercent >= 0 ? 'success' : 'error'" size="small">
                                            {{ item.targetChangePercent >= 0 ? '+' : '' }}{{ item.targetChangePercent?.toFixed(1) }}%
                                          </n-tag>
                                        </div>
                                        <n-divider style="margin: 8px 0;" />
                                        <div style="display: flex; justify-content: space-between; align-items: center;">
                                          <n-text depth="2">止损价</n-text>
                                          <n-text strong type="error">{{ item.stopLossPrice?.toFixed(2) }}</n-text>
                                        </div>
                                        <div style="display: flex; justify-content: space-between; align-items: center;">
                                          <n-text depth="2">止损幅度</n-text>
                                          <n-tag type="error" size="small">-{{ item.stopLossRate }}%</n-tag>
                                        </div>
                                      </n-space>
                                    </n-card>
                                  </n-grid-item>
                                </n-grid>
                              </div>
                            </n-tab-pane>
                          </n-tabs>
                        </div>
                      </n-collapse-item>
                      
                      <!-- 基本面分析 -->
                      <n-collapse-item title="基本面分析" name="fundamental">
                        <div style="padding: 16px 0;">
                          <n-grid :x-gap="16" :y-gap="16" :cols="1" responsive="screen">
                            <n-grid-item span="1">
                              <n-card size="small" title="基本面指标">
                                <div v-for="(item, index) in cardRecommendations.slice(0, 3)" :key="`fund-${item.stockCode}`" style="margin-bottom: 16px;">
                                  <n-text strong>{{ item.stockName }} ({{ item.stockCode }})</n-text>
                                  <div style="margin-top: 8px;">
                                    <n-tag size="small" v-if="item.pe" style="margin-right: 8px; margin-bottom: 4px;">
                                      PE: {{ item.pe.toFixed(2) }}
                                    </n-tag>
                                    <n-tag size="small" v-if="item.pb" style="margin-right: 8px; margin-bottom: 4px;">
                                      PB: {{ item.pb.toFixed(2) }}
                                    </n-tag>
                                    <n-tag size="small" v-if="item.roe" style="margin-right: 8px; margin-bottom: 4px;">
                                      ROE: {{ item.roe.toFixed(1) }}%
                                    </n-tag>
                                    <n-tag size="small" v-if="item.revenueGrowth" style="margin-right: 8px; margin-bottom: 4px;">
                                      营收增长: {{ item.revenueGrowth >= 0 ? '+' : '' }}{{ item.revenueGrowth.toFixed(1) }}%
                                    </n-tag>
                                    <n-tag size="small" v-if="item.profitGrowth" style="margin-right: 8px; margin-bottom: 4px;">
                                      利润增长: {{ item.profitGrowth >= 0 ? '+' : '' }}{{ item.profitGrowth.toFixed(1) }}%
                                    </n-tag>
                                  </div>
                                  <n-text v-if="item.fundamentalAnalysis" type="info" depth="3" style="display: block; margin-top: 8px;">
                                    {{ item.fundamentalAnalysis }}
                                  </n-text>
                                </div>
                              </n-card>
                            </n-grid-item>
                          </n-grid>
                        </div>
                      </n-collapse-item>

                      <!-- 筹码分析 -->
                      <n-collapse-item title="筹码分析" name="shareholder">
                        <div style="padding: 16px 0;">
                          <n-grid :x-gap="16" :y-gap="16" :cols="1" responsive="screen">
                            <n-grid-item span="1">
                              <n-card size="small" title="股东人数与筹码集中度">
                                <div v-for="(item, index) in cardRecommendations.slice(0, 3)" :key="`holder-${item.stockCode}`" style="margin-bottom: 16px;">
                                  <n-text strong>{{ item.stockName }} ({{ item.stockCode }})</n-text>
                                  <n-text v-if="item.shareHolderAnalysis" type="info" depth="3" style="display: block; margin-top: 8px;">
                                    {{ item.shareHolderAnalysis }}
                                  </n-text>
                                  <n-text v-else type="info" depth="3" style="display: block; margin-top: 8px; color: #999;">
                                    暂无筹码分析数据
                                  </n-text>
                                </div>
                              </n-card>
                            </n-grid-item>
                          </n-grid>
                        </div>
                      </n-collapse-item>

                      <!-- 舆情动态 -->
                      <n-collapse-item title="舆情动态" name="news">
                        <div style="padding: 16px 0;">
                          <n-grid :x-gap="16" :y-gap="16" :cols="1" responsive="screen">
                            <n-grid-item span="1">
                              <n-card size="small" title="相关新闻与资讯">
                                <div v-for="(item, index) in cardRecommendations.slice(0, 3)" :key="`news-${item.stockCode}`" style="margin-bottom: 16px;">
                                  <n-text strong>{{ item.stockName }} ({{ item.stockCode }})</n-text>
                                  <n-text v-if="item.newsAnalysis" type="info" depth="3" style="display: block; margin-top: 8px;">
                                    {{ item.newsAnalysis }}
                                  </n-text>
                                  <n-text v-else type="info" depth="3" style="display: block; margin-top: 8px; color: #999;">
                                    暂无舆情数据
                                  </n-text>
                                </div>
                              </n-card>
                            </n-grid-item>
                          </n-grid>
                        </div>
                      </n-collapse-item>

                      <!-- 风险分析 -->
                      <n-collapse-item title="风险分析" name="risk">
                        <div style="padding: 16px 0;">
                          <n-grid :x-gap="16" :y-gap="16" :cols="1" responsive="screen">
                            <n-grid-item span="1">
                              <n-card size="small" title="风险评估">
                                <div v-for="(item, index) in cardRecommendations" :key="`risk-${item.stockCode}`" style="margin-bottom: 16px;">
                                  <n-flex justify="space-between" align="center">
                                    <n-text strong>{{ item.stockName }}</n-text>
                                    <n-tag :type="item.riskLevel === 'low' ? 'success' : item.riskLevel === 'medium' ? 'warning' : 'error'" size="small">
                                      {{ item.riskLevel === 'low' ? '低风险' : item.riskLevel === 'medium' ? '中风险' : '高风险' }}
                                    </n-tag>
                                  </n-flex>
                                  <div style="margin-top: 8px;">
                                    <n-text type="info" depth="3">
                                      目标价: {{ item.targetPrice?.toFixed(2) }} ({{ item.targetChangePercent >= 0 ? '+' : '' }}{{ item.targetChangePercent?.toFixed(1) }}%)
                                    </n-text>
                                  </div>
                                  <n-alert v-if="item.riskTips" type="warning" size="small" :bordered="false" style="margin-top: 8px;">
                                    {{ item.riskTips }}
                                  </n-alert>
                                </div>
                              </n-card>
                            </n-grid-item>
                          </n-grid>
                        </div>
                      </n-collapse-item>
                    </n-collapse>
                  </div>
                </n-scrollbar>
              </div>

              <!-- 简洁列表视图 -->
              <div v-else class="simple-list-wrapper">
                <n-data-table
                  :columns="columns"
                  :data="simpleRecommendations"
                  :pagination="pagination"
                  :bordered="false"
                  striped
                  size="small"
                />
              </div>
            </div>
          </n-card>
        </n-col>
      </n-row>
    </div>

  <!-- 历史记录抽屉 -->
  <n-drawer v-model:show="historyVisible" width="50%" placement="right">
    <template #header>
      <n-text strong style="padding-left: 8px;">历史推荐记录</n-text>
    </template>
    <n-spin :show="loadingHistory">\n      <n-scrollbar style="height: calc(100vh - 120px);">
      <n-list v-if="historyList.length > 0" style="padding-left: 12px;">
        <n-list-item
          v-for="item in historyList"
          :key="item.ID"
          clickable
          @click="viewHistoryReport(item.ID)"
        >
          <template #suffix>
            <n-popconfirm
              positive-text="确认"
              negative-text="取消"
              @positive-click="handleDelete(item.ID)"
            >
              <template #trigger>
                <n-button
                  size="tiny"
                  type="error"
                  ghost
                  :loading="deletingIds.has(item.ID)"
                  @click.stop
                  style="margin-right: 8px"
                >
                  <template #icon>
                    <n-icon><component :is="TrashOutline"/></n-icon>
                  </template>
                  删除
                </n-button>
              </template>
              确定要删除这条历史记录吗？
            </n-popconfirm>
          </template>
          <n-thing>
            <template #header>
              <n-text>{{ item.QuerySummary }}</n-text>
            </template>
            <template #description>
              <n-space>
                <n-tag size="small" :type="item.Status === 'completed' ? 'success' : 'warning'">
                  {{ item.Status === 'completed' ? '已完成' : '处理中' }}
                </n-tag>
                <n-text type="info">
                  推荐数量: {{ item.RecommendCount }}
                </n-text>
                <n-text type="tertiary" depth="3">
                  {{ formatTime(item.CreatedAt) }}
                </n-text>
              </n-space>
            </template>
          </n-thing>
        </n-list-item>
      </n-list>
      <n-empty v-else description="暂无历史记录" style="padding-left: 12px;" />
      </n-scrollbar>
    </n-spin>
  </n-drawer>

  <!-- 加入持仓模态框 -->
  <n-modal v-model:show="showAddPositionModal" preset="card" title="加入持仓" style="width: 500px;">
    <n-form :model="addPositionForm" label-placement="left" label-width="100">
      <n-form-item label="股票代码">
        <n-input v-model:value="addPositionForm.stockCode" placeholder="股票代码" readonly />
      </n-form-item>
      <n-form-item label="股票名称">
        <n-input v-model:value="addPositionForm.stockName" placeholder="股票名称" readonly />
      </n-form-item>
      <n-form-item label="持股数量">
        <n-input-number v-model:value="addPositionForm.quantity" :min="1" placeholder="请输入持股数量" style="width: 100%;" />
      </n-form-item>
      <n-form-item label="买入价格">
        <n-input-number v-model:value="addPositionForm.buyPrice" :min="0" :precision="2" placeholder="请输入买入价格" style="width: 100%;" />
      </n-form-item>
      <n-form-item label="备注">
        <n-input v-model:value="addPositionForm.notes" type="textarea" placeholder="备注信息" />
      </n-form-item>
    </n-form>
    <template #footer>
      <n-flex justify="end">
        <n-button @click="showAddPositionModal = false" style="margin-right: 8px;">取消</n-button>
        <n-button type="primary" @click="confirmAddPosition">确认加入</n-button>
      </n-flex>
    </template>
  </n-modal>

</template>

<script setup>
import {ref, onMounted, onBeforeUnmount, h, computed, watch, nextTick} from 'vue'
import {
  AIStockPickChat,
  GetStockPickReports,
  GetStockPickReport,
  GetStockPickRecommendations,
  FollowStockFromReport,
  UnfollowStockFromReport,
  ExportStockPickReport,
  CheckStockFollowed,
  GetStockPickStats,
  DeleteStockPickReport
} from "../../wailsjs/go/main/App"
import KLineChart from "./KLineChart.vue"
import {EventsOn, EventsOff} from "../../wailsjs/runtime"
import {MdPreview} from "md-editor-v3"
import {
  NCard,
  NRow,
  NCol,
  NSpin,
  NSpace,
  NTag,
  NScrollbar,
  NInput,
  NButton,
  NButtonGroup,
  NText,
  NIcon,
  NDataTable,
  NDrawer,
  NList,
  NListItem,
  NThing,
  NDropdown,
  NEmpty,
  NPopconfirm,
  NAlert,
  NTabs,
  NTabPane,
  NDivider,
  NAvatar,
  NModal,
  NForm,
  NFormItem,
  NInputNumber,
  useMessage,
  useNotification
} from "naive-ui"
import {TimeOutline, DownloadOutline, TrashOutline, PersonOutline as UserOutline, SparklesOutline, RefreshOutline, CopyOutline, SendOutline} from '@vicons/ionicons5'

const message = useMessage()
const notify = useNotification()

// 缓存键名
const CACHE_KEY = 'ai-stock-pick-cache'

// 从缓存恢复状态
function restoreFromCache() {
  try {
    const cached = localStorage.getItem(CACHE_KEY)
    if (cached) {
      const data = JSON.parse(cached)
      // 检查缓存是否在1小时内
      if (data.timestamp && Date.now() - data.timestamp < 3600000) {
        if (data.messages && data.messages.length > 0) {
          messages.value = data.messages
        }
        if (data.recommendations && data.recommendations.length > 0) {
          recommendations.value = data.recommendations
        }
        if (data.fullReport) {
          fullReport.value = data.fullReport
        }
        if (data.reportId) {
          reportId.value = data.reportId
        }
        if (data.viewMode) {
          viewMode.value = data.viewMode
        }
        console.log('从缓存恢复了AI荐股状态')
      }
    }
  } catch (e) {
    console.warn('恢复缓存失败:', e)
  }
}

// 保存状态到缓存
function saveToCache() {
  try {
    const data = {
      timestamp: Date.now(),
      messages: messages.value,
      recommendations: recommendations.value,
      fullReport: fullReport.value,
      reportId: reportId.value,
      viewMode: viewMode.value
    }
    localStorage.setItem(CACHE_KEY, JSON.stringify(data))
  } catch (e) {
    console.warn('保存缓存失败:', e)
  }
}

// 清除缓存
function clearCache() {
  try {
    localStorage.removeItem(CACHE_KEY)
  } catch (e) {
    console.warn('清除缓存失败:', e)
  }
}

// 状态管理
const analyzing = ref(false)
const firstTokenReceived = ref(false)
const inputText = ref('')
const messagesContainer = ref(null)
const messages = ref([
  {
    id: 1,
    role: 'assistant',
    content: '请告诉我您的选股需求，例如：\n\n- 推荐今日资金流入大的科技股\n- 寻找市盈率低于20且业绩增长的银行股\n- 推荐近期创新高的新能源龙头股',
    timestamp: Date.now()
  }
])
const toolsList = ref([])
const toolStatus = ref({})
const recommendations = ref([])
const viewMode = ref('card') // 默认改为卡片视图
const fullReport = ref('')
const reportId = ref(null)
const darkTheme = ref(false)

// 滚动到底部
function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      // 直接操作原生滚动容器
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

// 历史记录相关
const historyVisible = ref(false)
const historyList = ref([])
const loadingHistory = ref(false)
const deletingIds = ref(new Set())

// 技术面分析当前选中的股票
const activeTechTab = ref('')

// 加入持仓相关
const showAddPositionModal = ref(false)
const addPositionForm = ref({
  stockCode: '',
  stockName: '',
  quantity: 100,
  buyPrice: 0,
  notes: ''
})

// 导出选项
const exportOptions = [
  {
    label: '导出为Markdown',
    key: 'markdown'
  }
]

// 简洁列表数据 - 使用与卡片视图相同的字段
const simpleRecommendations = computed(() => {
  return cardRecommendations.value.map((item, index) => ({
    key: item.stockCode,
    rank: index + 1,
    stockCode: item.stockCode,
    stockName: item.stockName,
    currentPrice: item.currentPrice,
    priceChange: item.priceChange,
    buyPriceRange: item.buyPriceRange,
    targetPrice: item.targetPrice,
    stopLossPrice: item.stopLossPrice,
    expectedReturn: item.expectedReturn,
    tradeSuggestion: item.tradeSuggestion,
    sectorConcept: item.sectorConcept,
    reason: item.reason,
    recommendedAt: item.recommendedAt,
    isFollowed: item.isFollowed
  }))
})

// 卡片视图数据 - 包含买卖点信息
const cardRecommendations = computed(() => {
  return recommendations.value.map((rec, index) => {
    // 计算目标涨幅和止损率
    const currentPrice = rec.current_price || 0
    const targetPrice = rec.target_price || (currentPrice * 1.1)
    const stopLossPrice = rec.stop_loss_price || (currentPrice * 0.95)
    const expectedReturn = currentPrice > 0 ? ((targetPrice - currentPrice) / currentPrice * 100) : 0
    const stopLossRate = currentPrice > 0 ? ((currentPrice - stopLossPrice) / currentPrice * 100) : 0

    // 解析技术指标
    let rsi = null, macd = null, kdj = null, trend = null
    if (rec.technical_analysis) {
      const techText = rec.technical_analysis
      if (techText.includes('RSI')) {
        const rsiMatch = techText.match(/RSI[^\d]*(\d+\.?\d*)/i)
        if (rsiMatch) rsi = parseFloat(rsiMatch[1])
      }
      if (techText.includes('MACD')) {
        const macdMatch = techText.match(/MACD[^\w]*(金叉|死叉|向上|向下|positive|negative)/i)
        if (macdMatch) macd = macdMatch[1]
      }
      if (techText.includes('KDJ')) {
        const kdjMatch = techText.match(/KDJ[^\w]*(金叉|死叉|超买|超卖)/i)
        if (kdjMatch) kdj = kdjMatch[1]
      }
      if (techText.includes('上升') || techText.includes('上涨') || techText.includes('up')) {
        trend = 'up'
      } else if (techText.includes('下降') || techText.includes('下跌') || techText.includes('down')) {
        trend = 'down'
      } else {
        trend = 'sideways'
      }
    }

    // 解析基本面指标
    let pe = null, pb = null, roe = null, revenueGrowth = null, profitGrowth = null
    if (rec.fundamental_analysis) {
      const fundText = rec.fundamental_analysis
      const peMatch = fundText.match(/PE[^\d]*(\d+\.?\d*)/i)
      if (peMatch) pe = parseFloat(peMatch[1])
      const pbMatch = fundText.match(/PB[^\d]*(\d+\.?\d*)/i)
      if (pbMatch) pb = parseFloat(pbMatch[1])
      const roeMatch = fundText.match(/ROE[^\d]*(\d+\.?\d*)/i)
      if (roeMatch) roe = parseFloat(roeMatch[1])
    }

    // 风险等级判断
    let riskLevel = 'medium'
    if (expectedReturn > 15) {
      riskLevel = 'high'
    } else if (expectedReturn < 5) {
      riskLevel = 'low'
    }

    // 在原始对象上添加新字段，保持引用完整性
    rec.stockName = rec.stock_name || '-'
    rec.stockCode = rec.stock_code || '-'
    rec.currentPrice = currentPrice
    rec.recommendedPrice = rec.recommended_price || currentPrice
    rec.previousClose = rec.previous_close || (currentPrice * 0.98)
    rec.priceChange = rec.price_change || 0
    rec.buyPriceRange = rec.buy_price_range || `${(currentPrice * 0.98).toFixed(2)}-${(currentPrice * 1.02).toFixed(2)}`
    rec.targetPrice = targetPrice
    rec.targetChangePercent = rec.target_change_percent || expectedReturn
    rec.stopLossPrice = stopLossPrice
    rec.expectedReturn = expectedReturn
    rec.stopLossRate = stopLossRate.toFixed(1)
    rec.sectorConcept = rec.sector_concept || rec.industry || ''
    rec.reason = rec.reason || rec.recommendation_reason || ''
    rec.riskTips = rec.risk_tips || rec.risk_warning || ''
    rec.tradeSuggestion = rec.trade_suggestion || rec.action || (expectedReturn > 5 ? '买入' : '观望')
    rec.remarks = rec.remarks || ''
    // 技术面分析数据
    rec.technicalAnalysis = rec.technical_analysis || ''
    rec.rsi = rsi
    rec.macd = macd
    rec.kdj = kdj
    rec.trend = trend
    // 基本面分析数据
    rec.fundamentalAnalysis = rec.fundamental_analysis || ''
    rec.pe = pe
    rec.pb = pb
    rec.roe = roe
    rec.revenueGrowth = revenueGrowth
    rec.profitGrowth = profitGrowth
    // 筹码分析数据
    rec.shareHolderAnalysis = rec.shareholder_analysis || ''
    // 舆情动态
    rec.newsAnalysis = rec.news_analysis || ''
    // 风险等级
    rec.riskLevel = riskLevel
    rec.recommendedAt = formatTime(rec.created_at || Date.now())
    rec.isFollowed = rec.is_followed || false

    return rec
  })
})

// 获取操作建议标签类型
function getTradeSuggestionType(suggestion) {
  if (!suggestion) return 'default'
  if (suggestion.includes('买入') || suggestion.includes('推荐')) return 'success'
  if (suggestion.includes('卖出') || suggestion.includes('减持')) return 'error'
  if (suggestion.includes('观望') || suggestion.includes('持有')) return 'warning'
  return 'default'
}

// 获取RSI标签类型
function getRsiType(rsi) {
  if (!rsi) return 'default'
  if (rsi > 70) return 'error' // 超买
  if (rsi < 30) return 'success' // 超卖
  return 'default'
}

// 表格列配置 - 简洁列表展示
const columns = [
  { title: '排名', key: 'rank', width: 50, fixed: 'left' },
  { title: '股票代码', key: 'stockCode', width: 90, fixed: 'left' },
  { title: '股票名称', key: 'stockName', width: 100 },
  {
    title: '现价',
    key: 'currentPrice',
    width: 80,
    render: (row) => h(NText, { type: 'info' }, { default: () => row.currentPrice.toFixed(2) })
  },
  {
    title: '涨跌幅',
    key: 'priceChange',
    width: 80,
    render: (row) => h(NText, { type: row.priceChange >= 0 ? 'success' : 'error' }, { default: () => `${row.priceChange >= 0 ? '+' : ''}${row.priceChange.toFixed(2)}%` })
  },
  { title: '买入区间', key: 'buyPriceRange', width: 120 },
  {
    title: '目标价',
    key: 'targetPrice',
    width: 80,
    render: (row) => h(NText, { type: 'info' }, { default: () => row.targetPrice.toFixed(2) })
  },
  {
    title: '止损价',
    key: 'stopLossPrice',
    width: 80,
    render: (row) => h(NText, { type: 'error' }, { default: () => row.stopLossPrice.toFixed(2) })
  },
  {
    title: '预期收益',
    key: 'expectedReturn',
    width: 90,
    render: (row) => h(NText, { type: row.expectedReturn >= 0 ? 'success' : 'error' }, { default: () => `${row.expectedReturn >= 0 ? '+' : ''}${row.expectedReturn.toFixed(1)}%` })
  },
  { title: '操作建议', key: 'tradeSuggestion', width: 80 },
  { title: '板块概念', key: 'sectorConcept', width: 100, ellipsis: { tooltip: true } },
  { title: '推荐理由', key: 'reason', width: 200, ellipsis: { tooltip: true } },
  { title: '推荐时间', key: 'recommendedAt', width: 140 },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    fixed: 'right',
    render: (row) => h('div', { style: { display: 'flex', gap: '4px' } }, [
      h(NButton, {
        size: 'tiny',
        type: 'info',
        onClick: () => handleAddToPosition(row)
      }, { default: () => '加入持仓' }),
      h(NButton, {
        size: 'tiny',
        type: row.isFollowed ? 'default' : 'primary',
        onClick: () => handleFollow(row.stockCode)
      }, { default: () => row.isFollowed ? '取消关注' : '关注' })
    ])
  }
]

// 分页配置
const pagination = ref({
  pageSize: 10
})

// 工具列表
const availableTools = [
  'ChoiceStockByIndicators',
  'QueryStockKLine',
  'GetFinancialReport',
  'QueryShareholderCount',
  'QueryStockNewsTool',
  'GetIndustryResearchReport',
  'QueryEconomicData',
  'QueryStockPriceInfo',
  'QueryInteractiveAnswerData'
]

// 初始化
onMounted(() => {
  initEventListeners()
  toolsList.value = availableTools
  restoreFromCache()
})

onBeforeUnmount(() => {
  cleanupEventListeners()
  saveToCache()
})

// 监听数据变化，自动保存缓存
watch([messages, recommendations, fullReport, viewMode], () => {
  saveToCache()
}, { deep: true, flush: 'post' })

// 事件监听
function initEventListeners() {
  EventsOn('ai-stock-pick-stream', handleStream)
  EventsOn('ai-stock-pick-start', handleStart)
  EventsOn('ai-stock-pick-tool', handleToolCall)
  EventsOn('ai-stock-pick-update', handleUpdate)
}

function cleanupEventListeners() {
  EventsOff('ai-stock-pick-stream')
  EventsOff('ai-stock-pick-start')
  EventsOff('ai-stock-pick-tool')
  EventsOff('ai-stock-pick-update')
}

// 处理回车发送
function handleEnter(e) {
  if (e.shiftKey) {
    // Shift+Enter 换行
    return
  }
  sendMessage()
}

// 发送消息
async function sendMessage() {
  const query = inputText.value.trim()
  if (!query) {
    message.warning('请输入选股需求')
    return
  }

  // 添加用户消息
  messages.value.push({
    role: 'user',
    content: query,
    timestamp: Date.now()
  })

  // 清空输入框
  inputText.value = ''

  // 添加AI响应占位
  const aiMessage = {
    role: 'assistant',
    content: '',
    timestamp: Date.now()
  }
  messages.value.push(aiMessage)

  // 开始分析
  analyzing.value = true
  firstTokenReceived.value = false
  fullReport.value = '正在分析市场数据...'

  try {
    const result = await AIStockPickChat(query, 0)
    reportId.value = parseInt(result)
    message.success('荐股分析完成')
  } catch (error) {
    message.error('荐股分析失败: ' + error)
    analyzing.value = false
    firstTokenReceived.value = false
  }
}

// 处理流式响应
function handleStream(data) {
  if (data.content) {
    // 收到第一个token时，停止loading状态
    if (!firstTokenReceived.value) {
      firstTokenReceived.value = true
      analyzing.value = false
    }
    // 更新最后一条AI消息
    const lastMessage = messages.value[messages.value.length - 1]
    if (lastMessage && lastMessage.role === 'assistant') {
      lastMessage.content += data.content
    }
    fullReport.value += data.content
  }
}

// 处理开始事件
function handleStart(data) {
  message.info(data.message)
}

// 处理工具调用
function handleToolCall(data) {
  toolStatus.value[data.tool_name] = data.status
  message.info(`调用工具: ${data.tool_name}`)
}

// 处理更新事件
function handleUpdate(data) {
  if (data.recommendations) {
    recommendations.value = data.recommendations
  }

  // 更新市场分析
  if (data.market_analysis) {
    // 检查当前报告是否已包含市场分析，如果没有则添加
    if (!fullReport.value.includes('## 市场环境分析')) {
      fullReport.value += '\n\n## 市场环境分析\n\n' + data.market_analysis + '\n\n'
    }
  }

  // 更新筛选逻辑
  if (data.filter_logic) {
    if (!fullReport.value.includes('## 筛选逻辑')) {
      fullReport.value += '## 筛选逻辑\n\n' + data.filter_logic + '\n\n'
    }
  }

  // 如果AI分析完成，显示成功消息并停止加载状态
  if (data.status === 'completed') {
    analyzing.value = false

    // 更新推荐股票部分（不再添加到完整报告视图，已保留在简洁列表和卡片视图中）
    let recMarkdown = ''
    if (data.recommendations && data.recommendations.length > 0) {
      recMarkdown = '## 推荐股票\n\n'
      data.recommendations.forEach((rec, index) => {
        recMarkdown += `${index + 1}. [${rec.stock_code}] ${rec.stock_name} - ${rec.reason || ''}\n`
        if (rec.current_price) {
          recMarkdown += `   - 当前价格：${rec.current_price.toFixed(2)}\n`
        }
        if (rec.price_change) {
          recMarkdown += `   - 涨跌幅：${rec.price_change >= 0 ? '+' : ''}${rec.price_change.toFixed(2)}%\n`
        }
        if (rec.technical_analysis) {
          recMarkdown += `   - 技术面分析：${rec.technical_analysis}\n`
        }
        if (rec.fundamental_analysis) {
          recMarkdown += `   - 基本面分析：${rec.fundamental_analysis}\n`
        }
        if (rec.shareholder_analysis) {
          recMarkdown += `   - 筹码分析：${rec.shareholder_analysis}\n`
        }
        if (rec.news_analysis) {
          recMarkdown += `   - 舆情动态：${rec.news_analysis}\n`
        }
        if (rec.target_change_percent) {
          recMarkdown += `   - 目标涨幅：${rec.target_change_percent}%\n`
        }
        if (rec.risk_tips) {
          recMarkdown += `   - 风险提示：${rec.risk_tips}\n`
        }
        recMarkdown += '\n'
      })
    }

    // 完整报告视图不再显示推荐股票列表（简洁列表和卡片视图已有）
    // 仅添加免责声明
    if (!fullReport.value.includes('本报告由AI智能分析生成')) {
      fullReport.value += '\n---\n\n本报告由AI智能分析生成，仅供参考，不构成投资建议。股市有风险，投资需谨慎。'
    }

    // 显示成功消息
    notify.success({
      title: 'AI分析完成',
      content: `成功生成${data.candidates_count || 0}个推荐股票，扫描了${data.total_scanned || 0}只股票`,
      duration: 5000
    })

    message.success(`AI荐股分析完成！推荐${data.candidates_count || 0}只股票`)
  }
}

// 获取工具状态
function getToolStatus(toolName) {
  const status = toolStatus.value[toolName]
  if (status === 'success') return 'success'
  if (status === 'running') return 'warning'
  if (status === 'failed') return 'error'
  return 'default'
}

// 显示历史记录
async function showHistory() {
  historyVisible.value = true
  await loadHistory()
}

// 加载历史记录
async function loadHistory() {
  loadingHistory.value = true
  try {
    const result = await GetStockPickReports(0, 20)

    console.log('GetStockPickReports 原始返回值:', result)
    console.log('返回值类型:', typeof result)

    // 新的返回格式: StockPickReportsResponse { items: [], total: number }
    let items = []

    if (result && result.items && Array.isArray(result.items)) {
      items = result.items
      console.log('从 result.items 获取，长度:', items.length)
      console.log('total:', result.total)
    } else {
      console.warn('GetStockPickReports 返回格式不正确:', result)
    }

    console.log('最终 items 数量:', items.length)

    if (items.length === 0) {
      message.info('暂无历史记录')
      historyList.value = []
      return
    }

    historyList.value = items
    message.success(`加载成功，共 ${items.length} 条记录`)
  } catch (error) {
    console.error('加载历史记录失败:', error)
    message.error('加载历史记录失败: ' + (error?.message || error || '未知错误'))
    historyList.value = []
  } finally {
    loadingHistory.value = false
  }
}

// 删除历史记录
async function handleDelete(id) {
  if (deletingIds.value.has(id)) {
    return
  }

  deletingIds.value.add(id)
  try {
    const result = await DeleteStockPickReport(id)
    message.success('删除成功')
    // 从列表中移除已删除的项
    const index = historyList.value.findIndex(item => item.ID === id)
    if (index !== -1) {
      historyList.value.splice(index, 1)
    }
  } catch (error) {
    console.error('删除失败:', error)
    message.error('删除失败: ' + (error?.message || error || '未知错误'))
  } finally {
    deletingIds.value.delete(id)
  }
}

// 查看历史报告
async function viewHistoryReport(id) {
  try {
    console.log('加载历史报告:', id)
    const report = await GetStockPickReport(id)
    if (report) {
      reportId.value = report.id
      fullReport.value = report.result
      const recs = await GetStockPickRecommendations(id)
      recommendations.value = recs || []
      historyVisible.value = false
      message.success('报告加载成功')
    }
  } catch (error) {
    console.error('加载报告失败:', error)
    const errorMsg = error?.message || error?.toString() || error || '未知错误'

    // 判断是否是记录不存在的错误
    if (errorMsg.includes('record not found') || errorMsg.includes('记录不存在') || errorMsg.includes('已被删除')) {
      message.error('报告记录不存在或已被删除，请刷新历史记录列表')
      // 自动刷新历史记录列表
      setTimeout(async () => {
        await loadHistory()
      }, 1000)
    } else {
      message.error('加载报告失败: ' + errorMsg)
    }
  }
}

// 格式化报告为Markdown
function formatReportToMarkdown(report) {
  let markdown = ''
  markdown += `# AI荐股报告\n\n`
  markdown += `**生成时间：** ${formatTime(report.created_at)}\n\n`
  markdown += `**选股需求：** ${report.user_query}\n\n`

  if (report.market_analysis) {
    markdown += `## 市场环境分析\n\n${report.market_analysis}\n\n`
  }

  if (report.filter_logic) {
    markdown += `## 筛选逻辑\n\n${report.filter_logic}\n\n`
  }

  markdown += `## 推荐股票\n\n${report.recommendations}\n\n`

  markdown += `---\n\n`
  markdown += `*本报告由AI智能分析生成，仅供参考，不构成投资建议。股市有风险，投资需谨慎。*\n`

  return markdown
}

// 格式化时间 - 支持多种时间格式
function formatTime(timestamp) {
  if (!timestamp) return ''

  let date
  if (typeof timestamp === 'string') {
    // 处理 Go 时间字符串格式 (如 "2024-01-15T10:30:00+08:00" 或 "2024-01-15T10:30:00Z")
    date = new Date(timestamp)
  } else if (typeof timestamp === 'number') {
    // 处理时间戳（毫秒或秒）
    date = timestamp > 1e10 ? new Date(timestamp) : new Date(timestamp * 1000)
  } else {
    date = new Date(timestamp)
  }

  if (isNaN(date.getTime())) return ''

  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 导出报告
async function exportReport(format) {
  if (!reportId.value) {
    message.warning('当前没有可导出的报告')
    return
  }

  try {
    const result = await ExportStockPickReport(reportId.value, format)
    message.success(`已导出: ${result}`)
  } catch (error) {
    message.error('导出失败: ' + error)
  }
}

// 一键关注/取消关注
async function handleFollow(stockCode) {
  try {
    // 找到对应股票
    const index = recommendations.value.findIndex(r => r.stock_code === stockCode)
    const isCurrentlyFollowed = index !== -1 && recommendations.value[index].is_followed

    let result
    if (isCurrentlyFollowed) {
      result = await UnfollowStockFromReport(reportId.value, stockCode)
      message.success(result)
      if (index !== -1) {
        recommendations.value[index].is_followed = false
      }
    } else {
      result = await FollowStockFromReport(reportId.value, stockCode)
      message.success(result)
      if (index !== -1) {
        recommendations.value[index].is_followed = true
      }
    }
  } catch (error) {
    message.error('操作失败: ' + error)
  }
}

// 加入持仓
function handleAddToPosition(item) {
  addPositionForm.value = {
    stockCode: item.stockCode || item.stock_code,
    stockName: item.stockName || item.stock_name,
    quantity: 100,
    buyPrice: item.currentPrice || item.current_price || item.recommendedPrice || item.recommended_price || 0,
    notes: `来自AI荐股: ${item.reason || item.recommendation_reason || ''}`
  }
  showAddPositionModal.value = true
}

// 确认添加持仓
async function confirmAddPosition() {
  if (!addPositionForm.value.stockCode || !addPositionForm.value.stockName) {
    message.warning('请选择股票')
    return
  }
  if (!addPositionForm.value.quantity || addPositionForm.value.quantity <= 0) {
    message.warning('请输入持股数量')
    return
  }
  if (!addPositionForm.value.buyPrice || addPositionForm.value.buyPrice <= 0) {
    message.warning('请输入买入价格')
    return
  }

  try {
    // TODO: Wails绑定生成后取消注释
    // await AddPositionFromRecommendation(
    //   addPositionForm.value.stockCode,
    //   addPositionForm.value.stockName,
    //   addPositionForm.value.buyPrice,
    //   addPositionForm.value.quantity,
    //   addPositionForm.value.notes
    // )
    message.success('成功加入持仓！')
    showAddPositionModal.value = false
  } catch (error) {
    message.error('加入持仓失败: ' + error)
  }
}

// 检查股票是否已关注
async function checkFollowStatus() {
  for (const rec of recommendations.value) {
    try {
      const followed = await CheckStockFollowed(rec.stock_code)
      rec.is_followed = followed
    } catch (error) {
      console.error('检查关注状态失败:', error)
    }
  }
}

// 获取荐股统计
async function getStats() {
  try {
    const stats = await GetStockPickStats()
    console.log('荐股统计:', stats)
  } catch (error) {
    console.error('获取统计失败:', error)
  }
}

// 清空输入框
function clearInput() {
  inputText.value = ''
  message.info('已清空输入')
}

// 清空聊天记录
function clearChat() {
  messages.value = [
    {
      id: 1,
      role: 'assistant',
      content: '请告诉我您的选股需求，例如：\n\n- 推荐今日资金流入大的科技股\n- 寻找市盈率低于20且业绩增长的银行股\n- 推荐近期创新高的新能源龙头股',
      timestamp: Date.now()
    }
  ]
  recommendations.value = []
  fullReport.value = ''
  reportId.value = null
  clearCache()
  message.success('已清空聊天记录')
}

// 复制最后一条AI消息
function copyLastMessage() {
  const aiMessages = messages.value.filter(m => m.role === 'assistant' && m.content)
  if (aiMessages.length > 0) {
    const lastMessage = aiMessages[aiMessages.length - 1]
    navigator.clipboard.writeText(lastMessage.content).then(() => {
      message.success('已复制到剪贴板')
    }).catch(() => {
      message.error('复制失败')
    })
  }
}

// 监听推荐股票变化，设置默认选中的技术面分析标签
watch(cardRecommendations, (newRecs) => {
  if (newRecs && newRecs.length > 0) {
    activeTechTab.value = newRecs[0].stockCode
  }
}, { immediate: true })

// 监听消息变化，自动滚动到底部
watch(messages, () => {
  scrollToBottom()
}, { deep: true })
</script>

<style scoped>
/* ========== 左侧对话区域新布局样式 ========== */

/* 列容器 */
.chat-column {
  height: 100%;
  min-height: 0;
}

/* 卡片容器 */
.chat-card {
  height: calc(100vh - 180px);
  display: flex;
  flex-direction: column;
}

.chat-card > :deep(.n-card__content) {
  flex: 1;
  overflow: hidden;
  padding: 16px;
  min-height: 0;
}

/* 主对话容器 - 使用 flex 垂直布局 */
.chat-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 工具列表区域 */
.tools-section {
  flex-shrink: 0;
  margin-bottom: 12px;
}

/* 消息列表区域 - 占据剩余空间并可滚动 */
.messages-section {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  margin-bottom: 12px;
  display: flex;
  flex-direction: column;
}

/* 原生滚动容器 - 使用原生滚动替代 n-scrollbar */
.messages-scroll-container {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: auto;
  /* 自定义滚动条样式 */
  scrollbar-width: thin;
  scrollbar-color: rgba(0,0,0,0.3) transparent;
}

/* Webkit 浏览器滚动条样式 */
.messages-scroll-container::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.messages-scroll-container::-webkit-scrollbar-track {
  background: transparent;
}

.messages-scroll-container::-webkit-scrollbar-thumb {
  background: rgba(0,0,0,0.2);
  border-radius: 3px;
}

.messages-scroll-container::-webkit-scrollbar-thumb:hover {
  background: rgba(0,0,0,0.3);
}

/* 右侧结果区域滚动条样式 */
.result-scrollbar {
  height: 100%;
  width: 100%;
}

.result-scrollbar > :deep(.n-scrollbar-container) {
  height: 100% !important;
}

.messages-content {
  padding: 4px;
  /* 确保内容正确流动 */
  min-width: 100%;
  box-sizing: border-box;
}

/* 底部输入区域 - 固定高度 */
.input-section {
  flex-shrink: 0;
}

.input-toolbar {
  margin-bottom: 8px;
}

.input-area {
  margin-bottom: 10px;
}

.send-button {
  margin-top: 0 !important;
}

/* 推荐结果内容区包装器 - 实现滚动 */
.result-content-wrapper {
  flex: 1;
  overflow: hidden;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.full-report-wrapper {
  flex: 1;
  overflow: hidden;
  min-height: 0;
  padding: 16px;
  display: flex;
  flex-direction: column;
}

.full-report-wrapper :deep(.md-editor-preview-wrapper) {
  padding-bottom: 40px;
}

.md-preview-content {
  max-width: 100%;
  box-sizing: border-box;
  padding-bottom: 60px;
}

.full-report-wrapper :deep(.md-editor-preview) {
  padding-bottom: 40px;
}

/* 滚动条容器 - 在flex布局中占据剩余空间 */
.scrollbar-container {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.simple-list-wrapper {
  flex: 1;
  overflow: auto;
  min-height: 0;
}

/* 卡片视图样式 */
.card-view-wrapper {
  flex: 1;
  overflow: hidden;
  min-height: 0;
  padding: 16px;
  display: flex;
  flex-direction: column;
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
  padding: 4px;
}

.stock-card {
  border-radius: 8px;
  transition: all 0.3s ease;
}

.stock-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

/* 卡片头部样式 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
}

.stock-title {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.stock-name {
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.stock-code {
  font-size: 12px;
  color: #999;
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
}

.suggestion-tag {
  font-weight: 500;
}

.stock-price {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.current-price {
  font-size: 20px;
  font-weight: 700;
}

.price-change {
  font-size: 12px;
  font-weight: 500;
}

.price-up {
  color: #d03050;
}

.price-down {
  color: #18a058;
}

/* 买卖点区域样式 */
.trading-points {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin: 12px 0;
  padding: 12px;
  background: #fafafa;
  border-radius: 8px;
}

.trading-point {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px;
  border-radius: 6px;
  background: white;
}

.point-label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #666;
}

.label-text {
  font-weight: 500;
}

.price-value {
  font-size: 16px;
  font-weight: 700;
}

.buy-price {
  color: #18a058;
}

.target-price {
  color: #2080f0;
}

.stop-loss-price {
  color: #d03050;
}

.expected-return {
  font-size: 11px;
  font-weight: 500;
  padding: 2px 6px;
  border-radius: 4px;
}

.expected-return.positive {
  color: #18a058;
  background: #e8f5e9;
}

.expected-return.negative {
  color: #d03050;
  background: #ffebee;
}

.stop-loss-rate {
  font-size: 11px;
  color: #d03050;
  margin-top: 2px;
}

/* 股票信息区域 */
.stock-info {
  margin: 12px 0;
}

.sector-concept {
  margin-bottom: 8px;
}

.reason {
  font-size: 13px;
  line-height: 1.6;
}

.reason-text {
  color: #333;
}

/* 风险提示区域 */
.risk-warning {
  margin: 8px 0;
}

/* 备注区域 */
.remarks {
  margin: 8px 0;
  padding: 8px 12px;
  background: #f5f5f5;
  border-radius: 4px;
}

.remarks-text {
  font-size: 12px;
  font-style: italic;
}

/* 分析部分样式 */
.analysis-section {
  margin: 8px 0;
}

.analysis-text {
  font-size: 13px;
  line-height: 1.7;
  color: #444;
  display: block;
}

/* 推荐时间 */
.recommend-time {
  margin: 8px 0;
  text-align: right;
}

.time-text {
  font-size: 11px;
  color: #999;
}

/* 卡片底部 */
.card-footer {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

/* 响应式调整 */
@media (max-width: 1200px) {
  .card-grid {
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  }
}

@media (max-width: 768px) {
  .card-grid {
    grid-template-columns: 1fr;
  }

  .trading-points {
    grid-template-columns: 1fr;
  }
}

/* ========== Markdown Preview 优化样式 ========== */
.md-preview-content {
  max-width: 100%;
  box-sizing: border-box;
  padding-bottom: 60px;
  line-height: 1.8;
  color: #333;
  font-size: 14px;
  text-align: left !important;
}

/* 确保所有内容左对齐，不居中 - 最高优先级 */
.md-preview-content :deep(.md-editor-preview),
.full-report-wrapper :deep(.md-editor-preview),
.md-preview-content :deep(.md-editor-preview-wrapper),
.full-report-wrapper :deep(.md-editor-preview-wrapper),
.md-preview-content :deep(.md-editor-preview *),
.full-report-wrapper :deep(.md-editor-preview *),
.md-preview-content :deep(.md-editor-preview-wrapper *),
.full-report-wrapper :deep(.md-editor-preview-wrapper *) {
  text-align: left !important;
}

/* 强制覆盖可能的居中样式 */
.md-preview-content :deep(.md-editor-preview h1),
.md-preview-content :deep(.md-editor-preview h2),
.md-preview-content :deep(.md-editor-preview h3),
.md-preview-content :deep(.md-editor-preview h4),
.md-preview-content :deep(.md-editor-preview h5),
.md-preview-content :deep(.md-editor-preview h6),
.md-preview-content :deep(.md-editor-preview p),
.md-preview-content :deep(.md-editor-preview li),
.md-preview-content :deep(.md-editor-preview div),
.md-preview-content :deep(.md-editor-preview blockquote),
.md-preview-content :deep(.md-editor-preview pre),
.md-preview-content :deep(.md-editor-preview table),
.md-preview-content :deep(.md-editor-preview ul),
.md-preview-content :deep(.md-editor-preview ol) {
  text-align: left !important;
}

/* 标题层级优化 */
.md-preview-content :deep(h1) {
  font-size: 22px;
  font-weight: 700;
  margin: 24px 0 16px 0;
  padding-bottom: 8px;
  border-bottom: 2px solid #e8e8e8;
  color: #1a1a1a;
}

.md-preview-content :deep(h2) {
  font-size: 18px;
  font-weight: 600;
  margin: 20px 0 12px 0;
  padding-bottom: 6px;
  border-bottom: 1px solid #f0f0f0;
  color: #2a2a2a;
}

.md-preview-content :deep(h3) {
  font-size: 16px;
  font-weight: 600;
  margin: 16px 0 10px 0;
  color: #333;
}

.md-preview-content :deep(h4),
.md-preview-content :deep(h5),
.md-preview-content :deep(h6) {
  font-size: 14px;
  font-weight: 500;
  margin: 12px 0 8px 0;
  color: #444;
}

/* 段落样式 */
.md-preview-content :deep(p) {
  margin: 10px 0;
  line-height: 1.8;
  text-align: left;
}

/* 表格美化 */
.md-preview-content :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 16px 0;
  font-size: 13px;
}

.md-preview-content :deep(th) {
  background: #f5f5f5;
  font-weight: 600;
  text-align: left;
  padding: 10px 12px;
  border: 1px solid #e0e0e0;
  color: #333;
}

.md-preview-content :deep(td) {
  padding: 10px 12px;
  border: 1px solid #e0e0e0;
  text-align: left;
}

.md-preview-content :deep(tr:nth-child(even)) {
  background: #fafafa;
}

.md-preview-content :deep(tr:hover) {
  background: #f0f7ff;
}

/* 代码块高亮 */
.md-preview-content :deep(pre) {
  background: #1e1e1e;
  border-radius: 8px;
  padding: 16px;
  margin: 16px 0;
  overflow-x: auto;
}

.md-preview-content :deep(pre code) {
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
}

.md-preview-content :deep(code) {
  background: #f4f4f4;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  color: #d03050;
}

.md-preview-content :deep(pre code) {
  background: transparent;
  padding: 0;
  color: inherit;
}

/* 列表优化 */
.md-preview-content :deep(ul),
.md-preview-content :deep(ol) {
  margin: 12px 0;
  padding-left: 24px;
}

.md-preview-content :deep(li) {
  margin: 6px 0;
  line-height: 1.7;
}

.md-preview-content :deep(ul li) {
  list-style-type: disc;
}

.md-preview-content :deep(ol li) {
  list-style-type: decimal;
}

.md-preview-content :deep(ul ul),
.md-preview-content :deep(ol ol),
.md-preview-content :deep(ul ol),
.md-preview-content :deep(ol ul) {
  margin: 6px 0;
  padding-left: 20px;
}

/* 引用块样式 */
.md-preview-content :deep(blockquote) {
  border-left: 4px solid #2080f0;
  background: #f0f7ff;
  padding: 12px 16px;
  margin: 16px 0;
  border-radius: 0 8px 8px 0;
  color: #444;
}

.md-preview-content :deep(blockquote p) {
  margin: 0;
}

/* 分隔线样式 */
.md-preview-content :deep(hr) {
  border: none;
  height: 1px;
  background: #e8e8e8;
  margin: 24px 0;
}

/* 链接样式 */
.md-preview-content :deep(a) {
  color: #2080f0;
  text-decoration: none;
  border-bottom: 1px solid transparent;
  transition: all 0.2s;
}

.md-preview-content :deep(a:hover) {
  color: #18a058;
  border-bottom-color: #18a058;
}

/* 强调样式 */
.md-preview-content :deep(strong),
.md-preview-content :deep(b) {
  font-weight: 600;
  color: #1a1a1a;
}

.md-preview-content :deep(em),
.md-preview-content :deep(i) {
  font-style: italic;
  color: #555;
}

/* 删除线 */
.md-preview-content :deep(s),
.md-preview-content :deep(del) {
  text-decoration: line-through;
  color: #999;
}

/* 图片样式 */
.md-preview-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  margin: 12px 0;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

/* 夜间主题适配 */
.md-preview-content.dark {
  color: #e0e0e0;
}

.md-preview-content.dark :deep(h1),
.md-preview-content.dark :deep(h2),
.md-preview-content.dark :deep(h3) {
  color: #f0f0f0;
  border-color: #444;
}

.md-preview-content.dark :deep(table) {
  border-color: #555;
}

.md-preview-content.dark :deep(th) {
  background: #333;
  border-color: #555;
  color: #f0f0f0;
}

.md-preview-content.dark :deep(td) {
  border-color: #555;
  color: #e0e0e0;
}

.md-preview-content.dark :deep(tr:nth-child(even)) {
  background: #2a2a2a;
}

.md-preview-content.dark :deep(tr:hover) {
  background: #333;
}

.md-preview-content.dark :deep(blockquote) {
  background: #2a2a2a;
  border-color: #2080f0;
  color: #e0e0e0;
}

.md-preview-content.dark :deep(code) {
  background: #333;
  color: #f0a020;
}

.md-preview-content.dark :deep(pre) {
  background: #1a1a1a;
}

.md-preview-content.dark :deep(pre code) {
  color: #e0e0e0;
}

.md-preview-content.dark :deep(a) {
  color: #40a0ff;
}

.md-preview-content.dark :deep(a:hover) {
  color: #60c0ff;
  border-bottom-color: #60c0ff;
}

.md-preview-content.dark :deep(hr) {
  background: #444;
}

/* ========== 对话UI样式 ========== */
.message-item {
  margin-bottom: 16px;
  opacity: 0;
  animation: messageSlideIn 0.3s ease forwards;
}

@keyframes messageSlideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message-wrapper {
  display: flex;
  margin-bottom: 4px;
}

.message-wrapper.user-message {
  justify-content: flex-end;
}

.message-wrapper.ai-message {
  justify-content: flex-start;
}

.message-content-wrapper {
  display: flex;
  align-items: flex-start;
  max-width: 90%;
  gap: 10px;
}

.message-wrapper.user-message .message-content-wrapper {
  flex-direction: row-reverse;
}

.message-avatar {
  flex-shrink: 0;
  margin-top: 4px;
}

.message-bubble {
  padding: 12px 16px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: box-shadow 0.2s ease;
}

.message-bubble:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.user-bubble {
  background: linear-gradient(135deg, #2080f0 0%, #18a058 100%);
  color: white;
  border-bottom-right-radius: 4px;
}

.user-bubble .message-text {
  color: white;
  line-height: 1.6;
}

.ai-bubble {
  background: #ffffff;
  color: #333;
  border-bottom-left-radius: 4px;
  border: 1px solid #f0f0f0;
}

.ai-bubble :deep(.md-editor-preview) {
  font-size: 14px;
  line-height: 1.7;
}

.ai-bubble :deep(.md-editor-preview p) {
  margin: 8px 0;
}

/* 确保 Markdown 内容不会破坏滚动 */
.ai-bubble :deep(.md-editor-preview-wrapper) {
  max-width: 100%;
  overflow-x: auto;
}

.ai-bubble :deep(.md-editor-preview) {
  word-wrap: break-word;
  overflow-wrap: break-word;
}

/* 确保代码块可以水平滚动 */
.ai-bubble :deep(pre) {
  overflow-x: auto;
  max-width: 100%;
}

/* 确保表格可以水平滚动 */
.ai-bubble :deep(.md-editor-preview-wrapper table) {
  display: block;
  overflow-x: auto;
}

.typing-indicator {
  padding: 16px 20px;
  min-width: 80px;
}

.typing-dots {
  display: flex;
  gap: 6px;
  align-items: center;
}

.typing-dots span {
  width: 8px;
  height: 8px;
  background: #18a058;
  border-radius: 50%;
  animation: typingBounce 1.4s infinite ease-in-out both;
}

.typing-dots span:nth-child(1) {
  animation-delay: -0.32s;
}

.typing-dots span:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes typingBounce {
  0%, 80%, 100% {
    transform: scale(0.8);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

.input-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 8px;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.input-toolbar:hover {
  opacity: 1;
}

.input-area {
  margin-bottom: 8px;
}

.chat-input {
  border-radius: 12px;
  transition: all 0.2s;
}

.chat-input:hover,
.chat-input:focus-within {
  box-shadow: 0 0 0 2px rgba(32, 128, 240, 0.2);
}

.send-button {
  border-radius: 10px;
  font-weight: 500;
  transition: all 0.2s;
}

.send-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(32, 128, 240, 0.3);
}

.send-button:active {
  transform: translateY(0);
}

/* 深色主题适配 */
@media (prefers-color-scheme: dark) {
  .ai-bubble {
    background: #1a1a1a;
    border-color: #333;
  }

  .ai-bubble :deep(.md-editor-preview) {
    color: #e0e0e0;
  }
}
</style>
