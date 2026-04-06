# Agent 友好的前端页面 - 使用指南

## 概述

本项目已完成 Agent 友好改造，添加了 `data-testid` 属性和全局测试 API，方便自动化测试 Agent 进行页面操作和验证。

## 核心功能

### 1. data-testid 属性

所有可交互元素都添加了 `data-testid` 属性，命名规范如下：

#### 页面容器
```
data-testid="page-{page-name}"
data-page="{page-name}"
data-loaded="true|false"
data-loading="true|false"
```

例如：
- `page-login`
- `page-dashboard`
- `page-employees`
- `page-tasks`
- `page-workflows`
- `page-channels`
- `page-channel-chat`

#### 按钮
```
data-testid="{entity}-{action}-btn"
data-action="{action}"
data-entity="{entity}"
```

例如：
- `employee-create-btn`
- `employee-edit-btn-{id}`
- `employee-delete-btn-{id}`
- `task-complete-btn-{id}`
- `workflow-execute-btn-{id}`

#### 输入框
```
data-testid="input-{field-name}"
data-input-name="{field-name}"
```

例如：
- `input-email`
- `input-password`
- `input-employee-name`
- `input-task-title`

#### 表格
```
data-testid="{entity}-table"
data-entity="{entity}"
```

例如：
- `employee-table`
- `task-table`
- `channel-table`
- `workflow-table`

#### 模态框
```
data-testid="{entity}-modal"
```

例如：
- `employee-modal`
- `task-modal`
- `channel-modal`

### 2. 全局测试 API

通过 `window.__CLAW_TEST__` 可以访问测试 API：

```javascript
// 获取所有消息
const messages = window.__CLAW_TEST__.getMessages();

// 获取最后一条消息
const lastMessage = window.__CLAW_TEST__.getLastMessage();

// 获取当前页面
const currentPage = window.__CLAW_TEST__.getCurrentPage();

// 查找元素
const element = window.__CLAW_TEST__.findElement('employee-create-btn');

// 等待元素出现
const element = await window.__CLAW_TEST__.waitForElement('employee-create-btn', 5000);

// 点击元素
window.__CLAW_TEST__.clickElement('employee-create-btn');

// 输入文本
window.__CLAW_TEST__.typeIntoElement('input-employee-name', '张三');

// 获取输入值
const value = window.__CLAW_TEST__.getElementValue('input-employee-name');

// 等待页面加载完成
const loaded = await window.__CLAW_TEST__.waitForPageLoad('employees', 10000);

// 休眠
await window.__CLAW_TEST__.sleep(1000);
```

### 3. 页面特定测试对象

每个页面都有特定的测试对象：

```javascript
// 员工页面
window.__TEST_EMPLOYEES__.openModal();
window.__TEST_EMPLOYEES__.closeModal();
window.__TEST_EMPLOYEES__.getEmployees();
window.__TEST_EMPLOYEES__.setEditingEmployee(employee);

// 频道页面
window.__TEST_CHANNELS__.openModal();
window.__TEST_CHANNELS__.closeModal();
window.__TEST_CHANNELS__.getChannels();

// 任务页面
window.__TEST_TASKS__.openModal();
window.__TEST_TASKS__.closeModal();
window.__TEST_TASKS__.getTasks();

// 工作流页面
window.__TEST_WORKFLOWS__.openModal();
window.__TEST_WORKFLOWS__.closeModal();
window.__TEST_WORKFLOWS__.getWorkflows();

// 频道聊天页面
window.__TEST_CHANNEL_CHAT__.getMessages();
window.__TEST_CHANNEL_CHAT__.getChannel();
await window.__TEST_CHANNEL_CHAT__.sendMessage('Hello');
```

## 测试脚本示例

### 示例 1：登录测试

```javascript
// 等待登录页面加载
await page.waitForSelector('[data-testid="page-login"]');

// 输入邮箱
await page.type('[data-testid="input-email"]', 'admin@example.com');

// 输入密码
await page.type('[data-testid="input-password"]', 'password123');

// 点击登录按钮
await page.click('[data-testid="login-submit-btn"]');

// 验证登录成功（通过消息或页面跳转）
await page.waitForNavigation();
await page.waitForSelector('[data-testid="page-dashboard"]');
```

