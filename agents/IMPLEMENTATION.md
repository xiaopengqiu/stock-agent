# OpenClaw + Claude Code Agent 集群实现方案

## 核心架构

基于你分享的实践案例，我们使用 **OpenClaw 作为编排层**，调度多个 **Claude Code 子代理** 形成完整的开发团队。

## 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                    🎛️ 编排中枢 (OpenClaw)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐      │
│  │  任务调度    │  │  上下文管理  │  │  会话状态管理    │      │
│  └──────────────┘  └──────────────┘  └──────────────────┘      │
└─────────────────────────────────────────────────────────────────┘
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

## 实现步骤

### 第一步：创建子代理配置文件

我们已经创建了:
1. ✅ `agents/orchestrator.yaml` - 编排中枢配置
2. ✅ `agents/sub-agents/data-collector.yaml` - 数据获取代理

还需要创建:
3. `agents/sub-agents/market-analyst.yaml` - 市场分析代理
4. `agents/sub-agents/stock-picker.yaml` - 选股策略代理
5. `agents/sub-agents/report-generator.yaml` - 报告生成代理

### 第二步：创建 OpenClaw 工作流配置

创建 `agents/workflows/stock-pick.yaml`:

```yaml
name: 智能荐股工作流
description: 完整的AI荐股流程，从市场分析到生成报告

steps:
  # 步骤1: 数据收集
  - id: collect_data
    name: 收集市场数据
    agent: data-collector
    action: fetchMarketData
    output: marketData
    
  # 步骤2: 市场分析
  - id: analyze_market
    name: 分析市场趋势
    agent: market-analyst
    action: analyzeMarket
    input: marketData
    output: marketAnalysis
    dependsOn: collect_data
    
  # 步骤3: 选股策略
  - id: pick_stocks
    name: 执行选股策略
    agent: stock-picker
    action: screenStocks
    input: marketAnalysis
    output: stockPicks
    dependsOn: analyze_market
    
  # 步骤4: 生成报告
  - id: generate_report
    name: 生成荐股报告
    agent: report-generator
    action: generateReport
    input: stockPicks
    output: finalReport
    dependsOn: pick_stocks
```

### 第三步：集成到现有代码库

#### 3.1 创建 Agent 调度器

在 `backend/agent/` 下创建 `scheduler.go`:

```go
package agent

import (
    "context"
    "fmt"
    "go-stock/backend/logger"
    "sync"
    "time"
)

// Agent 定义子代理接口
type Agent interface {
    Name() string
    Execute(ctx context.Context, input interface{}) (interface{}, error)
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
    ID       string
    Name     string
    Agent    Agent
    Input    interface{}
    Output   interface{}
    DependsOn []string
    Status   string // pending, running, completed, failed
    Error    error
}

// Workflow 工作流
type Workflow struct {
    Name  string
    Steps map[string]*WorkflowStep
    mu    sync.RWMutex
}

// Scheduler Agent调度器
type Scheduler struct {
    agents    map[string]Agent
    workflows map[string]*Workflow
    mu        sync.RWMutex
}

// NewScheduler 创建调度器
func NewScheduler() *Scheduler {
    return &Scheduler{
        agents:    make(map[string]Agent),
        workflows: make(map[string]*Workflow),
    }
}

// RegisterAgent 注册代理
func (s *Scheduler) RegisterAgent(agent Agent) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.agents[agent.Name()] = agent
    logger.SugaredLogger.Infof("Registered agent: %s", agent.Name())
}

// GetAgent 获取代理
func (s *Scheduler) GetAgent(name string) (Agent, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    agent, ok := s.agents[name]
    return agent, ok
}

// ExecuteWorkflow 执行工作流
func (s *Scheduler) ExecuteWorkflow(ctx context.Context, workflowName string, initialInput interface{}) (map[string]interface{}, error) {
    workflow, ok := s.workflows[workflowName]
    if !ok {
        return nil, fmt.Errorf("workflow not found: %s", workflowName)
    }

    // 拓扑排序确定执行顺序
    executionOrder, err := s.topologicalSort(workflow)
    if err != nil {
        return nil, err
    }

    results := make(map[string]interface{})
    results["initial"] = initialInput

    // 执行每个步骤
    for _, stepID := range executionOrder {
        step := workflow.Steps[stepID]
        
        // 准备输入
        var input interface{}
        if step.Input != nil {
            input = step.Input
        } else if len(step.DependsOn) > 0 {
            // 合并依赖步骤的输出
            mergedInput := make(map[string]interface{})
            for _, depID := range step.DependsOn {
                if depStep, ok := workflow.Steps[depID]; ok && depStep.Output != nil {
                    mergedInput[depID] = depStep.Output
                }
            }
            input = mergedInput
        } else {
            input = initialInput
        }

        // 执行步骤
        step.Status = "running"
        output, err := step.Agent.Execute(ctx, input)
        
        if err != nil {
            step.Status = "failed"
            step.Error = err
            logger.SugaredLogger.Errorf("Step %s failed: %v", step.Name, err)
            return results, err
        }
        
        step.Status = "completed"
        step.Output = output
        results[stepID] = output
        
        logger.SugaredLogger.Infof("Step %s completed successfully", step.Name)
    }

    return results, nil
}

// topologicalSort 拓扑排序
func (s *Scheduler) topologicalSort(workflow *Workflow) ([]string, error) {
    inDegree := make(map[string]int)
    graph := make(map[string][]string)
    
    // 初始化
    for id := range workflow.Steps {
        inDegree[id] = 0
        graph[id] = []string{}
    }
    
    // 构建图
    for id, step := range workflow.Steps {
        for _, dep := range step.DependsOn {
            graph[dep] = append(graph[dep], id)
            inDegree[id]++
        }
    }
    
    // Kahn算法
    var result []string
    queue := make([]string, 0)
    
    for id, degree := range inDegree {
        if degree == 0 {
            queue = append(queue, id)
        }
    }
    
    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]
        result = append(result, current)
        
        for _, neighbor := range graph[current] {
            inDegree[neighbor]--
            if inDegree[neighbor] == 0 {
                queue = append(queue, neighbor)
            }
        }
    }
    
    if len(result) != len(workflow.Steps) {
        return nil, fmt.Errorf("workflow has circular dependencies")
    }
    
    return result, nil
}

// CreateStockPickWorkflow 创建荐股工作流
func (s *Scheduler) CreateStockPickWorkflow() *Workflow {
    workflow := &Workflow{
        Name:  "智能荐股工作流",
        Steps: make(map[string]*WorkflowStep),
    }
    
    // 这里会在初始化时从配置中加载
    return workflow
}