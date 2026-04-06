import React from 'react';
import { Form, Input, Select, DatePicker, InputNumber, Switch, Radio, Checkbox } from 'antd';
import type { FormItemProps, InputProps, SelectProps, DatePickerProps, InputNumberProps, SwitchProps, RadioGroupProps } from 'antd';
import type { CheckboxGroupProps } from 'antd/es/checkbox';

const { TextArea } = Input;
const { Option } = Select;

interface FormFieldProps extends FormItemProps {
  'data-testid'?: string;
  'data-input-name'?: string;
  type?: 'text' | 'password' | 'email' | 'textarea' | 'select' | 'date' | 'number' | 'switch' | 'radio' | 'checkbox';
  options?: { label: string; value: string | number }[];
  inputProps?: InputProps | SelectProps | DatePickerProps | InputNumberProps | SwitchProps | RadioGroupProps | CheckboxGroupProps;
}

export const FormField: React.FC<FormFieldProps> = ({
  children,
  'data-testid': testId,
  'data-input-name': inputName,
  type = 'text',
  options,
  inputProps,
  ...props
}) => {
  const name = props.name as string;
  const computedTestId = testId || `input-${inputName || name}`;

  const renderInput = () => {
    const commonProps = {
      'data-testid': computedTestId,
      'data-input-name': inputName || name,
    };

    switch (type) {
      case 'password':
        return <Input.Password {...(inputProps as InputProps)} {...commonProps} />;
      case 'textarea':
        return <TextArea {...(inputProps as any)} {...commonProps} />;
      case 'select':
        return (
          <Select {...(inputProps as SelectProps)} {...commonProps}>
            {options?.map((opt) => (
              <Option key={opt.value} value={opt.value}>
                {opt.label}
              </Option>
            ))}
          </Select>
        );
      case 'date':
        return <DatePicker {...(inputProps as any)} {...commonProps} style={{ width: '100%' }} />;
      case 'number':
        return <InputNumber {...(inputProps as any)} {...commonProps} style={{ width: '100%' }} />;
      case 'switch':
        return <Switch {...(inputProps as any)} {...commonProps} />;
      case 'radio':
        return (
          <Radio.Group {...(inputProps as any)} {...commonProps}>
            {options?.map((opt) => (
              <Radio key={opt.value} value={opt.value}>
                {opt.label}
              </Radio>
            ))}
          </Radio.Group>
        );
      case 'checkbox':
        return (
          <Checkbox.Group {...(inputProps as any)} {...commonProps}>
            {options?.map((opt) => (
              <Checkbox key={opt.value} value={opt.value}>
                {opt.label}
              </Checkbox>
            ))}
          </Checkbox.Group>
        );
      case 'email':
        return <Input type="email" {...(inputProps as InputProps)} {...commonProps} />;
      default:
        return <Input {...(inputProps as InputProps)} {...commonProps} />;
    }
  };

  return (
    <Form.Item {...props} data-testid={`form-item-${name}`}>
      {children || renderInput()}
    </Form.Item>
  );
};

export default FormField;