### 示例 2：创建员工测试

```javascript
// 导航到员工页面
await page.click('[data-testid="nav-employees"]');
await page.waitForSelector('[data-testid="page-employees"][data-loaded="true"]');

// 点击新建按钮
await page.click('[data-testid="employee-create-btn"]');
await page.waitForSelector('[data-testid="employee-modal"]');

// 填写表单
await page.type('[data-testid="input-employee-name"]', '张三');
await page.type('[data-testid="input-employee-email"]', 'zhangsan@example.com');
await page.select('[data-testid="input-employee-type"]', 'employee');
await page.type('[data-testid="input-employee-password"]', 'password123');

// 提交
await page.click('.ant-modal-footer .ant-btn-primary');

// 验证成功消息
const lastMessage = await page.evaluate(() => {
  return window.__CLAW_TEST__.getLastMessage();
});
expect(lastMessage.type).toBe('success');
expect(lastMessage.content).toContain('创建成功');
```

### 示例 3：任务操作测试

```javascript
// 导航到任务页面
await page.click('[data-testid="nav-tasks"]');
await page.waitForSelector('[data-testid="page-tasks"][data-loaded="true"]');

// 搜索任务
await page.type('[data-testid="input-task-search"]', '测试任务');
await page.waitForTimeout(500);

// 完成任务
await page.click('[data-testid="task-complete-btn-{task-id}"]');

// 验证消息
const lastMessage = await page.evaluate(() => {
  return window.__CLAW_TEST__.getLastMessage();
});
expect(lastMessage.content).toBe('任务已完成');
```

## 组件列表

### 基础组件

| 组件 | 文件 | 功能 |
|------|------|------|
| Button | `components/common/Button.tsx` | 带 data-testid 的按钮 |
| FormField | `components/common/FormField.tsx` | 带 data-testid 的表单字段 |
| DataTable | `components/common/DataTable.tsx` | 带 data-testid 的表格 |
| PageContainer | `components/common/PageContainer.tsx` | 带加载状态的页面容器 |

### 工具

| 工具 | 文件 | 功能 |
|------|------|------|
| testStore | `stores/testStore.ts` | 测试状态管理 |
| messageInterceptor | `utils/messageInterceptor.ts` | 消息拦截器 |
| testExposer | `utils/testExposer.ts` | 测试 API 暴露 |

## 注意事项

1. **页面加载等待**：始终等待 `data-loaded="true"` 后再进行元素操作
2. **消息验证**：使用 `window.__CLAW_TEST__.getLastMessage()` 验证操作结果
3. **元素定位**：优先使用 `data-testid` 而非 CSS 选择器
4. **异步操作**：使用 `waitForElement` 或 `waitForPageLoad` 处理异步加载

## 文件变更清单

### 新增文件
- `src/components/common/Button.tsx`
- `src/components/common/FormField.tsx`
- `src/components/common/DataTable.tsx`
- `src/components/common/PageContainer.tsx`
- `src/components/common/index.ts`
- `src/stores/testStore.ts`
- `src/utils/messageInterceptor.ts`
- `src/utils/testExposer.ts`
- `src/types/global.d.ts`

### 修改文件
- `src/main.tsx` - 添加测试工具初始化
- `src/App.tsx` - 添加测试工具动态导入
- `src/pages/Login.tsx` - 添加 data-testid
- `src/pages/Dashboard.tsx` - 添加 data-testid
- `src/pages/Employees.tsx` - 添加 data-testid
- `src/pages/Channels.tsx` - 添加 data-testid
- `src/pages/ChannelChat.tsx` - 添加 data-testid
- `src/pages/Tasks.tsx` - 添加 data-testid
- `src/pages/Workflows.tsx` - 添加 data-testid
- `src/components/layout/MainLayout.tsx` - 添加 data-testid
- `src/components/layout/Header.tsx` - 添加 data-testid
- `src/components/layout/Sidebar.tsx` - 添加 data-testid
