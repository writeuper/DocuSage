import React from 'react'
import { Card, Form, Input, Button, Avatar, Space, Divider, Row, Col } from 'antd'
import { UserOutlined, MailOutlined, LockOutlined } from '@ant-design/icons'

const UserCenter: React.FC = () => {
  const [form] = Form.useForm()

  const onFinish = (values: any) => {
    console.log('Received values of form: ', values)
  }

  return (
    <div style={{ padding: '24px' }}>
      <h2>个人中心</h2>
      <Row gutter={[24, 24]}>
        <Col xs={24} md={8}>
          <Card title="个人信息">
            <div style={{ textAlign: 'center', marginBottom: '24px' }}>
              <Avatar size={80} icon={<UserOutlined />} />
              <h3 style={{ marginTop: '16px' }}>用户名</h3>
              <p style={{ color: '#666' }}>user@example.com</p>
            </div>
            <Divider />
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <strong>注册时间：</strong>2024-01-01
              </div>
              <div>
                <strong>最后登录：</strong>2024-12-16
              </div>
              <div>
                <strong>用户等级：</strong>普通用户
              </div>
            </Space>
          </Card>
        </Col>
        <Col xs={24} md={16}>
          <Card title="修改密码">
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
                <Input.Password prefix={<LockOutlined />} />
              </Form.Item>
              <Form.Item
                name="newPassword"
                label="新密码"
                rules={[{ required: true, message: '请输入新密码!' }]}
              >
                <Input.Password prefix={<LockOutlined />} />
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
                <Input.Password prefix={<LockOutlined />} />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit">
                  修改密码
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