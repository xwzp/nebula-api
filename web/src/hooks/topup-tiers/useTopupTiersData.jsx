import { useState, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';

export default function useTopupTiersData() {
  const { t } = useTranslation();
  const [allTiers, setAllTiers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [showEdit, setShowEdit] = useState(false);
  const [editingTier, setEditingTier] = useState(null);
  const [sheetPlacement, setSheetPlacement] = useState('right');

  const loadTiers = useCallback(async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/topup/admin/tiers');
      if (res.data?.success) {
        setAllTiers(res.data.data || []);
      } else {
        showError(res.data?.message || t('加载失败'));
      }
    } catch (e) {
      showError(t('加载失败'));
    }
    setLoading(false);
  }, [t]);

  useEffect(() => {
    loadTiers();
  }, [loadTiers]);

  const refresh = useCallback(() => {
    loadTiers();
  }, [loadTiers]);

  const openCreate = useCallback(() => {
    setEditingTier(null);
    setSheetPlacement('right');
    setShowEdit(true);
  }, []);

  const openEdit = useCallback((tier) => {
    setEditingTier(tier);
    setSheetPlacement('right');
    setShowEdit(true);
  }, []);

  const closeEdit = useCallback(() => {
    setShowEdit(false);
    setEditingTier(null);
  }, []);

  const setTierEnabled = useCallback(async (tierId, enabled) => {
    try {
      const res = await API.patch(`/api/topup/admin/tiers/${tierId}`, { enabled });
      if (res.data?.success) {
        showSuccess(t('操作成功'));
        refresh();
      } else {
        showError(res.data?.message || t('操作失败'));
      }
    } catch (e) {
      showError(t('操作失败'));
    }
  }, [t, refresh]);

  const deleteTier = useCallback(async (tierId) => {
    try {
      const res = await API.delete(`/api/topup/admin/tiers/${tierId}`);
      if (res.data?.success) {
        showSuccess(t('删除成功'));
        refresh();
      } else {
        showError(res.data?.message || t('删除失败'));
      }
    } catch (e) {
      showError(t('删除失败'));
    }
  }, [t, refresh]);

  const total = allTiers.length;
  const paginatedTiers = allTiers.slice((activePage - 1) * pageSize, activePage * pageSize);

  return {
    allTiers,
    paginatedTiers,
    loading,
    total,
    activePage,
    setActivePage,
    pageSize,
    setPageSize,
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
  };
}
