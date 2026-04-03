import React from 'react';
import { Card } from '@douyinfe/semi-ui';
import TopupTiersTable from './TopupTiersTable';
import TopupTiersActions from './TopupTiersActions';
import AddEditTopupTierModal from './modals/AddEditTopupTierModal';
import useTopupTiersData from '../../../hooks/topup-tiers/useTopupTiersData';

export default function TopupTiersPage() {
  const {
    paginatedTiers,
    loading,
    showEdit,
    editingTier,
    sheetPlacement,
    openCreate,
    openEdit,
    closeEdit,
    setTierEnabled,
    deleteTier,
    refresh,
    t,
  } = useTopupTiersData();

  return (
    <div>
      <Card
        title={t('充值档位管理')}
        headerExtraContent={<TopupTiersActions openCreate={openCreate} t={t} />}
        style={{ marginBottom: 16 }}
      >
        <TopupTiersTable
          tiers={paginatedTiers}
          loading={loading}
          t={t}
          openEdit={openEdit}
          setTierEnabled={setTierEnabled}
          deleteTier={deleteTier}
        />
      </Card>

      <AddEditTopupTierModal
        visible={showEdit}
        handleClose={closeEdit}
        editingTier={editingTier}
        placement={sheetPlacement}
        refresh={refresh}
      />
    </div>
  );
}
