import React, { useRef, useState } from 'react';
import {
  SideSheet,
  Form,
  Button,
  Card,
  InputNumber,
  Switch,
  Space,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../../../helpers';
import FeatureListEditor from '../../../common/FeatureListEditor';

export default function AddEditTopupTierModal({ visible, handleClose, editingTier, placement, refresh }) {
  const { t } = useTranslation();
  const formRef = useRef();
  const [loading, setLoading] = useState(false);
  const [features, setFeatures] = useState([]);

  const isEdit = editingTier?.id > 0;

  const handleOpen = () => {
    if (isEdit) {
      const tier = editingTier;
      setTimeout(() => {
        formRef.current?.setValues({
          title: tier.title || '',
          subtitle: tier.subtitle || '',
          tag: tier.tag || '',
          amount: tier.amount || 0,
          discount: tier.discount || 1,
          bonus_quota: tier.bonus_quota || 0,
          sort_order: tier.sort_order || 0,
          enabled: tier.enabled ?? true,
        });
        try {
          setFeatures(tier.features ? JSON.parse(tier.features) : []);
        } catch {
          setFeatures([]);
        }
      }, 100);
    } else {
      setTimeout(() => {
        formRef.current?.setValues({
          title: '',
          subtitle: '',
          tag: '',
          amount: 10,
          discount: 1,
          bonus_quota: 0,
          sort_order: 0,
          enabled: true,
        });
        setFeatures([]);
      }, 100);
    }
  };

  const handleSubmit = async () => {
    try {
      await formRef.current?.validate();
    } catch {
      return;
    }
    const values = formRef.current?.getValues();
    const body = {
      ...values,
      features: JSON.stringify(features),
    };

    setLoading(true);
    try {
      let res;
      if (isEdit) {
        res = await API.put(`/api/topup/admin/tiers/${editingTier.id}`, body);
      } else {
        res = await API.post('/api/topup/admin/tiers', body);
      }
      if (res.data?.success) {
        showSuccess(isEdit ? t('更新成功') : t('创建成功'));
        handleClose();
        refresh();
      } else {
        showError(res.data?.message || t('操作失败'));
      }
    } catch (e) {
      showError(t('操作失败'));
    }
    setLoading(false);
  };

  return (
    <SideSheet
      title={isEdit ? t('编辑充值档位') : t('创建充值档位')}
      visible={visible}
      onCancel={handleClose}
      placement={placement || 'right'}
      width={500}
      afterVisibleChange={(v) => v && handleOpen()}
      footer={
        <Space>
          <Button onClick={handleClose}>{t('取消')}</Button>
          <Button theme='solid' loading={loading} onClick={handleSubmit}>
            {isEdit ? t('更新') : t('创建')}
          </Button>
        </Space>
      }
    >
      <Form
        getFormApi={(api) => (formRef.current = api)}
        labelPosition='top'
        style={{ padding: '0 8px' }}
      >
        <Card title={t('基本信息')} style={{ marginBottom: 16 }}>
          <Form.Input field='title' label={t('标题')} rules={[{ required: true, message: t('标题不能为空') }]} />
          <Form.Input field='subtitle' label={t('副标题')} />
          <Form.Input field='tag' label={t('标签')} placeholder={t('如：最划算')} />
        </Card>

        <Card title={t('价格设置')} style={{ marginBottom: 16 }}>
          <Form.InputNumber field='amount' label={t('充值金额 (USD)')} min={1} rules={[{ required: true, message: t('金额不能为空') }]} />
          <Form.InputNumber field='discount' label={t('折扣率')} min={0.01} max={1} step={0.01}
            extraText={t('1.0 = 无折扣, 0.95 = 95折')} />
          <Form.InputNumber field='bonus_quota' label={t('赠送额度')} min={0}
            extraText={t('充值后额外赠送的配额，0 = 不赠送')} />
        </Card>

        <Card title={t('展示设置')} style={{ marginBottom: 16 }}>
          <Form.InputNumber field='sort_order' label={t('排序')} extraText={t('数字越大越靠前')} />
          <Form.Switch field='enabled' label={t('启用')} />
        </Card>

        <Card title={t('优势列表')} style={{ marginBottom: 16 }}>
          <FeatureListEditor value={features} onChange={setFeatures} />
        </Card>
      </Form>
    </SideSheet>
  );
}
