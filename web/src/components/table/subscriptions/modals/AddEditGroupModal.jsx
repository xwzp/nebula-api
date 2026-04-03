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

import React, { useRef, useState } from 'react';
import {
  SideSheet,
  Form,
  Button,
  Card,
  Space,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess } from '../../../../helpers';
import FeatureListEditor from '../../../common/FeatureListEditor';

const { Title } = Typography;

const AddEditGroupModal = ({
  visible,
  handleClose,
  editingGroup,
  placement = 'right',
  refresh,
  t,
}) => {
  const formRef = useRef();
  const [loading, setLoading] = useState(false);
  const [features, setFeatures] = useState([]);

  const isEdit = editingGroup?.id > 0;

  const handleOpen = () => {
    if (isEdit) {
      const g = editingGroup;
      setTimeout(() => {
        formRef.current?.setValues({
          title: g.title || '',
          subtitle: g.subtitle || '',
          tag: g.tag || '',
          sort_order: g.sort_order || 0,
          enabled: g.enabled ?? true,
        });
        try {
          setFeatures(g.features ? JSON.parse(g.features) : []);
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
        res = await API.put(
          `/api/subscription/admin/groups/${editingGroup.id}`,
          body,
        );
      } else {
        res = await API.post('/api/subscription/admin/groups', body);
      }
      if (res.data?.success) {
        showSuccess(isEdit ? t('更新成功') : t('创建成功'));
        handleClose();
        refresh?.();
      } else {
        showError(res.data?.message || t('操作失败'));
      }
    } catch {
      showError(t('操作失败'));
    }
    setLoading(false);
  };

  return (
    <SideSheet
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
            {isEdit ? t('编辑套餐组') : t('创建套餐组')}
          </Title>
        </Space>
      }
      visible={visible}
      onCancel={handleClose}
      placement={placement}
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
          <Form.Input
            field='title'
            label={t('套餐组标题')}
            placeholder={t('例如：Pro 套餐')}
            rules={[{ required: true, message: t('标题不能为空') }]}
          />
          <Form.Input
            field='subtitle'
            label={t('副标题')}
            placeholder={t('例如：适合专业用户')}
          />
          <Form.Input
            field='tag'
            label={t('标签')}
            placeholder={t('例如：推荐')}
          />
        </Card>

        <Card title={t('展示设置')} style={{ marginBottom: 16 }}>
          <Form.InputNumber
            field='sort_order'
            label={t('排序')}
            extraText={t('数字越大越靠前')}
          />
          <Form.Switch field='enabled' label={t('启用')} />
        </Card>

        <Card title={t('优势列表')} style={{ marginBottom: 16 }}>
          <FeatureListEditor value={features} onChange={setFeatures} />
        </Card>
      </Form>
    </SideSheet>
  );
};

export default AddEditGroupModal;
