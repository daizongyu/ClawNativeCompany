import React, { useState } from 'react';
import { Dropdown, Tag } from 'antd';
import { FolderOutlined, FolderAddOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';

interface ChannelTreeNodeTitleProps {
  node: {
    key: string | number;
    title?: string | React.ReactNode;
    doc_count?: number;
    child_count?: number;
  };
  onEdit: () => void;
  onDelete: () => void;
  onCreateChild: () => void;
}

export const ChannelTreeNodeTitle: React.FC<ChannelTreeNodeTitleProps> = ({
  node,
  onEdit,
  onDelete,
  onCreateChild,
}) => {
  const [menuVisible, setMenuVisible] = useState(false);

  const menuItems: MenuProps['items'] = [
    {
      key: 'create-child',
      label: '新建子频道',
      icon: <FolderAddOutlined />,
      onClick: () => {
        setMenuVisible(false);
        onCreateChild();
      },
    },
    {
      key: 'edit',
      label: '编辑频道',
      icon: <EditOutlined />,
      onClick: () => {
        setMenuVisible(false);
        onEdit();
      },
    },
    {
      type: 'divider',
    },
    {
      key: 'delete',
      label: '删除频道',
      icon: <DeleteOutlined />,
      danger: true,
      onClick: () => {
        setMenuVisible(false);
        onDelete();
      },
    },
  ];

  return (
    <Dropdown
      trigger={['contextMenu']}
      open={menuVisible}
      onOpenChange={setMenuVisible}
      menu={{ items: menuItems }}
    >
      <span
        className="channel-node-title"
        data-testid={`channel-node-${node.key}`}
        data-channel-id={node.key}
        data-channel-name={typeof node.title === 'string' ? node.title : ''}
      >
        <FolderOutlined style={{ marginRight: 8, color: '#1890ff' }} />
        <span className="channel-name">{node.title}</span>
        <span
          className="channel-node-count"
          style={{ marginLeft: 8, fontSize: 12, color: '#999' }}
        >
          {node.doc_count ?? 0} 文档
          {(node.child_count ?? 0) > 0 && (
            <Tag style={{ marginLeft: 4 }} color="blue">
              {node.child_count} 子频道
            </Tag>
          )}
        </span>
      </span>
    </Dropdown>
  );
};