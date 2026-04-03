/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useState, useRef } from 'react';
import {
  Avatar,
  Button,
  Card,
  Col,
  Collapsible,
  Form,
  Row,
  Select,
  SideSheet,
  Space,
  Spin,
  Switch,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import {
  IconClose,
  IconSave,
  IconSetting,
} from '@douyinfe/semi-icons';
import { CalendarClock, List } from 'lucide-react';
import { API, showError, showSuccess } from '../../../../helpers';
import {
  quotaToDisplayAmount,
  displayAmountToQuota,
} from '../../../../helpers/quota';
import { useIsMobile } from '../../../../hooks/common/useIsMobile';
import { parseFeatures } from '../../../../helpers/subscriptionFormat';
import FeatureListEditor from '../../../common/FeatureListEditor';

const { Text, Title } = Typography;

const resetPeriodOptions = [
  { value: 'never', label: '不重置' },
  { value: 'daily', label: '每天' },
  { value: 'weekly', label: '每周' },
  { value: 'monthly', label: '每月' },
  { value: 'custom', label: '自定义(秒)' },
];

const AddEditSubscriptionModal = ({
  visible,
  handleClose,
  editingPlan,
  placement = 'right',
  refresh,
  t,
}) => {
  const [loading, setLoading] = useState(false);
  const [groupOptions, setGroupOptions] = useState([]);
  const [groupLoading, setGroupLoading] = useState(false);
  const isMobile = useIsMobile();
  const formApiRef = useRef(null);
  const isEdit = editingPlan?.id !== undefined;
  const formKey = isEdit ? `edit-${editingPlan?.id}` : 'create-new';

  // Feature lists managed outside the form
  const [features, setFeatures] = useState([]);
  const [monthlyFeatures, setMonthlyFeatures] = useState([]);
  const [quarterlyFeatures, setQuarterlyFeatures] = useState([]);
  const [yearlyFeatures, setYearlyFeatures] = useState([]);

  // Period enabled state (managed outside form for controlling Collapsible)
  const [monthlyEnabled, setMonthlyEnabled] = useState(true);
  const [quarterlyEnabled, setQuarterlyEnabled] = useState(false);
  const [yearlyEnabled, setYearlyEnabled] = useState(false);

  const getInitValues = () => ({
    title: '',
    subtitle: '',
    tag: '',
    sort_order: 0,
    enabled: true,
    price_monthly: 0,
    currency: 'USD',
    total_amount: 0,
    quota_reset_period: 'monthly',
    quota_reset_custom_seconds: 0,
    upgrade_group: '',
    max_purchase_per_user: 1,
    // monthly
    monthly_enabled: true,
    monthly_stripe_price_id: '',
    monthly_creem_product_id: '',
    // quarterly
    quarterly_enabled: false,
    quarterly_discount: 0,
    quarterly_stripe_price_id: '',
    quarterly_creem_product_id: '',
    // yearly
    yearly_enabled: false,
    yearly_discount: 0,
    yearly_stripe_price_id: '',
    yearly_creem_product_id: '',
  });

  const buildFormValues = () => {
    const base = getInitValues();
    if (!isEdit) return base;
    const p = editingPlan || {};
    return {
      ...base,
      title: p.title || '',
      subtitle: p.subtitle || '',
      tag: p.tag || '',
      sort_order: Number(p.sort_order || 0),
      enabled: p.enabled !== false,
      price_monthly: Number(p.price_monthly || 0),
      currency: p.currency || 'USD',
      total_amount: Number(
        quotaToDisplayAmount(p.total_amount || 0).toFixed(2),
      ),
      quota_reset_period: p.quota_reset_period || 'monthly',
      quota_reset_custom_seconds: Number(p.quota_reset_custom_seconds || 0),
      upgrade_group: p.upgrade_group || '',
      max_purchase_per_user: Number(p.max_purchase_per_user || 1),
      // monthly
      monthly_enabled: p.monthly_enabled !== false,
      monthly_stripe_price_id: p.monthly_stripe_price_id || '',
      monthly_creem_product_id: p.monthly_creem_product_id || '',
      // quarterly
      quarterly_enabled: !!p.quarterly_enabled,
      quarterly_discount: Number(p.quarterly_discount || 0),
      quarterly_stripe_price_id: p.quarterly_stripe_price_id || '',
      quarterly_creem_product_id: p.quarterly_creem_product_id || '',
      // yearly
      yearly_enabled: !!p.yearly_enabled,
      yearly_discount: Number(p.yearly_discount || 0),
      yearly_stripe_price_id: p.yearly_stripe_price_id || '',
      yearly_creem_product_id: p.yearly_creem_product_id || '',
    };
  };

  // Initialize feature lists and period states when modal opens
  useEffect(() => {
    if (!visible) return;
    if (isEdit) {
      const p = editingPlan || {};
      setFeatures(parseFeatures(p.features));
      setMonthlyFeatures(parseFeatures(p.monthly_features));
      setQuarterlyFeatures(parseFeatures(p.quarterly_features));
      setYearlyFeatures(parseFeatures(p.yearly_features));
      setMonthlyEnabled(p.monthly_enabled !== false);
      setQuarterlyEnabled(!!p.quarterly_enabled);
      setYearlyEnabled(!!p.yearly_enabled);
    } else {
      setFeatures([]);
      setMonthlyFeatures([]);
      setQuarterlyFeatures([]);
      setYearlyFeatures([]);
      setMonthlyEnabled(true);
      setQuarterlyEnabled(false);
      setYearlyEnabled(false);
    }
  }, [visible, isEdit, editingPlan]);

  // Load upgrade group options
  useEffect(() => {
    if (!visible) return;
    setGroupLoading(true);
    API.get('/api/group')
      .then((res) => {
        if (res.data?.success) {
          setGroupOptions(res.data?.data || []);
        } else {
          setGroupOptions([]);
        }
      })
      .catch(() => setGroupOptions([]))
      .finally(() => setGroupLoading(false));
  }, [visible]);

  const calcPrice = (monthlyPrice, discount) => {
    const price = Number(monthlyPrice || 0);
    const disc = Number(discount || 0);
    return Number((price * (1 - disc / 100)).toFixed(2));
  };

  const submit = async (values) => {
    setLoading(true);
    try {
      const payload = {
        title: values.title,
        subtitle: values.subtitle || '',
        tag: values.tag || '',
        sort_order: Number(values.sort_order || 0),
        enabled: values.enabled !== false,
        features: features.length > 0 ? JSON.stringify(features) : '',
        price_monthly: Number(values.price_monthly || 0),
        currency: values.currency || 'USD',
        total_amount: displayAmountToQuota(values.total_amount),
        quota_reset_period: values.quota_reset_period || 'monthly',
        quota_reset_custom_seconds:
          values.quota_reset_period === 'custom'
            ? Number(values.quota_reset_custom_seconds || 0)
            : 0,
        upgrade_group: values.upgrade_group || '',
        max_purchase_per_user: Number(values.max_purchase_per_user || 0),
        // monthly
        monthly_enabled: monthlyEnabled,
        monthly_features: monthlyFeatures.length > 0 ? JSON.stringify(monthlyFeatures) : '',
        monthly_stripe_price_id: values.monthly_stripe_price_id || '',
        monthly_creem_product_id: values.monthly_creem_product_id || '',
        // quarterly
        quarterly_enabled: quarterlyEnabled,
        quarterly_discount: Number(values.quarterly_discount || 0),
        quarterly_features: quarterlyFeatures.length > 0 ? JSON.stringify(quarterlyFeatures) : '',
        quarterly_stripe_price_id: values.quarterly_stripe_price_id || '',
        quarterly_creem_product_id: values.quarterly_creem_product_id || '',
        // yearly
        yearly_enabled: yearlyEnabled,
        yearly_discount: Number(values.yearly_discount || 0),
        yearly_features: yearlyFeatures.length > 0 ? JSON.stringify(yearlyFeatures) : '',
        yearly_stripe_price_id: values.yearly_stripe_price_id || '',
        yearly_creem_product_id: values.yearly_creem_product_id || '',
      };

      if (isEdit) {
        const res = await API.put(
          `/api/subscription/admin/plans/${editingPlan.id}`,
          payload,
        );
        if (res.data?.success) {
          showSuccess(t('更新成功'));
          handleClose();
          refresh?.();
        } else {
          showError(res.data?.message || t('更新失败'));
        }
      } else {
        const res = await API.post('/api/subscription/admin/plans', payload);
        if (res.data?.success) {
          showSuccess(t('创建成功'));
          handleClose();
          refresh?.();
        } else {
          showError(res.data?.message || t('创建失败'));
        }
      }
    } catch (e) {
      showError(t('请求失败'));
    } finally {
      setLoading(false);
    }
  };

  const PeriodCard = ({ title, enabled, onToggle, children }) => (
    <Card
      className='!rounded-2xl shadow-sm border-0 mb-4'
      header={
        <div className='flex items-center justify-between w-full'>
          <Text className='text-base font-medium'>{title}</Text>
          <Switch
            checked={enabled}
            onChange={onToggle}
            size='small'
          />
        </div>
      }
      headerStyle={{ padding: '12px 16px' }}
    >
      <Collapsible isOpen={enabled}>
        {children}
      </Collapsible>
    </Card>
  );

  return (
    <SideSheet
      placement={placement}
      title={
        <Space>
          {isEdit ? (
            <Tag color='blue' shape='circle'>
              {t('更新')}
            </Tag>
          ) : (
            <Tag color='green' shape='circle'>
              {t('新建')}
            </Tag>
          )}
          <Title heading={4} className='m-0'>
            {isEdit ? t('编辑订阅套餐') : t('新建订阅套餐')}
          </Title>
        </Space>
      }
      bodyStyle={{ padding: '0' }}
      visible={visible}
      width={isMobile ? '100%' : 640}
      footer={
        <div className='flex justify-end bg-white'>
          <Space>
            <Button
              theme='solid'
              onClick={() => formApiRef.current?.submitForm()}
              icon={<IconSave />}
              loading={loading}
            >
              {t('提交')}
            </Button>
            <Button
              theme='light'
              type='primary'
              onClick={handleClose}
              icon={<IconClose />}
            >
              {t('取消')}
            </Button>
          </Space>
        </div>
      }
      closeIcon={null}
      onCancel={handleClose}
    >
      <Spin spinning={loading}>
        <Form
          key={formKey}
          initValues={buildFormValues()}
          getFormApi={(api) => (formApiRef.current = api)}
          onSubmit={submit}
        >
          {({ values }) => (
            <div className='p-2'>
              {/* Section 1: 基础信息 */}
              <Card className='!rounded-2xl shadow-sm border-0 mb-4'>
                <div className='flex items-center mb-2'>
                  <Avatar
                    size='small'
                    color='blue'
                    className='mr-2 shadow-md'
                  >
                    <List size={16} />
                  </Avatar>
                  <div>
                    <Text className='text-lg font-medium'>
                      {t('基础信息')}
                    </Text>
                    <div className='text-xs text-gray-600'>
                      {t('套餐名称、标签和展示设置')}
                    </div>
                  </div>
                </div>

                <Row gutter={12}>
                  <Col span={12}>
                    <Form.Input
                      field='title'
                      label={t('套餐标题')}
                      placeholder={t('例如：Pro 套餐')}
                      rules={[{ required: true, message: t('标题不能为空') }]}
                    />
                  </Col>
                  <Col span={12}>
                    <Form.Input
                      field='subtitle'
                      label={t('副标题')}
                      placeholder={t('例如：适合专业用户')}
                    />
                  </Col>
                  <Col span={12}>
                    <Form.Input
                      field='tag'
                      label={t('标签')}
                      placeholder={t('例如：推荐')}
                    />
                  </Col>
                  <Col span={6}>
                    <Form.InputNumber
                      field='sort_order'
                      label={t('排序')}
                      precision={0}
                      style={{ width: '100%' }}
                      extraText={t('数字越大越靠前')}
                    />
                  </Col>
                  <Col span={6}>
                    <Form.Switch
                      field='enabled'
                      label={t('启用状态')}
                      size='large'
                    />
                  </Col>
                </Row>
              </Card>

              {/* Section 2: 优势列表 */}
              <Card className='!rounded-2xl shadow-sm border-0 mb-4'>
                <div className='flex items-center mb-2'>
                  <Avatar
                    size='small'
                    color='cyan'
                    className='mr-2 shadow-md'
                  >
                    <List size={16} />
                  </Avatar>
                  <div>
                    <Text className='text-lg font-medium'>
                      {t('优势列表')}
                    </Text>
                    <div className='text-xs text-gray-600'>
                      {t('在定价页面展示的功能亮点')}
                    </div>
                  </div>
                </div>
                <FeatureListEditor value={features} onChange={setFeatures} />
              </Card>

              {/* Section 3: 订阅配置 */}
              <Card className='!rounded-2xl shadow-sm border-0 mb-4'>
                <div className='flex items-center mb-2'>
                  <Avatar
                    size='small'
                    color='green'
                    className='mr-2 shadow-md'
                  >
                    <IconSetting size={16} />
                  </Avatar>
                  <div>
                    <Text className='text-lg font-medium'>
                      {t('订阅配置')}
                    </Text>
                    <div className='text-xs text-gray-600'>
                      {t('定价、额度和重置周期')}
                    </div>
                  </div>
                </div>

                <Row gutter={12}>
                  <Col span={12}>
                    <Form.InputNumber
                      field='price_monthly'
                      label={t('月基准价')}
                      required
                      min={0}
                      precision={2}
                      rules={[{ required: true, message: t('请输入月基准价') }]}
                      style={{ width: '100%' }}
                      extraText={t('其他周期的实付价格将基于此价格和折扣计算')}
                    />
                  </Col>
                  <Col span={12}>
                    <Form.Input
                      field='currency'
                      label={t('币种')}
                      disabled
                      extraText={t('由全站货币展示设置统一控制')}
                    />
                  </Col>
                  <Col span={12}>
                    <Form.InputNumber
                      field='total_amount'
                      label={t('总额度')}
                      required
                      min={0}
                      precision={2}
                      rules={[{ required: true, message: t('请输入总额度') }]}
                      extraText={`${t('0 表示不限')} · ${t('原生额度')}：${displayAmountToQuota(
                        values.total_amount,
                      )}`}
                      style={{ width: '100%' }}
                    />
                  </Col>
                  <Col span={12}>
                    <Form.Select
                      field='quota_reset_period'
                      label={t('额度重置周期')}
                    >
                      {resetPeriodOptions.map((o) => (
                        <Select.Option key={o.value} value={o.value}>
                          {o.label}
                        </Select.Option>
                      ))}
                    </Form.Select>
                  </Col>
                  {values.quota_reset_period === 'custom' && (
                    <Col span={12}>
                      <Form.InputNumber
                        field='quota_reset_custom_seconds'
                        label={t('自定义重置秒数')}
                        required
                        min={60}
                        precision={0}
                        rules={[{ required: true, message: t('请输入秒数') }]}
                        style={{ width: '100%' }}
                      />
                    </Col>
                  )}
                  <Col span={12}>
                    <Form.Select
                      field='upgrade_group'
                      label={t('升级分组')}
                      showClear
                      loading={groupLoading}
                      placeholder={t('不升级')}
                      extraText={t(
                        '购买订阅后用户会升级到该分组；套餐失效后将回退到原分组。',
                      )}
                    >
                      <Select.Option value=''>{t('不升级')}</Select.Option>
                      {(groupOptions || []).map((g) => (
                        <Select.Option key={g} value={g}>
                          {g}
                        </Select.Option>
                      ))}
                    </Form.Select>
                  </Col>
                  <Col span={12}>
                    <Form.InputNumber
                      field='max_purchase_per_user'
                      label={t('购买上限')}
                      min={0}
                      precision={0}
                      extraText={t('0 表示不限')}
                      style={{ width: '100%' }}
                    />
                  </Col>
                </Row>
              </Card>

              {/* Section 4: 付款选项 */}
              <div className='flex items-center mb-2 mt-4 px-1'>
                <Avatar
                  size='small'
                  color='orange'
                  className='mr-2 shadow-md'
                >
                  <CalendarClock size={16} />
                </Avatar>
                <div>
                  <Text className='text-lg font-medium'>
                    {t('付款选项')}
                  </Text>
                  <div className='text-xs text-gray-600'>
                    {t('配置各付款周期的启用状态、折扣和支付渠道')}
                  </div>
                </div>
              </div>

              {/* 月付 */}
              <PeriodCard
                title={t('月付')}
                enabled={monthlyEnabled}
                onToggle={(val) => setMonthlyEnabled(val)}
              >
                <Row gutter={12}>
                  <Col span={24}>
                    <Text type='secondary' style={{ fontSize: 13, marginBottom: 8, display: 'block' }}>
                      {t('月付实际价格')}：{values.price_monthly || 0} {values.currency || 'USD'}
                    </Text>
                  </Col>
                  <Col span={12}>
                    <Form.Input
                      field='monthly_stripe_price_id'
                      label='Stripe Price ID'
                      placeholder='price_...'
                      showClear
                    />
                  </Col>
                  <Col span={12}>
                    <Form.Input
                      field='monthly_creem_product_id'
                      label='Creem Product ID'
                      placeholder='prod_...'
                      showClear
                    />
                  </Col>
                  <Col span={24}>
                    <div style={{ marginTop: 8 }}>
                      <Text strong style={{ fontSize: 13, marginBottom: 4, display: 'block' }}>
                        {t('月付专属优势（可选）')}
                      </Text>
                      <FeatureListEditor value={monthlyFeatures} onChange={setMonthlyFeatures} />
                    </div>
                  </Col>
                </Row>
              </PeriodCard>

              {/* 季付 */}
              <PeriodCard
                title={t('季付')}
                enabled={quarterlyEnabled}
                onToggle={(val) => setQuarterlyEnabled(val)}
              >
                <Row gutter={12}>
                  <Col span={12}>
                    <Form.InputNumber
                      field='quarterly_discount'
                      label={t('折扣百分比')}
                      min={0}
                      max={100}
                      precision={0}
                      style={{ width: '100%' }}
                      extraText={`${t('季付实际月价')}：${calcPrice(values.price_monthly, values.quarterly_discount)} ${values.currency || 'USD'} · ${t('季付总价')}：${(calcPrice(values.price_monthly, values.quarterly_discount) * 3).toFixed(2)}`}
                    />
                  </Col>
                  <Col span={12} />
                  <Col span={12}>
                    <Form.Input
                      field='quarterly_stripe_price_id'
                      label='Stripe Price ID'
                      placeholder='price_...'
                      showClear
                    />
                  </Col>
                  <Col span={12}>
                    <Form.Input
                      field='quarterly_creem_product_id'
                      label='Creem Product ID'
                      placeholder='prod_...'
                      showClear
                    />
                  </Col>
                  <Col span={24}>
                    <div style={{ marginTop: 8 }}>
                      <Text strong style={{ fontSize: 13, marginBottom: 4, display: 'block' }}>
                        {t('季付专属优势（可选）')}
                      </Text>
                      <FeatureListEditor value={quarterlyFeatures} onChange={setQuarterlyFeatures} />
                    </div>
                  </Col>
                </Row>
              </PeriodCard>

              {/* 年付 */}
              <PeriodCard
                title={t('年付')}
                enabled={yearlyEnabled}
                onToggle={(val) => setYearlyEnabled(val)}
              >
                <Row gutter={12}>
                  <Col span={12}>
                    <Form.InputNumber
                      field='yearly_discount'
                      label={t('折扣百分比')}
                      min={0}
                      max={100}
                      precision={0}
                      style={{ width: '100%' }}
                      extraText={`${t('年付实际月价')}：${calcPrice(values.price_monthly, values.yearly_discount)} ${values.currency || 'USD'} · ${t('年付总价')}：${(calcPrice(values.price_monthly, values.yearly_discount) * 12).toFixed(2)}`}
                    />
                  </Col>
                  <Col span={12} />
                  <Col span={12}>
                    <Form.Input
                      field='yearly_stripe_price_id'
                      label='Stripe Price ID'
                      placeholder='price_...'
                      showClear
                    />
                  </Col>
                  <Col span={12}>
                    <Form.Input
                      field='yearly_creem_product_id'
                      label='Creem Product ID'
                      placeholder='prod_...'
                      showClear
                    />
                  </Col>
                  <Col span={24}>
                    <div style={{ marginTop: 8 }}>
                      <Text strong style={{ fontSize: 13, marginBottom: 4, display: 'block' }}>
                        {t('年付专属优势（可选）')}
                      </Text>
                      <FeatureListEditor value={yearlyFeatures} onChange={setYearlyFeatures} />
                    </div>
                  </Col>
                </Row>
              </PeriodCard>
            </div>
          )}
        </Form>
      </Spin>
    </SideSheet>
  );
};

export default AddEditSubscriptionModal;
