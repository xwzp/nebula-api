import React, { useEffect, useState, useRef } from 'react';
import {
  Banner,
  Button,
  Form,
  Row,
  Col,
  Typography,
  Spin,
} from '@douyinfe/semi-ui';
const { Text } = Typography;
import {
  API,
  removeTrailingSlash,
  showError,
  showSuccess,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsPaymentGatewayAlipay(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    AlipayEnabled: false,
    AlipayAppId: '',
    AlipayPrivateKey: '',
    AlipayPublicKey: '',
    AlipayNotifyUrl: '',
    AlipayUnitPrice: 0,
    AlipayMinTopUp: 1,
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        AlipayEnabled:
          props.options.AlipayEnabled !== undefined
            ? props.options.AlipayEnabled
            : false,
        AlipayAppId: props.options.AlipayAppId || '',
        AlipayPrivateKey: props.options.AlipayPrivateKey || '',
        AlipayPublicKey: props.options.AlipayPublicKey || '',
        AlipayNotifyUrl: props.options.AlipayNotifyUrl || '',
        AlipayUnitPrice:
          props.options.AlipayUnitPrice !== undefined
            ? parseFloat(props.options.AlipayUnitPrice)
            : 0,
        AlipayMinTopUp:
          props.options.AlipayMinTopUp !== undefined
            ? parseFloat(props.options.AlipayMinTopUp)
            : 1,
      };
      setInputs(currentInputs);
      setOriginInputs({ ...currentInputs });
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitAlipaySetting = async () => {
    setLoading(true);
    try {
      const options = [];

      if (
        originInputs['AlipayEnabled'] !== inputs.AlipayEnabled &&
        inputs.AlipayEnabled !== undefined
      ) {
        options.push({
          key: 'AlipayEnabled',
          value: inputs.AlipayEnabled ? 'true' : 'false',
        });
      }
      if (inputs.AlipayAppId !== '') {
        options.push({ key: 'AlipayAppId', value: inputs.AlipayAppId });
      }
      if (inputs.AlipayPrivateKey !== '') {
        options.push({
          key: 'AlipayPrivateKey',
          value: inputs.AlipayPrivateKey,
        });
      }
      if (inputs.AlipayPublicKey !== '') {
        options.push({
          key: 'AlipayPublicKey',
          value: inputs.AlipayPublicKey,
        });
      }
      options.push({
        key: 'AlipayNotifyUrl',
        value: inputs.AlipayNotifyUrl || '',
      });
      if (
        inputs.AlipayUnitPrice !== undefined &&
        inputs.AlipayUnitPrice !== null
      ) {
        options.push({
          key: 'AlipayUnitPrice',
          value: inputs.AlipayUnitPrice.toString(),
        });
      }
      if (
        inputs.AlipayMinTopUp !== undefined &&
        inputs.AlipayMinTopUp !== null
      ) {
        options.push({
          key: 'AlipayMinTopUp',
          value: inputs.AlipayMinTopUp.toString(),
        });
      }

      const requestQueue = options.map((opt) =>
        API.put('/api/option/', {
          key: opt.key,
          value: opt.value,
        }),
      );

      const results = await Promise.all(requestQueue);

      const errorResults = results.filter((res) => !res.data.success);
      if (errorResults.length > 0) {
        errorResults.forEach((res) => {
          showError(res.data.message);
        });
      } else {
        showSuccess(t('更新成功'));
        setOriginInputs({ ...inputs });
        props.refresh?.();
      }
    } catch (error) {
      showError(t('更新失败'));
    }
    setLoading(false);
  };

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={handleFormChange}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={t('支付宝支付设置')}>
          <Text>
            {t('支付宝当面付模式（PC 扫码支付），需要在')}
            <a
              href='https://open.alipay.com'
              target='_blank'
              rel='noreferrer'
            >
              {t('支付宝开放平台')}
            </a>
            {t('创建应用并获取密钥。')}
          </Text>
          <Banner
            type='info'
            description={`${t('回调地址')}：${props.options.ServerAddress ? removeTrailingSlash(props.options.ServerAddress) : t('网站地址')}/api/pay/notify/alipay`}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch
                field='AlipayEnabled'
                size='default'
                checkedText='|'
                uncheckedText='O'
                label={t('启用支付宝支付')}
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='AlipayAppId'
                label={t('应用 AppID')}
                placeholder={t('支付宝开放平台应用 AppID')}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='AlipayNotifyUrl'
                label={t('自定义回调地址（可选）')}
                placeholder={t('留空则使用服务器地址 + /api/pay/notify/alipay')}
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={24} lg={24} xl={24}>
              <Form.TextArea
                field='AlipayPrivateKey'
                label={t('应用私钥（RSA2）')}
                placeholder={t(
                  '应用私钥内容，可在支付宝开放平台获取',
                )}
                autosize={{ minRows: 3, maxRows: 6 }}
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={24} lg={24} xl={24}>
              <Form.TextArea
                field='AlipayPublicKey'
                label={t('支付宝公钥')}
                placeholder={t(
                  '支付宝公钥内容，用于验签通知',
                )}
                autosize={{ minRows: 3, maxRows: 6 }}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='AlipayUnitPrice'
                precision={2}
                label={t('充值单价（CNY/单位，0 则使用全局价格）')}
                placeholder={t('例如：7.3')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='AlipayMinTopUp'
                label={t('最低充值数量')}
                placeholder={t('例如：1')}
              />
            </Col>
          </Row>
          <Button onClick={submitAlipaySetting}>
            {t('更新支付宝支付设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
