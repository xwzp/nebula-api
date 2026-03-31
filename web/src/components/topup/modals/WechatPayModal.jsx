import React from 'react';
import { QRCodeSVG } from 'qrcode.react';
import { X } from 'lucide-react';
import { SiWechat, SiAlipay } from 'react-icons/si';
import { useTranslation } from 'react-i18next';

// 支付方式主题配置 — 使用 inline style 保证颜色渲染
const THEME = {
  wechat: {
    name: '微信支付',
    scanTip: '请使用微信支付扫码完成支付',
    openTip: '请打开微信支付扫一扫',
    headerStyle: { background: 'linear-gradient(to right, #22c55e, #16a34a)', color: '#ffffff' },
    borderColor: '#10b981',
    pillStyle: { backgroundColor: '#dcfce7', color: '#15803d' },
    amountStyle: {
      background: 'linear-gradient(to right, #f0fdf4, #dcfce7)',
      border: '1px solid #bbf7d0',
    },
    amountTextColor: '#16a34a',
    tipStyle: { backgroundColor: '#f0fdf4', border: '1px solid #bbf7d0' },
    tipTextColor: '#15803d',
    Icon: SiWechat,
  },
  alipay: {
    name: '支付宝',
    scanTip: '请使用支付宝扫码完成支付',
    openTip: '请打开支付宝扫一扫',
    headerStyle: { background: 'linear-gradient(to right, #3b82f6, #2563eb)', color: '#ffffff' },
    borderColor: '#3b82f6',
    pillStyle: { backgroundColor: '#dbeafe', color: '#1d4ed8' },
    amountStyle: {
      background: 'linear-gradient(to right, #eff6ff, #dbeafe)',
      border: '1px solid #bfdbfe',
    },
    amountTextColor: '#2563eb',
    tipStyle: { backgroundColor: '#eff6ff', border: '1px solid #bfdbfe' },
    tipTextColor: '#1d4ed8',
    Icon: SiAlipay,
  },
};

const ScanPayModal = ({
  visible,
  onCancel,
  codeUrl,
  payMoney,
  topUpCount,
  renderQuotaWithAmount,
  paymentMethod = 'wechat',
}) => {
  const { t } = useTranslation();

  if (!visible) return null;

  const theme = THEME[paymentMethod] || THEME.wechat;
  const { Icon } = theme;

  return (
    <div className='fixed inset-0 z-[1100] flex items-center justify-center p-4'>
      {/* Backdrop */}
      <div
        className='absolute inset-0 bg-black/50 backdrop-blur-sm'
        onClick={onCancel}
      />

      {/* Modal */}
      <div
        className='relative w-full max-w-md rounded-2xl shadow-2xl overflow-hidden animate-[modalIn_0.3s_ease-out]'
        style={{ backgroundColor: '#ffffff' }}
      >
        {/* Close button */}
        <button
          onClick={onCancel}
          className='absolute top-4 right-4 z-10 p-2 rounded-full shadow-lg transition-all'
          style={{ backgroundColor: 'rgba(255,255,255,0.9)' }}
        >
          <X className='w-5 h-5 text-gray-600' />
        </button>

        {/* Header */}
        <div
          className='px-6 py-8 text-white'
          style={theme.headerStyle}
        >
          <div className='flex flex-col items-center gap-3'>
            <div
              className='w-16 h-16 rounded-2xl flex items-center justify-center'
              style={{ backgroundColor: 'rgba(255,255,255,0.2)' }}
            >
              <Icon className='w-10 h-10' style={{ color: '#ffffff' }} />
            </div>
            <h1 className='text-2xl text-center font-medium'>
              {t(theme.name)}
            </h1>
            <p className='text-center text-sm' style={{ opacity: 0.9 }}>
              {t(theme.scanTip)}
            </p>
          </div>
        </div>

        {/* QR Code */}
        <div
          className='px-6 py-8 flex flex-col items-center'
          style={{ backgroundColor: '#f9fafb' }}
        >
          {codeUrl && (
            <div
              className='p-5 rounded-2xl shadow-lg'
              style={{
                backgroundColor: '#ffffff',
                border: `4px solid ${theme.borderColor}`,
              }}
            >
              <QRCodeSVG value={codeUrl} size={220} level='H' />
            </div>
          )}
          <div
            className='mt-5 px-6 py-3 rounded-full'
            style={theme.pillStyle}
          >
            <p className='text-sm font-medium'>
              {t(theme.openTip)}
            </p>
          </div>
        </div>

        {/* Amount info */}
        <div className='px-6 py-6 space-y-3'>
          {payMoney > 0 && (
            <div
              className='flex items-center justify-between p-4 rounded-xl'
              style={{
                background: 'linear-gradient(to right, #fef2f2, #fee2e2)',
                border: '1px solid #fecaca',
              }}
            >
              <span className='text-gray-700 font-medium'>
                {t('实付金额')}
              </span>
              <span
                className='text-3xl font-bold'
                style={{ color: '#dc2626' }}
              >
                ¥{payMoney.toFixed(2)}
              </span>
            </div>
          )}

          {topUpCount > 0 && (
            <div
              className='flex items-center justify-between p-4 rounded-xl'
              style={theme.amountStyle}
            >
              <span className='text-gray-700 font-medium'>
                {t('到账额度')}
              </span>
              <span
                className='text-xl font-bold'
                style={{ color: theme.amountTextColor }}
              >
                {renderQuotaWithAmount
                  ? renderQuotaWithAmount(topUpCount)
                  : topUpCount}
              </span>
            </div>
          )}
        </div>

        {/* Tip */}
        <div className='px-6 pb-6'>
          <div
            className='rounded-lg p-4'
            style={theme.tipStyle}
          >
            <p
              className='text-xs leading-relaxed'
              style={{ color: theme.tipTextColor }}
            >
              <span className='font-bold'>{t('温馨提示')}:</span>{' '}
              {t(
                '支付完成后请勿关闭页面，系统将自动确认支付结果。如遇问题请联系客服。',
              )}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ScanPayModal;
