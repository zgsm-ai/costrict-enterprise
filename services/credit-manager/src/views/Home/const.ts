/**
 * @file home const
 */

import type { Router } from 'vue-router';
import type { PricingPlan } from './interface';
import type { QuotaTypeWithMarketingRules } from '@/api/bos/quota.bo';
import { FIRST_PURCHASE_ID } from '../Subscribe/const';

export const BING_TYPE = {
    github: 'github',
    sms: 'sms',
} as const;

export const STATUS = {
    success: 'success',
    error: 'error',
    undefined: undefined,
} as const;

export const TRANSFER_TYPE = {
    out: 'out',
    in: 'in',
} as const;

export const TRANSFER_IN_STATUS = {
    success: 'SUCCESS',
    partialSuccess: 'PARTIAL_SUCCESS',
    alreadyRedeemed: 'ALREADY_REDEEMED',
} as const;

// 发票状态值常量
export const INVOICE_STATUS = {
    PENDING: 0, // 待开票
    PROCESSING: 1, // 开票中
    COMPLETED: 2, // 已开票
    CANCELLED: 3, // 已作废
} as const;

// 发票相关常量生成函数，使用国际化
export const getInvoiceConstants = (t: (key: string) => string) => ({
    // 抬头类型
    HEADER_TYPE: {
        COMPANY: 'company',
        PERSONAL: 'personal',
    },
    // 发票类型
    INVOICE_TYPE: {
        VAT: 'vat',
        SPECIAL: 'special',
    },
    // 发票状态
    STATUS: INVOICE_STATUS,
    // 发票状态颜色
    STATUS_COLORS: {
        [INVOICE_STATUS.PENDING]: '#FFC300', // 待开票 - 橙色
        [INVOICE_STATUS.PROCESSING]: '#4A90E2', // 开票中 - 蓝色
        [INVOICE_STATUS.COMPLETED]: '#52C41A', // 已开票 - 绿色
        [INVOICE_STATUS.CANCELLED]: '#8C8C8C', // 已作废 - 灰色
    } as Record<number, string>,
    // 发票状态文本
    STATUS_TEXT: {
        [INVOICE_STATUS.PENDING]: t('invoiceStatus.pending'),
        [INVOICE_STATUS.PROCESSING]: t('invoiceStatus.processing'),
        [INVOICE_STATUS.COMPLETED]: t('invoiceStatus.completed'),
        [INVOICE_STATUS.CANCELLED]: t('invoiceStatus.cancelled'),
    } as Record<number, string>,
});

// 套餐配置生成函数，使用国际化 - 只返回免费套餐
export const getPricingPlans = (t: (key: string) => string): PricingPlan[] => [
    {
        title: t('pricingPlans.personalFree.title'),
        price: 0,
        description: t('pricingPlans.personalFree.description'),
        buttonText: t('pricingPlans.personalFree.buttonText'),
        buttonType: 'download',
        showTrafficLabel: false,
        features: [
            {
                text: t('pricingPlans.personalFree.features.newUserCredits'),
                available: true,
            },
            {
                text: t('pricingPlans.personalFree.features.freeCodeCompletion'),
                available: true,
            },
            {
                text: t('pricingPlans.personalFree.features.freeCodeReview'),
                available: true,
            },
            {
                text: t('pricingPlans.personalFree.features.advancedModelUnavailable'),
                available: false,
            },
        ],
        clickEvent() {
            window.open('https://costrict.ai/download');
        },
    },
];

// 根据API返回的配额类型数据生成套餐列表
export const generatePricingPlansFromAPI = (
    t: (key: string, params?: { [key: string]: unknown }) => string,
    router: Router,
    quotaTypes: QuotaTypeWithMarketingRules[],
): PricingPlan[] => {
    // 基础免费套餐
    const freePlan: PricingPlan = {
        title: t('pricingPlans.personalFree.title'),
        price: 0,
        description: t('pricingPlans.personalFree.description'),
        buttonText: t('pricingPlans.personalFree.buttonText'),
        buttonType: 'download',
        showTrafficLabel: false,
        features: [
            {
                text: t('pricingPlans.personalFree.features.newUserCredits'),
                available: true,
            },
            {
                text: t('pricingPlans.personalFree.features.freeCodeCompletion'),
                available: true,
            },
            {
                text: t('pricingPlans.personalFree.features.freeCodeReview'),
                available: true,
            },
            {
                text: t('pricingPlans.personalFree.features.advancedModelUnavailable'),
                available: false,
            },
        ],
        clickEvent() {
            window.open('https://costrict.ai/download');
        },
    };

    // 从API数据生成付费套餐
    const paidPlans = quotaTypes.map((quotaType, index) => {
        return {
            title:
                quotaType.display_name || `${t('pricingPlans.defaultPackageTitle')} ${index + 1}`,
            price: quotaType.amount,
            originalPrice:
                quotaType.original_amount && quotaType.original_amount > quotaType.amount
                    ? quotaType.original_amount
                    : undefined,
            description: t('pricingPlans.creditUnit'),
            buttonText: t('pricingPlans.purchaseButtonText'),
            buttonType: 'purchase' as const,
            showTrafficLabel: true,
            isFirstPurchase: quotaType.quota_marketing_rules_id === FIRST_PURCHASE_ID,
            features: [
                {
                    text: quotaType.bonus_credits
                        ? t('pricingPlans.creditFeature', {
                            total: quotaType.equivalent_credits,
                            bonus: quotaType.bonus_credits,
                            estimated: quotaType.estimated_requests,
                        })
                        : t('pricingPlans.creditFeatureWithoutBonus', {
                            total: quotaType.equivalent_credits,
                            estimated: quotaType.estimated_requests,
                        }),
                    available: true,
                },
                {
                    text: t('pricingPlans.freeCodeCompletion'),
                    available: true,
                },
                {
                    text: t('pricingPlans.freeCodeReview'),
                    available: true,
                },
                {
                    text: t('pricingPlans.advancedModelAvailable'),
                    available: true,
                },
            ],
            clickEvent() {
                router.push({
                    path: '/subscribe',
                    query: {
                        type: quotaType.id,
                    },
                });
            },
        };
    });

    // 合并免费套餐和付费套餐
    return [freePlan, ...paidPlans];
};
