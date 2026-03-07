/**
 * @file 订阅页面常量定义
 */

import type { PricingPlan, PaymentMethod } from './interface';
import AlipayIcon from '@/assets/subscribe/alipay.svg';
import BankpayIcon from '@/assets/subscribe/bankpay.svg';
import WechatpayIcon from '@/assets/subscribe/wechatpay.svg';

// 流量套餐配置生成函数，使用国际化
export const getPricingPlans = (t: (key: string) => string): PricingPlan[] => [
    {
        id: 'traffic-package-1',
        title: t('pricingPlans.trafficPackage1.title'),
        price: 25,
        originalPrice: 50,
        type: 1,
    },
    {
        id: 'traffic-package-2',
        title: t('pricingPlans.trafficPackage2.title'),
        price: 98,
        originalPrice: 200,
        type: 2,
    },
    {
        id: 'traffic-package-3',
        title: t('pricingPlans.trafficPackage3.title'),
        price: 248,
        originalPrice: 500,
        type: 3,
    },
];

// 支付方式配置生成函数，使用国际化
export const getPaymentMethods = (t: (key: string) => string): PaymentMethod[] => [
    {
        id: 'alipay',
        name: t('paymentMethods.alipay.name'),
        icon: AlipayIcon,
        description: t('paymentMethods.alipay.description'),
    },
    {
        id: 'wechat',
        name: t('paymentMethods.wechat.name'),
        icon: WechatpayIcon,
        description: t('paymentMethods.wechat.description'),
    },
    {
        id: 'unionpay',
        name: t('paymentMethods.unionpay.name'),
        icon: BankpayIcon,
        description: t('paymentMethods.unionpay.description'),
    },
];

export const PAY_TYPE = {
    alipay: 'alipay',
    wechat: 'wechat',
    unionpay: 'unionpay',
} as const;

export type TPayType = keyof typeof PAY_TYPE;

export const ORDER_STATUS = {
    COMPLETED: 'completed',
    FAILED: 'failed',
    CANCELLED: 'cancelled',
} as const;

export const POLLING_INTERVAL = 3000 as const;

export const CMB = 'CMB';

export const PRICING_PLAN = {
    'traffic-package-1': '流量套餐1',
    'traffic-package-2': '流量套餐2',
    'traffic-package-3': '流量套餐3',
} as const;

/**
 * 首次购买营销规则ID
 * 对应后端quota_marketing_rules表的rule_type字段
 * 注意：如果后端数据库中该ID发生变化，必须同步更新此常量
 */
export const FIRST_PURCHASE_ID = 1;
