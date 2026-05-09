import React, { useState, useEffect, useCallback } from 'react';
import { Modal, List, Button, Tag, Spin, message, Empty } from 'antd';
import ReactMarkdown from 'react-markdown';
import { documentApi, DocumentVersion } from '../../services/document';

interface DocumentHistoryModalProps {
  visible: boolean;
  documentId: string;
  documentTitle: string;
  onCancel: () => void;
  onRestore: (newVersion: number) => void;
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

export const DocumentHistoryModal: React.FC<DocumentHistoryModalProps> = ({
  visible,
  documentId,
  documentTitle,
  onCancel,
  onRestore,
}) => {
  const [versions, setVersions] = useState<DocumentVersion[]>([]);
  const [loading, setLoading] = useState(false);
  const [previewVisible, setPreviewVisible] = useState(false);
  const [previewContent, setPreviewContent] = useState('');
  const [previewVersion, setPreviewVersion] = useState<number>(0);
  const [previewLoading, setPreviewLoading] = useState(false);

  // 加载版本列表
  const loadVersions = useCallback(async (id: string) => {
    setLoading(true);
    try {
      const res = await documentApi.getVersions(id);
      if (res.code === 0) {
        setVersions(res.data.versions || []);
      } else {
        message.error(res.message || '加载版本历史失败');
      }
    } catch (error) {
      message.error('加载版本历史失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // 弹窗打开时加载版本列表
  useEffect(() => {
    if (visible && documentId) {
      loadVersions(documentId);
    } else if (!visible) {
      setVersions([]);
      setPreviewVisible(false);
    }
  }, [visible, documentId, loadVersions]);

  // 查看版本内容
  const handlePreview = useCallback(async (version: number) => {
    setPreviewLoading(true);
    setPreviewVersion(version);
    setPreviewVisible(true);
    try {
      const res = await documentApi.getVersionContent(documentId, version);
      if (res.code === 0) {
        setPreviewContent(res.data.content || '');
      } else {
        message.error(res.message || '加载版本内容失败');
        setPreviewVisible(false);
      }
    } catch (error) {
      message.error('加载版本内容失败');
      setPreviewVisible(false);
    } finally {
      setPreviewLoading(false);
    }
  }, [documentId]);

  // 恢复版本
  const handleRestore = useCallback(async (version: number) => {
    Modal.confirm({
      title: '确认恢复',
      content: `恢复到版本 v${version} 将覆盖当前内容，是否继续？`,
      okText: '确认恢复',
      cancelText: '取消',
      onOk: async () => {
        try {
          const res = await documentApi.restoreVersion(documentId, version);
          if (res.code === 0) {
            message.success(`已恢复到版本 v${version}`);
            onRestore(res.data.version);
            loadVersions(documentId);  // 刷新列表
          } else {
            message.error(res.message || '恢复失败');
          }
        } catch (error) {
          message.error('恢复失败');
        }
      },
    });
  }, [documentId, onRestore, loadVersions]);

  // 从预览弹窗恢复
  const handleRestoreFromPreview = useCallback(async () => {
    setPreviewVisible(false);
    handleRestore(previewVersion);
  }, [previewVersion, handleRestore]);

  // 当前版本号（列表中第一个版本）
  const currentVersion = versions.length > 0 ? versions[0].version : 0;

  return (
    <>
      {/* 版本列表弹窗 */}
      <Modal
        title={`历史版本 - ${documentTitle}`}
        open={visible}
        onCancel={onCancel}
        footer={null}
        width={700}
        destroyOnClose
        data-testid="document-history-modal"
        data-modal-type="history"
      >
        <Spin spinning={loading}>
          {versions.length === 0 && !loading ? (
            <Empty
              description="暂无版本历史"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          ) : (
            <List
              dataSource={versions}
              renderItem={(v) => (
                <List.Item
                  actions={[
                    <Button
                      size="small"
                      onClick={() => handlePreview(v.version)}
                      data-testid={`view-version-${v.version}`}
                      data-action="preview-version"
                    >
                      查看
                    </Button>,
                    v.version !== currentVersion && (
                      <Button
                        size="small"
                        type="primary"
                        onClick={() => handleRestore(v.version)}
                        data-testid={`restore-version-${v.version}`}
                        data-action="restore-version"
                      >
                        恢复
                      </Button>
                    ),
                  ]}
                  data-testid={`version-item-${v.version}`}
                >
                  <List.Item.Meta
                    avatar={<Tag color={v.version === currentVersion ? 'green' : 'blue'}>v{v.version}</Tag>}
                    title={
                      v.version === currentVersion
                        ? '当前版本'
                        : `版本 v${v.version}`
                    }
                    description={v.summary || '无摘要'}
                  />
                  <div className="version-meta" style={{ textAlign: 'right', minWidth: 120 }}>
                    <div>{v.editor_name}</div>
                    <div style={{ fontSize: 12, color: '#999' }}>{formatTime(v.created_at)}</div>
                  </div>
                </List.Item>
              )}
              data-testid="version-list"
            />
          )}
        </Spin>
      </Modal>

      {/* 版本预览弹窗 */}
      <Modal
        title={`版本 v${previewVersion} 内容预览`}
        open={previewVisible}
        onCancel={() => setPreviewVisible(false)}
        footer={[
          <Button key="close" onClick={() => setPreviewVisible(false)}>
            关闭
          </Button>,
          previewVersion !== currentVersion && (
            <Button
              key="restore"
              type="primary"
              onClick={handleRestoreFromPreview}
              data-testid="restore-from-preview-btn"
              data-action="restore-version"
            >
              恢复此版本
            </Button>
          ),
        ]}
        width={800}
        destroyOnClose
        data-testid="version-preview-modal"
        data-modal-type="preview"
      >
        <Spin spinning={previewLoading}>
          <div 
            className="version-preview-content" 
            style={{ 
              maxHeight: 400, 
              overflow: 'auto',
              padding: 16,
              background: '#f5f5f5',
              borderRadius: 4,
            }}
          >
            {previewContent ? (
              <ReactMarkdown>{previewContent}</ReactMarkdown>
            ) : (
              <Empty description="无内容" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            )}
          </div>
        </Spin>
      </Modal>
    </>
  );
};