import React from 'react'
import { Card, Row, Col, Statistic, Progress, List, Tag, Space } from 'antd'
import { 
  UserOutlined, 
  FileTextOutlined, 
  CloudServerOutlined,
  ApiOutlined 
} from '@ant-design/icons'

const MonitorPanel: React.FC = () => {
  const systemStats = [
    {
      title: '在线用户',
      value: 1128,
      prefix: <UserOutlined />,
      suffix: '人',
    },
    {
      title: '文档总数',
      value: 25680,
      prefix: <FileTextOutlined />,
      suffix: '篇',
    },
    {
      title: 'API调用',
      value: 89234,
      prefix: <ApiOutlined />,
      suffix: '次',
    },
    {
      title: '存储使用',
      value: 68.5,
      prefix: <CloudServerOutlined />,
      suffix: '%',
    },
  ]

  const recentLogs = [
    {
      time: '2024-12-16 15:30:22',
      type: 'info',
      message: '用户 user@example.com 登录系统',
    },
    {
      time: '2024-12-16 15:28:15',
      type: 'warning',
      message: 'API调用频率接近限制',
    },
    {
      time: '2024-12-16 15:25:08',
      type: 'success',
      message: '文档上传成功: report.pdf',
    },
    {
      time: '2024-12-16 15:22:33',
      type: 'error',
      message: '数据库连接超时',
    },
  ]

  const getLogTypeColor = (type: string) => {
    switch (type) {
      case 'success': return 'green'
      case 'warning': return 'orange'
      case 'error': return 'red'
      default: return 'blue'
    }
  }

  return (
    <div style={{ padding: '24px' }}>
      <h2>监控面板</h2>
      
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        {systemStats.map((stat, index) => (
          <Col xs={24} sm={12} md={6} key={index}>
            <Card>
              <Statistic
                title={stat.title}
                value={stat.value}
                prefix={stat.prefix}
                suffix={stat.suffix}
              />
            </Card>
          </Col>
        ))}
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card title="系统资源使用情况">
            <div style={{ marginBottom: '16px' }}>
              <div style={{ marginBottom: '8px' }}>CPU使用率</div>
              <Progress percent={45} status="active" />
            </div>
            <div style={{ marginBottom: '16px' }}>
              <div style={{ marginBottom: '8px' }}>内存使用率</div>
              <Progress percent={68} status="active" />
            </div>
            <div style={{ marginBottom: '16px' }}>
              <div style={{ marginBottom: '8px' }}>磁盘使用率</div>
              <Progress percent={72} />
            </div>
            <div>
              <div style={{ marginBottom: '8px' }}>网络带宽</div>
              <Progress percent={35} status="active" />
            </div>
          </Card>
        </Col>
        
        <Col xs={24} md={12}>
          <Card title="系统日志">
            <List
              dataSource={recentLogs}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    title={
                      <Space>
                        <span>{item.time}</span>
                        <Tag color={getLogTypeColor(item.type)}>
                          {item.type.toUpperCase()}
                        </Tag>
                      </Space>
                    }
                    description={item.message}
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default MonitorPanel