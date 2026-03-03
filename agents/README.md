# 🤖 OpenClaw + Claude Code Agent 集群系统

> 基于文章《OpenClaw + Claude Code 超强教程》的实战实现

## 📋 系统架构

这套系统实现了**一个人搭建完整开发团队**的效果，通过 OpenClaw 编排多个 Claude Code 子代理，实现高效的 AI 驱动开发。

```
┌─────────────────────────────────────────────────────────────┐
│                    🎛️ OpenClaw 编排中枢                       │
│              任务调度 · 上下文管理 · 结果聚合                   │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│ 数据获取代理  │    │ 分析推理代理  │    │ 选股策略代理  │
│  (Claude)    │    │   (Claude)   │    │   (Claude)   │
└───────────────┘    └───────────────┘    └───────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              ▼
                    ┌───────────────┐
                    │ 报告生成代理  │
                    │   (Claude)   │
                    └───────────────┘
```

## 🎯 核心功能

### 1. 智能荐股工作流
8个步骤的完整荐股流程：
1. **收集市场数据** - 获取行情、新闻、资金流向
2. **收集个股数据** - 获取关注股票的详细数据
3. **分析市场趋势** - AI分析市场整体趋势和情绪
4. **识别热点板块** - 分析当前市场热点
5. **筛选优质股票** - 根据策略筛选符合条件的股票
6. **优化股票组合** - 控制风险，优化选股结果
7. **生成荐股报告** - 生成专业的荐股分析报告
8. **格式化买卖点建议** - 格式化买卖点建议，便于展示

### 2. 子代理集群

#### 数据获取代理 (data-collector)
- **职责**: 获取股票实时数据、K线、资讯等
- **能力**: 
  - 获取股票实时价格
  - 获取K线历史数据
  - 获取市场资讯
  - 获取资金流向数据
  - 获取龙虎榜数据

#### 市场分析代理 (market-analyst)
- **职责**: 基于AI分析市场整体趋势和情绪
- **能力**:
  - 分析市场趋势
  - 识别热点板块
  - 情绪分析
  - 风险评估

#### 选股策略代理 (stock-picker)
- **职责**: 根据策略筛选优质股票
- **能力**:
  - 技术指标筛选
  - 基本面分析
  - 多因子选股
  - 组合优化

#### 报告生成代理 (report-generator)
- **职责**: 生成专业的荐股分析报告
- **能力**:
  - 报告撰写
  - 买卖点格式化
  - 可视化图表生成
  - 风险提示

## 🚀 快速开始

### 1. 启动 OpenClaw 网关

确保 OpenClaw 网关正在运行：

```bash
openclaw gateway start
```

### 2. 执行智能荐股工作流

#### 方法一：使用 OpenClaw CLI

```bash
# 执行荐股工作流
openclaw workflow run stock-pick --input '{"market": "A", "strategy": "multiFactor"}'
```

#### 方法二：使用 Claude Code 直接调用

由于你只安装了 Claude Code，我们可以通过以下方式实现：

```bash
# 1. 启动数据收集代理
claude-code --agent agents/sub-agents/data-collector.yaml

# 2. 启动市场分析代理
claude-code --agent agents/sub-agents/market-analyst.yaml

# 3. 启动选股策略代理
claude-code --agent agents/sub-agents/stock-picker.yaml

# 4. 启动报告生成代理
claude-code --agent agents/sub-agents/report-generator.yaml
```

#### 方法三：使用 Python 脚本编排 (推荐)

创建一个 Python 脚本来编排所有子代理：

