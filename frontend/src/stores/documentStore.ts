import { create } from 'zustand';
import { ChannelNode } from '../services/channel';

export interface Document {
  id: string;
  title: string;
  content: string;
  summary: string;
  author_id: string;
  author_name: string;
  editor_id?: string;
  editor_name?: string;
  version: number;
  file_size: number;
  channel_id: string;
  created_at: string;
  updated_at: string;
}

export interface DocumentVersion {
  id: string;
  version: number;
  summary: string;
  editor_name: string;
  created_at: string;
}

interface DocumentState {
  // 频道树
  channels: ChannelNode[];
  selectedChannelId: string | null;
  selectedChannel: ChannelNode | null;
  channelLoading: boolean;

  // 文档列表
  documents: Document[];
  selectedDocument: Document | null;
  documentLoading: boolean;
  documentPagination: {
    current: number;
    pageSize: number;
    total: number;
  };
  searchKeyword: string;

  // 编辑器
  editorVisible: boolean;
  editingDocumentId: string | null;
  editingDocumentTitle: string;

  // 弹窗
  createChannelModalVisible: boolean;
  createChannelParentId: string | null;
  createChannelParentName: string;
  createDocumentModalVisible: boolean;
  historyModalVisible: boolean;
  historyDocumentId: string | null;
  historyDocumentTitle: string;

  // Actions
  setChannels: (channels: ChannelNode[]) => void;
  selectChannel: (id: string | null) => void;
  setSelectedChannel: (channel: ChannelNode | null) => void;
  setChannelLoading: (loading: boolean) => void;
  setDocuments: (documents: Document[]) => void;
  setDocumentLoading: (loading: boolean) => void;
  setDocumentPagination: (pagination: { current: number; pageSize: number; total: number }) => void;
  selectDocument: (doc: Document | null) => void;
  setSearchKeyword: (keyword: string) => void;
  openEditor: (documentId: string, title: string) => void;
  closeEditor: () => void;
  openCreateChannelModal: (parentId?: string, parentName?: string) => void;
  closeCreateChannelModal: () => void;
  openCreateDocumentModal: () => void;
  closeCreateDocumentModal: () => void;
  openHistoryModal: (documentId: string, title: string) => void;
  closeHistoryModal: () => void;
}

export const useDocumentStore = create<DocumentState>((set) => ({
  // 频道树
  channels: [],
  selectedChannelId: null,
  selectedChannel: null,
  channelLoading: false,

  // 文档列表
  documents: [],
  selectedDocument: null,
  documentLoading: false,
  documentPagination: {
    current: 1,
    pageSize: 20,
    total: 0,
  },
  searchKeyword: '',

  // 编辑器
  editorVisible: false,
  editingDocumentId: null,
  editingDocumentTitle: '',

  // 弹窗
  createChannelModalVisible: false,
  createChannelParentId: null,
  createChannelParentName: '',
  createDocumentModalVisible: false,
  historyModalVisible: false,
  historyDocumentId: null,
  historyDocumentTitle: '',

  // Actions
  setChannels: (channels) => set({ channels }),
  selectChannel: (id) => set({ selectedChannelId: id, selectedDocument: null }),
  setSelectedChannel: (channel) => set({ selectedChannel: channel }),
  setChannelLoading: (loading) => set({ channelLoading: loading }),

  setDocuments: (documents) => set({ documents }),
  setDocumentLoading: (loading) => set({ documentLoading: loading }),
  setDocumentPagination: (pagination) => set({ documentPagination: pagination }),
  selectDocument: (doc) => set({ selectedDocument: doc }),
  setSearchKeyword: (keyword) => set({ searchKeyword: keyword }),

  openEditor: (id, title) => set({
    editorVisible: true,
    editingDocumentId: id,
    editingDocumentTitle: title,
  }),
  closeEditor: () => set({
    editorVisible: false,
    editingDocumentId: null,
    editingDocumentTitle: '',
  }),

  openCreateChannelModal: (parentId, parentName) => set({
    createChannelModalVisible: true,
    createChannelParentId: parentId || null,
    createChannelParentName: parentName || '',
  }),
  closeCreateChannelModal: () => set({
    createChannelModalVisible: false,
    createChannelParentId: null,
    createChannelParentName: '',
  }),

  openCreateDocumentModal: () => set({ createDocumentModalVisible: true }),
  closeCreateDocumentModal: () => set({ createDocumentModalVisible: false }),

  openHistoryModal: (documentId, title) => set({
    historyModalVisible: true,
    historyDocumentId: documentId,
    historyDocumentTitle: title,
  }),
  closeHistoryModal: () => set({
    historyModalVisible: false,
    historyDocumentId: null,
    historyDocumentTitle: '',
  }),
}));