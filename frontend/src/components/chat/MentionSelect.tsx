import React, { useEffect, useRef } from 'react';
import { List, Avatar } from 'antd';
import { UserOutlined } from '@ant-design/icons';

interface Member {
  employee_id: string;
  employee_name?: string;
  email?: string;
  role?: string;
  employee?: {
    id: string;
    name: string;
    email: string;
    type: string;
  };
}

interface MentionSelectProps {
  visible: boolean;
  keyword: string;
  members: Member[];
  currentUserId?: string;
  onSelect: (member: Member) => void;
  onCancel: () => void;
}

const MentionSelect: React.FC<MentionSelectProps> = ({
  visible,
  keyword,
  members,
  currentUserId,
  onSelect,
  onCancel,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);

  // 获取成员显示名称
  const getMemberName = (member: Member) => {
    return member.employee_name || member.employee?.name || member.employee_id;
  };

  // 获取成员邮箱
  const getMemberEmail = (member: Member) => {
    return member.email || member.employee?.email || '';
  };

  // 过滤成员列表（排除当前用户）
  const filteredMembers = members.filter((member) => {
    // 排除当前用户
    if (currentUserId && member.employee_id === currentUserId) return false;
    // 关键词过滤
    if (!keyword) return true;
    const searchLower = keyword.toLowerCase();
    const name = getMemberName(member);
    const email = getMemberEmail(member);
    return (
      name.toLowerCase().includes(searchLower) ||
      email.toLowerCase().includes(searchLower)
    );
  });

  // 点击外部关闭
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        onCancel();
      }
    };

    if (visible) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [visible, onCancel]);

  if (!visible) return null;

  return (
    <div
      ref={containerRef}
      style={{
        position: 'absolute',
        bottom: '100%',
        left: 0,
        right: 0,
        marginBottom: 8,
        background: '#fff',
        borderRadius: 8,
        boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
        zIndex: 1000,
        maxHeight: 300,
        overflow: 'auto',
      }}
      data-testid="mention-select"
    >
      {filteredMembers.length === 0 ? (
        <div style={{ padding: 20, textAlign: 'center', color: '#999' }}>
          未找到匹配的成员
        </div>
      ) : (
        <List
          size="small"
          dataSource={filteredMembers}
          renderItem={(member) => (
            <List.Item
              style={{
                cursor: 'pointer',
                padding: '8px 16px',
                transition: 'background 0.2s',
              }}
              onClick={() => onSelect(member)}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background = '#f5f5f5';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'transparent';
              }}
              data-testid={`mention-option-${member.employee_id}`}
            >
              <List.Item.Meta
                avatar={
                  <Avatar
                    icon={<UserOutlined />}
                    style={{
                      backgroundColor: '#1890ff',
                    }}
                  />
                }
                title={getMemberName(member)}
                description={getMemberEmail(member)}
              />
            </List.Item>
          )}
        />
      )}
    </div>
  );
};

export default MentionSelect;
