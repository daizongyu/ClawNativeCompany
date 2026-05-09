import React from 'react';
import { Table, Button, Input, Empty, Space, Tag, Tooltip, Spin } from 'antd';
import { FileTextOutlined, FileAddOutlined, EditOutlined, HistoryOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { Document } from '../../services/document';

interface DocumentListProps {
  channelId: string | null;
  documents: Document[];
  loading: boolean;
  pagination: {
    current: number;
    pageSize: number;
    total: number;
  };
  onPageChange: (page: number, pageSize: number) => void;
  onCreateDocument: () => void;
  onEditDocument: (docId: string, title: string) => void;
  onDeleteDocument: (docId: string) => void;
  onViewHistory: (docId: string, title: string) => void;
  onSearch: (keyword: string) => void;
  searchKeyword: string;
}

// 格式化文件大小
function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

// 格式化时间
function formatTime(time: string): string {
  const date = new Date(time);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  const hours = date.getHours().toString().padStart(2, '0');
  const minutes = date.getMinutes().toString().padStart(2, '0');

  if (diffDays === 0) return '今天 ' + hours + ':' + minutes;
  if (diffDays === 1) return '昨天 ' + hours + ':' + minutes;
  if (diffDays === 2) return '前天 ' + hours + ':' + minutes;
  if (diffDays < 7) return `${diffDays} 天前`;
  
  const year = date.getFullYear();
  const month = (date.getMonth() + 1).toString().padStart(2, '0');
  const day = date.getDate().toString().padStart(2, '0');
  return `${year}-${month}-${day} ${hours}:${minutes}`;
}

// 表格列定义
function getColumns(
  onEdit: (id: string, title: string) => void,
  onDelete: (id: string) => void,
  onViewHistory: (id: string, title: string) => void
): ColumnsType<Document> {
  return [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      width: 250,
      render: (title: string, record) => (
        <Space>
          <FileTextOutlined style={{ color: '#1890ff' }} />
          <span 
            className="document-title" 
            data-testid={`doc-title-${record.id}`}
            style={{ cursor: 'pointer' }}
            onClick={() => onEdit(record.id, title)}
          >
            {title}
          </span>
        </Space>
      ),
    },
    {
      title: '摘要',
      dataIndex: 'summary',
      key: 'summary',
      ellipsis: true,
      width: 300,
      render: (summary: string) => summary || '-',
    },
    {
      title: '创建者',
      dataIndex: 'author_name',
      key: 'author_name',
      width: 100,
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 80,
      render: (version: number) => <Tag color="blue">v{version}</Tag>,
    },
    {
      title: '大小',
      dataIndex: 'file_size',
      key: 'file_size',
      width: 80,
      render: (size: number) => formatFileSize(size),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 150,
      render: (time: string) => formatTime(time),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="编辑">
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => onEdit(record.id, record.title)}
              data-testid={`doc-edit-btn-${record.id}`}
              data-action="edit-document"
            />
          </Tooltip>
          <Tooltip title="历史版本">
            <Button
              type="text"
              size="small"
              icon={<HistoryOutlined />}
              onClick={() => onViewHistory(record.id, record.title)}
              data-testid={`doc-history-btn-${record.id}`}
              data-action="view-history"
            />
          </Tooltip>
          <Tooltip title="删除">
            <Button
              type="text"
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={() => onDelete(record.id)}
              data-testid={`doc-delete-btn-${record.id}`}
              data-action="delete-document"
            />
          </Tooltip>
        </Space>
      ),
    },
  ];
}

export const DocumentList: React.FC<DocumentListProps> = ({
  channelId,
  documents,
  loading,
  pagination,
  onPageChange,
  onCreateDocument,
  onEditDocument,
  onDeleteDocument,
  onViewHistory,
  onSearch,
  searchKeyword,
}) => {
  // 未选择频道时显示提示
  if (!channelId) {
    return (
      <div className="document-list-empty" data-testid="document-list-empty">
        <Empty
          description="请选择一个频道查看文档"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      </div>
    );
  }

  // 频道内无文档时显示空状态
  if (!loading && documents.length === 0 && !searchKeyword) {
    return (
      <div className="document-list-container" data-testid="document-list">
        <div className="document-list-toolbar" style={{ marginBottom: 16 }}>
          <Space>
            <Button
              type="primary"
              icon={<FileAddOutlined />}
              onClick={onCreateDocument}
              data-testid="create-document-btn"
              data-action="create-document"
            >
              创建文档
            </Button>
          </Space>
        </div>
        <Empty
          description="该频道暂无文档"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        >
          <Button
            type="primary"
            onClick={onCreateDocument}
            data-testid="create-first-document-btn"
          >
            创建第一个文档
          </Button>
        </Empty>
      </div>
    );
  }

  return (
    <div className="document-list-container" data-testid="document-list">
      {/* 操作栏 */}
      <div className="document-list-toolbar" style={{ marginBottom: 16 }}>
        <Space>
          <Button
            type="primary"
            icon={<FileAddOutlined />}
            onClick={onCreateDocument}
            data-testid="create-document-btn"
            data-action="create-document"
          >
            创建文档
          </Button>
          
          <Input.Search
            placeholder="搜索文档标题"
            value={searchKeyword}
            onChange={(e) => onSearch(e.target.value)}
            style={{ width: 200 }}
            allowClear
            data-testid="document-search-input"
          />
        </Space>
      </div>

      {/* 文档表格 */}
      <Spin spinning={loading}>
        <Table
          dataSource={documents}
          rowKey="id"
          columns={getColumns(onEditDocument, onDeleteDocument, onViewHistory)}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            onChange: onPageChange,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 个文档`,
          }}
          data-testid="document-table"
          data-entity="document"
          onRow={(record) => ({
            onClick: () => onEditDocument(record.id, record.title),
            style: { cursor: 'pointer' },
          })}
        />
      </Spin>
    </div>
  );
};