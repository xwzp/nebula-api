import React, { useEffect, useState, useRef } from 'react';
import {
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
  verifyJSON,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsPaymentGatewayEpay(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    PayAddress: '',
    EpayId: '',
    EpayKey: '',
    EpayUnitPrice: 0,
    EpayMinTopUp: 0,
    PayMethods: '',
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        PayAddress: props.options.PayAddress || '',
        EpayId: props.options.EpayId || '',
        EpayKey: props.options.EpayKey || '',
        EpayUnitPrice:
          props.options.EpayUnitPrice !== undefined
            ? parseFloat(props.options.EpayUnitPrice)
            : 0,
        EpayMinTopUp:
          props.options.EpayMinTopUp !== undefined
            ? parseFloat(props.options.EpayMinTopUp)
            : 0,
        PayMethods: props.options.PayMethods || '',
      };

      setInputs(currentInputs);
      setOriginInputs({ ...currentInputs });
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitEpaySetting = async () => {
    if (originInputs['PayMethods'] !== inputs.PayMethods) {
      if (!verifyJSON(inputs.PayMethods)) {
        showError(t('充值方式设置不是合法的 JSON 字符串'));
        return;
      }
    }

    setLoading(true);
    try {
      const options = [];

      options.push({
        key: 'PayAddress',
        value: removeTrailingSlash(inputs.PayAddress),
      });
      if (inputs.EpayId !== '') {
        options.push({ key: 'EpayId', value: inputs.EpayId });
      }
      if (inputs.EpayKey !== undefined && inputs.EpayKey !== '') {
        options.push({ key: 'EpayKey', value: inputs.EpayKey });
      }
      if (inputs.EpayUnitPrice !== undefined && inputs.EpayUnitPrice !== null) {
        options.push({
          key: 'EpayUnitPrice',
          value: inputs.EpayUnitPrice.toString(),
        });
      }
      if (inputs.EpayMinTopUp !== undefined && inputs.EpayMinTopUp !== null) {
        options.push({
          key: 'EpayMinTopUp',
          value: inputs.EpayMinTopUp.toString(),
        });
      }
      if (originInputs['PayMethods'] !== inputs.PayMethods) {
        options.push({ key: 'PayMethods', value: inputs.PayMethods });
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
        <Form.Section text={t('易支付设置')}>
          <Text>
            {t('易支付（Epay）网关设置，默认使用上方服务器地址作为回调地址。')}
          </Text>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='PayAddress'
                label={t('支付地址')}
                placeholder={t('例如：https://yourdomain.com')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='EpayId'
                label={t('易支付商户 ID')}
                placeholder={t('例如：0001')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='EpayKey'
                label={t('易支付商户密钥')}
                placeholder={t('敏感信息不会发送到前端显示')}
                type='password'
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='EpayUnitPrice'
                precision={2}
                label={t('充值单价（0 则使用全局价格）')}
                placeholder={t('例如：7.3')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='EpayMinTopUp'
                label={t('最低充值数量（0 则使用全局）')}
                placeholder={t('例如：1')}
              />
            </Col>
          </Row>
          <Form.TextArea
            field='PayMethods'
            label={t('充值方式设置')}
            placeholder={t('为一个 JSON 文本')}
            autosize
          />
          <Button onClick={submitEpaySetting}>{t('更新易支付设置')}</Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
