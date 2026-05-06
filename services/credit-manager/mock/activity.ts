/**
 * @file activity mock
 */

import type { MockMethod } from 'vite-plugin-mock';
import { successResponse } from './_utils';

export default [
    {
        url: '/operational_activities/api/v1/user/me',
        method: 'get',
        response: () =>
            successResponse({
                userId: 'user-mock-001',
                username: 'mockuser',
                displayName: 'Mock User',
                casdoorId: 'casdoor-mock-001',
                createdTime: '2025-01-01T00:00:00Z',
                registerOrder: 1,
                registerOrderPercent: 0.01,
                latestActivity: '2026-04-29T10:00:00Z',
                totalLatencyMs: 1500,
                totalTokens: 100000,
                totalRequests: 5000,
                totalUsageDays: 120,
                modelStats: '{"gpt-4":5000,"claude-3":3000}',
                modeStats: '{"chat":4000,"code":2000}',
                updatedAt: '2026-04-29T10:00:00Z',
                creditUsage: 3500,
                creditUsageOrder: 10,
                creditUsageOrderPercent: 0.1,
                accessId: 'access-mock-001',
                isInner: 0,
                modelStatsJson: { 'gpt-4': 5000, 'claude-3': 3000 },
                modeStatsJson: { chat: 4000, code: 2000 },
                identity: 'user',
            }),
    },
    {
        url: '/operational_activities/api/v1/share/record',
        method: 'post',
        response: () =>
            successResponse({
                userId: 'user-mock-001',
                accessId: 'access-mock-001',
                username: 'mockuser',
                shareNum: 5,
                inviteCode: 'INVITE2025MOCK',
                updatedAt: '2026-04-29T10:00:00Z',
            }),
    },
] as MockMethod[];
