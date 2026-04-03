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

import React, { useMemo } from 'react';
import { Button, Empty, Table, Typography } from '@douyinfe/semi-ui';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import { IconPlus } from '@douyinfe/semi-icons';
import CardTable from '../../common/ui/CardTable';
import { getGroupColumns, getPlanVariantColumns } from './SubscriptionsColumnDefs';

const { Text } = Typography;

const SubscriptionsTable = ({
  groups,
  loading,
  compactMode,
  openEditGroup,
  setGroupEnabled,
  deleteGroup,
  openCreatePlan,
  openEditPlan,
  setPlanEnabled,
  enableEpay,
  t,
}) => {
  const groupColumns = useMemo(() => {
    return getGroupColumns({
      t,
      openEditGroup,
      setGroupEnabled,
      deleteGroup,
    });
  }, [t, openEditGroup, setGroupEnabled, deleteGroup]);

  const tableColumns = useMemo(() => {
    return compactMode
      ? groupColumns.map((col) => {
          if (col.dataIndex === 'operate') {
            const { fixed, ...rest } = col;
            return rest;
          }
          return col;
        })
      : groupColumns;
  }, [compactMode, groupColumns]);

  const planColumns = useMemo(() => {
    return getPlanVariantColumns({
      t,
      openEditPlan,
      setPlanEnabled,
      enableEpay,
    });
  }, [t, openEditPlan, setPlanEnabled, enableEpay]);

  const expandedRowRender = (group) => {
    const plans = group?.plans || [];
    return (
      <div style={{ padding: '8px 0 8px 16px' }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 8,
          }}
        >
          <Text strong type='secondary'>
            {t('计费周期')}
          </Text>
          <Button
            icon={<IconPlus />}
            size='small'
            theme='light'
            onClick={() => openCreatePlan(group.id)}
          >
            {t('添加周期')}
          </Button>
        </div>
        {plans.length > 0 ? (
          <Table
            columns={planColumns}
            dataSource={plans}
            pagination={false}
            rowKey='id'
            size='small'
            scroll={{ x: 'max-content' }}
          />
        ) : (
          <Text type='tertiary' style={{ padding: '16px 0', display: 'block', textAlign: 'center' }}>
            {t('暂无计费周期，点击上方按钮添加')}
          </Text>
        )}
      </div>
    );
  };

  return (
    <CardTable
      columns={tableColumns}
      dataSource={groups}
      scroll={compactMode ? undefined : { x: 'max-content' }}
      pagination={false}
      hidePagination={true}
      loading={loading}
      rowKey='id'
      expandedRowRender={expandedRowRender}
      empty={
        <Empty
          image={<IllustrationNoResult style={{ width: 150, height: 150 }} />}
          darkModeImage={
            <IllustrationNoResultDark style={{ width: 150, height: 150 }} />
          }
          description={t('暂无套餐组')}
          style={{ padding: 30 }}
        />
      }
      className='overflow-hidden'
      size='middle'
    />
  );
};

export default SubscriptionsTable;
