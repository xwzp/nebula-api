import { Button, Popconfirm, Switch, Tag } from '@douyinfe/semi-ui';
import { IconEdit } from '@douyinfe/semi-icons';

export function getTopupTiersColumns({ t, openEdit, setTierEnabled, deleteTier }) {
  return [
    {
      title: t('标题'),
      dataIndex: 'title',
      key: 'title',
      render: (text, record) => (
        <div>
          <span style={{ fontWeight: 500 }}>{text}</span>
          {record.tag && (
            <Tag color='purple' size='small' style={{ marginLeft: 6 }}>
              {record.tag}
            </Tag>
          )}
          {record.subtitle && (
            <div style={{ fontSize: 12, color: 'var(--semi-color-text-2)' }}>
              {record.subtitle}
            </div>
          )}
        </div>
      ),
    },
    {
      title: t('金额'),
      dataIndex: 'amount',
      key: 'amount',
      width: 100,
      render: (val) => `$${val}`,
    },
    {
      title: t('折扣'),
      dataIndex: 'discount',
      key: 'discount',
      width: 100,
      render: (val) => {
        if (!val || val >= 1) return '-';
        const zhe = Math.round(val * 100);
        return <Tag color='green'>{zhe % 10 === 0 ? `${zhe / 10}折` : `${zhe}折`}</Tag>;
      },
    },
    {
      title: t('赠送额度'),
      dataIndex: 'bonus_quota',
      key: 'bonus_quota',
      width: 100,
      render: (val) => (val > 0 ? `+${val}` : '-'),
    },
    {
      title: t('排序'),
      dataIndex: 'sort_order',
      key: 'sort_order',
      width: 80,
    },
    {
      title: t('状态'),
      dataIndex: 'enabled',
      key: 'enabled',
      width: 80,
      render: (val, record) => (
        <Switch
          checked={val}
          onChange={(checked) => setTierEnabled(record.id, checked)}
          size='small'
        />
      ),
    },
    {
      title: t('操作'),
      key: 'operate',
      width: 120,
      render: (_, record) => (
        <div style={{ display: 'flex', gap: 4 }}>
          <Button
            icon={<IconEdit />}
            theme='borderless'
            size='small'
            onClick={() => openEdit(record)}
          />
          <Popconfirm
            title={t('确认删除？')}
            onConfirm={() => deleteTier(record.id)}
          >
            <Button theme='borderless' size='small' type='danger'>
              {t('删除')}
            </Button>
          </Popconfirm>
        </div>
      ),
    },
  ];
}
