import React, { useState, useMemo, useCallback } from 'react';
import { Tree, Input, Button, Empty, Spin } from 'antd';
import { FolderOutlined, FolderAddOutlined } from '@ant-design/icons';
import type { DataNode } from 'antd/es/tree';
import { ChannelNode } from '../../services/channel';
import { ChannelTreeNodeTitle } from './ChannelTreeNodeTitle';
import './ChannelTree.css';

interface ChannelTreeProps {
  channels: ChannelNode[];
  selectedId: string | null;
  onSelect: (channelId: string) => void;
  onCreateChannel: () => void;
  onEditChannel: (channelId: string) => void;
  onDeleteChannel: (channelId: string) => void;
  onCreateChild: (parentId: string, parentName: string) => void;
  loading?: boolean;
}

// 将后端返回的频道树数据转换为 Ant Design Tree 格式（纯函数，不修改原数据）
function renderTreeData(channels: ChannelNode[]): DataNode[] {
  return channels.map((ch) => ({
    key: ch.id,
    title: ch.name,
    icon: <FolderOutlined />,
    children: ch.children?.length > 0
      ? renderTreeData(ch.children)
      : undefined,
    // 扩展字段用于 titleRender
    doc_count: ch.doc_count,
    child_count: ch.child_count,
  } as DataNode & { doc_count: number; child_count: number }));
}

// 搜索过滤（递归，返回新数组不修改原数据）
function filterChannels(channels: ChannelNode[], keyword: string): ChannelNode[] {
  if (!keyword) return channels;

  const lowerKeyword = keyword.toLowerCase();
  
  return channels.reduce<ChannelNode[]>((acc, ch) => {
    // 名称匹配
    if (ch.name.toLowerCase().includes(lowerKeyword)) {
      acc.push({
        ...ch,
        children: ch.children ? filterChannels(ch.children, '') : [],
      });
      return acc;
    }
    // 子频道匹配
    if (ch.children?.length > 0) {
      const filteredChildren = filterChannels(ch.children, keyword);
      if (filteredChildren.length > 0) {
        acc.push({
          ...ch,
          children: filteredChildren,
        });
      }
    }
    return acc;
  }, []);
}

export const ChannelTree: React.FC<ChannelTreeProps> = ({
  channels,
  selectedId,
  onSelect,
  onCreateChannel,
  onEditChannel,
  onDeleteChannel,
  onCreateChild,
  loading,
}) => {
  const [expandedKeys, setExpandedKeys] = useState<string[]>([]);
  const [searchKeyword, setSearchKeyword] = useState('');

  // 搜索过滤后的频道（使用 useMemo 缓存）
  const filteredChannels = useMemo(
    () => filterChannels(channels, searchKeyword),
    [channels, searchKeyword]
  );

  // 自动展开所有匹配的节点
  const autoExpandKeys = useMemo(() => {
    if (!searchKeyword) return expandedKeys;
    const keys: string[] = [];
    const collectKeys = (nodes: ChannelNode[]) => {
      nodes.forEach((node) => {
        if (node.children?.length > 0) {
          keys.push(node.id);
          collectKeys(node.children);
        }
      });
    };
    collectKeys(filteredChannels);
    return keys;
  }, [searchKeyword, expandedKeys, filteredChannels]);

  // 处理展开事件（使用 useCallback 缓存）
  const handleExpand = useCallback((keys: React.Key[]) => {
    setExpandedKeys(keys as string[]);
  }, []);

  // 处理选择事件
  const handleSelect = useCallback((keys: React.Key[]) => {
    if (keys.length > 0) {
      onSelect(keys[0] as string);
    }
  }, [onSelect]);

  // 转换为 Tree 数据格式（使用 useMemo 缓存）
  const treeData = useMemo(
    () => renderTreeData(filteredChannels),
    [filteredChannels]
  );

  if (loading) {
    return (
      <div className="channel-tree-container" data-testid="channel-tree">
        <Spin tip="加载频道..." />
      </div>
    );
  }

  if (channels.length === 0) {
    return (
      <div className="channel-tree-container" data-testid="channel-tree">
        <Empty
          description="暂无频道"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        >
          <Button
            type="primary"
            icon={<FolderAddOutlined />}
            onClick={onCreateChannel}
            data-testid="create-first-channel-btn"
          >
            创建第一个频道
          </Button>
        </Empty>
      </div>
    );
  }

  return (
    <div className="channel-tree-container" data-testid="channel-tree">
      {/* 搜索框 */}
      <Input.Search
        placeholder="搜索频道"
        value={searchKeyword}
        onChange={(e) => setSearchKeyword(e.target.value)}
        style={{ marginBottom: 12 }}
        allowClear
        data-testid="channel-tree-search"
      />

      {/* 树形组件 */}
      <Tree
        showLine
        showIcon
        expandedKeys={searchKeyword ? autoExpandKeys : expandedKeys}
        onExpand={handleExpand}
        selectedKeys={selectedId ? [selectedId] : []}
        onSelect={(keys) => handleSelect(keys)}
        treeData={treeData}
        titleRender={(node: any) => (
          <ChannelTreeNodeTitle
            node={node}
            onEdit={() => onEditChannel(node.key as string)}
            onDelete={() => onDeleteChannel(node.key as string)}
            onCreateChild={() => onCreateChild(node.key as string, node.title as string)}
          />
        )}
        data-testid="channel-tree-list"
      />

      {/* 创建按钮 */}
      <Button
        type="dashed"
        block
        icon={<FolderAddOutlined />}
        onClick={onCreateChannel}
        style={{ marginTop: 12 }}
        data-testid="create-channel-btn"
        data-action="create-channel"
      >
        新建频道
      </Button>
    </div>
  );
};