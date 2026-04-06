import { message as antMessage } from 'antd';
import { testStore } from '../stores/testStore';

// 保存原始方法
const originalSuccess = antMessage.success;
const originalError = antMessage.error;
const originalWarning = antMessage.warning;
const originalInfo = antMessage.info;

// 拦截消息方法
export const messageInterceptor = {
  init: () => {
    // 拦截 success
    antMessage.success = (content: any, duration?: any, onClose?: () => void) => {
      const msgContent = typeof content === 'string' ? content : content?.content || '';
      testStore.addMessage('success', msgContent);
      return originalSuccess(content, duration, onClose);
    };

    // 拦截 error
    antMessage.error = (content: any, duration?: any, onClose?: () => void) => {
      const msgContent = typeof content === 'string' ? content : content?.content || '';
      testStore.addMessage('error', msgContent);
      return originalError(content, duration, onClose);
    };

    // 拦截 warning
    antMessage.warning = (content: any, duration?: any, onClose?: () => void) => {
      const msgContent = typeof content === 'string' ? content : content?.content || '';
      testStore.addMessage('warning', msgContent);
      return originalWarning(content, duration, onClose);
    };

    // 拦截 info
    antMessage.info = (content: any, duration?: any, onClose?: () => void) => {
      const msgContent = typeof content === 'string' ? content : content?.content || '';
      testStore.addMessage('info', msgContent);
      return originalInfo(content, duration, onClose);
    };
  },

  // 恢复原始方法（用于测试）
  restore: () => {
    antMessage.success = originalSuccess;
    antMessage.error = originalError;
    antMessage.warning = originalWarning;
    antMessage.info = originalInfo;
  },
};

// 导出带拦截的消息对象
export const message = antMessage;

// 自动初始化
if (typeof window !== 'undefined') {
  messageInterceptor.init();
}

export default messageInterceptor;
