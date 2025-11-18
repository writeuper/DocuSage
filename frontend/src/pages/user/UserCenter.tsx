import React from 'react'
import { Card, Form, Input, Button, Avatar, Space, Divider, Row, Col, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useSelector } from 'react-redux'
import { RootState } from '@/store'

const UserCenter: React.FC = () => {
  const [form] = Form.useForm()
  const user = useSelector((state: RootState) => state.user.userInfo)

  const onFinish = (values: any) => {
    console.log('Received values of form: ', values)
    message.success('密码修改功能开发中')
  }

  return (
    <div style={{ padding: '24px' }}>
      <h2>个人中心</h2>
      <Row gutter={[24, 24]}>
        <Col xs={24} md={8}>
          <Card title="个人信息">
            <div style={{ textAlign: 'center', marginBottom: '24px' }}>
              <Avatar 
                size={80} 
                icon={<UserOutlined />}
                style={{ backgroundColor: user?.role === 'admin' ? '#f50' : '#1890ff' }}
              />
              <h3 style={{ marginTop: '16px' }}>{user?.username || '未知用户'}</h3>
              <p style={{ color: '#666' }}>{user?.email || 'user@example.com'}</p>
            </div>
            <Divider />
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <strong>用户ID：</strong>{user?.id || 'N/A'}
              </div>
              <div>
                <strong>用户角色：</strong>
                <span style={{ 
                  color: user?.role === 'admin' ? '#f50' : '#1890ff',
                  fontWeight: 'bold'
                }}>
                  {user?.role === 'admin' ? '管理员' : '普通用户'}
                </span>
              </div>
              <div>
                <strong>账户状态：</strong>
                <span style={{ color: user?.status === 'active' ? '#52c41a' : '#f50' }}>
                  {user?.status === 'active' ? '正常' : '已锁定'}
                </span>
              </div>
              <div>
                <strong>邮箱：</strong>{user?.email || '未设置'}
              </div>
              <div>
                <strong>全名：</strong>{user?.full_name || '未设置'}
              </div>
            </Space>
          </Card>
        </Col>
        <Col xs={24} md={16}>
          <Card title="账户安全" extra={<span style={{ color: '#666', fontSize: '12px' }}>定期修改密码，保护账户安全</span>}>
            <Form
              form={form}
              name="changePassword"
              onFinish={onFinish}
              layout="vertical"
            >
              <Form.Item
                name="currentPassword"
                label="当前密码"
                rules={[{ required: true, message: '请输入当前密码!' }]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="请输入当前密码" />
              </Form.Item>
              <Form.Item
                name="newPassword"
                label="新密码"
                rules={[
                  { required: true, message: '请输入新密码!' },
                  { min: 6, message: '密码长度至少6位!' }
                ]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="请输入新密码（至少6位）" />
              </Form.Item>
              <Form.Item
                name="confirmPassword"
                label="确认新密码"
                dependencies={['newPassword']}
                rules={[
                  { required: true, message: '请确认新密码!' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('newPassword') === value) {
                        return Promise.resolve()
                      }
                      return Promise.reject(new Error('两次输入的密码不一致!'))
                    },
                  }),
                ]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="请再次输入新密码" />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" style={{ marginRight: '8px' }}>
                  修改密码
                </Button>
                <Button onClick={() => form.resetFields()}>
                  重置
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default UserCenter