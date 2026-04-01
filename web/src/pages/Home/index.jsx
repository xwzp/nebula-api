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

import React, { useContext, useEffect, useState, useRef, useCallback } from 'react';
import { API, showError, copy, showSuccess } from '../../helpers';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { StatusContext } from '../../context/Status';
import { useActualTheme } from '../../context/Theme';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { motion, AnimatePresence } from 'motion/react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import {
  Copy,
  Check,
  Sparkles,
  ChevronRight,
  TerminalSquare,
  Box,
  Globe,
  FileText,
  DollarSign,
  RefreshCcw,
  Map,
  MessageCircle,
  Bot,
  ShieldCheck,
  Users,
  Mail,
  Megaphone,
  Smartphone,
} from 'lucide-react';
import {
  Moonshot,
  OpenAI,
  XAI,
  Zhipu,
  Volcengine,
  Cohere,
  Claude,
  Gemini,
  Suno,
  Minimax,
  Wenxin,
  Spark,
  Qingyan,
  DeepSeek,
  Qwen,
  Midjourney,
  Grok,
  AzureAI,
  Hunyuan,
  Xinference,
} from '@lobehub/icons';
import NoticeModal from '../../components/layout/NoticeModal';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

// ============ Subscription Plans Data ============
const SUBSCRIPTION_PLANS = [
  {
    name: 'Plus',
    priceMonthly: 100,
    priceYearly: 80,
    descKey: '适合个人开发者与小微项目',
    features: [
      '500K Tokens/月',
      '普通节点响应',
      '基础售后支持',
      '社区交流群',
    ],
  },
  {
    name: 'Pro',
    priceMonthly: 200,
    priceYearly: 160,
    descKey: '适合专业团队与中型业务',
    features: [
      '1.5M Tokens/月',
      '高速专线节点',
      '优先工单响应',
      '一对一技术指导',
      '请求并发升级',
    ],
    highlight: true,
  },
  {
    name: 'Max',
    priceMonthly: 500,
    priceYearly: 400,
    descKey: '适合企业级核心业务与大规模应用',
    features: [
      '5M Tokens/月',
      '企业专属极速节点',
      '7x24小时VIP支持',
      '私有化部署咨询',
      '无限制并发',
    ],
  },
];

const PAY_AS_YOU_GO_PLANS = [
  { price: 100, amount: 100, discount: null, tag: null },
  { price: 200, amount: 200, discount: null, tag: null },
  { price: 500, amount: 500, discount: 0.96, tag: '最推荐' },
  { price: 1000, amount: 1000, discount: 0.95, tag: '最具性价比' },
];

// ============ TOC Sections ============
const SECTIONS = [
  { id: 'hero', label: '首页概览' },
  { id: 'pricing-sub', label: '订阅方案' },
  { id: 'pricing-payg', label: '按量付费' },
  { id: 'quick-config', label: '极速配置' },
  { id: 'industry', label: '行业案例' },
];

