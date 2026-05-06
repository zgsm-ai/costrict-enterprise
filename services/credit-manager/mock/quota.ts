/**
 * @file quota mock
 */

import type { MockMethod, Recordable } from 'vite-plugin-mock';
import { successResponse } from './_utils';

const mockQuotaList = [
    { amount: 5000, expiry_date: '2026-12-31 23:59:59' },
    { amount: 3000, expiry_date: '2026-06-30 23:59:59' },
    { amount: 2000, expiry_date: '2025-12-31 23:59:59' },
];

const mockAuditRecords = [
    {
        amount: 1000,
        operation: 'recharge',
        voucher_code: 'VC001',
        related_user: 'admin',
        strategy_name: 'New User Bonus',
        expiry_date: '2026-12-31 23:59:59',
        create_time: '2026-04-01 10:00:00',
        details: {
            operation: 'recharge',
            items: [{ amount: 1000, expiry_date: '2026-12-31 23:59:59', status: 'active' }],
        },
    },
    {
        amount: 500,
        operation: 'bonus',
        voucher_code: 'VC002',
        related_user: 'system',
        strategy_name: 'Monthly Bonus',
        expiry_date: '2026-06-30 23:59:59',
        create_time: '2026-04-15 14:30:00',
        details: {
            operation: 'bonus',
            items: [{ amount: 500, expiry_date: '2026-06-30 23:59:59', status: 'active' }],
        },
    },
];

const mockUsageRecords = Array.from({ length: 25 }, (_, i) => ({
    id: i + 1,
    user_id: 'user-mock-001',
    model: ['gpt-4', 'claude-3', 'gpt-3.5'][i % 3],
    mode: ['chat', 'code', 'completion'][i % 3],
    tokens: Math.floor(Math.random() * 5000) + 100,
    credits_used: Math.floor(Math.random() * 100) + 1,
    package: 'Standard',
    record_time: `2026-04-${String((i % 28) + 1).padStart(2, '0')} ${String(9 + (i % 12)).padStart(2, '0')}:00:00`,
    create_time: `2026-04-${String((i % 28) + 1).padStart(2, '0')} ${String(9 + (i % 12)).padStart(2, '0')}:00:00`,
    update_time: `2026-04-${String((i % 28) + 1).padStart(2, '0')} ${String(9 + (i % 12)).padStart(2, '0')}:00:00`,
}));

export default [
    {
        url: '/quota-manager/api/v1/quota',
        method: 'get',
        response: () =>
            successResponse({
                total_quota: 10000,
                used_quota: 3500,
                quota_list: mockQuotaList,
                is_star: 'true',
            }),
    },
    {
        url: '/quota-manager/api/v1/quota/audit',
        method: 'get',
        response: ({ query }: { query: Recordable }) => {
            const page = Number(query.page) || 1;
            const pageSize = Number(query.page_size) || 10;
            const total = mockAuditRecords.length;
            const start = (page - 1) * pageSize;
            const end = start + pageSize;
            return successResponse({
                total,
                records: mockAuditRecords.slice(start, end),
            });
        },
    },
    {
        url: '/quota-manager/api/v1/quota/transfer-out',
        method: 'post',
        response: () =>
            successResponse({
                voucher_code: 'VC-OUT-123456',
                related_user: 'user-mock-001',
                operation: 'transfer_out',
                quota_list: [{ amount: 500, expiry_date: '2026-12-31 23:59:59' }],
            }),
    },
    {
        url: '/quota-manager/api/v1/quota/transfer-in',
        method: 'post',
        response: () =>
            successResponse({
                giver_id: 'giver-mock-001',
                giver_name: 'Alice',
                giver_phone: '13800138001',
                giver_github: 'alice-github',
                receiver_id: 'user-mock-001',
                quota_list: [
                    {
                        amount: 500,
                        expiry_date: '2026-12-31 23:59:59',
                        is_expired: false,
                        success: true,
                    },
                ],
                voucher_code: 'VC-IN-123456',
                operation: 'transfer_in',
                status: 'success',
                message: 'Transfer completed successfully',
            }),
    },
    {
        url: '/quota-manager/api/v1/usage/statistics',
        method: 'get',
        response: ({ query }: { query: Recordable }) => {
            const page = Number(query.page) || 1;
            const pageSize = Number(query.page_size) || 10;
            const total = mockUsageRecords.length;
            const start = (page - 1) * pageSize;
            const end = start + pageSize;
            return successResponse({
                records: mockUsageRecords.slice(start, end),
                total,
                page,
                page_size: pageSize,
            });
        },
    },
] as MockMethod[];
