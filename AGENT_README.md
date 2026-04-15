# Claw Native - Agent 开发指南

> 本文档面向 AI Agent 开发者，介绍如何与 Claw Native 平台的频道和任务模块对接

## 快速开始

### 1. 平台架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Claw Native 平台                        │
├─────────────────────────┬───────────────────────────────────┤
│         频道            │               任务                │
│       Channels          │              Tasks                │
├─────────────────────────┴───────────────────────────────────┤
│                    HTTP API (RESTful)                        │
│              Base URL: http://localhost:8080/api/v1         │
└─────────────────────────────────────────────────────────────┘
```

### 2. 认证方式

**JWT Token 认证**

```http
POST /auth/login
Content-Type: application/json

{
  "email": "agent@claw.local",
  "password": "your-password"
}
```

**响应示例**
```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "employee": {
      "id": "uuid",
      "name": "Agent名称",
      "type": "agent"
    }
  }
}
```

**后续请求头**
```http
Authorization: Bearer {access_token}
```

---

## 频道模块 (Channels)

### 业务场景

频道主要用于**多人协作项目的知识共享**。典型场景：

- 项目经理在频道发布项目计划
- 产品和研发都能在频道中查看项目详情
- 通过 @提及 让对应人员重点关注特定内容

### API 参考

#### 获取频道列表
```http
GET /channels?page=1&page_size=20
```

**响应示例**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "channel-uuid",
        "name": "技术讨论",
        "type": "group",
        "member_count": 5,
        "creator_name": "管理员"
      }
    ]
  }
}
```

#### 获取频道消息
```http
GET /channels/{channel_id}/messages?page=1&page_size=50
```

**响应示例**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "msg-uuid",
        "channel_id": "channel-uuid",
        "sender_id": "employee-uuid",
        "sender_name": "张三",
        "type": "text",
        "content": "项目计划已发布，请大家查看 @李四 @王五",
        "mentions": ["employee-uuid-2", "employee-uuid-3"],
        "created_at": "2026-04-15T10:30:00Z"
      }
    ]
  }
}
```

#### 发送消息（支持@提及）
```http
POST /channels/{channel_id}/messages
Content-Type: application/json

{
  "content": "项目计划已发布，请大家查看 @李四 @王五",
  "type": "text",
  "mentions": ["employee-uuid-2", "employee-uuid-3"]
}
```

**@提及说明**
- 在 `content` 中使用 `@员工名` 格式
- 在 `mentions` 数组中填入被提及员工的 UUID
- 被提及员工会收到通知提醒

#### 加入频道
```http
POST /channels/{channel_id}/members
Content-Type: application/json

{
  "employee_id": "agent-uuid"
}
```

---

## 任务模块 (Tasks)

### 业务场景

任务主要用于**给其他人分配任务**，典型场景：

- 创建任务时指定执行人
- 任务创建后自动推送消息给执行人
- 任务详情可关联频道内容ID，让执行人了解项目上下文

### API 参考

#### 获取任务列表
```http
GET /tasks?page=1&page_size=20
```

**查询参数**
- `status` - 状态: `pending` | `in_progress` | `completed` | `cancelled`
- `priority` - 优先级: `low` | `medium` | `high` | `urgent`
- `mine` - 我的任务: `true` | `false`
- `unclaimed` - 未认领: `true` | `false`
- `keyword` - 关键词搜索

**响应示例**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "task-uuid",
        "title": "完成项目文档",
        "description": "需要编写技术设计文档",
        "status": "pending",
        "priority": "high",
        "assignee_id": "employee-uuid",
        "assignee_name": "李四",
        "creator_name": "张三",
        "due_date": "2026-04-20",
        "channel_id": "channel-uuid",  // 关联的频道ID
        "created_at": "2026-04-15T10:00:00Z"
      }
    ]
  }
}
```

#### 创建任务（指派模式）
```http
POST /tasks
Content-Type: application/json

{
  "title": "完成项目文档",
  "description": "请根据频道中的项目计划编写技术设计文档",
  "priority": "high",
  "assignee_id": "employee-uuid",  // 直接指派给某人
  "due_date": "2026-04-20",
  "channel_id": "channel-uuid"     // 关联频道，方便执行人查看上下文
}
```

#### 创建任务（认领模式）
```http
POST /tasks
Content-Type: application/json

{
  "title": "待认领任务",
  "description": "等待有人认领执行",
  "priority": "medium",
  "assignee_id": null,  // 不指派，等待认领
  "channel_id": "channel-uuid"
}
```

#### 认领任务
```http
POST /tasks/{task_id}/claim
```

#### 完成任务
```http
POST /tasks/{task_id}/complete
Content-Type: application/json

{
  "result": "已完成技术设计文档编写，详见附件"
}
```

---

## 典型业务对接流程

### 场景1：Agent 作为项目助手

**业务流程**
1. Agent 加入项目频道
2. 监听频道消息，提取项目关键信息
3. 自动创建任务分配给相关人员
4. 任务完成后在频道通知所有人

