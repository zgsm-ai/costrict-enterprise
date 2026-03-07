/**
 * @file home type
 */
import type { QuotaList } from '@/api/bos/quota.bo';
import type { VNode } from 'vue';

export interface RowData extends QuotaList {
    id: number;
    transferAmount: number;
}

export interface FormModel {
    receiverId: string;
    redeemCode: string;
    checkedRowKeys: number[];
}

export interface Feature {
    text: string;
    available: boolean;
}

export interface PricingPlan {
    title: string;
    price: number;
    originalPrice?: number;
    description: string;
    buttonText: string;
    buttonType: 'download' | 'purchase';
    showTrafficLabel?: boolean;
    features: Feature[];
    clickEvent: () => void;
    isFirstPurchase?: boolean;
}

export interface GuideStep {
    title: string | (() => VNode);
    content?: string | (() => VNode);
    tips?: string | (() => VNode);
    imageTextPairs?: ImageTextPair[];
}
export interface ImageTextPair {
    imgUrl?: string | (() => VNode);
    text?: string | (() => VNode);
}

export interface InvoiceModalProps {
    show: boolean;
    orderNo?: string;
    amount?: number | null;
}

export interface InvoiceModalFormData {
    headerType: string;
    invoiceType: string;
    title: string;
    taxNumber: string;
    address: string;
    phone: string;
    bank: string;
    bankAccount: string;
    email: string;
}