```python
# orchestrator.py
import subprocess
import json
import asyncio
from concurrent.futures import ThreadPoolExecutor

class OpenClawOrchestrator:
    """OpenClaw 编排器 - 调度多个 Claude Code 子代理"""
    
    def __init__(self):
        self.agents = {}
        self.workflow_results = {}
    
    def register_agent(self, name, config_path):
        """注册子代理"""
        self.agents[name] = {
            'config': config_path,
            'status': 'idle'
        }
        print(f"✅ 注册代理: {name}")
    
    async def execute_agent(self, agent_name, input_data):
        """执行子代理"""
        agent = self.agents[agent_name]
        agent['status'] = 'running'
        
        # 构建 Claude Code 命令
        cmd = [
            'claude-code',
            '--agent', agent['config'],
            '--input', json.dumps(input_data),
            '--output-format', 'json'
        ]
        
        try:
            # 执行命令
            process = await asyncio.create_subprocess_exec(
                *cmd,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )
            
            stdout, stderr = await process.communicate()
            
            if process.returncode == 0:
                result = json.loads(stdout.decode())
                agent['status'] = 'completed'
                print(f"✅ 代理 {agent_name} 执行完成")
                return result
            else:
                error_msg = stderr.decode()
                agent['status'] = 'failed'
                print(f"❌ 代理 {agent_name} 执行失败: {error_msg}")
                raise Exception(f"Agent {agent_name} failed: {error_msg}")
                
        except Exception as e:
            agent['status'] = 'failed'
            print(f"❌ 代理 {agent_name} 执行异常: {str(e)}")
            raise
    
    async def run_workflow(self, workflow_name, initial_input):
        """执行工作流"""
        print(f"\n🚀 开始执行工作流: {workflow_name}\n")
        
        # 这里根据工作流名称执行不同的步骤序列
        if workflow_name == "stock-pick":
            return await self._run_stock_pick_workflow(initial_input)
        else:
            raise ValueError(f"Unknown workflow: {workflow_name}")
    
    async def _run_stock_pick_workflow(self, initial_input):
        """执行荐股工作流"""
        results = {}
        
        # Step 1: 收集市场数据
        print("📊 Step 1: 收集市场数据...")
        market_data = await self.execute_agent('data-collector', {
            'action': 'fetchMarketData',
            'input': initial_input
        })
        results['marketData'] = market_data
        
        # Step 2: 收集个股数据
        print("📈 Step 2: 收集个股数据...")
        stock_data = await self.execute_agent('data-collector', {
            'action': 'fetchStockData',
            'input': initial_input
        })
        results['stockData'] = stock_data
        
        # Step 3: 市场分析
        print("🧠 Step 3: 分析市场趋势...")
        market_analysis = await self.execute_agent('market-analyst', {
            'action': 'analyzeMarket',
            'input': {
                'marketData': market_data,
                'analysisType': ['trend', 'sentiment', 'hotSectors']
            }
        })
        results['marketAnalysis'] = market_analysis
        
        # Step 4: 识别热点板块
        print("🔥 Step 4: 识别热点板块...")
        hot_sectors = await self.execute_agent('market-analyst', {
            'action': 'identifyHotSectors',
            'input': {
                'marketData': market_data,
                'marketAnalysis': market_analysis
            }
        })
        results['hotSectors'] = hot_sectors
        
        # Step 5: 选股策略
        print("🎯 Step 5: 筛选优质股票...")
        stock_picks = await self.execute_agent('stock-picker', {
            'action': 'screenStocks',
            'input': {
                'stockData': stock_data,
                'marketAnalysis': market_analysis,
                'hotSectors': hot_sectors,
                'strategy': 'multiFactor'
            }
        })
        results['stockPicks'] = stock_picks
        
        # Step 6: 组合优化
        print("⚖️ Step 6: 优化股票组合...")
        optimized_stocks = await self.execute_agent('stock-picker', {
            'action': 'optimizePortfolio',
            'input': {
                'stockPicks': stock_picks,
                'constraints': {
                    'maxStocks': 10,
                    'riskLevel': 'medium',
                    'diversification': true
                }
            }
        })
        results['optimizedStocks'] = optimized_stocks
        
        # Step 7: 生成报告
        print("📝 Step 7: 生成荐股报告...")
        report = await self.execute_agent('report-generator', {
            'action': 'generateReport',
            'input': {
                'optimizedStocks': optimized_stocks,
                'marketAnalysis': market_analysis,
                'hotSectors': hot_sectors,
                'reportTemplate': 'professional'
            }
        })
        results['report'] = report
        
        # Step 8: 格式化输出
        print("✨ Step 8: 格式化买卖点建议...")
        formatted_report = await self.execute_agent('report-generator', {
            'action': 'formatSuggestions',
            'input': {
                'report': report,
                'format': {
                    'type': 'table',
                    'includeReason': true,
                    'includeRisk': true
                }
            }
        })
        results['formattedReport'] = formatted_report
        
        print("\n✅ 荐股工作流执行完成！")
        return results
    
    def get_results(self):
        """获取所有执行结果"""
        return self.workflow_results


# ========== 使用示例 ==========

async def main():
    """主函数示例"""
    # 创建编排器
    orchestrator = OpenClawOrchestrator()
    
    # 注册子代理
    orchestrator.register_agent('data-collector', 'agents/sub-agents/data-collector.yaml')
    orchestrator.register_agent('market-analyst', 'agents/sub-agents/market-analyst.yaml')
    orchestrator.register_agent('stock-picker', 'agents/sub-agents/stock-picker.yaml')
    orchestrator.register_agent('report-generator', 'agents/sub-agents/report-generator.yaml')
    
    # 执行荐股工作流
    try:
        results = await orchestrator.run_workflow('stock-pick', {
            'market': 'A',
            'strategy': 'multiFactor',
            'riskLevel': 'medium'
        })
        
        # 输出最终结果
        print("\n" + "="*60)
        print("🎯 荐股报告生成完成")
        print("="*60)
        print(json.dumps(results.get('formattedReport'), indent=2, ensure_ascii=False))
        
    except Exception as e:
        print(f"❌ 工作流执行失败: {str(e)}")
        raise


if __name__ == "__main__":
    # 运行主函数
    asyncio.run(main())
