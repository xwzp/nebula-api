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
import React, { useState, useEffect, useMemo } from 'react';
import {
  Modal,
  Table,
  Badge,
  Typography,
  Toast,
  Empty,
  Button,
  Input,
  Tag,
} from '@douyinfe/semi-ui';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import { Coins } from 'lucide-react';
import { IconSearch } from '@douyinfe/semi-icons';
import { API, timestamp2string } from '../../../helpers';
import { isAdmin } from '../../../helpers/utils';
import { useIsMobile } from '../../../hooks/common/useIsMobile';
const { Text } = Typography;

// 状态映射配置
const STATUS_CONFIG = {
  success: { type: 'success', key: '成功' },
  pending: { type: 'warning', key: '待支付' },
  failed: { type: 'danger', key: '失败' },
  expired: { type: 'danger', key: '已过期' },
  cancelled: { type: 'tertiary', key: '已取消' },
};

// 支付方式映射
const PAYMENT_METHOD_MAP = {
  stripe: 'Stripe',
  creem: 'Creem',
  waffo: 'Waffo',
  alipay: '支付宝',
  wxpay: '微信',
  wechat: '微信支付',
};

const TopupHistoryModal = ({ visible, onCancel, t, onShowWechatQr, onShowAlipayQr }) => {
  const [loading, setLoading] = useState(false);
  const [topups, setTopups] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [keyword, setKeyword] = useState('');
  const isMobile = useIsMobile();

  const loadTopups = async (currentPage, currentPageSize) => {
    setLoading(true);
    try {
      const base = isAdmin() ? '/api/user/topup' : '/api/user/topup/self';
      const qs =
        `p=${currentPage}&page_size=${currentPageSize}` +
        (keyword ? `&keyword=${encodeURIComponent(keyword)}` : '');
      const endpoint = `${base}?${qs}`;
      const res = await API.get(endpoint);
      const { success, message, data } = res.data;
      if (success) {
        setTopups(data.items || []);
        setTotal(data.total || 0);
      } else {
        Toast.error({ content: message || t('加载失败') });
      }
    } catch (error) {
      Toast.error({ content: t('加载账单失败') });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (visible) {
      loadTopups(page, pageSize);
    }
  }, [visible, page, pageSize, keyword]);

  const handlePageChange = (currentPage) => {
    setPage(currentPage);
  };

  const handlePageSizeChange = (currentPageSize) => {
    setPageSize(currentPageSize);
    setPage(1);
  };

  const handleKeywordChange = (value) => {
    setKeyword(value);
    setPage(1);
  };

  // 管理员补单
  const handleAdminComplete = async (tradeNo) => {
    try {
      const res = await API.post('/api/user/topup/complete', {
        trade_no: tradeNo,
      });
      const { success, message } = res.data;
      if (success) {
        Toast.success({ content: t('补单成功') });
        await loadTopups(page, pageSize);
      } else {
        Toast.error({ content: message || t('补单失败') });
      }
    } catch (e) {
      Toast.error({ content: t('补单失败') });
    }
  };

  const confirmAdminComplete = (tradeNo) => {
    Modal.confirm({
      title: t('确认补单'),
      content: t('是否将该订单标记为成功并为用户入账？'),
      onOk: () => handleAdminComplete(tradeNo),
    });
  };

  // 取消订单
  const handleCancelOrder = async (tradeNo) => {
    try {
      const res = await API.post('/api/user/topup/cancel', {
        trade_no: tradeNo,
      });
      const { success, message, data } = res.data;
      if (success || message === 'success') {
        Toast.success({ content: t('订单已取消') });
        await loadTopups(page, pageSize);
      } else {
        Toast.error({ content: data || message || t('取消失败') });
      }
    } catch (e) {
      Toast.error({ content: t('取消失败') });
    }
  };

  const confirmCancelOrder = (tradeNo) => {
    Modal.confirm({
      title: t('确认取消'),
      content: t('确定要取消该订单吗？'),
      onOk: () => handleCancelOrder(tradeNo),
    });
  };

  // 重新显示微信支付二维码
  const handleShowQrCode = async (tradeNo, paymentMethod) => {
    try {
      const res = await API.get(`/api/user/topup/qrcode?trade_no=${encodeURIComponent(tradeNo)}`);
      const { message, data } = res.data;
      if (message === 'success' && data?.code_url) {
        if (paymentMethod === 'alipay') {
          onShowAlipayQr?.(data.code_url, data.pay_money || 0);
        } else {
          onShowWechatQr?.(data.code_url, data.pay_money || 0);
        }
      } else {
        Toast.error({ content: data || t('获取二维码失败') });
      }
    } catch (e) {
      Toast.error({ content: t('获取二维码失败') });
    }
  };

  // 渲染状态徽章
  const renderStatusBadge = (status) => {
    const config = STATUS_CONFIG[status] || { type: 'primary', key: status };
    return (
      <span className='flex items-center gap-2'>
        <Badge dot type={config.type} />
        <span>{t(config.key)}</span>
      </span>
    );
  };

  // 渲染支付方式
  const renderPaymentMethod = (pm) => {
    const displayName = PAYMENT_METHOD_MAP[pm];
    return <Text>{displayName ? t(displayName) : pm || '-'}</Text>;
  };

  const isSubscriptionTopup = (record) => {
    const tradeNo = (record?.trade_no || '').toLowerCase();
    return Number(record?.amount || 0) === 0 && tradeNo.startsWith('sub');
  };

  // 检查是否为管理员
  const userIsAdmin = useMemo(() => isAdmin(), []);

  const columns = useMemo(() => {
    const baseColumns = [
      {
        title: t('订单号'),
        dataIndex: 'trade_no',
        key: 'trade_no',
        render: (text) => <Text copyable>{text}</Text>,
      },
    ];

    // 管理员显示用户名列
    if (userIsAdmin) {
      baseColumns.push({
        title: t('用户'),
        dataIndex: 'username',
        key: 'username',
        render: (username, record) => (
          <Text>{username || `UID:${record.user_id}`}</Text>
        ),
      });
    }

    baseColumns.push(
      {
        title: t('支付方式'),
        dataIndex: 'payment_method',
        key: 'payment_method',
        render: renderPaymentMethod,
      },
      {
        title: t('充值额度'),
        dataIndex: 'amount',
        key: 'amount',
        render: (amount, record) => {
          if (isSubscriptionTopup(record)) {
            return (
              <Tag color='purple' shape='circle' size='small'>
                {t('订阅套餐')}
              </Tag>
            );
          }
          return (
            <span className='flex items-center gap-1'>
              <Coins size={16} />
              <Text>{amount}</Text>
            </span>
          );
        },
      },
      {
        title: t('支付金额'),
        dataIndex: 'money',
        key: 'money',
        render: (money) => <Text type='danger'>¥{money.toFixed(2)}</Text>,
      },
      {
        title: t('状态'),
        dataIndex: 'status',
        key: 'status',
        render: renderStatusBadge,
      },
      {
        title: t('操作'),
        key: 'action',
        render: (_, record) => {
          if (record.status !== 'pending') return null;
          const actions = [];
          // 管理员补单
          if (userIsAdmin) {
            actions.push(
              <Button
                key='complete'
                size='small'
                type='primary'
                theme='outline'
                onClick={() => confirmAdminComplete(record.trade_no)}
              >
                {t('补单')}
              </Button>,
            );
          }
          // 取消订单（用户和管理员都可）
          actions.push(
            <Button
              key='cancel'
              size='small'
              type='danger'
              theme='outline'
              onClick={() => confirmCancelOrder(record.trade_no)}
            >
              {t('取消')}
            </Button>,
          );
          // 微信支付可重新显示二维码（仅用户侧）
          if (record.payment_method === 'wechat' && onShowWechatQr) {
            actions.push(
              <Button
                key='qr'
                size='small'
                theme='outline'
                onClick={() => handleShowQrCode(record.trade_no)}
              >
                {t('付款')}
              </Button>,
            );
          }
          // 支付宝可重新显示二维码
          if (record.payment_method === 'alipay' && onShowAlipayQr) {
            actions.push(
              <Button
                key='qr'
                size='small'
                theme='outline'
                onClick={() => handleShowQrCode(record.trade_no, 'alipay')}
              >
                {t('付款')}
              </Button>,
            );
          }
          return actions.length > 0 ? (
            <div className='flex gap-1 flex-wrap'>{actions}</div>
          ) : null;
        },
      },
      {
        title: t('创建时间'),
        dataIndex: 'create_time',
        key: 'create_time',
        render: (time) => timestamp2string(time),
      },
    );

    return baseColumns;
  }, [t, userIsAdmin]);

  return (
    <Modal
      title={t('充值账单')}
      visible={visible}
      onCancel={onCancel}
      footer={null}
      size={isMobile ? 'full-width' : 'large'}
    >
      <div className='mb-3'>
        <Input
          prefix={<IconSearch />}
          placeholder={t('订单号')}
          value={keyword}
          onChange={handleKeywordChange}
          showClear
        />
      </div>
      <Table
        columns={columns}
        dataSource={topups}
        loading={loading}
        rowKey='id'
        pagination={{
          currentPage: page,
          pageSize: pageSize,
          total: total,
          showSizeChanger: true,
          pageSizeOpts: [10, 20, 50, 100],
          onPageChange: handlePageChange,
          onPageSizeChange: handlePageSizeChange,
        }}
        size='small'
        empty={
          <Empty
            image={<IllustrationNoResult style={{ width: 150, height: 150 }} />}
            darkModeImage={
              <IllustrationNoResultDark style={{ width: 150, height: 150 }} />
            }
            description={t('暂无充值记录')}
            style={{ padding: 30 }}
          />
        }
      />
    </Modal>
  );
};

export default TopupHistoryModal;
