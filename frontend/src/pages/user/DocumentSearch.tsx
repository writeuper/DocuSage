import React, { useState } from 'react'
import { Input, Button, Card, List, Tag, Typography, Empty, Spin, Select, DatePicker, Space } from 'antd'
import { SearchOutlined, FileTextOutlined, ClockCircleOutlined, FilterOutlined } from '@ant-design/icons'
import { useQuery } from 'react-query'
import { searchDocuments } from '@/services/queryService'
import type { SearchResult } from '@/types'

const { Title, Paragraph } = Typography
const { Search } = Input
const { Option } = Select
const { RangePicker } = DatePicker

const DocumentSearch: React.FC = () => {
  const [searchText, setSearchText] = useState('')
  const [filters, setFilters] = useState({
    docType: '',
    department: '',
    dateRange: []
  })

  const { data, isLoading, refetch } = useQuery<SearchResult>(
    ['search', searchText, filters],
    async () => {
      const response = await searchDocuments({ query: searchText, ...filters })
      return response.data
    },
    { enabled: false }
  )

  const handleSearch = () => {
    if (searchText.trim()) {
      refetch()
    }
  }

  const handleFilterChange = (key: string, value: any) => {
    setFilters(prev => ({ ...prev, [key]: value }))
  }

  return (
    <div>
      <Title level={2}>文档检索</Title>
      
      <Card style={{ marginBottom: '16px' }}>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap', alignItems: 'end' }}>
          <Search
            placeholder="输入您的问题或关键词"
            allowClear
            enterButton={<Button type="primary" icon={<SearchOutlined />}>搜索</Button>}
            size="large"
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onPressEnter={handleSearch}
            style={{ flex: 1, minWidth: '300px' }}
          />
          
          <Select
            placeholder="文档类型"
            style={{ width: 150 }}
            onChange={(value) => handleFilterChange('docType', value)}
            allowClear
          >
            <Option value="pdf">PDF文档</Option>
            <Option value="doc">Word文档</Option>
            <Option value="xls">Excel表格</Option>
            <Option value="txt">文本文件</Option>
          </Select>
          
          <Select
            placeholder="所属部门"
            style={{ width: 150 }}
            onChange={(value) => handleFilterChange('department', value)}
            allowClear
          >
            <Option value="tech">技术部</Option>
            <Option value="hr">人力资源部</Option>
            <Option value="marketing">市场部</Option>
            <Option value="sales">销售部</Option>
          </Select>
          
          <RangePicker
            placeholder={['开始日期', '结束日期']}
            onChange={(dates) => handleFilterChange('dateRange', dates)}
          />
          
          <Button
            type="default"
            icon={<FilterOutlined />}
            onClick={handleSearch}
          >
            应用筛选
          </Button>
        </div>
      </Card>

      {isLoading ? (
        <div className="loading-container">
          <Spin size="large" tip="正在检索中..." />
        </div>
      ) : data?.results.length === 0 ? (
        <Empty description="暂无检索结果" />
      ) : (
        <div>
          {data?.answer && (
            <Card title="智能回答" style={{ marginBottom: '16px' }}>
              <Paragraph>{data.answer}</Paragraph>
            </Card>
          )}
          
          <List
            dataSource={data?.results}
            renderItem={(item) => (
              <List.Item>
                <Card className="result-card" title={
                  <Space>
                    <FileTextOutlined />
                    <span>{item.title}</span>
                    <Tag color="blue">{item.similarity}%</Tag>
                  </Space>
                }>
                  <div style={{ marginBottom: '8px' }}>
                    <Tag color="default">{item.docType}</Tag>
                    <Tag color="green">{item.department}</Tag>
                    <Tag icon={<ClockCircleOutlined />} color="default">
                      {item.uploadDate}
                    </Tag>
                  </div>
                  <Paragraph ellipsis={{ rows: 3 }}>
                    {item.highlightedText}
                  </Paragraph>
                  <Button type="link" href={item.url} target="_blank">
                    查看原文
                  </Button>
                </Card>
              </List.Item>
            )}
          />
        </div>
      )}
    </div>
  )
}

export default DocumentSearch