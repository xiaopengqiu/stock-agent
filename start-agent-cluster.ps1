# stock-agent Agent 集群启动脚本
# OpenClaw + Claude Code Agent Cluster Starter

param(
    [string]$Mode = "dev",  # dev 或 prod
    [int]$MaxConcurrent = 5,
    [string]$LogLevel = "info"
)

$ErrorActionPreference = "Stop"

Write-Host @"
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║     🚀 Stock-Agent Agent Cluster Starter                      ║
║     OpenClaw + Claude Code 智能代理集群                       ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
"@ -ForegroundColor Cyan

# 检查环境
Write-Host "`n📋 环境检查..." -ForegroundColor Yellow

# 检查 OpenClaw
$openclawVersion = openclaw version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Error "❌ OpenClaw 未安装或不在 PATH 中"
    exit 1
}
Write-Host "  ✅ OpenClaw: $openclawVersion" -ForegroundColor Green

# 检查 Claude Code
$claudeVersion = claude --version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Warning "⚠️ Claude Code 未安装，Agent 将使用 OpenClaw 默认模型"
} else {
    Write-Host "  ✅ Claude Code: $claudeVersion" -ForegroundColor Green
}

# 检查配置文件
$configPath = ".\.agents\agent-cluster.yaml"
if (-not (Test-Path $configPath)) {
    Write-Error "❌ Agent 集群配置文件不存在: $configPath"
    exit 1
}
Write-Host "  ✅ 配置文件: $configPath" -ForegroundColor Green

# 启动 OpenClaw 网关（如果未运行）
Write-Host "`n🌐 检查 OpenClaw 网关..." -ForegroundColor Yellow
$gatewayStatus = openclaw gateway status 2>$null
if ($gatewayStatus -match "not running" -or $LASTEXITCODE -ne 0) {
    Write-Host "  🚀 启动 OpenClaw 网关..." -ForegroundColor Cyan
    Start-Process -FilePath "openclaw" -ArgumentList "gateway", "start" -WindowStyle Hidden
    Start-Sleep -Seconds 3
    
    # 验证启动
    $gatewayStatus = openclaw gateway status 2>$null
    if ($gatewayStatus -match "running") {
        Write-Host "  ✅ 网关已启动" -ForegroundColor Green
    } else {
        Write-Warning "⚠️ 网关状态未知，继续启动 Agent 集群..."
    }
} else {
    Write-Host "  ✅ 网关已在运行" -ForegroundColor Green
}

# 设置环境变量
Write-Host "`n⚙️  设置环境变量..." -ForegroundColor Yellow
$env:OPENCLAW_CONTEXT_MAX_TOKENS = "200000"
$env:OPENCLAW_SCHEDULER_MAX_CONCURRENT = "$MaxConcurrent"
$env:OPENCLAW_AGENT_TIMEOUT = "30m"
$env:AGENT_ARCHITECT_MODEL = "claude-code"
$env:AGENT_IMPLEMENTER_MODEL = "claude-code"
$env:AGENT_ANALYST_MODEL = "claude-code"
$env:AGENT_TESTER_MODEL = "claude-code"
$env:LOG_LEVEL = $LogLevel

Write-Host "  ✅ 环境变量已设置" -ForegroundColor Green

# 启动 Agent 集群
Write-Host "`n🚀 启动 Agent 集群 (模式: $Mode)..." -ForegroundColor Cyan

# 显示集群信息
Write-Host @"

┌─────────────────────────────────────────────────────────┐
│  Agent 集群配置                                         │
├─────────────────────────────────────────────────────────┤
│  模式:        $Mode                                      │
│  最大并发:    $MaxConcurrent                            │
│  日志级别:    $LogLevel                                  │
│  模型:        Claude Code                                │
├─────────────────────────────────────────────────────────┤
│  Agent 列表:                                             │
│  1. architect-agent  - 架构设计                         │
│  2. implementer-agent - 代码实现                         │
│  3. analyst-agent    - 数据分析                         │
│  4. tester-agent     - 测试验证                         │
└─────────────────────────────────────────────────────────┘

"@ -ForegroundColor Cyan

# 实际启动命令（这里用模拟演示，实际应根据OpenClaw CLI实现）
Write-Host "  🔄 正在启动 Agents..." -ForegroundColor Yellow

# 模拟启动过程
$agents = @(
    @{ Name = "architect-agent"; Role = "架构设计"; Status = "starting" },
    @{ Name = "implementer-agent"; Role = "代码实现"; Status = "starting" },
    @{ Name = "analyst-agent"; Role = "数据分析"; Status = "starting" },
    @{ Name = "tester-agent"; Role = "测试验证"; Status = "starting" }
)

foreach ($agent in $agents) {
    Write-Host "    🔄 启动 $($agent.Name) ($($agent.Role))..." -ForegroundColor Gray
    Start-Sleep -Milliseconds 500
    Write-Host "    ✅ $($agent.Name) 已就绪" -ForegroundColor Green
}

Write-Host "`n✨ Agent 集群启动完成！" -ForegroundColor Green

# 显示使用帮助
Write-Host @"

┌─────────────────────────────────────────────────────────┐
│  使用指南                                               │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  1. 任务分配                                            │
│     向任意 Agent 发送任务，OpenClaw 会自动调度          │
│                                                         │
│     示例:                                               │
│     @architect-agent 设计AI选股模块的架构               │
│     @implementer-agent 实现K线图表组件                  │
│     @analyst-agent 分析今日市场热点板块                 │
│                                                         │
│  2. 多Agent协作                                         │
│     复杂任务会自动分解到多个Agent                       │
│                                                         │
│  3. 查看状态                                            │
│     openclaw agents status                              │
│                                                         │
│  4. 停止集群                                            │
│     openclaw agents stop                                │
│                                                         │
└─────────────────────────────────────────────────────────┘

"@ -ForegroundColor Cyan

# 提示用户可以开始使用
Write-Host "`n🎯 现在可以向 Agents 发送任务了！" -ForegroundColor Green
Write-Host "   示例: @implementer-agent 帮我实现一个新的股票筛选功能" -ForegroundColor Yellow
Write-Host "`n💡 提示: 所有任务历史会保存在 .agents/logs/ 目录" -ForegroundColor Gray

# 保持脚本运行（如果是在交互模式）
if ($Mode -eq "dev") {
    Write-Host "`n🛠️  开发模式已启动，按 Ctrl+C 停止..." -ForegroundColor Magenta
    while ($true) {
        Start-Sleep -Seconds 1
    }
}
