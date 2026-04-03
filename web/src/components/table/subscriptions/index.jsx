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

import React from 'react';
import CardPro from '../../common/ui/CardPro';
import SubscriptionsTable from './SubscriptionsTable';
import SubscriptionsActions from './SubscriptionsActions';
import SubscriptionsDescription from './SubscriptionsDescription';
import AddEditSubscriptionModal from './modals/AddEditSubscriptionModal';
import { useSubscriptionsData } from '../../../hooks/subscriptions/useSubscriptionsData';
import { useIsMobile } from '../../../hooks/common/useIsMobile';
import { createCardProPagination } from '../../../helpers/utils';

const SubscriptionsPage = () => {
  const subscriptionsData = useSubscriptionsData();
  const isMobile = useIsMobile();

  const {
    plans,
    planCount,
    loading,

    // Plan modal
    showPlanEdit,
    editingPlan,
    openCreatePlan,
    openEditPlan,
    closePlanEdit,
    setPlanEnabled,
    deletePlan,

    sheetPlacement,

    compactMode,
    setCompactMode,
    refresh,
    t,
  } = subscriptionsData;

  return (
    <>
      <AddEditSubscriptionModal
        visible={showPlanEdit}
        handleClose={closePlanEdit}
        editingPlan={editingPlan}
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
              <SubscriptionsActions openCreatePlan={openCreatePlan} t={t} />
            </div>
          </div>
        }
        paginationArea={createCardProPagination({
          currentPage: subscriptionsData.activePage,
          pageSize: subscriptionsData.pageSize,
          total: planCount,
          onPageChange: subscriptionsData.handlePageChange,
          onPageSizeChange: subscriptionsData.handlePageSizeChange,
          isMobile,
          t,
        })}
        t={t}
      >
        <SubscriptionsTable
          plans={plans}
          loading={loading}
          compactMode={compactMode}
          openEditPlan={openEditPlan}
          setPlanEnabled={setPlanEnabled}
          deletePlan={deletePlan}
          t={t}
        />
      </CardPro>
    </>
  );
};

export default SubscriptionsPage;
