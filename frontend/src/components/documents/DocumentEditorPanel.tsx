import React, { useRef, useState, useEffect, useCallback } from 'react';
import { Button, Space, Tag, Spin, Modal, message } from 'antd';
import { FileTextOutlined, HistoryOutlined, SaveOutlined, CloseOutlined } from '@ant-design/icons';
import Vditor from 'vditor';
import 'vditor/dist/index.css';
import './DocumentEditorPanel.css';  // 在vditor CSS之后导入，确保样式覆盖生效
import { documentApi } from '../../services/document';

interface DocumentEditorPanelProps {
  visible: boolean;
  documentId: string | null;
  documentTitle: string;
  onClose: () => void;
  onSaveSuccess: () => void;
  onOpenHistory: () => void;
}

// 格式化文件大小
function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

// 注入tooltip样式覆盖（解决CSS伪元素无法被CSS文件覆盖的问题）
function injectTooltipStyles() {
  const styleId = 'vditor-tooltip-override-styles';
  if (document.getElementById(styleId)) return; // 已存在则不重复注入
  
  const style = document.createElement('style');
  style.id = styleId;
  style.textContent = `
    /* Tooltip向右显示 */
    .vditor-toolbar button.vditor-tooltipped::after {
      position: absolute !important;
      left: 100% !important;
      top: 50% !important;
      right: auto !important;
      bottom: auto !important;
      transform: translateY(-50%) !important;
      margin-left: 12px !important;
      margin-top: 0 !important;
      margin-right: 0 !important;
      margin-bottom: 0 !important;
      /* 确保文字完整显示 */
      white-space: nowrap !important;
      height: auto !important;
      min-height: 20px !important;
      line-height: 20px !important;
      padding: 6px 10px !important;
      font-size: 13px !important;
      max-width: 200px !important;
      overflow: visible !important;
      text-overflow: clip !important;
    }
    
    /* Tooltip箭头紧贴tooltip左侧 */
    .vditor-toolbar button.vditor-tooltipped::before {
      position: absolute !important;
      left: 100% !important;
      top: 50% !important;
      right: auto !important;
      bottom: auto !important;
      transform: translateY(-50%) !important;
      margin-left: 6px !important;
      margin-top: 0 !important;
      margin-right: 0 !important;
      margin-bottom: 0 !important;
      /* 箭头指向按钮（左边） */
      border-right: 7px solid #4b4b4b !important;
      border-left: none !important;
      border-top: 7px solid transparent !important;
      border-bottom: 7px solid transparent !important;
    }
  `;
  document.head.appendChild(style);
}

