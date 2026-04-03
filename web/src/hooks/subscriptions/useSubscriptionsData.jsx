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

import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import { useTableCompactMode } from '../common/useTableCompactMode';

export const useSubscriptionsData = () => {
  const { t } = useTranslation();
  const [compactMode, setCompactMode] = useTableCompactMode('subscriptions');

  // State management — groups with nested plans
  const [allGroups, setAllGroups] = useState([]);
  const [loading, setLoading] = useState(true);

  // Pagination (client-side)
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // Drawer states for group editing
  const [showGroupEdit, setShowGroupEdit] = useState(false);
  const [editingGroup, setEditingGroup] = useState(null);

  // Drawer states for plan variant editing
  const [showPlanEdit, setShowPlanEdit] = useState(false);
  const [editingPlan, setEditingPlan] = useState(null);
  const [editingPlanGroupId, setEditingPlanGroupId] = useState(null);

  const [sheetPlacement, setSheetPlacement] = useState('right');

  // Load groups with plans
  const loadGroups = useCallback(async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/subscription/admin/groups');
      if (res.data?.success) {
        const next = res.data.data || [];
        setAllGroups(next);
        const totalPages = Math.max(1, Math.ceil(next.length / pageSize));
        setActivePage((p) => Math.min(p || 1, totalPages));
      } else {
        showError(res.data?.message || t('加载失败'));
      }
    } catch (e) {
      showError(t('请求失败'));
    } finally {
      setLoading(false);
    }
  }, [pageSize, t]);

  const refresh = useCallback(async () => {
    await loadGroups();
  }, [loadGroups]);

  useEffect(() => {
    loadGroups();
  }, [loadGroups]);

  // Group actions
  const openCreateGroup = useCallback(() => {
    setEditingGroup(null);
    setSheetPlacement('right');
    setShowGroupEdit(true);
  }, []);

  const openEditGroup = useCallback((group) => {
    setEditingGroup(group);
    setSheetPlacement('right');
    setShowGroupEdit(true);
  }, []);

  const closeGroupEdit = useCallback(() => {
    setShowGroupEdit(false);
    setEditingGroup(null);
  }, []);

  const setGroupEnabled = useCallback(async (groupId, enabled) => {
    setLoading(true);
    try {
      const res = await API.patch(`/api/subscription/admin/groups/${groupId}`, { enabled: !!enabled });
      if (res.data?.success) {
        showSuccess(enabled ? t('已启用') : t('已禁用'));
        await loadGroups();
      } else {
        showError(res.data?.message || t('操作失败'));
      }
    } catch (e) {
      showError(t('请求失败'));
    } finally {
      setLoading(false);
    }
  }, [t, loadGroups]);

  const deleteGroup = useCallback(async (groupId) => {
    try {
      const res = await API.delete(`/api/subscription/admin/groups/${groupId}`);
      if (res.data?.success) {
        showSuccess(t('删除成功'));
        await loadGroups();
      } else {
        showError(res.data?.message || t('删除失败'));
      }
    } catch (e) {
      showError(t('删除失败'));
    }
  }, [t, loadGroups]);

  // Plan variant actions
  const openCreatePlan = useCallback((groupId) => {
    setEditingPlan(null);
    setEditingPlanGroupId(groupId);
    setSheetPlacement('right');
    setShowPlanEdit(true);
  }, []);

  const openEditPlan = useCallback((plan) => {
    setEditingPlan(plan);
    setEditingPlanGroupId(plan.group_id);
    setSheetPlacement('right');
    setShowPlanEdit(true);
  }, []);

  const closePlanEdit = useCallback(() => {
    setShowPlanEdit(false);
    setEditingPlan(null);
    setEditingPlanGroupId(null);
  }, []);

  const setPlanEnabled = useCallback(async (planId, enabled) => {
    setLoading(true);
    try {
      const res = await API.patch(`/api/subscription/admin/plans/${planId}`, { enabled: !!enabled });
      if (res.data?.success) {
        showSuccess(enabled ? t('已启用') : t('已禁用'));
        await loadGroups();
      } else {
        showError(res.data?.message || t('操作失败'));
      }
    } catch (e) {
      showError(t('请求失败'));
    } finally {
      setLoading(false);
    }
  }, [t, loadGroups]);

  const groupCount = allGroups.length;
  const groups = allGroups.slice(
    Math.max(0, (activePage - 1) * pageSize),
    Math.max(0, (activePage - 1) * pageSize) + pageSize,
  );

  return {
    // Data state
    groups,
    groupCount,
    loading,

    // Group modal state
    showGroupEdit,
    editingGroup,
    openCreateGroup,
    openEditGroup,
    closeGroupEdit,
    setGroupEnabled,
    deleteGroup,

    // Plan modal state
    showPlanEdit,
    editingPlan,
    editingPlanGroupId,
    openCreatePlan,
    openEditPlan,
    closePlanEdit,
    setPlanEnabled,

    sheetPlacement,

    // UI state
    compactMode,
    setCompactMode,

    // Pagination
    activePage,
    pageSize,
    handlePageChange: setActivePage,
    handlePageSizeChange: (size) => { setPageSize(size); setActivePage(1); },

    // Actions
    refresh,

    // Translation
    t,
  };
};
