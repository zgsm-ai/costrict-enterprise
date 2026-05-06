/**
 * @file auth mock
 */

import type { MockMethod } from 'vite-plugin-mock';
import { successResponse } from './_utils';

export default [
    {
        url: '/oidc-auth/api/v1/manager/token',
        method: 'get',
        response: () =>
            successResponse({
                access_token: 'mock_access_token_12345',
                message: 'Token obtained successfully',
                state: 'state',
                success: true,
            }),
    },
    {
        url: '/oidc-auth/api/v1/manager/userinfo',
        method: 'get',
        response: () =>
            successResponse({
                email: 'mock@example.com',
                githubID: 'github-mock-001',
                githubName: 'mockuser',
                phone: '13800138000',
                username: 'Mock User',
                uuid: 'user-mock-001',
                employee_number: 'E001',
                isPrivate: false,
            }),
    },
    {
        url: '/oidc-auth/api/v1/manager/bind/account',
        method: 'get',
        response: () =>
            successResponse({
                url: 'https://example.com/bind?state=mock',
                message: 'Binding URL generated',
                success: true,
            }),
    },
    {
        url: '/oidc-auth/api/v1/manager/invite-code',
        method: 'get',
        response: () =>
            successResponse({
                invite_code: 'INVITE2025MOCK',
            }),
    },
    {
        url: '/oidc-auth/api/v1/manager/login',
        method: 'get',
        response: () =>
            successResponse({
                url: 'https://example.com/oauth/authorize?client_id=mock&redirect_uri=https%3A%2F%2Flocalhost%3A9527%2Flogin',
            }),
    },
] as MockMethod[];