export const DocumentEditorPanel: React.FC<DocumentEditorPanelProps> = ({
  visible,
  documentId,
  documentTitle,
  onClose,
  onSaveSuccess,
  onOpenHistory,
}) => {
  const vditorRef = useRef<Vditor | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [content, setContent] = useState('');
  const [currentVersion, setCurrentVersion] = useState(1);
  const [hasChanges, setHasChanges] = useState(false);

  // 加载文档内容
  const loadDocumentContent = useCallback(async (id: string) => {
    setLoading(true);
    try {
      const res = await documentApi.getById(id);
      if (res.code === 0) {
        setContent(res.data.content || '');
        setCurrentVersion(res.data.version);
        setHasChanges(false);
      } else {
        message.error(res.message || '加载文档失败');
      }
    } catch (error) {
      message.error('加载文档失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // 加载文档内容
  useEffect(() => {
    if (visible && documentId) {
      loadDocumentContent(documentId);
    } else if (!visible) {
      // 关闭时销毁编辑器
      if (vditorRef.current) {
        vditorRef.current.destroy();
        vditorRef.current = null;
      }
      setContent('');
      setHasChanges(false);
    }
  }, [visible, documentId, loadDocumentContent]);

  // 初始化编辑器
  useEffect(() => {
    if (!containerRef.current || !visible || loading || !documentId) return;

    // 如果编辑器已存在且内容已更新，更新编辑器内容
    if (vditorRef.current) {
      vditorRef.current.setValue(content);
      return;
    }

    vditorRef.current = new Vditor(containerRef.current, {
      height: 'calc(100vh - 120px)',
      width: '100%',
      mode: 'wysiwyg',  // 所见即所得模式
      placeholder: '开始编辑文档...',
      value: content,
      theme: 'classic',
      icon: 'ant',
      
      // 工具栏配置
      toolbar: [
        'headings',
        'bold',
        'italic',
        'strike',
        '|',
        'line',
        'quote',
        '|',
        'list',
        'ordered-list',
        'check',
        '|',
        'code',
        'inline-code',
        'insert-before',
        'insert-after',
        '|',
        'table',
        '|',
        'undo',
        'redo',
        '|',
        'edit-mode',
        {
          name: 'more',
          toolbar: [
            'both',
            'code-theme',
            'content-theme',
            'export',
            'outline',
            'preview',
            'devtools',
          ],
        },
      ],
      
      // 输入回调（检测变化）
      input: (value) => {
        setHasChanges(value !== content);
      },
      
      // 初始化后设置内容并修复工具栏样式
      after: () => {
        vditorRef.current?.setValue(content);
        // 清除vditor自动设置的居中padding
        const toolbar = document.querySelector('.vditor-toolbar') as HTMLElement;
        if (toolbar) {
          toolbar.style.paddingLeft = '';
          toolbar.style.paddingRight = '';
        }
        
        // 动态注入tooltip样式，确保覆盖vditor默认样式
        injectTooltipStyles();
      },
      
      // 缓存（自动保存）
      cache: {
        enable: true,
        id: `doc-${documentId}`,
      },
      
      // 预览配置
      preview: {
        markdown: {
          toc: true,
          mark: true,
        },
      },
      
      // 计数器
      counter: {
        enable: true,
        type: 'text',
      },
    });

    return () => {
      if (vditorRef.current) {
        vditorRef.current.destroy();
        vditorRef.current = null;
      }
    };
  }, [visible, loading, content, documentId]);

  // 保存文档
  const handleSave = useCallback(async () => {
    if (!vditorRef.current || !documentId) return;

    setSaving(true);
    const newContent = vditorRef.current.getValue();

    try {
      const res = await documentApi.saveContent(documentId, {
        content: newContent,
        expected_version: currentVersion,
      });

      if (res.code === 409) {
        // 版本冲突
        Modal.confirm({
          title: '版本冲突',
          content: '文档已被他人修改，是否加载最新版本？您的修改将丢失。',
          okText: '加载最新版本',
          cancelText: '取消',
          onOk: () => {
            loadDocumentContent(documentId);
          },
        });
      } else if (res.code === 0) {
        // 保存成功
        message.success('保存成功');
        setCurrentVersion(res.data.version);
        setContent(newContent);
        setHasChanges(false);
        onSaveSuccess();
      } else {
        message.error(res.message || '保存失败');
      }
    } catch (error: any) {
      // 检查是否是版本冲突错误
      if (error.response?.status === 409) {
        Modal.confirm({
          title: '版本冲突',
          content: '文档已被他人修改，是否加载最新版本？您的修改将丢失。',
          okText: '加载最新版本',
          cancelText: '取消',
          onOk: () => {
            loadDocumentContent(documentId);
          },
        });
      } else {
        message.error('保存失败');
      }
    } finally {
      setSaving(false);
    }
  }, [documentId, currentVersion, loadDocumentContent, onSaveSuccess]);

  // 快捷键监听
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault();
        if (visible && documentId) {
          handleSave();
        }
      }
    };

    if (visible) {
      window.addEventListener('keydown', handleKeyDown);
    }

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [visible, documentId, handleSave]);

  // 关闭时检查未保存
  const handleClose = useCallback(() => {
    if (hasChanges) {
      Modal.confirm({
        title: '未保存的修改',
        content: '文档有未保存的修改，是否保存后再关闭？',
        okText: '保存并关闭',
        cancelText: '不保存直接关闭',
        onOk: async () => {
          await handleSave();
          onClose();
        },
        onCancel: onClose,
      });
    } else {
      onClose();
    }
  }, [hasChanges, handleSave, onClose]);

  // 获取当前内容大小
  const contentSize = vditorRef.current?.getValue()?.length || content.length || 0;

  if (!visible) return null;

  return (
    <div
      className="document-editor-panel visible"
      data-testid="document-editor-panel"
      data-document-id={documentId || ''}
    >
      {/* 顶部工具栏 */}
      <div className="editor-toolbar">
        <div className="editor-title">
          <FileTextOutlined style={{ marginRight: 8 }} />
          <span data-testid="editor-document-title">{documentTitle}</span>
          <Tag color={hasChanges ? 'orange' : 'green'} style={{ marginLeft: 12 }}>
            {hasChanges ? '未保存' : `v${currentVersion}`}
          </Tag>
        </div>
        <div className="editor-actions">
          <Space>
            <Button
              icon={<HistoryOutlined />}
              onClick={onOpenHistory}
              data-testid="editor-history-btn"
              data-action="view-history"
            >
              历史版本
            </Button>
            <Button
              type="primary"
              icon={<SaveOutlined />}
              loading={saving}
              onClick={handleSave}
              data-testid="editor-save-btn"
              data-action="save-document"
            >
              保存 (Ctrl+S)
            </Button>
            <Button
              icon={<CloseOutlined />}
              onClick={handleClose}
              data-testid="editor-close-btn"
              data-action="close-editor"
            >
              关闭
            </Button>
          </Space>
        </div>
      </div>

      {/* 编辑器容器 */}
      <Spin spinning={loading}>
        <div
          ref={containerRef}
          id={`vditor-${documentId || 'new'}`}
          className="vditor-container"
          data-testid="vditor-editor"
        />
      </Spin>

      {/* 底部状态栏 */}
      <div className="editor-status-bar">
        <span>版本: v{currentVersion}</span>
        <span>大小: {formatFileSize(contentSize)}</span>
        <span>快捷键: Ctrl+S 保存</span>
      </div>
    </div>
  );
};