# Claw 前端

基于 React + TypeScript + Vite + Ant Design 构建的前端应用。

## 技术栈

- **框架**: React 18
- **语言**: TypeScript
- **构建工具**: Vite 5
- **UI 组件库**: Ant Design 5
- **状态管理**: Zustand
- **HTTP 客户端**: Axios
- **路由**: React Router 6
- **日期处理**: Day.js

## 项目结构

```
src/
├── api/              # API 封装
│   ├── auth.ts       # 认证相关
│   ├── channel.ts    # 频道管理
│   ├── dashboard.ts  # 仪表盘
│   ├── employee.ts   # 员工管理
│   ├── message.ts    # 消息管理
│   ├── task.ts       # 任务管理
│   └── workflow.ts   # 工作流管理
├── components/       # 组件
│   └── layout/       # 布局组件
│       ├── Header.tsx
│       ├── MainLayout.tsx
│       └── Sidebar.tsx
├── pages/            # 页面
│   ├── ChannelChat.tsx   # 频道聊天
│   ├── Channels.tsx      # 频道列表
│   ├── Dashboard.tsx     # 仪表盘
│   ├── Employees.tsx     # 员工管理
│   ├── Login.tsx         # 登录
│   ├── Tasks.tsx         # 任务管理
│   └── Workflows.tsx     # 工作流管理
├── stores/           # 状态管理
│   └── auth.ts       # 认证状态
├── utils/            # 工具函数
│   ├── request.ts    # HTTP 请求封装
│   └── websocket.ts  # WebSocket 管理
├── App.tsx           # 应用入口
├── main.tsx          # 渲染入口
└── index.css         # 全局样式
```

## 功能模块

### ✅ 已完成

- [x] 登录/认证
- [x] 主布局（侧边栏 + 头部）
- [x] 仪表盘（统计数据展示）
- [x] 员工管理（CRUD + API Key 管理）
- [x] 频道管理（列表、创建、编辑、删除）
- [x] 频道聊天（消息列表、发送、撤回）
- [x] 任务管理（CRUD + 筛选 + 认领/完成）
- [x] 工作流管理（CRUD + 执行历史）
- [x] WebSocket 管理器

### 🚧 待完善

- [ ] WebSocket 实时消息推送
- [ ] 消息通知系统
- [ ] 文件上传/下载
- [ ] 个人设置页面
- [ ] 系统设置页面
- [ ] 响应式适配
- [ ] 单元测试

## 开发命令

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 预览生产构建
npm run preview

# 代码检查
npm run lint
```

## 环境变量

复制 `.env.example` 为 `.env.local` 并配置：

```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws
```

## API 对接

所有 API 封装在 `src/api/` 目录下，统一使用 `request.ts` 中的 axios 实例。

### 响应格式

```typescript
{
  code: number;      // 0 表示成功
  message: string;   // 错误信息
  data: any;         // 响应数据
}
```

### 认证

使用 JWT Token，存储在 Zustand 和 localStorage 中，自动附加到请求头。

## WebSocket

使用 `src/utils/websocket.ts` 中的 WebSocketManager：

```typescript
import { useWebSocket } from '../utils/websocket';

const { connect, subscribe, send } = useWebSocket();

// 连接
connect();

// 订阅消息
const unsubscribe = subscribe('message', (data) => {
  console.log('收到消息:', data);
});

// 发送消息
send('ping', { time: Date.now() });
```

## 路由配置

| 路径 | 页面 | 说明 |
|------|------|------|
| /login | Login | 登录页（公开） |
| / | Dashboard | 仪表盘 |
| /employees | Employees | 员工管理 |
| /channels | Channels | 频道列表 |
| /channels/:id | ChannelChat | 频道聊天 |
| /tasks | Tasks | 任务管理 |
| /workflows | Workflows | 工作流管理 |

## 注意事项

1. 所有页面组件默认受保护，需要登录才能访问
2. API 请求会自动处理 token 过期（401 跳转登录）
3. 表单验证使用 Ant Design 的 Form 组件
4. 日期处理统一使用 Day.js
