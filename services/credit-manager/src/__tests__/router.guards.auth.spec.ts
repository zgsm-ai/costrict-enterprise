// src/__tests__/router.guards.auth.spec.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { Router, RouteLocationNormalized } from 'vue-router';
import { setupAuthGuard } from '@/router/guards/auth';
import { authService } from '@/services/auth';
import { PUBLIC_ROUTES } from '@/router';

vi.mock('@/services/auth', () => ({
    authService: {
        isAuthenticated: vi.fn(),
        authenticate: vi.fn(),
    },
}));

vi.mock('@/router', () => ({
    PUBLIC_ROUTES: ['/credit-reward-plan', '/credit-md-preview'],
}));

vi.mock('@/utils/token', () => ({
    tokenManager: {
        cleanUrlState: vi.fn(),
    },
}));

function makeRoute(path: string, query: Record<string, string> = {}): RouteLocationNormalized {
    const search = Object.keys(query).length
        ? '?' + new URLSearchParams(query).toString()
        : '';
    return {
        path,
        fullPath: path + search,
        query,
        hash: '',
        name: undefined,
        params: {},
        matched: [],
        meta: {},
        redirectedFrom: undefined,
    } as unknown as RouteLocationNormalized;
}

function makeRouter(): Router {
    const guards: ((to: any, from: any, next: any) => void)[] = [];
    const afterGuards: ((to: any) => void)[] = [];
    return {
        beforeEach: vi.fn((fn) => guards.push(fn)),
        afterEach: vi.fn((fn) => afterGuards.push(fn)),
        replace: vi.fn(),
        _guards: guards,
        _afterGuards: afterGuards,
    } as unknown as Router;
}

async function runGuard(
    router: Router,
    to: RouteLocationNormalized,
    from: RouteLocationNormalized = makeRoute('/')
) {
    const next = vi.fn();
    await (router as any)._guards[0](to, from, next);
    return next;
}

const flushPromises = () => new Promise(resolve => setTimeout(resolve, 0));

describe('setupAuthGuard - auth_redirect', () => {
    let router: Router;

    beforeEach(() => {
        router = makeRouter();
        setupAuthGuard(router);
        localStorage.clear();
        vi.clearAllMocks();
    });

    it('认证失败时，将目标 fullPath 写入 localStorage["auth_redirect"]', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(false);
        vi.mocked(authService.authenticate).mockResolvedValue({ success: false });

        await runGuard(router, makeRoute('/credits'));
        await flushPromises();

        expect(localStorage.getItem('auth_redirect')).toBe('/credits');
    });

    it('认证失败时，带 query 的 fullPath 完整写入', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(false);
        vi.mocked(authService.authenticate).mockResolvedValue({ success: false });

        await runGuard(router, makeRoute('/credits', { tab: 'history' }));
        await flushPromises();

        expect(localStorage.getItem('auth_redirect')).toBe('/credits?tab=history');
    });

    it('访问 /login 时不写入 auth_redirect', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(false);

        await runGuard(router, makeRoute('/login'));
        await flushPromises();

        expect(localStorage.getItem('auth_redirect')).toBeNull();
    });

    it('访问公开路由时不写入 auth_redirect', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(false);

        await runGuard(router, makeRoute('/credit-reward-plan'));
        await flushPromises();

        expect(localStorage.getItem('auth_redirect')).toBeNull();
    });

    it('已登录时不写入 auth_redirect', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(true);

        await runGuard(router, makeRoute('/credits'));
        await flushPromises();

        expect(localStorage.getItem('auth_redirect')).toBeNull();
    });
});
