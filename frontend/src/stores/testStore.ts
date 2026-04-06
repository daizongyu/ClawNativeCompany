import { create } from 'zustand';

export interface TestMessage {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  content: string;
  timestamp: number;
}

export interface TestElement {
  id: string;
  element: Element;
  timestamp: number;
}

interface TestState {
  // 消息列表
  messages: TestMessage[];
  // 当前页面
  currentPage: string;
  // 页面加载状态
  pageLoading: Record<string, boolean>;
  // 元素映射
  elements: Record<string, Element>;
  
  // Actions
  addMessage: (type: TestMessage['type'], content: string) => void;
  clearMessages: () => void;
  setCurrentPage: (page: string) => void;
  setPageLoading: (page: string, loading: boolean) => void;
  registerElement: (id: string, element: Element) => void;
  unregisterElement: (id: string) => void;
  
  // Getters for test API
  getMessages: () => TestMessage[];
  getLastMessage: () => TestMessage | null;
  getCurrentPage: () => string;
  isPageLoading: (page: string) => boolean;
  findElement: (id: string) => Element | null;
  waitForElement: (id: string, timeout?: number) => Promise<Element | null>;
}

export const useTestStore = create<TestState>((set, get) => ({
  messages: [],
  currentPage: '',
  pageLoading: {},
  elements: {},

  addMessage: (type, content) => {
    const message: TestMessage = {
      id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      type,
      content,
      timestamp: Date.now(),
    };
    set((state) => ({
      messages: [...state.messages, message],
    }));
  },

  clearMessages: () => set({ messages: [] }),

  setCurrentPage: (page) => set({ currentPage: page }),

  setPageLoading: (page, loading) =>
    set((state) => ({
      pageLoading: { ...state.pageLoading, [page]: loading },
    })),

  registerElement: (id, element) =>
    set((state) => ({
      elements: { ...state.elements, [id]: element },
    })),

  unregisterElement: (id) =>
    set((state) => {
      const newElements = { ...state.elements };
      delete newElements[id];
      return { elements: newElements };
    }),

  // Getters
  getMessages: () => get().messages,

  getLastMessage: () => {
    const { messages } = get();
    return messages.length > 0 ? messages[messages.length - 1] : null;
  },

  getCurrentPage: () => get().currentPage,

  isPageLoading: (page) => !!get().pageLoading[page],

  findElement: (id) => get().elements[id] || null,

  waitForElement: async (id, timeout = 5000) => {
    const checkInterval = 100;
    const maxAttempts = timeout / checkInterval;
    let attempts = 0;

    return new Promise((resolve) => {
      const check = () => {
        const element = get().findElement(id);
        if (element) {
          resolve(element);
          return;
        }
        // Also try DOM query as fallback
        const domElement = document.querySelector(`[data-testid="${id}"]`);
        if (domElement) {
          resolve(domElement);
          return;
        }
        attempts++;
        if (attempts >= maxAttempts) {
          resolve(null);
          return;
        }
        setTimeout(check, checkInterval);
      };
      check();
    });
  },
}));

// 导出便捷方法
export const testStore = {
  addMessage: (type: TestMessage['type'], content: string) =>
    useTestStore.getState().addMessage(type, content),
  clearMessages: () => useTestStore.getState().clearMessages(),
  setCurrentPage: (page: string) => useTestStore.getState().setCurrentPage(page),
  setPageLoading: (page: string, loading: boolean) =>
    useTestStore.getState().setPageLoading(page, loading),
  getMessages: () => useTestStore.getState().getMessages(),
  getLastMessage: () => useTestStore.getState().getLastMessage(),
  getCurrentPage: () => useTestStore.getState().getCurrentPage(),
  findElement: (id: string) => useTestStore.getState().findElement(id),
  waitForElement: (id: string, timeout?: number) =>
    useTestStore.getState().waitForElement(id, timeout),
};

export default useTestStore;
