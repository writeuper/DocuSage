import React from 'react'
import { Card, Row, Col, Button, Space } from 'antd'
import { ToolOutlined, FileTextOutlined, RobotOutlined } from '@ant-design/icons'

const SceneTools: React.FC = () => {
  return (
    <div style={{ padding: '24px' }}>
      <h2>场景工具</h2>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={8}>
          <Card 
            title={
              <Space>
                <RobotOutlined />
                智能问答
              </Space>
            }
            extra={<Button type="primary">开始使用</Button>}
          >
            <p>基于AI的智能问答系统，帮助您快速找到答案</p>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <Card 
            title={
              <Space>
                <FileTextOutlined />
                文档摘要
              </Space>
            }
            extra={<Button type="primary">开始使用</Button>}
          >
            <p>自动生成文档摘要，快速了解文档核心内容</p>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <Card 
            title={
              <Space>
                <ToolOutlined />
                数据分析
              </Space>
            }
            extra={<Button type="primary">开始使用</Button>}
          >
            <p>深入分析文档数据，提供洞察和建议</p>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default SceneTools