// 全局类型声明

import { ClawTestAPI } from '../utils/testExposer';

declare global {
  interface Window {
    __CLAW_TEST__?: ClawTestAPI;
    __TEST_EMPLOYEES__?: {
      openModal: () => void;
      closeModal: () => void;
      getEmployees: () => any[];
      setEditingEmployee: (emp: any) => void;
    };
    __TEST_CHANNELS__?: {
      openModal: () => void;
      closeModal: () => void;
      getChannels: () => any[];
      setEditingChannel: (ch: any) => void;
    };
    __TEST_TASKS__?: {
      openModal: () => void;
      closeModal: () => void;
      getTasks: () => any[];
      setEditingTask: (task: any) => void;
    };
    __TEST_WORKFLOWS__?: {
      openModal: () => void;
      closeModal: () => void;
      getWorkflows: () => any[];
      setEditingWorkflow: (wf: any) => void;
    };
    __TEST_CHANNEL_CHAT__?: {
      getMessages: () => any[];
      getChannel: () => any;
      sendMessage: (content: string) => Promise<void>;
    };
  }
}

export {};
