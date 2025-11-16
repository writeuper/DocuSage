import React from 'react'
import { Card, Statistic, Row, Col, Typography } from 'antd'
import { BookOutlined, SearchOutlined, FileTextOutlined, UserOutlined } from '@ant-design/icons'

const { Title, Paragraph } = Typography

const Home: React.FC = () => {
  return (
    <div>
      <Title level={2}>欢迎使用 DocuSage 智能文档助手</Title>
      <Paragraph>高效检索、智能问答、精准溯源 - 让企业文档焕发新生</Paragraph>
      
      <Row gutter={16} style={{ marginTop: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic 
              title="文档总量" 
              value={1258} 
              prefix={<BookOutlined />}
              suffix="份"
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic 
              title="今日查询" 
              value={342} 
              prefix={<SearchOutlined />}
              suffix="次"
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic 
              title="本月新增" 
              value={89} 
              prefix={<FileTextOutlined />}
              suffix="份"
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic 
              title="活跃用户" 
              value={256} 
              prefix={<UserOutlined />}
              suffix="人"
            />
          </Card>
        </Col>
      </Row>
      
      <Card title="快速开始" style={{ marginTop: '24px' }}>
        <ul>
          <li style={{ marginBottom: '8px' }}>1. 前往「文档检索」页面，输入自然语言问题</li>
          <li style={{ marginBottom: '8px' }}>2. 在「场景工具」中使用架构图查看、文档摘要等功能</li>
          <li style={{ marginBottom: '8px' }}>3. 在「个人中心」查看您的查询历史和统计信息</li>
        </ul>
      </Card>
    </div>
  )
}

export default Home