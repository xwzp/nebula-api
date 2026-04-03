import React from 'react';
import { Button, Input, Select, Space, Typography } from '@douyinfe/semi-ui';
import { IconPlus, IconDelete, IconChevronUp, IconChevronDown } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

const iconOptions = [
  { value: 'check', label: '✓ Check' },
  { value: 'x', label: '✗ Cross' },
  { value: 'info', label: 'ℹ Info' },
];

const styleOptions = [
  { value: 'default', label: 'Default' },
  { value: 'highlight', label: 'Highlight' },
  { value: 'disabled', label: 'Disabled' },
];

export default function FeatureListEditor({ value = [], onChange }) {
  const { t } = useTranslation();
  const items = Array.isArray(value) ? value : [];

  const updateItem = (index, field, val) => {
    const newItems = [...items];
    newItems[index] = { ...newItems[index], [field]: val };
    onChange(newItems);
  };

  const addItem = () => {
    onChange([...items, { text: '', icon: 'check', style: 'default' }]);
  };

  const removeItem = (index) => {
    onChange(items.filter((_, i) => i !== index));
  };

  const moveItem = (index, direction) => {
    const newItems = [...items];
    const targetIndex = index + direction;
    if (targetIndex < 0 || targetIndex >= newItems.length) return;
    [newItems[index], newItems[targetIndex]] = [newItems[targetIndex], newItems[index]];
    onChange(newItems);
  };

  return (
    <div>
      <Text strong style={{ marginBottom: 8, display: 'block' }}>
        {t('优势列表')}
      </Text>
      {items.map((item, index) => (
        <Space key={index} style={{ marginBottom: 8, width: '100%' }} align='start'>
          <Input
            value={item.text}
            onChange={(val) => updateItem(index, 'text', val)}
            placeholder={t('功能描述')}
            style={{ width: 200 }}
          />
          <Select
            value={item.icon}
            onChange={(val) => updateItem(index, 'icon', val)}
            optionList={iconOptions}
            style={{ width: 110 }}
          />
          <Select
            value={item.style}
            onChange={(val) => updateItem(index, 'style', val)}
            optionList={styleOptions}
            style={{ width: 110 }}
          />
          <Button
            icon={<IconChevronUp />}
            theme='borderless'
            disabled={index === 0}
            onClick={() => moveItem(index, -1)}
          />
          <Button
            icon={<IconChevronDown />}
            theme='borderless'
            disabled={index === items.length - 1}
            onClick={() => moveItem(index, 1)}
          />
          <Button
            icon={<IconDelete />}
            theme='borderless'
            type='danger'
            onClick={() => removeItem(index)}
          />
        </Space>
      ))}
      <Button icon={<IconPlus />} theme='light' onClick={addItem} style={{ marginTop: 4 }}>
        {t('添加优势')}
      </Button>
    </div>
  );
}