// ============ TOC Navigation Component ============
const TocNavigation = ({ activeSection, onNavigate }) => {
  return (
    <div className='hidden lg:flex fixed right-8 top-1/2 -translate-y-1/2 z-50 flex-col py-5 px-3 bg-white/60 dark:bg-zinc-800/60 backdrop-blur-xl border border-slate-200/50 dark:border-zinc-700/50 shadow-xl shadow-slate-200/20 dark:shadow-black/20 rounded-2xl gap-2'>
      <div className='text-[10px] font-black text-slate-400 dark:text-zinc-500 uppercase tracking-widest mb-1 px-3'>
        本页导航
      </div>
      <div className='relative flex flex-col'>
        {SECTIONS.map((section) => {
          const isActive = activeSection === section.id;
          return (
            <button
              key={section.id}
              onClick={() => onNavigate(section.id)}
              className={cn(
                'relative z-10 px-4 py-2 text-sm font-medium transition-colors duration-300 text-left outline-none rounded-xl whitespace-nowrap group',
                isActive
                  ? 'text-purple-700 dark:text-purple-400'
                  : 'text-slate-500 dark:text-zinc-400 hover:text-slate-900 dark:hover:text-white',
              )}
            >
              {isActive && (
                <motion.div
                  layoutId='toc-indicator'
                  className='absolute inset-0 bg-white dark:bg-zinc-700 rounded-xl shadow-sm border border-slate-200/60 dark:border-zinc-600/60 z-[-1]'
                  transition={{ type: 'spring', stiffness: 300, damping: 30 }}
                />
              )}
              <span className='relative z-10 flex items-center gap-2'>
                <div
                  className={cn(
                    'w-1.5 h-1.5 rounded-full transition-all duration-300',
                    isActive
                      ? 'bg-purple-500 shadow-[0_0_8px_rgba(168,85,247,0.8)]'
                      : 'bg-slate-300 dark:bg-zinc-500 group-hover:bg-slate-400',
                  )}
                />
                {section.label}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
};

// ============ Hero Section ============
const HeroSection = ({ serverAddress, onCopy, copied, t }) => {
  return (
    <section
      id='hero'
      className='md:min-h-[calc(100vh-64px)] w-full flex items-center justify-center px-4 md:px-6 pt-24 pb-12 md:pt-0 md:pb-0'
    >
      <div className='max-w-5xl mx-auto text-center flex flex-col items-center'>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.7, ease: 'easeOut' }}
        className='mb-6 inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-white dark:bg-zinc-800 border border-slate-200 dark:border-zinc-700 shadow-sm text-sm font-medium text-slate-600 dark:text-zinc-300'
      >
        <Sparkles className='w-4 h-4 text-purple-500' />
        {t('全新极速体验，更聪明的模型网关')}
      </motion.div>

      <motion.h1
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.7, delay: 0.1, ease: 'easeOut' }}
        className='text-3xl sm:text-4xl md:text-6xl lg:text-7xl font-extrabold tracking-tight mb-6 md:mb-8 leading-tight text-slate-900 dark:text-white'
      >
        {t('原生的支持')} <br className='hidden md:block' />
        <span className='text-transparent bg-clip-text bg-gradient-to-r from-pink-500 via-purple-500 to-blue-500 animate-gradient-x'>
          {t('极致生产工具链')}
        </span>
      </motion.h1>

      <motion.p
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.7, delay: 0.2, ease: 'easeOut' }}
        className='text-base md:text-xl text-slate-500 dark:text-zinc-400 mb-8 md:mb-12 max-w-2xl px-2'
      >
        {t(
          '深度优化 API 路由，确保在 CLI 环境下依然拥有流畅的流式交互体验。开箱即用、价格实惠、专业运营，让开发者只关注代码本身。',
        )}
      </motion.p>

      {/* BASE_URL Copy Box */}
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.7, delay: 0.3, ease: 'easeOut' }}
        className='w-full max-w-2xl bg-white dark:bg-zinc-800 rounded-2xl p-2 border border-slate-200/60 dark:border-zinc-700/60 shadow-[0_8px_30px_rgb(0,0,0,0.04)] dark:shadow-[0_8px_30px_rgb(0,0,0,0.2)] flex flex-col sm:flex-row items-center gap-2 relative overflow-hidden group'
      >
        <div className='absolute inset-0 bg-gradient-to-r from-pink-500/10 via-purple-500/10 to-blue-500/10 opacity-0 group-hover:opacity-100 transition-opacity duration-500' />

        <div className='flex-1 flex items-center gap-3 px-4 py-3 bg-slate-50/50 dark:bg-zinc-900/50 rounded-xl w-full relative z-10 font-mono text-sm text-slate-600 dark:text-zinc-300'>
          <span className='text-slate-400 dark:text-zinc-500 select-none'>
            BASE_URL
          </span>
          <div className='flex-1 overflow-x-auto whitespace-nowrap scrollbar-hide'>
            {serverAddress}{' '}
            <span className='text-purple-500 ml-1'>/v1/models</span>
          </div>
        </div>

        <button
          onClick={onCopy}
          className='px-6 py-3 bg-slate-900 dark:bg-white hover:bg-slate-800 dark:hover:bg-zinc-200 text-white dark:text-slate-900 rounded-xl font-medium transition-all active:scale-95 flex items-center gap-2 w-full sm:w-auto justify-center z-10'
        >
          {copied ? (
            <Check className='w-4 h-4' />
          ) : (
            <Copy className='w-4 h-4' />
          )}
          {copied ? t('已复制') : t('复制终端')}
        </button>
      </motion.div>

      {/* Model Providers */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 1, delay: 0.5 }}
        className='mt-12 md:mt-20 w-full overflow-hidden'
      >
        <p className='text-xs md:text-sm font-medium text-slate-400 dark:text-zinc-500 mb-6 uppercase tracking-wider'>
          {t('支持全球顶级大模型')}
        </p>
        <div className='flex flex-wrap justify-center gap-6 md:gap-16'>
          {[
            { Icon: Gemini, name: 'Gemini', color: true },
            { Icon: OpenAI, name: 'OpenAI', color: false },
            { Icon: Claude, name: 'Anthropic', color: true },
            { Icon: DeepSeek, name: 'DeepSeek', color: true },
          ].map(({ Icon, name, color }, i) => (
            <div
              key={i}
              className='flex flex-col items-center gap-2 md:gap-3 text-slate-600 dark:text-zinc-400 hover:text-purple-600 dark:hover:text-purple-400 transition-colors group w-[60px] md:w-auto'
            >
              <div className='w-8 h-8 md:w-12 md:h-12 flex items-center justify-center transition-transform group-hover:scale-110 duration-300'>
                {color && Icon.Color ? <Icon.Color size={44} /> : <Icon size={44} />}
              </div>
              <span className='text-xs md:text-sm font-semibold tracking-wide'>{name}</span>
            </div>
          ))}
          <div className='flex flex-col items-center gap-2 md:gap-3 text-slate-600 dark:text-zinc-400 w-[60px] md:w-auto'>
            <div className='w-8 h-8 md:w-12 md:h-12 flex items-center justify-center font-black text-xl md:text-2xl'>
              30+
            </div>
            <span className='text-xs md:text-sm font-semibold tracking-wide'>More</span>
          </div>
        </div>
      </motion.div>
      </div>
    </section>
  );
};

