import React from 'react';
import { Spin } from 'antd';

interface PageContainerProps {
  children: React.ReactNode;
  'data-testid'?: string;
  'data-page'?: string;
  loading?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

export const PageContainer: React.FC<PageContainerProps> = ({
  children,
  'data-testid': testId,
  'data-page': page,
  loading = false,
  className,
  style,
}) => {
  const computedTestId = testId || `page-${page}`;

  return (
    <div
      data-testid={computedTestId}
      data-page={page}
      data-loaded={!loading}
      data-loading={loading}
      className={className}
      style={{ minHeight: '100%', ...style }}
    >
      <Spin spinning={loading} size="large" tip="加载中...">
        {children}
      </Spin>
    </div>
  );
};

export default PageContainer;
