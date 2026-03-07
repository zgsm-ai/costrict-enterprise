/**
 * @file 订阅页面类型定义
 */

export interface Feature {
    text: string;
    available: boolean;
}

export interface PricingPlan {
    id: string;
    title: string;
    price: number;
    originalPrice?: number;
    type: number;
    isFirstPurchase?: boolean;
}

export interface PaymentMethod {
    id: string;
    name: string;
    icon: string;
    description: string;
    popular?: boolean;
}

export interface PaymentModalProps {
    modelValue: boolean;
    selectedPaymentMethod: string | null;
    paymentMethods: Array<{
        id: string;
        name: string;
        icon: string;
    }>;
    orderInfo?: {
        id: number;
        order_id: string;
        user_id: string;
        amount: number;
        credit_count: number;
        credit_expire_date: string;
        status: string;
        quota_type: string;
        description: string;
        order_source: string;
        jd_order_id: string;
        invoice_status: number;
        created_at: string;
        updated_at: string;
    } | null;
}

export interface PaymentModalEmits {
    (event: 'update:modelValue', value: boolean): void;
    (event: 'close'): void;
}
