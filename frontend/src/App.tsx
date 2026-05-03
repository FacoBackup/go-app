import React, {useEffect, useState} from 'react';
import {Layout, Card, Form, Input, Select, Button, Table, List, Typography, Space, message, Modal, Row, Col, Tag} from 'antd';
import {DeleteOutlined, PlusOutlined, ReloadOutlined, HistoryOutlined} from '@ant-design/icons';
import './App.css';

const {Header, Content, Sider} = Layout;
const {Title, Text} = Typography;
const {Option} = Select;

const API_BASE_URL = 'http://localhost:8080/v1';

function App() {
    const [deviceId, setDeviceId] = useState('DEV001');
    const [gas, setGas] = useState('H2');
    const [health, setHealth] = useState<any>(null);
    const [series, setSeries] = useState<any[]>([]);
    const [devices, setDevices] = useState<string[]>([]);
    const [loading, setLoading] = useState(false);
    const [form] = Form.useForm();

    useEffect(() => {
        fetchDevices();
    }, []);

    const fetchDevices = async () => {
        try {
            const response = await fetch(`${API_BASE_URL}/devices`);
            if (response.ok) {
                const data = await response.json();
                setDevices(data);
            }
        } catch (error) {
            console.error('Error fetching devices:', error);
            message.error('Erro ao buscar dispositivos');
        }
    };

    const postMeasurement = async () => {
        try {
            const values = await form.validateFields();
            setLoading(true);
            const response = await fetch(`${API_BASE_URL}/measurements`, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify([{
                    device_id: values.deviceId,
                    timestamp: new Date().toISOString(),
                    gas: values.gas,
                    value_ppm: parseFloat(values.value),
                    patient_ref: values.patientRef
                }])
            });
            if (response.ok) {
                message.success('Medição enviada com sucesso!');
                fetchDevices();
            } else {
                message.error('Erro ao enviar medição');
            }
        } catch (error) {
            console.error('Validation failed:', error);
        } finally {
            setLoading(false);
        }
    };

    const fetchHealth = async () => {
        setLoading(true);
        try {
            const response = await fetch(`${API_BASE_URL}/devices/${deviceId}/health`);
            if (response.ok) {
                const data = await response.json();
                setHealth(data);
            } else {
                message.error('Erro ao buscar status de saúde');
            }
        } catch (error) {
            message.error('Erro de conexão: ' + error);
        } finally {
            setLoading(false);
        }
    };

    const fetchSeries = async () => {
        setLoading(true);
        try {
            const now = new Date();
            const hourAgo = new Date(now.getTime() - 60 * 60 * 1000);
            const start = hourAgo.toISOString();
            const end = now.toISOString();

            const response = await fetch(`${API_BASE_URL}/devices/${deviceId}/series?gas=${gas}&start=${start}&end=${end}`);
            if (response.ok) {
                const data = await response.json();
                setSeries(data);
            } else {
                message.error('Erro ao buscar série temporal');
            }
        } catch (error) {
            message.error('Erro de conexão: ' + error);
        } finally {
            setLoading(false);
        }
    };

    const deleteDevice = (id: string) => {
        Modal.confirm({
            title: `Excluir todos os dados do dispositivo ${id}?`,
            content: 'Esta ação não pode ser desfeita.',
            okText: 'Sim',
            okType: 'danger',
            cancelText: 'Não',
            onOk: async () => {
                try {
                    const response = await fetch(`${API_BASE_URL}/devices/${id}`, {
                        method: 'DELETE'
                    });
                    if (response.ok) {
                        message.success(`Dados do dispositivo ${id} excluídos.`);
                        fetchDevices();
                        if (deviceId === id) {
                            setHealth(null);
                            setSeries([]);
                        }
                    } else {
                        message.error('Erro ao excluir dados');
                    }
                } catch (error) {
                    message.error('Erro de conexão: ' + error);
                }
            }
        });
    };

    const columns = [
        {
            title: 'Horário (Minuto)',
            dataIndex: 'timestamp',
            key: 'timestamp',
            render: (text: string) => new Date(text).toLocaleTimeString(),
        },
        {
            title: 'Média PPM',
            dataIndex: 'value_ppm',
            key: 'value_ppm',
            render: (val: number) => val.toFixed(2),
        },
    ];

    return (
        <Layout style={{minHeight: '100vh'}}>
            <Header style={{display: 'flex', alignItems: 'center'}}>
                <Title level={3} style={{color: 'white', margin: 0}}>Telemetria de Gases</Title>
            </Header>
            <Layout>
                <Sider width={250} theme="light">
                    <div style={{padding: '16px'}}>
                        <Title level={4}>Dispositivos</Title>
                        <List
                            size="small"
                            bordered
                            dataSource={devices}
                            renderItem={id => (
                                <List.Item
                                    style={{
                                        cursor: 'pointer',
                                        backgroundColor: id === deviceId ? '#e6f7ff' : 'transparent'
                                    }}
                                    onClick={() => {
                                        setDeviceId(id);
                                        form.setFieldsValue({deviceId: id});
                                    }}
                                    actions={[
                                        <Button
                                            type="text"
                                            danger
                                            icon={<DeleteOutlined/>}
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                deleteDevice(id);
                                            }}
                                        />
                                    ]}
                                >
                                    {id}
                                </List.Item>
                            )}
                        />
                    </div>
                </Sider>
                <Content style={{padding: '24px', backgroundColor: '#f0f2f5'}}>
                    <Row gutter={[24, 24]}>
                        <Col xs={24} lg={12}>
                            <Card title={<span><PlusOutlined style={{marginRight: 8}}/>Enviar Medição</span>}>
                                <Form
                                    form={form}
                                    layout="vertical"
                                    initialValues={{deviceId, patientRef: 'PAT001', gas, value: '10.5'}}
                                    onFinish={postMeasurement}
                                >
                                    <Form.Item name="deviceId" label="Device ID" rules={[{required: true}]}>
                                        <Input />
                                    </Form.Item>
                                    <Form.Item name="patientRef" label="Patient Ref" rules={[{required: true}]}>
                                        <Input />
                                    </Form.Item>
                                    <Form.Item name="gas" label="Gás" rules={[{required: true}]}>
                                        <Select onChange={setGas}>
                                            <Option value="H2">H2</Option>
                                            <Option value="CH4">CH4</Option>
                                            <Option value="H2S">H2S</Option>
                                        </Select>
                                    </Form.Item>
                                    <Form.Item name="value" label="Valor (PPM)" rules={[{required: true}]}>
                                        <Input type="number" />
                                    </Form.Item>
                                    <Button type="primary" htmlType="submit" loading={loading} block>
                                        Enviar
                                    </Button>
                                </Form>
                            </Card>
                        </Col>

                        <Col xs={24} lg={12}>
                            <Card
                                title={`Status: ${deviceId}`}
                                extra={<Button icon={<ReloadOutlined/>} onClick={fetchHealth} loading={loading}>Atualizar</Button>}
                            >
                                {health ? (
                                    <Space direction="vertical" style={{width: '100%'}}>
                                        <Text>
                                            <Text strong>Última Medição: </Text>
                                            {health.last_measurement ?
                                                `${health.last_measurement.value_ppm} ${health.last_measurement.gas} (${new Date(health.last_measurement.timestamp).toLocaleTimeString()})` :
                                                'Nenhuma'}
                                        </Text>
                                        <Text>
                                            <Text strong>Taxa (Última Hora): </Text>
                                            {health.rate_last_hour.toFixed(2)} m/min
                                        </Text>
                                        <Text>
                                            <Text strong>Status: </Text>
                                            <Tag color={health.stale ? 'red' : 'green'}>
                                                {health.stale ? 'STALE (> 5 min)' : 'ACTIVE'}
                                            </Tag>
                                        </Text>
                                    </Space>
                                ) : (
                                    <Text type="secondary">Clique em Atualizar para ver o status.</Text>
                                )}
                            </Card>
                        </Col>

                        <Col span={24}>
                            <Card
                                title={`Série Temporal (${gas}) - Última Hora`}
                                extra={<Button icon={<HistoryOutlined/>} onClick={fetchSeries} loading={loading}>Buscar Série</Button>}
                            >
                                <Table
                                    columns={columns}
                                    dataSource={series}
                                    rowKey={(record, index) => index?.toString() || ''}
                                    pagination={{pageSize: 5}}
                                    size="small"
                                />
                            </Card>
                        </Col>
                    </Row>
                </Content>
            </Layout>
        </Layout>
    );
}

export default App;
