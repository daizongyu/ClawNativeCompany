import React, { useEffect, useCallback } from 'react';
import { Layout, Card, Empty, Button, Space, message, Modal } from 'antd';
import { FileTextOutlined, FolderOutlined, FolderAddOutlined } from '@ant-design/icons';
import { useDocumentStore } from '../../stores/documentStore';
import { channelApi, ChannelNode } from '../../services/channel';
import { documentApi } from '../../services/document';
import { ChannelTree } from '../../components/documents/ChannelTree';
import { CreateChannelModal } from '../../components/documents/CreateChannelModal';
import { DocumentList } from '../../components/documents/DocumentList';
import { CreateDocumentModal } from '../../components/documents/CreateDocumentModal';
import { DocumentEditorPanel } from '../../components/documents/DocumentEditorPanel';
import '../../components/documents/ChannelTree.css';
import '../../components/documents/DocumentEditorPanel.css';

const { Sider, Content } = Layout;

// 查找频道节点
function findChannelNode(channels: ChannelNode[], id: string): ChannelNode | null {
  for (const ch of channels) {
    if (ch.id === id) return ch;
    if (ch.children?.length > 0) {
      const found = findChannelNode(ch.children, id);
      if (found) return found;
    }
  }
  return null;
}

const DocumentsPage: React.FC = () => {
  const {
    // 频道树
    channels,
    selectedChannelId,
    selectedChannel,
    channelLoading,
    createChannelModalVisible,
    createChannelParentId,
    createChannelParentName,
    setChannels,
    selectChannel,
    setSelectedChannel,
    setChannelLoading,
    openCreateChannelModal,
    closeCreateChannelModal,
    
    // 文档列表
    documents,
    documentLoading,
    documentPagination,
    searchKeyword,
    setDocuments,
    setDocumentLoading,
    setDocumentPagination,
    setSearchKeyword,
    
    // 编辑器
    editorVisible,
    editingDocumentId,
    editingDocumentTitle,
    openEditor,
    closeEditor,
    
    // 创建文档弹窗
    createDocumentModalVisible,
    openCreateDocumentModal,
    closeCreateDocumentModal,
    
    // 历史弹窗（Phase 4 实现）
    // historyModalVisible,
    // historyDocumentId,
    // historyDocumentTitle,
    openHistoryModal,
    // closeHistoryModal,
  } = useDocumentStore();

  // 加载频道树
  const loadChannelTree = useCallback(async () => {
    setChannelLoading(true);
    try {
      const res = await channelApi.getTree();
      if (res.code === 0) {
        setChannels(res.data.channels || []);
      } else {
        message.error(res.message || '加载频道失败');
      }
    } catch (error) {
      message.error('加载频道失败');
    } finally {
      setChannelLoading(false);
    }
  }, [setChannels, setChannelLoading]);

  // 加载文档列表
  const loadDocuments = useCallback(async (channelId: string, keyword?: string) => {
    setDocumentLoading(true);
    try {
      const res = await documentApi.listByChannel(channelId, {
        keyword,
        page: documentPagination.current,
        page_size: documentPagination.pageSize,
      });
      if (res.code === 0) {
        setDocuments(res.data.list || []);
        setDocumentPagination({
          current: res.data.page,
          pageSize: res.data.page_size,
          total: res.data.total,
        });
      } else {
        message.error(res.message || '加载文档失败');
      }
    } catch (error) {
      message.error('加载文档失败');
    } finally {
      setDocumentLoading(false);
    }
  }, [documentPagination.current, documentPagination.pageSize, setDocuments, setDocumentLoading, setDocumentPagination]);

  // 初始化加载频道树
  useEffect(() => {
    loadChannelTree();
  }, [loadChannelTree]);

  // 选择频道时加载文档列表
  useEffect(() => {
    if (selectedChannelId) {
      const channel = findChannelNode(channels, selectedChannelId);
      setSelectedChannel(channel);
      loadDocuments(selectedChannelId, searchKeyword);
    } else {
      setSelectedChannel(null);
      setDocuments([]);
    }
  }, [selectedChannelId, channels, searchKeyword, loadDocuments, setSelectedChannel, setDocuments]);

  // 搜索时重新加载
  const handleSearch = useCallback((keyword: string) => {
    setSearchKeyword(keyword);
    if (selectedChannelId) {
      loadDocuments(selectedChannelId, keyword);
    }
  }, [selectedChannelId, loadDocuments, setSearchKeyword]);

  // 分页变化
  const handlePageChange = useCallback((page: number, pageSize: number) => {
    setDocumentPagination({ current: page, pageSize, total: documentPagination.total });
    if (selectedChannelId) {
      loadDocuments(selectedChannelId, searchKeyword);
    }
  }, [selectedChannelId, searchKeyword, documentPagination.total, loadDocuments, setDocumentPagination]);

  // 创建频道成功
  const handleCreateChannelSuccess = useCallback((newChannelId: string) => {
    closeCreateChannelModal();
    loadChannelTree();
    selectChannel(newChannelId);
  }, [closeCreateChannelModal, loadChannelTree, selectChannel]);

  // 编辑频道
  const handleEditChannel = useCallback((_channelId: string) => {
    message.info('编辑频道功能待实现');
  }, []);

  // 删除频道
  const handleDeleteChannel = useCallback((channelId: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除频道将同时删除其下所有文档和子频道，是否继续？',
      okText: '确认删除',
      okButtonProps: { danger: true },
      cancelText: '取消',
      onOk: async () => {
        try {
          const res = await channelApi.delete(channelId);
          if (res.code === 0) {
            message.success('频道删除成功');
            loadChannelTree();
            if (selectedChannelId === channelId) {
              selectChannel(null);
            }
          } else {
            message.error(res.message || '删除失败');
          }
        } catch (error) {
          message.error('删除失败');
        }
      },
    });
  }, [loadChannelTree, selectedChannelId, selectChannel]);

  // 创建子频道
  const handleCreateChild = useCallback((parentId: string, parentName: string) => {
    openCreateChannelModal(parentId, parentName);
  }, [openCreateChannelModal]);

  // 创建根频道
  const handleCreateRootChannel = useCallback(() => {
    openCreateChannelModal();
  }, [openCreateChannelModal]);

  // 创建文档成功
  const handleCreateDocumentSuccess = useCallback((docId: string, title: string) => {
    closeCreateDocumentModal();
    loadDocuments(selectedChannelId!, searchKeyword);
    // 创建成功后自动打开编辑器
    openEditor(docId, title);
  }, [closeCreateDocumentModal, loadDocuments, selectedChannelId, searchKeyword, openEditor]);

  // 编辑文档
  const handleEditDocument = useCallback((docId: string, title: string) => {
    openEditor(docId, title);
  }, [openEditor]);

  // 删除文档
  const handleDeleteDocument = useCallback((docId: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除文档将同时删除其所有版本历史，是否继续？',
      okText: '确认删除',
      okButtonProps: { danger: true },
      cancelText: '取消',
      onOk: async () => {
        try {
          const res = await documentApi.delete(docId);
          if (res.code === 0) {
            message.success('文档删除成功');
            loadDocuments(selectedChannelId!, searchKeyword);
          } else {
            message.error(res.message || '删除失败');
          }
        } catch (error) {
          message.error('删除失败');
        }
      },
    });
  }, [loadDocuments, selectedChannelId, searchKeyword]);

  // 查看历史版本
  const handleViewHistory = useCallback((docId: string, title: string) => {
    openHistoryModal(docId, title);
  }, [openHistoryModal]);

  // 保存成功后刷新
  const handleSaveSuccess = useCallback(() => {
    if (selectedChannelId) {
      loadDocuments(selectedChannelId, searchKeyword);
    }
  }, [loadDocuments, selectedChannelId, searchKeyword]);

  return (
    <Layout className="documents-page" style={{ height: 'calc(100vh - 64px)', background: '#f5f5f5' }}>
      {/* 左侧频道树 */}
      <Sider
        width={280}
        style={{ background: '#fff', borderRight: '1px solid #e8e8e8' }}
        className="channel-sider"
      >
        <ChannelTree
          channels={channels}
          selectedId={selectedChannelId}
          onSelect={selectChannel}
          onCreateChannel={handleCreateRootChannel}
          onEditChannel={handleEditChannel}
          onDeleteChannel={handleDeleteChannel}
          onCreateChild={handleCreateChild}
          loading={channelLoading}
        />
      </Sider>

      {/* 中间内容区 */}
      <Content style={{ padding: 24, overflow: 'auto' }}>
        {selectedChannel ? (
          <div className="channel-content">
            {/* 频道头部 */}
            <Card style={{ marginBottom: 16 }}>
              <div className="channel-header">
                <Space>
                  <FolderOutlined style={{ fontSize: 24, color: '#1890ff' }} />
                  <span style={{ fontSize: 20, fontWeight: 500 }} data-testid="selected-channel-name">
                    {selectedChannel.name}
                  </span>
                </Space>
              </div>
              <div className="channel-meta" style={{ marginTop: 8, color: '#666' }}>
                <Space split={<span>|</span>}>
                  <span>路径: {selectedChannel.path}</span>
                  <span>
                    <FileTextOutlined /> {selectedChannel.doc_count} 文档
                  </span>
                  <span>
                    <FolderOutlined /> {selectedChannel.child_count} 子频道
                  </span>
                </Space>
              </div>
              {selectedChannel.description && (
                <div style={{ marginTop: 8 }}>
                  {selectedChannel.description}
                </div>
              )}
            </Card>

            {/* 文档列表 */}
            <Card>
              <DocumentList
                channelId={selectedChannelId}
                documents={documents}
                loading={documentLoading}
                pagination={documentPagination}
                onPageChange={handlePageChange}
                onCreateDocument={openCreateDocumentModal}
                onEditDocument={handleEditDocument}
                onDeleteDocument={handleDeleteDocument}
                onViewHistory={handleViewHistory}
                onSearch={handleSearch}
                searchKeyword={searchKeyword}
              />
            </Card>
          </div>
        ) : (
          <div className="empty-state" style={{ textAlign: 'center', paddingTop: 100 }}>
            <Empty
              description="请选择一个频道查看文档"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            >
              <Button
                type="primary"
                icon={<FolderAddOutlined />}
                onClick={handleCreateRootChannel}
                data-testid="create-channel-btn-empty"
              >
                创建频道
              </Button>
            </Empty>
          </div>
        )}
      </Content>

      {/* 创建频道弹窗 */}
      <CreateChannelModal
        visible={createChannelModalVisible}
        parentId={createChannelParentId}
        parentName={createChannelParentName}
        onCancel={closeCreateChannelModal}
        onSuccess={handleCreateChannelSuccess}
      />

      {/* 创建文档弹窗 */}
      <CreateDocumentModal
        visible={createDocumentModalVisible}
        channelId={selectedChannelId || ''}
        channelName={selectedChannel?.name || ''}
        onCancel={closeCreateDocumentModal}
        onSuccess={handleCreateDocumentSuccess}
      />

      {/* 文档编辑器面板 */}
      <DocumentEditorPanel
        visible={editorVisible}
        documentId={editingDocumentId}
        documentTitle={editingDocumentTitle}
        onClose={closeEditor}
        onSaveSuccess={handleSaveSuccess}
        onOpenHistory={() => handleViewHistory(editingDocumentId!, editingDocumentTitle)}
      />
    </Layout>
  );
};

export default DocumentsPage;