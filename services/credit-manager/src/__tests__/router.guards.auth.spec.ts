import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createMemoryHistory, createRouter } from 'vue-router';
import { setupAuthGuard } from '@/router/guards/auth';
import { authService } from '@/services/auth';
import { getLoginReturnTo, getLoginServiceUrl } from '@/utils/login';

vi.mock('@/services/auth', () => ({ authService: { isAuthenticated: vi.fn() } }));
vi.mock('@/router', () => ({ PUBLIC_ROUTES: ['/credit-reward-plan'] }));

function makeRouter() {
    const component = { template: '<div />' };
    const router = createRouter({
        history: createMemoryHistory('/credit/manager'),
        routes: ['/', '/login', '/credits', '/credit-reward-plan'].map((path) => ({
            path,
            component,
        })),
    });
    setupAuthGuard(router);
    return router;
}

describe('Cookie route guard and login return URL', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        sessionStorage.clear();
        vi.stubEnv('BASE_URL', '/credit/manager');
        window.history.replaceState({}, '', '/credit/manager?state=legacy-state&tab=usage#details');
    });

    it('preserves the original complete entry URL for loginservice when Cookie is missing', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(false);
        const original = window.location.href;
        const router = makeRouter();
        await router.push('/?state=legacy-state&tab=usage#details');
        expect(router.currentRoute.value.path).toBe('/login');
        expect(getLoginReturnTo()).toBe(original);
        const login = new URL(getLoginServiceUrl());
        expect(login.pathname).toBe('/login-service/api/v1/web/login');
        expect(login.searchParams.get('return_to')).toBe(original);
    });

    it('leaves state and tab intact when Cookie exists', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(true);
        const router = makeRouter();
        await router.push('/?state=legacy-state&tab=usage');
        expect(router.currentRoute.value.fullPath).toBe('/?state=legacy-state&tab=usage');
        expect(sessionStorage.length).toBe(0);
    });

    it('remembers the target of an in-app navigation after Cookie disappears', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(true);
        const router = makeRouter();
        await router.push('/');
        vi.mocked(authService.isAuthenticated).mockResolvedValue(false);
        await router.push('/credits?tab=history');
        expect(getLoginReturnTo()).toBe(
            `${window.location.origin}/credit/manager/credits?tab=history`,
        );
    });

    it('allows public and login routes without Cookie', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(false);
        const router = makeRouter();
        await router.push('/credit-reward-plan');
        expect(router.currentRoute.value.path).toBe('/credit-reward-plan');
        await router.push('/login');
        expect(router.currentRoute.value.path).toBe('/login');
        expect(sessionStorage.length).toBe(0);
    });

    it('rejects return URLs outside this application', () => {
        for (const url of [
            'https://other.example/credit/manager',
            `${window.location.origin}/other`,
        ]) {
            sessionStorage.setItem('login_service_return_to', url);
            expect(getLoginReturnTo()).toBe(`${window.location.origin}/credit/manager`);
        }
    });
});
