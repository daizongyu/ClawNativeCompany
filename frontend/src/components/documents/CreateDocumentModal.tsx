import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, Alert, message } from 'antd';
import { documentApi } from '../../services/document';

interface CreateDocumentModalProps {
  visible: boolean;
  channelId: string;
  channelName: string;
  onCancel: () => void;
  onSuccess: (documentId: string, title: string) => void;
}

export const CreateDocumentModal: React.FC<CreateDocumentModalProps> = ({
  visible,
  channelId,
  channelName,
  onCancel,
  onSuccess,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  // 弹窗打开时重置表单
  useEffect(() => {
    if (visible) {
      form.resetFields();
    }
  }, [visible, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      const res = await documentApi.create(channelId, {
        title: values.title,
        content: '',  // 创建时内容为空，编辑时填写
      });

      if (res.code === 0) {
        message.success('文档创建成功');
        onSuccess(res.data.id, res.data.title);
        form.resetFields();
      } else {
        message.error(res.message || '创建失败');
      }
    } catch (error: any) {
      if (error.errorFields) {
        // 表单验证错误，不需要处理
        return;
      }
      message.error('创建失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={`在 "${channelName}" 创建文档`}
      open={visible}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={loading}
      okText="创建并编辑"
      cancelText="取消"
      width={400}
      destroyOnClose
      data-testid="create-document-modal"
      data-modal-type="create-document"
    >
      <Form form={form} layout="vertical" data-testid="create-document-form">
        <Form.Item
          name="title"
          label="文档标题"
          rules={[
            { required: true, message: '请输入文档标题' },
            { max: 200, message: '标题最多200个字符' },
          ]}
          data-testid="doc-title-field"
        >
          <Input
            placeholder="例如：PRD 文档"
            autoFocus
            data-testid="doc-title-input"
          />
        </Form.Item>

        <Alert
          type="info"
          message="创建后将自动打开编辑器，您可以开始编写内容"
          style={{ marginTop: 8 }}
        />
      </Form>
    </Modal>
  );
};