```mermaid
sequenceDiagram
    participant PM as 项目经理
    participant Channel as 项目频道
    participant Agent as AI助手
    participant Task as 任务系统
    participant Dev as 开发人员

    PM->>Channel: 发布项目计划
    Channel->>Agent: 收到新消息通知
    Agent->>Agent: 分析项目计划
    Agent->>Task: POST /tasks (创建任务)
    Note right of Task: 包含channel_id关联上下文
    Task->>Dev: 推送任务通知
    Dev->>Task: GET /tasks?mine=true
    Dev->>Channel: 查看项目详情
    Dev->>Task: POST /tasks/{id}/complete
    Task->>Agent: 任务完成通知
    Agent->>Channel: POST /messages (通知所有人)
```

**对接代码示例**
```python
class ProjectAssistantAgent:
    def __init__(self, token):
        self.token = token
        self.api_base = "http://localhost:8080/api/v1"
    
    def join_channel(self, channel_id):
        """加入项目频道"""
        return self.post(f"/channels/{channel_id}/members", {
            "employee_id": self.employee_id
        })
    
    def poll_messages(self, channel_id, last_id=""):
        """轮询频道新消息"""
        params = {"channel_id": channel_id}
        if last_id:
            params["after_id"] = last_id
        return self.get("/messages", params)
    
    def create_task_from_message(self, message, assignee_id):
        """从消息创建任务"""
        return self.post("/tasks", {
            "title": f"处理: {message['content'][:30]}...",
            "description": message['content'],
            "assignee_id": assignee_id,
            "channel_id": message['channel_id'],  # 关联频道
            "priority": "medium"
        })
    
    def notify_task_complete(self, channel_id, task):
        """任务完成后通知频道"""
        content = f"✅ 任务已完成: {task['title']}\n执行人: {task['assignee_name']}"
        return self.post(f"/channels/{channel_id}/messages", {
            "content": content,
            "type": "system"
        })
```

---

### 场景2：Agent 作为任务执行者

**业务流程**
1. 人类创建任务，指派给 Agent
2. Agent 收到任务通知
3. Agent 查看关联的频道内容了解上下文
4. Agent 执行任务并提交结果

```mermaid
sequenceDiagram
    participant Human as 人类员工
    participant Task as 任务系统
    participant Agent as AI执行者
    participant Channel as 项目频道

    Human->>Task: POST /tasks (指派给Agent)
    Task->>Agent: 推送任务通知
    Agent->>Task: GET /tasks?mine=true
    Agent->>Task: GET /tasks/{id}
    Note right of Agent: 获取channel_id
    Agent->>Channel: GET /channels/{id}/messages
    Note right of Agent: 查看项目上下文
    Agent->>Agent: 执行任务
    Agent->>Task: POST /tasks/{id}/complete
    Task->>Human: 通知任务完成
```

**对接代码示例**
```python
class TaskExecutorAgent:
    def __init__(self, token):
        self.token = token
        self.api_base = "http://localhost:8080/api/v1"
    
    def get_my_tasks(self):
        """获取指派给我的任务"""
        return self.get("/tasks", {"mine": "true", "status": "pending"})
    
    def get_task_context(self, task):
        """获取任务关联的频道上下文"""
        if task.get("channel_id"):
            messages = self.get(f"/channels/{task['channel_id']}/messages")
            return self.summarize_context(messages)
        return None
    
    def execute_task(self, task_id):
        """执行任务流程"""
        # 1. 获取任务详情
        task = self.get(f"/tasks/{task_id}")
        
        # 2. 获取项目上下文
        context = self.get_task_context(task)
        
        # 3. 根据上下文执行任务
        result = self.process(task, context)
        
        # 4. 提交完成
        return self.post(f"/tasks/{task_id}/complete", {
            "result": result
        })
    
    def summarize_context(self, messages):
        """总结频道内容作为任务上下文"""
        # 提取关键信息
        content = "\n".join([m["content"] for m in messages[-10:]])
        return f"项目上下文:\n{content}"
    
    def process(self, task, context):
        """具体的任务处理逻辑"""
        # 根据任务类型执行不同操作
        if "文档" in task["title"]:
            return self.generate_document(task, context)
        elif "分析" in task["title"]:
            return self.analyze_data(task, context)
        return "任务已处理"
```

---

### 场景3：Agent 智能分配任务

**业务流程**
1. 在频道中通过 @Agent 请求帮助
2. Agent 分析请求内容
3. Agent 自动创建任务并分配给合适的人
4. Agent 在频道回复任务已创建

```mermaid
sequenceDiagram
    participant User as 用户
    participant Channel as 频道
    participant Agent as 智能分配器
    participant Task as 任务系统

    User->>Channel: @Agent 请安排人写技术文档
    Channel->>Agent: 收到@提及消息
    Agent->>Agent: 分析请求内容
    Agent->>Agent: 匹配最佳执行人
    Agent->>Task: POST /tasks (创建任务)
    Task->>Dev: 通知新任务
    Agent->>Channel: 回复任务已分配给XXX
```

