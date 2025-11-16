import React from 'react'
import { Card, Form, Input, Button, Switch, Select, Space, Divider } from 'antd'
import { SaveOutlined, ReloadOutlined } from '@ant-design/icons'

const { Option } = Select

const SystemConfig: React.FC = () => {
  const [form] = Form.useForm()

  const onFinish = (values: any) => {
    console.log('Received values of form: ', values)
  }

  const onReset = () => {
    form.resetFields()
  }

  return (
    <div style={{ padding: '24px' }}>
      <h2>系统配置</h2>
      
      <Card title="基础设置" style={{ marginBottom: '24px' }}>
        <Form
          form={form}
          name="basicConfig"
          onFinish={onFinish}
          layout="vertical"
        >
          <Form.Item
            name="systemName"
            label="系统名称"
            initialValue="企业智能文档知识库助手"
          >
            <Input />
          </Form.Item>
          
          <Form.Item
            name="systemVersion"
            label="系统版本"
            initialValue="v1.0.0"
          >
            <Input />
          </Form.Item>

          <Form.Item
            name="maxUploadSize"
            label="最大上传文件大小(MB)"
            initialValue={100}
          >
            <Input type="number" />
          </Form.Item>

          <Form.Item
            name="supportedFormats"
            label="支持的文件格式"
            initialValue={['pdf', 'doc', 'docx', 'txt']}
          >
            <Select mode="multiple" placeholder="请选择支持的文件格式">
              <Option value="pdf">PDF</Option>
              <Option value="doc">DOC</Option>
              <Option value="docx">DOCX</Option>
              <Option value="txt">TXT</Option>
              <Option value="md">MD</Option>
            </Select>
          </Form.Item>
        </Form>
      </Card>

      <Card title="AI配置" style={{ marginBottom: '24px' }}>
        <Form layout="vertical">
          <Form.Item
            name="aiEnabled"
            label="启用AI功能"
            valuePropName="checked"
            initialValue={true}
          >
            <Switch />
          </Form.Item>

          <Form.Item
            name="aiModel"
            label="AI模型"
            initialValue="gpt-3.5-turbo"
          >
            <Select>
              <Option value="gpt-3.5-turbo">GPT-3.5 Turbo</Option>
              <Option value="gpt-4">GPT-4</Option>
              <Option value="claude-3">Claude-3</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="maxTokens"
            label="最大Token数"
            initialValue={2048}
          >
            <Input type="number" />
          </Form.Item>
        </Form>
      </Card>

      <Card title="安全设置">
        <Form layout="vertical">
          <Form.Item
            name="sessionTimeout"
            label="会话超时时间(分钟)"
            initialValue={30}
          >
            <Input type="number" />
          </Form.Item>

          <Form.Item
            name="maxLoginAttempts"
            label="最大登录尝试次数"
            initialValue={5}
          >
            <Input type="number" />
          </Form.Item>

          <Form.Item
            name="enableTwoFactor"
            label="启用双因素认证"
            valuePropName="checked"
            initialValue={false}
          >
            <Switch />
          </Form.Item>
        </Form>
      </Card>

      <Divider />

      <Space>
        <Button type="primary" htmlType="submit" icon={<SaveOutlined />}>
          保存配置
        </Button>
        <Button icon={<ReloadOutlined />} onClick={onReset}>
          重置
        </Button>
      </Space>
    </div>
  )
}

export default SystemConfig