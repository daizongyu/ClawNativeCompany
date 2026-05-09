import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, Radio, Alert, message } from 'antd';
import { channelApi } from '../../services/channel';

interface CreateChannelModalProps {
  visible: boolean;
  parentId: string | null;
  parentName: string;
  onCancel: () => void;
  onSuccess: (channelId: string) => void;
}

export const CreateChannelModal: React.FC<CreateChannelModalProps> = ({
  visible,
  parentId,
  parentName,
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

      const res = await channelApi.create({
        name: values.name,
        type: values.type,
        description: values.description,
        parent_id: parentId || undefined,
      });

      if (res.code === 0) {
        message.success(parentId ? '子频道创建成功' : '频道创建成功');
        onSuccess(res.data.id);
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
      title={parentId ? `在 "${parentName}" 下创建子频道` : '创建频道'}
      open={visible}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={loading}
      okText="创建"
      cancelText="取消"
      width={400}
      destroyOnClose
      data-testid="create-channel-modal"
      data-modal-type="create-channel"
    >
      <Form form={form} layout="vertical" data-testid="create-channel-form">
        {/* 频道名称 */}
        <Form.Item
          name="name"
          label="频道名称"
          rules={[
            { required: true, message: '请输入频道名称' },
            { max: 100, message: '名称最多100个字符' },
            { pattern: /^[^\/]+$/, message: '名称不能包含 /' },
          ]}
          data-testid="channel-name-field"
        >
          <Input
            placeholder="例如：产品需求"
            autoFocus
            data-testid="channel-name-input"
          />
        </Form.Item>

        {/* 频道类型 */}
        <Form.Item
          name="type"
          label="频道类型"
          initialValue="public"
          data-testid="channel-type-field"
        >
          <Radio.Group data-testid="channel-type-radio">
            <Radio value="public" data-testid="channel-type-public">
              公开频道
            </Radio>
            <Radio value="private" data-testid="channel-type-private">
              私有频道
            </Radio>
          </Radio.Group>
        </Form.Item>

        {/* 描述 */}
        <Form.Item
          name="description"
          label="描述"
          data-testid="channel-desc-field"
        >
          <Input.TextArea
            rows={2}
            placeholder="频道用途说明（可选）"
            data-testid="channel-desc-input"
          />
        </Form.Item>

        {/* 提示 */}
        {parentId && (
          <Alert
            type="info"
            message={`子频道将继承父频道 "${parentName}" 的权限设置`}
            style={{ marginTop: 8 }}
          />
        )}
      </Form>
    </Modal>
  );
};