// ============ Subscription Pricing Section ============
const SubscriptionSection = ({ t }) => {
  const [subCycle, setSubCycle] = useState('year');

  return (
    <section id='pricing-sub' className='md:min-h-[calc(100vh-64px)] w-full flex items-center justify-center px-4 md:px-6 py-12'>
      <div className='w-full max-w-6xl mx-auto'>
      <div className='text-center mb-10 md:mb-12 px-2'>
        <h2 className='text-2xl md:text-4xl font-bold mb-4 text-slate-900 dark:text-white'>
          {t('简单透明的计费方式 — 订阅方案')}
        </h2>
        <p className='text-slate-500 dark:text-zinc-400 text-base md:text-lg'>
          {t('包月或包年订阅，享受更稳定、更独享的专属资源与权益')}
        </p>
      </div>

      {/* Month/Year Toggle */}
      <div className='flex justify-center mb-10 md:mb-12'>
        <div className='bg-slate-100/80 dark:bg-zinc-800/80 backdrop-blur p-1 rounded-full inline-flex max-w-full overflow-x-auto scrollbar-hide'>
          <button
            onClick={() => setSubCycle('month')}
            className={cn(
              'px-6 md:px-8 py-2.5 rounded-full text-xs md:text-sm font-medium transition-all duration-300 relative whitespace-nowrap',
              subCycle === 'month'
                ? 'text-slate-900 dark:text-white shadow-sm'
                : 'text-slate-500 dark:text-zinc-400 hover:text-slate-700',
            )}
          >
            {subCycle === 'month' && (
              <motion.div
                layoutId='sub-tab-bg'
                className='absolute inset-0 bg-white dark:bg-zinc-700 rounded-full -z-10 shadow-sm'
              />
            )}
            {t('按月支付')}
          </button>
          <button
            onClick={() => setSubCycle('year')}
            className={cn(
              'px-6 md:px-8 py-2.5 rounded-full text-xs md:text-sm font-medium transition-all duration-300 relative whitespace-nowrap',
              subCycle === 'year'
                ? 'text-slate-900 dark:text-white shadow-sm'
                : 'text-slate-500 dark:text-zinc-400 hover:text-slate-700',
            )}
          >
            {subCycle === 'year' && (
              <motion.div
                layoutId='sub-tab-bg'
                className='absolute inset-0 bg-white dark:bg-zinc-700 rounded-full -z-10 shadow-sm'
              />
            )}
            {t('按年支付')}{' '}
            <span className='ml-1 text-xs text-purple-600 dark:text-purple-400 font-bold bg-purple-100 dark:bg-purple-900/40 px-1.5 py-0.5 rounded-md'>
              8折
            </span>
          </button>
        </div>
      </div>

      {/* Plan Cards */}
      <div className='grid grid-cols-1 md:grid-cols-3 gap-6 md:gap-8 max-w-5xl mx-auto px-2'>
        {SUBSCRIPTION_PLANS.map((plan, idx) => {
          const currentPrice =
            subCycle === 'year' ? plan.priceYearly : plan.priceMonthly;
          return (
            <motion.div
              key={idx}
              whileHover={{ y: -5 }}
              className={cn(
                'bg-white dark:bg-zinc-800 rounded-3xl p-6 md:p-8 border transition-all duration-300 flex flex-col relative',
                plan.highlight
                  ? 'border-purple-500/30 shadow-2xl shadow-purple-500/10 md:scale-105 z-10'
                  : 'border-slate-200/60 dark:border-zinc-700/60 shadow-lg shadow-slate-200/20 dark:shadow-black/20',
              )}
            >
              {plan.highlight && (
                <div className='absolute top-0 left-0 right-0 h-1.5 bg-gradient-to-r from-pink-500 via-purple-500 to-blue-500 rounded-t-3xl' />
              )}

              <h3 className='text-xl font-bold mb-2 text-slate-900 dark:text-white'>
                {plan.name}
              </h3>
              <p className='text-sm text-slate-500 dark:text-zinc-400 min-h-[40px] mb-6'>
                {t(plan.descKey)}
              </p>

              <div className='flex items-baseline gap-1 mb-2'>
                <span className='text-4xl font-extrabold text-slate-900 dark:text-white'>
                  ¥{currentPrice}
                </span>
                <span className='text-slate-500 dark:text-zinc-400'>
                  /{t('月')}
                </span>
              </div>

              <div className='h-6 mb-6 pb-8 border-b border-slate-100 dark:border-zinc-700 flex items-center'>
                {subCycle === 'year' && (
                  <span className='text-sm text-slate-400 line-through'>
                    {t('原价')} ¥{plan.priceMonthly}/{t('月')}
                  </span>
                )}
              </div>

              <ul className='space-y-4 mb-8 flex-1'>
                {plan.features.map((feat, i) => (
                  <li
                    key={i}
                    className='flex items-start gap-3 text-sm text-slate-700 dark:text-zinc-300'
                  >
                    <Check
                      className={cn(
                        'w-5 h-5 shrink-0',
                        plan.highlight
                          ? 'text-purple-500'
                          : 'text-slate-400 dark:text-zinc-500',
                      )}
                    />
                    <span>{t(feat)}</span>
                  </li>
                ))}
              </ul>

              <button
                className={cn(
                  'w-full py-3.5 rounded-xl font-medium transition-colors text-sm',
                  plan.highlight
                    ? 'bg-gradient-to-r from-purple-600 to-blue-600 text-white hover:opacity-90 shadow-md shadow-purple-500/20'
                    : 'bg-slate-100 dark:bg-zinc-700 text-slate-900 dark:text-white hover:bg-slate-200 dark:hover:bg-zinc-600',
                )}
              >
                {t('订阅')} {plan.name}
              </button>
            </motion.div>
          );
        })}
      </div>
      </div>
    </section>
  );
};

