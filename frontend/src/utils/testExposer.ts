import { testStore, useTestStore } from '../stores/testStore';

// 测试 API 接口定义
export interface ClawTestAPI {
  // 消息相关
  getMessages: () => { type: string; content: string; timestamp: number }[];
  getLastMessage: () => { type: string; content: string; timestamp: number } | null;
  clearMessages: () => void;
  
  // 页面相关
  getCurrentPage: () => string;
  setCurrentPage: (page: string) => void;
  isPageLoading: (page: string) => boolean;
  
  // 元素相关
  findElement: (id: string) => Element | null;
  waitForElement: (id: string, timeout?: number) => Promise<Element | null>;
  clickElement: (id: string) => boolean;
  typeIntoElement: (id: string, value: string) => boolean;
  getElementValue: (id: string) => string | null;
  
  // 工具方法
  sleep: (ms: number) => Promise<void>;
  waitForPageLoad: (page: string, timeout?: number) => Promise<boolean>;
}

// 创建测试 API 对象
const createTestAPI = (): ClawTestAPI => ({
  // 消息相关
  getMessages: () => {
    const messages = testStore.getMessages();
    return messages.map((msg) => ({
      type: msg.type,
      content: msg.content,
      timestamp: msg.timestamp,
    }));
  },

  getLastMessage: () => {
    const msg = testStore.getLastMessage();
    return msg
      ? {
          type: msg.type,
          content: msg.content,
          timestamp: msg.timestamp,
        }
      : null;
  },

  clearMessages: () => {
    testStore.clearMessages();
  },

  // 页面相关
  getCurrentPage: () => testStore.getCurrentPage(),

  setCurrentPage: (page: string) => {
    testStore.setCurrentPage(page);
  },

  isPageLoading: (page: string) => {
    const state = useTestStore.getState();
    return state.isPageLoading(page);
  },

  // 元素相关
  findElement: (id: string) => {
    // 先从 store 查找
    const fromStore = testStore.findElement(id);
    if (fromStore) return fromStore;
    // 再从 DOM 查找
    return document.querySelector(`[data-testid="${id}"]`);
  },

  waitForElement: async (id: string, timeout = 5000) => {
    return await testStore.waitForElement(id, timeout);
  },

  clickElement: (id: string) => {
    const element = document.querySelector(`[data-testid="${id}"]`) as HTMLElement;
    if (element) {
      element.click();
      return true;
    }
    return false;
  },

  typeIntoElement: (id: string, value: string) => {
    const element = document.querySelector(`[data-testid="${id}"]`) as HTMLInputElement | HTMLTextAreaElement;
    if (element) {
      // 设置值
      element.value = value;
      
      // 触发 focus 事件
      element.dispatchEvent(new FocusEvent('focus', { bubbles: true }));
      
      // 触发 input 事件（React 使用这个）
      const inputEvent = new Event('input', { bubbles: true });
      // 对于 React，需要设置 inputType
      Object.defineProperty(inputEvent, 'inputType', { value: 'insertText' });
      element.dispatchEvent(inputEvent);
      
      // 触发 change 事件
      element.dispatchEvent(new Event('change', { bubbles: true }));
      
      // 触发 blur 事件
      element.dispatchEvent(new FocusEvent('blur', { bubbles: true }));
      
      // 触发 keydown/keyup 事件序列（可选，某些组件需要）
      element.dispatchEvent(new KeyboardEvent('keydown', { bubbles: true, key: value.slice(-1) }));
      element.dispatchEvent(new KeyboardEvent('keyup', { bubbles: true, key: value.slice(-1) }));
      
      return true;
    }
    return false;
  },

  getElementValue: (id: string) => {
    const element = document.querySelector(`[data-testid="${id}"]`) as HTMLInputElement;
    return element ? element.value : null;
  },

  // 工具方法
  sleep: (ms: number) => new Promise((resolve) => setTimeout(resolve, ms)),

  waitForPageLoad: async (page: string, timeout = 10000) => {
    const checkInterval = 100;
    const maxAttempts = timeout / checkInterval;
    let attempts = 0;

    return new Promise((resolve) => {
      const check = () => {
        const pageElement = document.querySelector(`[data-testid="page-${page}"]`);
        if (pageElement) {
          const isLoaded = pageElement.getAttribute('data-loaded') === 'true';
          if (isLoaded) {
            resolve(true);
            return;
          }
        }
        attempts++;
        if (attempts >= maxAttempts) {
          resolve(false);
          return;
        }
        setTimeout(check, checkInterval);
      };
      check();
    });
  },
});

// 暴露到全局
declare global {
  interface Window {
    __CLAW_TEST__?: ClawTestAPI;
  }
}

export const testExposer = {
  init: () => {
    if (typeof window !== 'undefined') {
      window.__CLAW_TEST__ = createTestAPI();
      console.log('[TestExposer] __CLAW_TEST__ API exposed to window');
    }
  },
};

// 自动初始化
if (typeof window !== 'undefined') {
  testExposer.init();
}

export default testExposer;
