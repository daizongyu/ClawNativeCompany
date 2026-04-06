import React from 'react';
import { Button as AntButton } from 'antd';
import type { ButtonProps as AntButtonProps } from 'antd';

interface ButtonProps extends AntButtonProps {
  testId?: string;
  action?: string;
  entity?: string;
}

export const Button: React.FC<ButtonProps> = ({
  children,
  testId,
  action,
  entity,
  ...props
}) => {
  // 自动生成 testid: 如果提供了 action 和 entity，组合成 {entity}-{action}-btn
  const computedTestId = testId || (action && entity ? `${entity}-${action}-btn` : action);

  return (
    <AntButton
      {...props}
      data-testid={computedTestId}
      data-action={action}
      data-entity={entity}
    >
      {children}
    </AntButton>
  );
};

export default Button;