// ============ Pay-As-You-Go Section ============
const PayAsYouGoSection = ({ t }) => {
  return (
    <section
      id='pricing-payg'
      className='md:min-h-[calc(100vh-64px)] w-full flex items-center justify-center px-4 md:px-6 py-12'
    >
      <div className='w-full max-w-6xl mx-auto'>
      <div className='text-center mb-10 md:mb-12 px-2'>
        <h2 className='text-2xl md:text-4xl font-bold mb-4 text-slate-900 dark:text-white'>
          {t('按量付费充值')}
        </h2>
        <p className='text-slate-500 dark:text-zinc-400 text-base md:text-lg'>
          {t('随充随用，支持所有模型无缝切换调用，额度永久有效')}
        </p>
      </div>

      <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6 px-2'>
        {PAY_AS_YOU_GO_PLANS.map((plan, idx) => {
          const actualPrice = plan.discount
            ? plan.price * plan.discount
            : plan.price;
          return (
            <motion.div
              key={idx}
              whileHover={{ y: -5 }}
              className={cn(
                'relative bg-white dark:bg-zinc-800 rounded-3xl p-5 md:p-6 border transition-all duration-300 flex flex-col',
                plan.tag === '最推荐'
                  ? 'border-purple-500/50 shadow-xl shadow-purple-500/10 ring-1 ring-purple-500/20'
                  : plan.tag === '最具性价比'
                    ? 'border-pink-500/50 shadow-xl shadow-pink-500/10 ring-1 ring-pink-500/20'
                    : 'border-slate-200/60 dark:border-zinc-700/60 shadow-lg shadow-slate-200/20 dark:shadow-black/20 hover:border-slate-300',
              )}
            >
              {plan.tag && (
                <div
                  className={cn(
                    'absolute -top-3 left-1/2 -translate-x-1/2 px-4 py-1 rounded-full text-xs font-bold shadow-sm whitespace-nowrap text-white',
                    plan.tag === '最推荐'
                      ? 'bg-gradient-to-r from-purple-500 to-blue-500'
                      : 'bg-gradient-to-r from-pink-500 to-purple-500',
                  )}
                >
                  {t(plan.tag)}
                </div>
              )}

              <div className='mb-8 mt-4 text-center'>
                <div className='text-slate-500 dark:text-zinc-400 font-medium mb-2'>
                  {t('充值额度')}
                </div>
                <div className='text-4xl font-extrabold flex items-center justify-center gap-1 text-slate-900 dark:text-white'>
                  <span className='text-2xl'>¥</span>
                  {plan.amount}
                </div>
              </div>

              <div className='flex-1'>
                <div className='bg-slate-50 dark:bg-zinc-900 rounded-xl p-4 mb-6'>
                  <div className='flex justify-between items-center mb-2'>
                    <span className='text-sm text-slate-500 dark:text-zinc-400'>
                      {t('实际支付')}
                    </span>
                    {plan.discount && (
                      <span className='text-xs font-bold text-emerald-600 dark:text-emerald-400 bg-emerald-100 dark:bg-emerald-900/40 px-2 py-0.5 rounded-md'>
                        {plan.discount === 0.96
                          ? '96折'
                          : plan.discount === 0.95
                            ? '95折'
                            : `${plan.discount * 10}折`}
                      </span>
                    )}
                  </div>
                  <div className='flex items-end gap-2'>
                    <span className='text-2xl font-bold text-slate-900 dark:text-white'>
                      ¥{actualPrice.toFixed(0)}
                    </span>
                    {plan.discount && (
                      <span className='text-sm text-slate-400 line-through mb-1'>
                        ¥{plan.price}
                      </span>
                    )}
                  </div>
                </div>

                <ul className='space-y-3 mb-8'>
                  <li className='flex items-center gap-2 text-sm text-slate-600 dark:text-zinc-300'>
                    <Check className='w-4 h-4 text-purple-500' />{' '}
                    {t('额度永久有效')}
                  </li>
                  <li className='flex items-center gap-2 text-sm text-slate-600 dark:text-zinc-300'>
                    <Check className='w-4 h-4 text-purple-500' />{' '}
                    {t('支持所有模型调用')}
                  </li>
                  <li className='flex items-center gap-2 text-sm text-slate-600 dark:text-zinc-300'>
                    <Check className='w-4 h-4 text-purple-500' />{' '}
                    {t('包含增值税发票')}
                  </li>
                </ul>
              </div>

              <button
                className={cn(
                  'w-full py-3 rounded-xl font-medium transition-colors text-sm',
                  plan.tag
                    ? 'bg-slate-900 dark:bg-white text-white dark:text-slate-900 hover:bg-slate-800 dark:hover:bg-zinc-200'
                    : 'bg-slate-100 dark:bg-zinc-700 text-slate-900 dark:text-white hover:bg-slate-200 dark:hover:bg-zinc-600',
                )}
              >
                {t('立即充值')}
              </button>
            </motion.div>
          );
        })}
      </div>
      </div>
    </section>
  );
};

// ============ Quick Config / Terminal Section ============
const QuickConfigSection = ({ t }) => {
  return (
    <section id='quick-config' className='md:min-h-[calc(100vh-64px)] w-full flex items-center justify-center px-4 md:px-6 py-12 md:py-0'>
      <div className='w-full max-w-6xl mx-auto'>
      <div className='grid md:grid-cols-2 gap-8 md:gap-12 items-center px-2'>
        <motion.div
          initial={{ opacity: 0, x: -30 }}
          whileInView={{ opacity: 1, x: 0 }}
          viewport={{ once: true }}
          className='space-y-6 md:space-y-8'
        >
          <div className='inline-flex items-center gap-2 text-purple-600 dark:text-purple-400 font-medium text-xs md:text-sm tracking-widest uppercase'>
            <TerminalSquare className='w-4 h-4' /> Terminal Ready
          </div>
          <h2 className='text-3xl md:text-4xl font-bold text-slate-900 dark:text-white'>
            {t('1秒配置，光速开始编码')}
          </h2>
          <p className='text-slate-500 dark:text-zinc-400 text-base md:text-lg leading-relaxed'>
            {t(
              '兼容所有标准的 OpenAI 及 Anthropic API 格式客户端。支持 Cursor, Cline, AutoGPT, LangChain 等众多主流工具与框架。',
            )}
          </p>
          <div className='flex flex-col sm:flex-row gap-3 md:gap-4'>
            <Link to='/console' className='w-full sm:w-auto'>
              <button className='px-6 py-3 rounded-full bg-slate-900 dark:bg-white text-white dark:text-slate-900 font-medium hover:bg-slate-800 dark:hover:bg-zinc-200 transition-colors flex items-center justify-center gap-2 w-full sm:w-auto'>
                {t('快速接入')} <ChevronRight className='w-4 h-4' />
              </button>
            </Link>
            <button className='px-6 py-3 rounded-full bg-white dark:bg-zinc-800 border border-slate-200 dark:border-zinc-700 text-slate-700 dark:text-zinc-300 font-medium hover:bg-slate-50 dark:hover:bg-zinc-700 transition-colors w-full sm:w-auto'>
              {t('浏览文档')}
            </button>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true }}
          className='bg-white/60 dark:bg-zinc-800/60 backdrop-blur-xl border border-white/40 dark:border-zinc-700/40 shadow-2xl rounded-2xl p-4 md:p-6 font-mono text-xs md:text-sm overflow-x-auto relative w-full'
        >
          <div className='flex gap-2 mb-6 border-b border-slate-100 dark:border-zinc-700 pb-4'>
            <div className='w-3 h-3 rounded-full bg-red-400' />
            <div className='w-3 h-3 rounded-full bg-amber-400' />
            <div className='w-3 h-3 rounded-full bg-emerald-400' />
          </div>
          <div className='space-y-3 text-slate-600 dark:text-zinc-300'>
            <div className='text-slate-400 dark:text-zinc-500'>
              # 1秒配置 CLAUDE CLI
            </div>
            <div>
              <span className='text-purple-500'>export</span>{' '}
              <span className='text-blue-500'>ANTHROPIC_BASE_URL</span>=
              <span className='text-emerald-500'>
                &quot;https://api.nebulatrip.com&quot;
              </span>
            </div>
            <div>
              <span className='text-purple-500'>export</span>{' '}
              <span className='text-blue-500'>ANTHROPIC_API_KEY</span>=
              <span className='text-emerald-500'>
                &quot;sk-nebula-...&quot;
              </span>
            </div>
            <br />
            <div className='text-slate-400 dark:text-zinc-500'>
              # 以光速开始对话
            </div>
            <div className='flex items-center gap-2'>
              <span className='text-pink-500'>$</span>
              <span>claude</span>
              <motion.div
                animate={{ opacity: [1, 0, 1] }}
                transition={{ duration: 1, repeat: Infinity }}
                className='w-2 h-4 bg-slate-400 dark:bg-zinc-500'
              />
            </div>
          </div>
        </motion.div>
      </div>
      </div>
    </section>
  );
};

