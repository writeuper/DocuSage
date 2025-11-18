/*
 * @Author: lixinda
 * @Description: 
 * @File: 
 * @Date: 2025-11-13 20:02:41
 */
import React, { useState } from 'react'
import { Form, Input, Button, Card, Typography, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useDispatch } from 'react-redux'
import { login } from '@/store/slices/userSlice'
import { login as loginApi, getUserInfo } from '@/services/authService'

const { Title, Paragraph } = Typography
const { Item } = Form

const Login: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const dispatch = useDispatch()

  const handleSubmit = async (values: { username: string; password: string }) => {
    setLoading(true)
    try {
      // 先进行登录获取token
      const loginResponse = await loginApi(values)
      
      // 保存token到localStorage
      localStorage.setItem('token', loginResponse.data.access_token)
      
      // 使用token调用getUserInfo获取完整用户信息
      const userInfoResponse = await getUserInfo()
      
      // 构建完整的用户信息对象，包含token
      const userData = {
        ...userInfoResponse.data,
        token: loginResponse.data.access_token
      }
      
      // 更新Redux状态
      dispatch(login(userData))
      message.success('登录成功')
      
      // 根据用户角色跳转
      if (userInfoResponse.data.role === 'admin') {
        navigate('/admin')
      } else {
        navigate('/')
      }
    } catch (error) {
      message.error('登录失败，请检查用户名和密码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f0f2f5' }}>
      <Card style={{ width: 400, padding: '24px' }}>
        <Title level={2} style={{ textAlign: 'center', marginBottom: '24px' }}>DocuSage 登录</Title>
        <Paragraph style={{ textAlign: 'center', marginBottom: '24px', color: '#666' }}>
          企业智能文档知识库助手
        </Paragraph>
        
        <Form
          name="login"
          initialValues={{ remember: true }}
          onFinish={handleSubmit}
        >
          <Item
            name="username"
            rules={[{ required: true, message: '请输入用户名!' }]}
            style={{ marginBottom: '16px' }}
          >
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Item>
          
          <Item
            name="password"
            rules={[{ required: true, message: '请输入密码!' }]}
            style={{ marginBottom: '24px' }}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Item>
          
          <Item>
            <Button
              type="primary"
              htmlType="submit"
              className="login-form-button"
              style={{ width: '100%' }}
              loading={loading}
            >
              登录
            </Button>
          </Item>
        </Form>
      </Card>
    </div>
  )
}

export default Login