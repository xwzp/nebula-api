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

import React, { useContext } from 'react';
import { Banner } from '@douyinfe/semi-ui';
import CardPro from '../../common/ui/CardPro';
import SubscriptionsTable from './SubscriptionsTable';
import SubscriptionsActions from './SubscriptionsActions';
import SubscriptionsDescription from './SubscriptionsDescription';
import AddEditGroupModal from './modals/AddEditGroupModal';
import AddEditSubscriptionModal from './modals/AddEditSubscriptionModal';
import { useSubscriptionsData } from '../../../hooks/subscriptions/useSubscriptionsData';
import { useIsMobile } from '../../../hooks/common/useIsMobile';
import { createCardProPagination } from '../../../helpers/utils';
import { StatusContext } from '../../../context/Status';

const SubscriptionsPage = () => {
  const subscriptionsData = useSubscriptionsData();
  const isMobile = useIsMobile();
  const [statusState] = useContext(StatusContext);
  const enableEpay = !!statusState?.status?.enable_online_topup;

  const {
    groups,
    groupCount,
    loading,

    // Group modal
    showGroupEdit,
    editingGroup,
    openCreateGroup,
    openEditGroup,
    closeGroupEdit,
    setGroupEnabled,
    deleteGroup,

    // Plan modal
    showPlanEdit,
    editingPlan,
    editingPlanGroupId,
    openCreatePlan,
    openEditPlan,
    closePlanEdit,
    setPlanEnabled,

    sheetPlacement,

    compactMode,
    setCompactMode,
    refresh,
    t,
  } = subscriptionsData;

  return (
    <>
      <AddEditGroupModal
        visible={showGroupEdit}
        handleClose={closeGroupEdit}
        editingGroup={editingGroup}
        placement={sheetPlacement}
        refresh={refresh}
        t={t}
      />

      <AddEditSubscriptionModal
        visible={showPlanEdit}
        handleClose={closePlanEdit}
        editingPlan={editingPlan}
        groupId={editingPlanGroupId}
        placement={sheetPlacement}
        refresh={refresh}
        t={t}
      />

      <CardPro
        type='type1'
        descriptionArea={
          <SubscriptionsDescription
            compactMode={compactMode}
            setCompactMode={setCompactMode}
            t={t}
          />
        }
        actionsArea={
          <div className='flex flex-col md:flex-row justify-between items-start md:items-center gap-2 w-full'>
            <div className='order-1 md:order-0 w-full md:w-auto'>
              <SubscriptionsActions openCreateGroup={openCreateGroup} t={t} />
            </div>
            <Banner
              type='info'
              description={t('套餐组包含多个计费周期（月付/年付等），展开可管理各周期')}
              closeIcon={null}
              className='!rounded-lg order-2 md:order-1'
              style={{ maxWidth: '100%' }}
            />
          </div>
        }
        paginationArea={createCardProPagination({
          currentPage: subscriptionsData.activePage,
          pageSize: subscriptionsData.pageSize,
          total: groupCount,
          onPageChange: subscriptionsData.handlePageChange,
          onPageSizeChange: subscriptionsData.handlePageSizeChange,
          isMobile,
          t,
        })}
        t={t}
      >
        <SubscriptionsTable
          groups={groups}
          loading={loading}
          compactMode={compactMode}
          openEditGroup={openEditGroup}
          setGroupEnabled={setGroupEnabled}
          deleteGroup={deleteGroup}
          openCreatePlan={openCreatePlan}
          openEditPlan={openEditPlan}
          setPlanEnabled={setPlanEnabled}
          enableEpay={enableEpay}
          t={t}
        />
      </CardPro>
    </>
  );
};

export default SubscriptionsPage;
