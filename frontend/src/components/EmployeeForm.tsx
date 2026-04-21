import React, { useEffect } from 'react';
import {
  Modal,
  Form,
  Input,
  Select,
  Radio,
  message,
} from 'antd';
import { employeeApi, Employee, CreateEmployeeRequest, UpdateEmployeeRequest } from '../services/employee';

interface EmployeeFormProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  initialValues?: Partial<Employee>;
  mode: 'create' | 'edit';
}

const EmployeeForm: React.FC<EmployeeFormProps> = ({
  visible,
  onCancel,
  onSuccess,
  initialValues,
  mode,
}) => {
  const [form] = Form.useForm();
  const isEdit = mode === 'edit';

  useEffect(() => {
    if (visible) {
      if (initialValues) {
        form.setFieldsValue({
          ...initialValues,
          skills: initialValues.skills?.join(', ') || '',
        });
      } else {
        form.resetFields();
        form.setFieldsValue({ type: 'human', status: 'active' });
      }
    }
  }, [visible, initialValues, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();

      // 处理技能字段
      const skills = values.skills
        ? values.skills.split(/[,，]/).map((s: string) => s.trim()).filter(Boolean)
        : [];

      const data = {
        ...values,
        skills,
      };

      if (isEdit && initialValues?.id) {
        await employeeApi.update(initialValues.id, data as UpdateEmployeeRequest);
        message.success('员工信息已更新');
      } else {
        await employeeApi.create(data as CreateEmployeeRequest);
        message.success('员工已创建');
      }

      onSuccess();
    } catch (err: any) {
      message.error(err.message || '操作失败');
    }
  };

  return (
    <Modal
      title={isEdit ? '编辑员工' : '创建员工'}
      open={visible}
      onCancel={onCancel}
      onOk={handleSubmit}
      width={600}
      data-testid="employee-form-modal"
    >
      <Form
        form={form}
        layout="vertical"
      >
        {/* 用户名 */}
        <Form.Item
          name="username"
          label="用户名"
          rules={[
            { required: true, message: '请输入用户名' },
            { min: 3, message: '用户名至少3个字符' },
            { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线' },
          ]}
        >
          <Input
            placeholder="例如：zhangsan"
            disabled={isEdit}
            data-testid="employee-username-input"
          />
        </Form.Item>

        {/* 显示名称 */}
        <Form.Item
          name="display_name"
          label="显示名称"
          rules={[{ required: true, message: '请输入显示名称' }]}
        >
          <Input
            placeholder="例如：张三"
            data-testid="employee-displayname-input"
          />
        </Form.Item>

        {/* 类型 */}
        <Form.Item
          name="type"
          label="类型"
          rules={[{ required: true, message: '请选择类型' }]}
        >
          <Radio.Group disabled={isEdit} data-testid="employee-type-radio">
            <Radio value="human">人类员工</Radio>
            <Radio value="agent">Agent</Radio>
          </Radio.Group>
        </Form.Item>

        {/* 邮箱 */}
        <Form.Item
          name="email"
          label="邮箱"
          rules={[
            { required: true, message: '请输入邮箱' },
            { type: 'email', message: '请输入有效的邮箱地址' },
          ]}
        >
          <Input
            placeholder="例如：zhangsan@example.com"
            data-testid="employee-email-input"
          />
        </Form.Item>

        {/* 密码（仅创建时） */}
        {!isEdit && (
          <Form.Item
            name="password"
            label="密码"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少6个字符' },
            ]}
          >
            <Input.Password
              placeholder="请输入密码"
              data-testid="employee-password-input"
            />
          </Form.Item>
        )}

        {/* 职能 */}
        <Form.Item
          name="role"
          label="职能"
        >
          <Input
            placeholder="例如：开发工程师、产品经理"
            data-testid="employee-role-input"
          />
        </Form.Item>

        {/* 部门 */}
        <Form.Item
          name="department"
          label="部门"
        >
          <Input
            placeholder="例如：技术部、产品部"
            data-testid="employee-department-input"
          />
        </Form.Item>

        {/* 职位 */}
        <Form.Item
          name="position"
          label="职位"
        >
          <Input
            placeholder="例如：高级工程师、总监"
            data-testid="employee-position-input"
          />
        </Form.Item>

        {/* 电话 */}
        <Form.Item
          name="phone"
          label="电话"
        >
          <Input
            placeholder="例如：13800138000"
            data-testid="employee-phone-input"
          />
        </Form.Item>

        {/* 头像 */}
        <Form.Item
          name="avatar"
          label="头像 URL"
        >
          <Input
            placeholder="https://example.com/avatar.jpg"
            data-testid="employee-avatar-input"
          />
        </Form.Item>

        {/* 技能 */}
        <Form.Item
          name="skills"
          label="技能"
          extra="多个技能用逗号分隔"
        >
          <Input.TextArea
            placeholder="例如：Go, React, Python"
            rows={2}
            data-testid="employee-skills-input"
          />
        </Form.Item>

        {/* 状态 */}
        {isEdit && (
          <Form.Item
            name="status"
            label="状态"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select data-testid="employee-status-select">
              <Select.Option value="active">活跃</Select.Option>
              <Select.Option value="inactive">停用</Select.Option>
            </Select>
          </Form.Item>
        )}
      </Form>
    </Modal>
  );
};

export default EmployeeForm;
