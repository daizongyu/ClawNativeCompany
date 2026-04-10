import React, { useState, useEffect, useRef } from 'react';
import { List, Avatar, Spin } from 'antd';
import { UserOutlined } from '@ant-design/icons';
import { employeeApi } from '../../api/employee';

interface Employee {
  id: string;
  name: string;
  email: string;
  type: 'human' | 'agent';
  avatar?: string;
}

interface MentionSelectProps {
  visible: boolean;
  keyword: string;
  onSelect: (employee: Employee) => void;
  onCancel: () => void;
}

const MentionSelect: React.FC<MentionSelectProps> = ({
  visible,
  keyword,
  onSelect,
  onCancel,
}) => {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [loading, setLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // 搜索员工
  useEffect(() => {
    if (!visible) return;

    const searchEmployees = async () => {
      setLoading(true);
      try {
        const res = await employeeApi.list({
          page: 1,
          pageSize: 10,
          keyword: keyword || undefined,
        });
        if (res.code === 0) {
          const list = res.data?.list || res.data?.items || [];
          setEmployees(list);
        }
      } catch (error) {
        console.error('搜索员工失败:', error);
      } finally {
        setLoading(false);
      }
    };

    // 防抖搜索
    const timer = setTimeout(searchEmployees, 200);
    return () => clearTimeout(timer);
  }, [visible, keyword]);

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
      {loading ? (
        <div style={{ padding: 20, textAlign: 'center' }}>
          <Spin size="small" />
        </div>
      ) : employees.length === 0 ? (
        <div style={{ padding: 20, textAlign: 'center', color: '#999' }}>
          未找到匹配的员工
        </div>
      ) : (
        <List
          size="small"
          dataSource={employees}
          renderItem={(employee) => (
            <List.Item
              style={{
                cursor: 'pointer',
                padding: '8px 16px',
                transition: 'background 0.2s',
              }}
              onClick={() => onSelect(employee)}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background = '#f5f5f5';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'transparent';
              }}
              data-testid={`mention-option-${employee.id}`}
            >
              <List.Item.Meta
                avatar={
                  <Avatar
                    icon={<UserOutlined />}
                    style={{
                      backgroundColor: employee.type === 'agent' ? '#87d068' : '#1890ff',
                    }}
                  />
                }
                title={employee.name}
                description={employee.email}
              />
            </List.Item>
          )}
        />
      )}
    </div>
  );
};

export default MentionSelect;
