import React from 'react';
import { Table } from 'antd';
import type { TableProps } from 'antd';

interface DataTableProps<T> extends TableProps<T> {
  'data-testid'?: string;
  'data-entity'?: string;
}

export const DataTable = <T extends object>({
  'data-testid': testId,
  'data-entity': entity,
  ...props
}: DataTableProps<T>): React.ReactElement => {
  const computedTestId = testId || `${entity}-table`;

  return (
    <Table
      {...props}
      data-testid={computedTestId}
      data-entity={entity}
    />
  );
};

export default DataTable;