// ============ Industry Solutions Bento Grid ============
const IndustrySection = ({ t }) => {
  return (
    <section id='industry' className='md:min-h-[calc(100vh-64px)] w-full flex items-center justify-center px-4 md:px-6 py-12 md:overflow-y-auto'>
      <div className='w-full max-w-6xl mx-auto'>
      {/* Section Header */}
      <div className='text-center mb-12 md:mb-20 flex flex-col items-center relative px-2'>
        <div className='absolute top-10 left-1/2 -translate-x-1/2 w-[150%] md:w-full max-w-2xl h-32 bg-gradient-to-r from-orange-500/20 via-pink-500/20 to-purple-500/20 blur-[60px] md:blur-[80px] -z-10 pointer-events-none' />

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className='flex items-center justify-center gap-3 md:gap-4 mb-6 md:mb-8'
        >
          <div className='h-[2px] w-8 md:w-24 bg-gradient-to-r from-transparent to-orange-400/50 rounded-full' />
          <div className='w-12 h-12 md:w-14 md:h-14 rounded-2xl bg-gradient-to-br from-orange-400 via-pink-500 to-purple-500 p-[2px] shadow-xl shadow-orange-500/20 shrink-0'>
            <div className='w-full h-full bg-white/90 dark:bg-zinc-800/90 backdrop-blur-sm rounded-[14px] flex items-center justify-center'>
              <Sparkles className='w-5 h-5 md:w-7 md:h-7 text-pink-500' />
            </div>
          </div>
          <div className='h-[2px] w-8 md:w-24 bg-gradient-to-l from-transparent to-purple-400/50 rounded-full' />
        </motion.div>

        <motion.h2
          initial={{ opacity: 0, y: 10 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className='text-3xl md:text-5xl font-black mb-4 md:mb-6 tracking-tight leading-tight flex flex-col md:flex-row items-center justify-center gap-1 md:gap-4'
        >
          <span className='text-transparent bg-clip-text bg-gradient-to-r from-orange-600 via-pink-500 to-purple-600 pb-1'>
            {t('行业落地案例')}
          </span>
          <span className='hidden md:block text-slate-300 dark:text-zinc-600 font-light pb-1'>
            |
          </span>
          <span className='text-slate-900 dark:text-white pb-1'>
            {t('组团社、地接社')}
          </span>
        </motion.h2>

        <motion.div
          initial={{ opacity: 0, y: 10 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className='text-xl md:text-4xl font-bold text-slate-800 dark:text-zinc-200 mb-6 md:mb-8 flex flex-wrap items-center justify-center gap-x-2 gap-y-3'
        >
          {t('进入人工智能时代，一定要养')}{' '}
          <span className='relative inline-block ml-0.5 md:ml-1 mt-1 md:mt-0'>
            <span className='relative z-10 text-white px-3 py-1 md:px-4 md:py-1.5 font-black tracking-wide whitespace-nowrap'>
              "星云龙虾"
            </span>
            <span className='absolute inset-0 bg-gradient-to-r from-orange-500 via-pink-500 to-purple-500 rounded-xl shadow-lg shadow-pink-500/30 -z-0 rotate-[-2deg] scale-[1.02] md:scale-105' />
          </span>
        </motion.div>

        <motion.p
          initial={{ opacity: 0, y: 10 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className='text-slate-600 dark:text-zinc-400 text-base md:text-xl max-w-3xl leading-relaxed mb-4 md:mb-6'
        >
          {t('基于最先进的')}{' '}
          <strong className='text-slate-900 dark:text-white'>
            Claude Opus {t('模型')} + {t('旅行社自有数据与行业经验')}
          </strong>
          {t('，教您从小白开始，')}
          <strong className='text-purple-600 dark:text-purple-400'>
            {t('一周内即可养成')}
          </strong>
          {t(
            '并搭建公司从产品运营、市场、行程定制、供应商管理等整体人工智能架构。',
          )}
        </motion.p>
      </div>

      {/* Bento Grid */}
      <div className='grid grid-cols-1 md:grid-cols-12 gap-4 md:gap-6 auto-rows-[minmax(180px,auto)] px-2 md:px-0'>
        {/* 行程定制 (Span 8) */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          whileHover={{ scale: 1.02 }}
          viewport={{ once: true }}
          className='md:col-span-8 bg-white dark:bg-zinc-800 rounded-3xl p-6 md:p-8 border border-slate-200/60 dark:border-zinc-700/60 shadow-xl shadow-slate-200/40 dark:shadow-black/20 relative overflow-hidden group hover:border-purple-500/30 transition-colors'
        >
          <div className='absolute top-0 right-0 p-6 md:p-8 opacity-5 group-hover:opacity-10 transition-opacity'>
            <Map className='w-24 h-24 md:w-32 md:h-32' />
          </div>
          <div className='relative z-10 h-full flex flex-col justify-center'>
            <div className='flex items-center gap-3 mb-5 md:mb-6'>
              <div className='w-10 h-10 rounded-xl bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 flex items-center justify-center'>
                <Map className='w-5 h-5' />
              </div>
              <h3 className='text-xl md:text-2xl font-bold text-slate-900 dark:text-white'>
                {t('超级行程定制')}
              </h3>
            </div>
            <div className='grid sm:grid-cols-2 gap-3 md:gap-4'>
              <div className='bg-slate-50 dark:bg-zinc-900 rounded-2xl p-4 flex items-start gap-3'>
                <Globe className='w-5 h-5 text-blue-500 shrink-0 mt-0.5' />
                <p className='text-sm md:text-base text-slate-700 dark:text-zinc-300 font-medium'>
                  {t('全球任何国家')}{' '}
                  <span className='text-blue-600 dark:text-blue-400 font-bold'>
                    5{t('分钟')}
                  </span>{' '}
                  {t('之内定制一国或多国行程')}
                </p>
              </div>
              <div className='bg-slate-50 dark:bg-zinc-900 rounded-2xl p-4 flex items-start gap-3'>
                <FileText className='w-5 h-5 text-purple-500 shrink-0 mt-0.5' />
                <p className='text-sm md:text-base text-slate-700 dark:text-zinc-300 font-medium'>
                  <span className='text-purple-600 dark:text-purple-400 font-bold'>
                    5{t('分钟')}
                  </span>{' '}
                  {t('内自动生成精美路书与行程图文')}
                </p>
              </div>
              <div className='bg-slate-50 dark:bg-zinc-900 rounded-2xl p-4 flex items-start gap-3 sm:col-span-2'>
                <DollarSign className='w-5 h-5 text-emerald-500 shrink-0 mt-0.5' />
                <p className='text-sm md:text-base text-slate-700 dark:text-zinc-300 font-medium'>
                  <span className='text-emerald-600 dark:text-emerald-400 font-bold'>
                    5{t('分钟')}
                  </span>{' '}
                  {t('内精确核算各项成本：用车、酒店、门票、机票、活动')}
                </p>
              </div>
            </div>
          </div>
        </motion.div>

        {/* 全平台产品自动化 (Span 4) */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          whileHover={{ scale: 1.02 }}
          viewport={{ once: true }}
          className='md:col-span-4 bg-gradient-to-br from-slate-900 to-slate-800 dark:from-zinc-900 dark:to-zinc-800 text-white rounded-3xl p-6 md:p-8 shadow-xl relative overflow-hidden'
        >
          <div className='absolute -right-4 -bottom-4 opacity-10'>
            <RefreshCcw className='w-32 h-32 md:w-40 md:h-40' />
          </div>
          <div className='relative z-10 h-full flex flex-col'>
            <div className='flex items-center gap-3 mb-5 md:mb-6'>
              <div className='w-10 h-10 rounded-xl bg-white/10 flex items-center justify-center shrink-0'>
                <Box className='w-5 h-5 text-white' />
              </div>
              <h3 className='text-xl md:text-2xl font-bold'>
                {t('全平台产品自动化')}
              </h3>
            </div>
            <ul className='space-y-3 md:space-y-4 flex-1'>
              <li className='flex items-start gap-3'>
                <Check className='w-5 h-5 text-emerald-400 shrink-0 mt-0.5' />
                <span className='text-sm md:text-base text-slate-300'>
                  {t('自动录入飞猪、携程、途牛产品')}
                </span>
              </li>
              <li className='flex items-start gap-3'>
                <Check className='w-5 h-5 text-emerald-400 shrink-0 mt-0.5' />
                <span className='text-sm md:text-base text-slate-300'>
                  {t('自动录入独立站产品与行程')}
                </span>
              </li>
              <li className='flex items-start gap-3'>
                <Check className='w-5 h-5 text-emerald-400 shrink-0 mt-0.5' />
                <span className='text-sm md:text-base text-slate-300'>
                  {t('自动录入 Viator, GetYourGuide, Trip.com 等')}
                </span>
              </li>
            </ul>
          </div>
        </motion.div>

        {/* 智能客服与销售 (Span 5) */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          whileHover={{ scale: 1.02 }}
          viewport={{ once: true }}
          className='md:col-span-5 bg-white dark:bg-zinc-800 rounded-3xl p-6 md:p-8 border border-slate-200/60 dark:border-zinc-700/60 shadow-xl shadow-slate-200/40 dark:shadow-black/20 relative overflow-hidden group hover:border-pink-500/30 transition-colors'
        >
          <div className='flex items-center gap-3 mb-5 md:mb-6'>
            <div className='w-10 h-10 rounded-xl bg-pink-100 dark:bg-pink-900/30 text-pink-600 dark:text-pink-400 flex items-center justify-center shrink-0'>
              <MessageCircle className='w-5 h-5' />
            </div>
            <h3 className='text-xl md:text-2xl font-bold text-slate-900 dark:text-white'>
              {t('智能客服与销售')}
            </h3>
          </div>
          <div className='space-y-3 md:space-y-4'>
            <div className='flex items-start gap-3 p-3 rounded-xl hover:bg-slate-50 dark:hover:bg-zinc-700 transition-colors'>
              <Bot className='w-5 h-5 text-pink-500 shrink-0 mt-0.5' />
              <div>
                <div className='text-sm md:text-base font-bold text-slate-800 dark:text-zinc-200 mb-1'>
                  {t('多平台全自动回复')}
                </div>
                <p className='text-xs md:text-sm text-slate-500 dark:text-zinc-400'>
                  {t('毫秒级自动回复千牛、携程、小红书等线上平台咨询')}
                </p>
              </div>
            </div>
            <div className='flex items-start gap-3 p-3 rounded-xl hover:bg-slate-50 dark:hover:bg-zinc-700 transition-colors'>
              <Smartphone className='w-5 h-5 text-green-500 shrink-0 mt-0.5' />
              <div>
                <div className='text-sm md:text-base font-bold text-slate-800 dark:text-zinc-200 mb-1'>
                  {t('微信私域管家')}
                </div>
                <p className='text-xs md:text-sm text-slate-500 dark:text-zinc-400'>
                  {t('智能回复微信顾客咨询，24小时不间断')}
                </p>
              </div>
            </div>
            <div className='flex items-start gap-3 p-3 rounded-xl hover:bg-slate-50 dark:hover:bg-zinc-700 transition-colors'>
              <ShieldCheck className='w-5 h-5 text-blue-500 shrink-0 mt-0.5' />
              <div>
                <div className='text-sm md:text-base font-bold text-slate-800 dark:text-zinc-200 mb-1'>
                  {t('行中智能监控')}
                </div>
                <p className='text-xs md:text-sm text-slate-500 dark:text-zinc-400'>
                  {t('实时自动监控并回复微信客服服务群，保障服务质量')}
                </p>
              </div>
            </div>
          </div>
        </motion.div>

        {/* 内容与市场 (Span 4) */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          whileHover={{ scale: 1.02 }}
          viewport={{ once: true }}
          className='md:col-span-4 bg-gradient-to-br from-orange-100 to-amber-50 dark:from-orange-900/20 dark:to-amber-900/10 rounded-3xl p-6 md:p-8 border border-orange-200/50 dark:border-orange-800/30 shadow-xl shadow-orange-500/10 relative overflow-hidden'
        >
          <div className='absolute -right-6 -top-6 opacity-20'>
            <Megaphone className='w-24 h-24 md:w-32 md:h-32 text-orange-500' />
          </div>
          <div className='flex items-center gap-3 mb-5 md:mb-6 relative z-10'>
            <div className='w-10 h-10 rounded-xl bg-orange-500 text-white flex items-center justify-center shadow-lg shadow-orange-500/30 shrink-0'>
              <Megaphone className='w-5 h-5' />
            </div>
            <h3 className='text-xl md:text-2xl font-bold text-slate-900 dark:text-white'>
              {t('内容与市场')}
            </h3>
          </div>
          <div className='bg-white/60 dark:bg-zinc-800/60 backdrop-blur-md rounded-2xl p-4 md:p-5 border border-white dark:border-zinc-700 relative z-10'>
            <p className='text-sm md:text-base text-slate-700 dark:text-zinc-300 font-medium leading-relaxed'>
              {t('打破内容生产瓶颈，一键批量根据产品生成高质量的：')}
            </p>
            <div className='mt-4 flex flex-wrap gap-2'>
              <span className='px-2 py-1 md:px-3 md:py-1.5 bg-white dark:bg-zinc-800 rounded-lg shadow-sm text-xs md:text-sm font-bold text-red-500'>
                {t('小红书图文')}
              </span>
              <span className='px-2 py-1 md:px-3 md:py-1.5 bg-white dark:bg-zinc-800 rounded-lg shadow-sm text-xs md:text-sm font-bold text-green-600'>
                {t('公众号爆款')}
              </span>
              <span className='px-2 py-1 md:px-3 md:py-1.5 bg-white dark:bg-zinc-800 rounded-lg shadow-sm text-xs md:text-sm font-bold text-slate-900 dark:text-white'>
                {t('抖音文案')}
              </span>
            </div>
          </div>
        </motion.div>

        {/* 智能供应商 (Span 3) */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          whileHover={{ scale: 1.02 }}
          viewport={{ once: true }}
          className='md:col-span-3 bg-white dark:bg-zinc-800 rounded-3xl p-6 md:p-8 border border-slate-200/60 dark:border-zinc-700/60 shadow-xl shadow-slate-200/40 dark:shadow-black/20 relative overflow-hidden group hover:border-indigo-500/30 transition-colors'
        >
          <div className='flex items-center gap-3 mb-5 md:mb-6'>
            <div className='w-10 h-10 rounded-xl bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 flex items-center justify-center shrink-0'>
              <Users className='w-5 h-5' />
            </div>
            <h3 className='text-xl font-bold text-slate-900 dark:text-white'>
              {t('智能供应商')}
            </h3>
          </div>
          <p className='text-slate-600 dark:text-zinc-400 mb-5 md:mb-6 text-sm leading-relaxed'>
            {t('化身金牌采购，主动出击寻找全球优质资源。')}
          </p>
          <div className='space-y-3 mt-auto'>
            <div className='flex items-center gap-2 text-xs md:text-sm font-medium text-slate-700 dark:text-zinc-300 bg-slate-50 dark:bg-zinc-900 px-3 py-2.5 md:py-2 rounded-lg'>
              <Mail className='w-4 h-4 text-indigo-500 shrink-0' /> WhatsApp &
              Email {t('寻源')}
            </div>
            <div className='flex items-center gap-2 text-xs md:text-sm font-medium text-slate-700 dark:text-zinc-300 bg-slate-50 dark:bg-zinc-900 px-3 py-2.5 md:py-2 rounded-lg'>
              <Bot className='w-4 h-4 text-indigo-500 shrink-0' />{' '}
              {t('高效拟人全自动沟通')}
            </div>
            <div className='flex items-center gap-2 text-xs md:text-sm font-medium text-slate-700 dark:text-zinc-300 bg-slate-50 dark:bg-zinc-900 px-3 py-2.5 md:py-2 rounded-lg'>
              <DollarSign className='w-4 h-4 text-indigo-500 shrink-0' />{' '}
              {t('自动咨询与比对报价')}
            </div>
          </div>
        </motion.div>
      </div>
      </div>
    </section>
  );
};

// ============ Main Home Component ============
const Home = () => {
  const { t, i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const actualTheme = useActualTheme();
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [noticeVisible, setNoticeVisible] = useState(false);
  const [copied, setCopied] = useState(false);
  const [activeSection, setActiveSection] = useState('hero');
  const isMobile = useIsMobile();
  const snapContainerRef = useRef(null);

  const serverAddress =
    statusState?.status?.server_address || `${window.location.origin}`;

  // ---- Custom homepage content loading ----
  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);

      if (data.startsWith('https://')) {
        const iframe = document.querySelector('iframe');
        if (iframe) {
          iframe.onload = () => {
            iframe.contentWindow.postMessage({ themeMode: actualTheme }, '*');
            iframe.contentWindow.postMessage({ lang: i18n.language }, '*');
          };
        }
      }
    } else {
      showError(message);
      setHomePageContent('加载首页内容失败...');
    }
    setHomePageContentLoaded(true);
  };

  // ---- Copy handler ----
  const handleCopy = async () => {
    const ok = await copy(serverAddress);
    if (ok) {
      showSuccess(t('已复制到剪切板'));
    }
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  // ---- Notice check ----
  useEffect(() => {
    const checkNoticeAndShow = async () => {
      const lastCloseDate = localStorage.getItem('notice_close_date');
      const today = new Date().toDateString();
      if (lastCloseDate !== today) {
        try {
          const res = await API.get('/api/notice');
          const { success, data } = res.data;
          if (success && data && data.trim() !== '') {
            setNoticeVisible(true);
          }
        } catch (error) {
          console.error('获取公告失败:', error);
        }
      }
    };
    checkNoticeAndShow();
  }, []);

  useEffect(() => {
    displayHomePageContent().then();
  }, []);

  // ---- TOC navigate: scroll snap container to target section ----
  const handleTocNavigate = useCallback((sectionId) => {
    const el = document.getElementById(sectionId);
    const container = snapContainerRef.current;
    if (!el || !container) return;
    el.scrollIntoView({ behavior: 'smooth', block: 'start' });
    setActiveSection(sectionId);
  }, []);

  // ---- Scroll spy: 监听 snap container 的 scroll 事件 ----
  useEffect(() => {
    const container = snapContainerRef.current;
    if (!container) return;

    const updateActiveSection = () => {
      const containerRect = container.getBoundingClientRect();
      const triggerY = containerRect.top + containerRect.height * 0.3;

      for (let i = SECTIONS.length - 1; i >= 0; i--) {
        const section = document.getElementById(SECTIONS[i].id);
        if (!section) continue;
        if (section.getBoundingClientRect().top <= triggerY) {
          setActiveSection(SECTIONS[i].id);
          break;
        }
      }
    };

    container.addEventListener('scroll', updateActiveSection, { passive: true });
    updateActiveSection();
    return () => container.removeEventListener('scroll', updateActiveSection);
  }, [homePageContentLoaded, homePageContent]);

  return (
    <div className='w-full overflow-x-hidden'>
      <NoticeModal
        visible={noticeVisible}
        onClose={() => setNoticeVisible(false)}
        isMobile={isMobile}
      />

      {homePageContentLoaded && homePageContent === '' ? (
        <>
          {/* TOC Navigation — fixed, outside scroll container */}
          <TocNavigation activeSection={activeSection} onNavigate={handleTocNavigate} />

          {/* Scroll Snap Container */}
          <div
            ref={snapContainerRef}
            className='md:h-[calc(100vh-64px)] md:overflow-y-auto scroll-snap-container bg-[#FAFAFA] dark:bg-zinc-950 text-slate-900 dark:text-white font-sans selection:bg-purple-500/30 relative'
          >
            {/* Background Grid */}
            <div className='fixed inset-0 z-0 pointer-events-none home-bg-grid' />

            <HeroSection
              serverAddress={serverAddress}
              onCopy={handleCopy}
              copied={copied}
              t={t}
            />
            <SubscriptionSection t={t} />
            <PayAsYouGoSection t={t} />
            <QuickConfigSection t={t} />
            <IndustrySection t={t} />

            {/* Footer — desktop only (mobile uses PageLayout footer) */}
            <footer className='hidden md:block w-full bg-white dark:bg-zinc-900 border-t border-slate-200 dark:border-zinc-800 py-8 px-6 text-sm text-slate-500 dark:text-zinc-400'>
              <div className='max-w-6xl mx-auto flex flex-col md:flex-row justify-between items-center gap-4'>
                <div>&copy; {new Date().getFullYear()} Nebula API. {t('版权所有')}</div>
                <div>{t('设计与开发由')} <a href='https://github.com/xwzp/nebula-api' className='text-purple-600 dark:text-purple-400 font-medium hover:underline'>New API</a></div>
              </div>
            </footer>
          </div>
        </>
      ) : (
        <div className='overflow-x-hidden w-full'>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              className='w-full h-screen border-none'
            />
          ) : (
            <div
              className='mt-[60px]'
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            />
          )}
        </div>
      )}
    </div>
  );
};

export default Home;