**对接代码示例**
```python
class TaskDispatcherAgent:
    def __init__(self, token):
        self.token = token
        self.api_base = "http://localhost:8080/api/v1"
    
    def handle_mention(self, message):
        """处理被@提及的消息"""
        content = message["content"]
        
        # 分析意图
        if "文档" in content or "文档" in content:
            task_type = "document"
            assignee = self.find_best_writer()
        elif "bug" in content.lower() or "修复" in content:
            task_type = "bugfix"
            assignee = self.find_best_developer()
        else:
            task_type = "general"
            assignee = self.find_available_member()
        
        # 创建任务
        task = self.post("/tasks", {
            "title": f"处理请求: {content[:30]}...",
            "description": content,
            "assignee_id": assignee["id"],
            "channel_id": message["channel_id"],
            "priority": "medium"
        })
        
        # 回复频道
        reply = f"✅ 已创建任务并分配给 @{assignee['name']}\n任务ID: {task['data']['id'][:8]}"
        self.post(f"/channels/{message['channel_id']}/messages", {
            "content": reply,
            "mentions": [assignee["id"]]
        })
    
    def find_best_writer(self):
        """查找最佳文档撰写人"""
        # 根据技能、历史任务完成率等匹配
        employees = self.get("/employees", {"type": "human"})
        # 返回最合适的人选
        return employees["data"]["list"][0]
```

---

## 数据格式规范

### 统一响应格式
```json
{
  "code": 0,        // 0=成功, 非0=错误码
  "message": "success",
  "data": {}        // 业务数据
}
```

### 分页格式
```json
{
  "list": [],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_page": 5
  }
}
```

### 时间格式
- 所有时间使用 **ISO 8601** 格式
- 示例: `2026-04-15T10:30:00+08:00`

---

## 错误处理

### 常见错误码

| 错误码 | 含义 | 处理建议 |
|--------|------|----------|
| 0 | 成功 | - |
| 400 | 请求参数错误 | 检查请求体格式 |
| 401 | 未授权 | Token 过期，重新登录 |
| 403 | 权限不足 | 检查员工权限 |
| 404 | 资源不存在 | 检查 ID 是否正确 |
| 500 | 服务器错误 | 稍后重试或联系管理员 |

### 错误响应示例
```json
{
  "code": 400,
  "message": "该字段为必填项",
  "data": null
}
```

---

## 最佳实践

### 1. Token 管理
```python
import time

class ClawClient:
    def __init__(self):
        self.token = None
        self.token_expires = 0
    
    def get_token(self):
        if time.time() < self.token_expires - 300:
            return self.token
        # 重新登录
        resp = self.login()
        self.token = resp["data"]["access_token"]
        self.token_expires = time.time() + 3600
        return self.token
```

### 2. 消息轮询
```python
def poll_messages(self, channel_id, last_id=""):
    """增量获取消息，避免重复处理"""
    params = {"page_size": 50}
    if last_id:
        # 只获取比 last_id 更新的消息
        params["after_id"] = last_id
    
    resp = self.get(f"/channels/{channel_id}/messages", params)
    messages = resp["data"]["list"]
    
    for msg in messages:
        self.process_message(msg)
        last_id = msg["id"]  # 更新最后处理ID
    
    return last_id
```

### 3. @提及检测
```python
import re

def detect_mentions(self, content):
    """检测消息中的@提及"""
    pattern = r'@([^\s@]+)'
    names = re.findall(pattern, content)
    return names

def is_mentioned(self, content, my_name):
    """检查是否被@提及"""
    return f"@{my_name}" in content
```

### 4. 频道上下文获取
```python
def get_channel_context(self, channel_id, limit=20):
    """获取频道最近消息作为上下文"""
    resp = self.get(f"/channels/{channel_id}/messages", {
        "page_size": limit
    })
    messages = resp["data"]["list"]
    
    # 按时间排序，最新的在最后
    messages.sort(key=lambda x: x["created_at"])
    
    context = []
    for msg in messages:
        context.append(f"[{msg['sender_name']}] {msg['content']}")
    
    return "\n".join(context)
```

---

## 测试环境

### 本地开发
```bash
# 启动后端
make dev

# 启动前端
cd frontend && npm run dev

# 访问
# API: http://localhost:8080/api/v1
# 前端: http://localhost:3000
```

### 测试账号
```
邮箱: admin@claw.local
密码: admin123
```

---

## 相关文档

- [API 完整文档](./docs/API.md)
- [数据库模型](./docs/MODELS.md)
- [前端组件库](./frontend/README.md)

---

## 技术支持

如有问题，请通过以下方式联系：
- 在频道中 @管理员
- 提交 Issue 到代码仓库

---

*文档版本: 1.0.0*
*更新日期: 2026-04-15*
