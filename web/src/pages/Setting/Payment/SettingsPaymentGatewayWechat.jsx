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

export default function SettingsPaymentGatewayWechat(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    WechatPayEnabled: false,
    WechatPayMchId: '',
    WechatPayMchApiV3Key: '',
    WechatPayMchSerialNo: '',
    WechatPayMchPrivateKey: '',
    WechatPayAppId: '',
    WechatPayNotifyUrl: '',
    WechatPayUnitPrice: 0,
    WechatPayMinTopUp: 1,
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        WechatPayEnabled:
          props.options.WechatPayEnabled !== undefined
            ? props.options.WechatPayEnabled
            : false,
        WechatPayMchId: props.options.WechatPayMchId || '',
        WechatPayMchApiV3Key: props.options.WechatPayMchApiV3Key || '',
        WechatPayMchSerialNo: props.options.WechatPayMchSerialNo || '',
        WechatPayMchPrivateKey: props.options.WechatPayMchPrivateKey || '',
        WechatPayAppId: props.options.WechatPayAppId || '',
        WechatPayNotifyUrl: props.options.WechatPayNotifyUrl || '',
        WechatPayUnitPrice:
          props.options.WechatPayUnitPrice !== undefined
            ? parseFloat(props.options.WechatPayUnitPrice)
            : 0,
        WechatPayMinTopUp:
          props.options.WechatPayMinTopUp !== undefined
            ? parseFloat(props.options.WechatPayMinTopUp)
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

  const submitWechatSetting = async () => {
    setLoading(true);
    try {
      const options = [];

      if (
        originInputs['WechatPayEnabled'] !== inputs.WechatPayEnabled &&
        inputs.WechatPayEnabled !== undefined
      ) {
        options.push({
          key: 'WechatPayEnabled',
          value: inputs.WechatPayEnabled ? 'true' : 'false',
        });
      }
      if (inputs.WechatPayMchId !== '') {
        options.push({ key: 'WechatPayMchId', value: inputs.WechatPayMchId });
      }
      if (inputs.WechatPayMchApiV3Key !== '') {
        options.push({
          key: 'WechatPayMchApiV3Key',
          value: inputs.WechatPayMchApiV3Key,
        });
      }
      if (inputs.WechatPayMchSerialNo !== '') {
        options.push({
          key: 'WechatPayMchSerialNo',
          value: inputs.WechatPayMchSerialNo,
        });
      }
      if (inputs.WechatPayMchPrivateKey !== '') {
        options.push({
          key: 'WechatPayMchPrivateKey',
          value: inputs.WechatPayMchPrivateKey,
        });
      }
      if (inputs.WechatPayAppId !== '') {
        options.push({ key: 'WechatPayAppId', value: inputs.WechatPayAppId });
      }
      options.push({
        key: 'WechatPayNotifyUrl',
        value: inputs.WechatPayNotifyUrl || '',
      });
      if (
        inputs.WechatPayUnitPrice !== undefined &&
        inputs.WechatPayUnitPrice !== null
      ) {
        options.push({
          key: 'WechatPayUnitPrice',
          value: inputs.WechatPayUnitPrice.toString(),
        });
      }
      if (
        inputs.WechatPayMinTopUp !== undefined &&
        inputs.WechatPayMinTopUp !== null
      ) {
        options.push({
          key: 'WechatPayMinTopUp',
          value: inputs.WechatPayMinTopUp.toString(),
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
        <Form.Section text={t('微信支付设置')}>
          <Text>
            {t('微信支付 Native 模式（PC 扫码支付），需要在')}
            <a
              href='https://pay.weixin.qq.com'
              target='_blank'
              rel='noreferrer'
            >
              {t('微信支付商户平台')}
            </a>
            {t('获取商户凭证。')}
          </Text>
          <Banner
            type='info'
            description={`${t('回调地址')}：${props.options.ServerAddress ? removeTrailingSlash(props.options.ServerAddress) : t('网站地址')}/api/pay/notify/wechat`}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Switch
                field='WechatPayEnabled'
                size='default'
                checkedText='|'
                uncheckedText='O'
                label={t('启用微信支付')}
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WechatPayMchId'
                label={t('商户号 (mch_id)')}
                placeholder={t('微信支付商户号')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WechatPayAppId'
                label={t('AppID')}
                placeholder={t('关联的公众号或小程序 AppID')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WechatPayMchSerialNo'
                label={t('商户证书序列号')}
                placeholder={t('API 证书的序列号')}
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='WechatPayMchApiV3Key'
                label={t('APIv3 密钥')}
                placeholder={t('APIv3 密钥，用于验签和解密通知')}
                type='password'
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='WechatPayNotifyUrl'
                label={t('自定义回调地址（可选）')}
                placeholder={t('留空则使用服务器地址 + /api/pay/notify/wechat')}
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={24} lg={24} xl={24}>
              <Form.TextArea
                field='WechatPayMchPrivateKey'
                label={t('商户私钥 (PEM)')}
                placeholder={t(
                  '-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----',
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
                field='WechatPayUnitPrice'
                precision={2}
                label={t('充值单价（CNY/单位，0 则使用全局价格）')}
                placeholder={t('例如：7.3')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='WechatPayMinTopUp'
                label={t('最低充值数量')}
                placeholder={t('例如：1')}
              />
            </Col>
          </Row>
          <Button onClick={submitWechatSetting}>
            {t('更新微信支付设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
