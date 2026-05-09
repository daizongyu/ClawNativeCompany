import request from '../utils/request';

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

export interface DocumentListResponse {
  list: Document[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreateDocumentRequest {
  title: string;
  content?: string;
}

export interface SaveContentRequest {
  content: string;
  expected_version: number;
}

export const documentApi = {
  // 获取文档列表
  listByChannel: (channelId: string, params?: {
    keyword?: string;
    page?: number;
    page_size?: number;
  }): Promise<any> => {
    let url = `/documents?channel_id=${channelId}&page=${params?.page || 1}&page_size=${params?.page_size || 20}`;
    if (params?.keyword) url += `&keyword=${encodeURIComponent(params.keyword)}`;
    return request.get(url);
  },

  // 创建文档
  create: (channelId: string, data: CreateDocumentRequest): Promise<any> => {
    return request.post(`/documents?channel_id=${channelId}`, data);
  },

  // 获取文档详情
  getById: (id: string): Promise<any> => {
    return request.get(`/documents/${id}`);
  },

  // 保存文档内容
  saveContent: (id: string, data: SaveContentRequest): Promise<any> => {
    return request.put(`/documents/${id}/content`, data);
  },

  // 删除文档
  delete: (id: string): Promise<any> => {
    return request.delete(`/documents/${id}`);
  },

  // 获取版本列表
  getVersions: (id: string, page?: number): Promise<any> => {
    return request.get(`/documents/${id}/versions?page=${page || 1}`);
  },

  // 获取版本内容
  getVersionContent: (id: string, version: number): Promise<any> => {
    return request.get(`/documents/${id}/versions/${version}`);
  },

  // 恢复版本
  restoreVersion: (id: string, version: number): Promise<any> => {
    return request.post(`/documents/${id}/versions/${version}/restore`);
  },
};