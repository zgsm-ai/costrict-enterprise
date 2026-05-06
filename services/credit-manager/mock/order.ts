/**
 * @file order mock
 */

import type { MockMethod, Recordable } from 'vite-plugin-mock';
import { successResponse } from './_utils';

const mockOrders = [
    {
        id: 1,
        order_id: 'ORD20250429001',
        user_id: 'user-mock-001',
        amount: 99.0,
        status: 'completed',
        quota_type: 'standard',
        order_source: 'web',
        credit_count: 1000,
        credit_expire_date: '2026-12-31',
        created_at: '2026-04-20T10:00:00Z',
        updated_at: '2026-04-20T10:00:00Z',
        invoice_status: 1,
        description: 'Standard quota package',
        jd_order_id: '',
    },
    {
        id: 2,
        order_id: 'ORD20250415002',
        user_id: 'user-mock-001',
        amount: 299.0,
        status: 'completed',
        quota_type: 'premium',
        order_source: 'web',
        credit_count: 5000,
        credit_expire_date: '2027-04-15',
        created_at: '2026-04-15T14:30:00Z',
        updated_at: '2026-04-15T14:30:00Z',
        invoice_status: 0,
        description: 'Premium quota package',
        jd_order_id: '',
    },
    {
        id: 3,
        order_id: 'ORD20250401003',
        user_id: 'user-mock-001',
        amount: 29.9,
        status: 'completed',
        quota_type: 'starter',
        order_source: 'web',
        credit_count: 200,
        credit_expire_date: '2026-06-01',
        created_at: '2026-04-01T09:00:00Z',
        updated_at: '2026-04-01T09:00:00Z',
        invoice_status: 2,
        description: 'Starter quota package',
        jd_order_id: '',
    },
];

const mockQuotaTypes = [
    {
        id: 1,
        display_name: 'Starter',
        credit_count: 200,
        amount: 29.9,
        valid_days: 90,
        quota_marketing_rules_id: null,
        original_amount: 49.9,
        equivalent_credits: 200,
        bonus_credits: 0,
        estimated_requests: 1000,
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z',
        marketing_rules: null,
    },
    {
        id: 2,
        display_name: 'Standard',
        credit_count: 1000,
        amount: 99.0,
        valid_days: 365,
        quota_marketing_rules_id: 1,
        original_amount: 199.0,
        equivalent_credits: 1000,
        bonus_credits: 100,
        estimated_requests: 5000,
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z',
        marketing_rules: {
            id: 1,
            rule_type: 1,
            pay_amount: 99.0,
            pay_discount: 0.5,
            status: 1,
            apply_user_type: 3,
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
        },
    },
    {
        id: 3,
        display_name: 'Premium',
        credit_count: 5000,
        amount: 299.0,
        valid_days: 365,
        quota_marketing_rules_id: 2,
        original_amount: 599.0,
        equivalent_credits: 5000,
        bonus_credits: 500,
        estimated_requests: 25000,
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z',
        marketing_rules: {
            id: 2,
            rule_type: 2,
            pay_amount: 299.0,
            pay_discount: 0.5,
            status: 1,
            apply_user_type: 3,
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
        },
    },
];

export default [
    {
        url: '/quota-order-manager/api/v1/orders',
        method: 'get',
        response: ({ query }: { query: Recordable }) => {
            const page = Number(query.page) || 1;
            const pageSize = Number(query.page_size) || 10;
            const total = mockOrders.length;
            const start = (page - 1) * pageSize;
            const end = start + pageSize;
            return successResponse({
                orders: mockOrders.slice(start, end),
                total,
                limit: pageSize,
                offset: (page - 1) * pageSize,
            });
        },
    },
    {
        url: '/quota-order-manager/api/v1/orders',
        method: 'post',
        response: () =>
            successResponse({
                id: 4,
                order_id: 'ORD20250429004',
                user_id: 'user-mock-001',
                amount: 99.0,
                credit_count: 1000,
                credit_expire_date: '2026-12-31',
                status: 'pending',
                quota_type: 'standard',
                description: 'Standard quota package',
                order_source: 'web',
                jd_order_id: '',
                invoice_status: 0,
                created_at: '2026-04-29T10:00:00Z',
                updated_at: '2026-04-29T10:00:00Z',
            }),
    },
    {
        url: '/quota-order-manager/api/v1/payment/initiate',
        method: 'post',
        response: () =>
            successResponse({
                version: '1.0',
                encoding: 'UTF-8',
                sign: 'mock_sign_xxx',
                signMethod: 'RSA',
                returnCode: 'SUCCESS',
                respCode: '00',
                qrCode: 'weixin://wxpay/bizpayurl?pr=mock123',
                cmbOrderId: 'CMB20250429001',
                txnTime: '20260429100000',
                biz_content: {
                    merId: 'M001',
                    orderId: 'ORD20250429004',
                    cmbOrderId: 'CMB20250429001',
                    qrCode: 'weixin://wxpay/bizpayurl?pr=mock123',
                    txnTime: '20260429100000',
                },
            }),
    },
    {
        url: '/quota-order-manager/api/v1/invoices',
        method: 'post',
        response: () =>
            successResponse({
                invoice_id: 1,
                title_type: 1,
                invoice_type: 1,
                invoice_title: 'Mock Company Ltd.',
                taxpayer_id: '91110000123456789X',
                company_address: 'Beijing, China',
                company_phone: '010-12345678',
                bank_name: 'Industrial and Commercial Bank of China',
                bank_account: '6222021234567890123',
                receive_email: 'invoice@example.com',
                order_id: 'ORD20250429001',
                amount: 99.0,
                invoice_content: 'Software Service',
                invoice_status: 0,
                apply_time: '2026-04-29T10:00:00Z',
                issue_time: null,
            }),
    },
    {
        url: '/quota-order-manager/api/v1/orders/:id',
        method: 'get',
        response: ({ query }: { query: Recordable }) => {
            const orderId = query.id as string;
            const order = mockOrders.find((o) => o.order_id === orderId) || mockOrders[0];
            return successResponse(order);
        },
    },
    {
        url: '/quota-order-manager/api/v1/quotas/types',
        method: 'get',
        response: () => successResponse(mockQuotaTypes),
    },
    {
        url: '/quota-order-manager/api/v1/quotas/types/:id',
        method: 'get',
        response: ({ query }: { query: Recordable }) => {
            const id = Number(query.id);
            const type = mockQuotaTypes.find((t) => t.id === id) || mockQuotaTypes[0];
            return successResponse(type);
        },
    },
] as MockMethod[];
