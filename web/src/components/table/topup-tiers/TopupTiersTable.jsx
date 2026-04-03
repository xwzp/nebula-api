import React, { useMemo } from 'react';
import { Table } from '@douyinfe/semi-ui';
import { getTopupTiersColumns } from './TopupTiersColumnDefs';

export default function TopupTiersTable({ tiers, loading, t, openEdit, setTierEnabled, deleteTier }) {
  const columns = useMemo(
    () => getTopupTiersColumns({ t, openEdit, setTierEnabled, deleteTier }),
    [t, openEdit, setTierEnabled, deleteTier],
  );

  return (
    <Table
      columns={columns}
      dataSource={tiers}
      loading={loading}
      rowKey='id'
      pagination={false}
      empty={t('暂无数据')}
    />
  );
}
