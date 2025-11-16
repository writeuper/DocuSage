import React from 'react'
import { Card, Row, Col, Button, Space, Table, Modal, message } from 'antd'
import { 
  ReloadOutlined, 
  DownloadOutlined, 
  UploadOutlined,
  DatabaseOutlined,
  FileTextOutlined
} from '@ant-design/icons'

const OperationTools: React.FC = () => {
  const handleBackup = () => {
    Modal.confirm({
      title: '确认备份',
      content: '确定要执行系统备份吗？这可能需要几分钟时间。',
      onOk() {
        message.success('备份任务已开始执行')
      },
    })
  }

  const handleRestore = () => {
    Modal.confirm({
      title: '确认恢复',
      content: '确定要执行系统恢复吗？这将覆盖当前数据。',
      onOk() {
        message.success('恢复任务已开始执行')
      },
    })
  }

  const handleReindex = () => {
    Modal.confirm({
      title: '确认重建索引',
      content: '确定要重建全文索引吗？这可能需要较长时间。',
      onOk() {
        message.success('索引重建任务已开始')
      },
    })
  }

  const handleClearCache = () => {
    message.success('缓存清理完成')
  }

  const maintenanceTasks = [
    {
      key: '1',
      name: '每日备份',
      status: 'success',
      lastRun: '2024-12-16 02:00:00',
      nextRun: '2024-12-17 02:00:00',
    },
    {
      key: '2',
      name: '索引优化',
      status: 'running',
      lastRun: '2024-12-16 01:00:00',
      nextRun: '2024-12-17 01:00:00',
    },
    {
      key: '3',
      name: '日志清理',
      status: 'pending',
      lastRun: '2024-12-15 03:00:00',
      nextRun: '2024-12-17 03:00:00',
    },
  ]

  const taskColumns = [
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusMap = {
          success: { text: '成功', color: 'green' },
          running: { text: '运行中', color: 'blue' },
          pending: { text: '等待中', color: 'orange' },
          failed: { text: '失败', color: 'red' },
        }
        const config = statusMap[status as keyof typeof statusMap]
        return <span style={{ color: config.color }}>{config.text}</span>
      },
    },
    {
      title: '上次运行',
      dataIndex: 'lastRun',
      key: 'lastRun',
    },
    {
      title: '下次运行',
      dataIndex: 'nextRun',
      key: 'nextRun',
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      <h2>运维工具</h2>
      
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} md={12}>
          <Card title="数据管理">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Button 
                type="primary" 
                icon={<DownloadOutlined />}
                onClick={handleBackup}
                block
              >
                系统备份
              </Button>
              <Button 
                icon={<UploadOutlined />}
                onClick={handleRestore}
                block
              >
                系统恢复
              </Button>
              <Button 
                icon={<DatabaseOutlined />}
                onClick={handleReindex}
                block
              >
                重建索引
              </Button>
              <Button 
                icon={<FileTextOutlined />}
                onClick={handleClearCache}
                block
              >
                清理缓存
              </Button>
            </Space>
          </Card>
        </Col>
        
        <Col xs={24} md={12}>
          <Card title="系统维护">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Button icon={<ReloadOutlined />}>
                重启服务
              </Button>
              <Button icon={<ReloadOutlined />}>
                更新配置
              </Button>
              <Button icon={<ReloadOutlined />}>
                清理临时文件
              </Button>
              <Button icon={<ReloadOutlined />}>
                健康检查
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>

      <Card title="定时任务">
        <Table
          columns={taskColumns}
          dataSource={maintenanceTasks}
          pagination={false}
          size="small"
        />
      </Card>
    </div>
  )
}

export default OperationTools