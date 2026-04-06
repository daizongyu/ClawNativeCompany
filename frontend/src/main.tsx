import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './index.css';

// 初始化测试工具（仅在开发环境）
if (import.meta.env.DEV || import.meta.env.VITE_ENABLE_TEST_API === 'true') {
  import('./utils/messageInterceptor');
  import('./utils/testExposer');
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
