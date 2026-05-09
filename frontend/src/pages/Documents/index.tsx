import React, { useEffect, useCallback } from 'react';
import { Layout, Card, Empty, Button, Space, message, Modal } from 'antd';
import { FileTextOutlined, FolderOutlined, FolderAddOutlined } from '@ant-design/icons';
import { useDocumentStore } from '../../stores/documentStore';
import { channelApi, ChannelNode } from '../../services/channel';
import { ChannelTree } from '../../components/documents/ChannelTree';
import { CreateChannelModal } from '../../components/documents/CreateChannelModal';

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

  // 初始化加载频道树
  useEffect(() => {
    loadChannelTree();
  }, [loadChannelTree]);

  // 选择频道时更新详情
  useEffect(() => {
    if (selectedChannelId) {
      const channel = findChannelNode(channels, selectedChannelId);
      setSelectedChannel(channel);
    } else {
      setSelectedChannel(null);
    }
  }, [selectedChannelId, channels, setSelectedChannel]);

  // 创建频道成功
  const handleCreateChannelSuccess = useCallback((newChannelId: string) => {
    closeCreateChannelModal();
    loadChannelTree();
    // 创建成功后自动选择新频道
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

            {/* 文档列表占位 */}
            <Card>
              <Empty
                description="Phase 3 将实现文档列表和编辑器"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              >
                <Space>
                  <Button
                    type="primary"
                    icon={<FileTextOutlined />}
                    data-testid="create-document-btn-placeholder"
                  >
                    创建文档 (Phase 3)
                  </Button>
                  <Button
                    icon={<FolderAddOutlined />}
                    onClick={() => handleCreateChild(selectedChannel.id, selectedChannel.name)}
                  >
                    创建子频道
                  </Button>
                </Space>
              </Empty>
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
    </Layout>
  );
};

export default DocumentsPage;