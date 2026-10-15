import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { AuthService } from '@/services/auth';
import { useUserStore } from '@/store/user';
import { tokenManager } from '@/utils/token';
import { userService } from '@/services/user';
import { getUserToken } from '@/api/mods/quota.mod';

vi.mock('@/utils/token', () => ({
    tokenManager: { getToken: vi.fn(), clearToken: vi.fn() },
}));
vi.mock('@/services/user', () => ({ userService: { fetchUserInfo: vi.fn() } }));
vi.mock('@/api/mods/quota.mod', () => ({ getUserToken: vi.fn() }));

describe('Cookie authentication', () => {
    beforeEach(() => {
        setActivePinia(createPinia());
        vi.resetAllMocks();
        window.history.replaceState({}, '', '/credit/manager?state=old-state&tab=usage');
    });

    it('accepts an existing Cookie without fetching token or user information or changing URL', async () => {
        vi.mocked(tokenManager.getToken).mockReturnValue('cookie-token');
        const service = new AuthService();
        const url = window.location.href;
        expect(await service.authenticate()).toEqual({ success: true });
        expect(useUserStore().isTokenInitialized).toBe(true);
        expect(await service.isAuthenticated()).toBe(true);
        expect(getUserToken).not.toHaveBeenCalled();
        expect(userService.fetchUserInfo).not.toHaveBeenCalled();
        expect(window.location.href).toBe(url);
    });

    it('does not accept state without a Cookie, even if the store was previously initialized', async () => {
        useUserStore().updateTokenInitialized(true);
        expect(await new AuthService().isAuthenticated()).toBe(false);
        expect(useUserStore().isTokenInitialized).toBe(false);
        expect(getUserToken).not.toHaveBeenCalled();
        expect(tokenManager.clearToken).not.toHaveBeenCalled();
    });

    it('hydrates profile from JWT and clears stale fields when the Cookie changes', async () => {
        const payload = btoa(
            JSON.stringify({
                display_name: 'Alice',
                sub: 'user-1',
                universal_id: 'quota-1',
                employee_number: 'E001',
            }),
        );
        vi.mocked(tokenManager.getToken).mockReturnValue(`header.${payload}.signature`);
        const service = new AuthService();
        await service.authenticate();
        expect(useUserStore()).toMatchObject({
            userName: 'Alice',
            userId: 'user-1',
            universalId: 'quota-1',
            employeeNumber: 'E001',
        });
        vi.mocked(tokenManager.getToken).mockReturnValue(
            `header.${btoa(JSON.stringify({ sub: 'user-2' }))}.signature`,
        );
        await service.authenticate();
        expect(useUserStore()).toMatchObject({
            userName: '',
            userId: 'user-2',
            universalId: '',
            employeeNumber: '',
        });
        vi.mocked(tokenManager.getToken).mockReturnValue(undefined);
        await service.authenticate();
        expect(useUserStore().userId).toBe('');
        expect(tokenManager.clearToken).not.toHaveBeenCalled();
    });

    it('only clears Cookie and profile on explicit logout', async () => {
        useUserStore().$patch({ userName: 'previous user', isTokenInitialized: true });
        await new AuthService().logout();
        expect(tokenManager.clearToken).toHaveBeenCalledOnce();
        expect(useUserStore().userName).toBe('');
        expect(useUserStore().isTokenInitialized).toBe(false);
    });
});
