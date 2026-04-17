import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { authService, AuthService } from '@/services/auth';
import { tokenManager } from '@/utils/token';
import { userService } from '@/services/user';
import { getUserToken } from '@/api/mods/quota.mod';
import { setActivePinia, createPinia } from 'pinia';
import { useUserStore } from '@/store/user';

// Mock dependencies
vi.mock('@/utils/token', () => ({
    tokenManager: {
        getHashToken: vi.fn(),
        getToken: vi.fn(),
        setToken: vi.fn(),
        clearToken: vi.fn(),
        cleanUrlState: vi.fn(),
    },
}));

vi.mock('@/services/user', () => ({
    userService: {
        fetchUserInfo: vi.fn(),
    },
    UserService: vi.fn(),
}));

vi.mock('@/api/mods/quota.mod', () => ({
    getUserToken: vi.fn(),
}));

const mockedGetUserToken = vi.mocked(getUserToken);
const mockedFetchUserInfo = vi.mocked(userService.fetchUserInfo);

describe('AuthService', () => {
    let service: AuthService;

    beforeEach(() => {
        setActivePinia(createPinia());
        service = new AuthService();
        vi.clearAllMocks();
        // Reset window.location
        if (typeof window !== 'undefined') {
            // @ts-ignore
            delete window.location;
            // @ts-ignore
            window.location = { href: 'https://example.com' };
        }
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    describe('authenticate', () => {
        it('should return error when authentication is already in progress', async () => {
            // Set authenticating flag manually
            (service as any).isAuthenticating = true;

            const result = await service.authenticate();

            expect(result.success).toBe(false);
            expect(result.error).toBe('Authentication in progress');
        });

        it('should handle hash token authentication successfully', async () => {
            vi.mocked(tokenManager.getHashToken).mockReturnValue('hash-token-123');
            vi.mocked(tokenManager.getToken).mockReturnValue('existing-token');
            mockedGetUserToken.mockResolvedValue({
                data: { access_token: 'new-access-token' },
            } as any);
            mockedFetchUserInfo.mockResolvedValue({
                phoneNumber: '13800138000',
                userId: 'user-123',
                employeeNumber: 'EMP001',
                githubName: 'testuser',
                userName: 'Test User',
                isPrivate: false,
            });

            const result = await service.authenticate();

            expect(result.success).toBe(true);
            expect(result.user).toBeDefined();
            expect(result.user?.userId).toBe('user-123');
            expect(tokenManager.setToken).toHaveBeenCalledWith('new-access-token');
        });

        it('should handle hash token authentication failure', async () => {
            vi.mocked(tokenManager.getHashToken).mockReturnValue('hash-token-123');
            mockedGetUserToken.mockRejectedValue(new Error('API Error'));

            const result = await service.authenticate();

            expect(result.success).toBe(false);
            expect(result.error).toBe('Hash token authentication failed');
        });

        it('should handle regular authentication with existing token', async () => {
            vi.mocked(tokenManager.getHashToken).mockReturnValue(null);
            vi.mocked(tokenManager.getToken).mockReturnValue('existing-token');
            mockedFetchUserInfo.mockResolvedValue({
                phoneNumber: '13800138000',
                userId: 'user-123',
                employeeNumber: 'EMP001',
                githubName: 'testuser',
                userName: 'Test User',
                isPrivate: false,
            });

            const result = await service.authenticate();

            expect(result.success).toBe(true);
            expect(result.user).toBeDefined();
        });

        it('should fail regular authentication when no token exists', async () => {
            vi.mocked(tokenManager.getHashToken).mockReturnValue(null);
            vi.mocked(tokenManager.getToken).mockReturnValue(undefined);

            const result = await service.authenticate();

            expect(result.success).toBe(false);
            expect(result.error).toBe('No token available');
        });

        it('should handle regular authentication failure', async () => {
            vi.mocked(tokenManager.getHashToken).mockReturnValue(null);
            vi.mocked(tokenManager.getToken).mockReturnValue('invalid-token');
            mockedFetchUserInfo.mockRejectedValue(new Error('Unauthorized'));

            const result = await service.authenticate();

            expect(result.success).toBe(false);
            expect(result.error).toBe('Regular authentication failed');
        });

        it('should handle unexpected errors', async () => {
            vi.mocked(tokenManager.getHashToken).mockImplementation(() => {
                throw new Error('Unexpected error');
            });

            const result = await service.authenticate();

            expect(result.success).toBe(false);
            expect(result.error).toBe('Authentication failed');
        });
    });

    describe('isAuthenticated', () => {
        it('should return true when token is initialized and token exists', async () => {
            const store = useUserStore();
            store.updateTokenInitialized(true);
            vi.mocked(tokenManager.getToken).mockReturnValue('valid-token');

            const result = await service.isAuthenticated();

            expect(result).toBe(true);
        });

        it('should return false when token is not initialized', async () => {
            const store = useUserStore();
            store.updateTokenInitialized(false);
            vi.mocked(tokenManager.getToken).mockReturnValue('valid-token');

            const result = await service.isAuthenticated();

            expect(result).toBe(false);
        });

        it('should return false when token does not exist', async () => {
            const store = useUserStore();
            store.updateTokenInitialized(true);
            vi.mocked(tokenManager.getToken).mockReturnValue(undefined);

            const result = await service.isAuthenticated();

            expect(result).toBe(false);
        });
    });

    describe('logout', () => {
        it('should clear token and reset authentication state', async () => {
            const store = useUserStore();
            store.updateTokenInitialized(true);

            await service.logout();

            expect(tokenManager.clearToken).toHaveBeenCalled();
            expect(store.isTokenInitialized).toBe(false);
        });

        it('should handle errors gracefully', async () => {
            vi.mocked(tokenManager.clearToken).mockImplementation(() => {
                throw new Error('Clear token error');
            });
            const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

            await service.logout();

            expect(consoleErrorSpy).toHaveBeenCalledWith('Logout error');
            consoleErrorSpy.mockRestore();
        });
    });
});

describe('AuthService - auth_redirect', () => {
    let service: AuthService;

    beforeEach(() => {
        setActivePinia(createPinia());
        service = new AuthService();
        vi.clearAllMocks();
        localStorage.clear();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('认证成功后，若 localStorage 有 auth_redirect，应跳转并清除', async () => {
        localStorage.setItem('auth_redirect', '/credits');
        vi.mocked(tokenManager.getHashToken).mockReturnValue(null);
        vi.mocked(tokenManager.getToken).mockReturnValue('valid-token');
        mockedFetchUserInfo.mockResolvedValue({
            phoneNumber: '13800138000',
            userId: 'user-123',
            employeeNumber: 'EMP001',
            githubName: 'testuser',
            userName: 'Test User',
            isPrivate: false,
        });

        const mockRouter = { replace: vi.fn() };
        const result = await service.authenticate(mockRouter as any);

        expect(result.success).toBe(true);
        expect(mockRouter.replace).toHaveBeenCalledWith('/credits');
        expect(localStorage.getItem('auth_redirect')).toBeNull();
    });

    it('认证成功后，若 localStorage 无 auth_redirect，不执行跳转', async () => {
        vi.mocked(tokenManager.getHashToken).mockReturnValue(null);
        vi.mocked(tokenManager.getToken).mockReturnValue('valid-token');
        mockedFetchUserInfo.mockResolvedValue({
            phoneNumber: '13800138000',
            userId: 'user-123',
            employeeNumber: 'EMP001',
            githubName: 'testuser',
            userName: 'Test User',
            isPrivate: false,
        });

        const mockRouter = { replace: vi.fn() };
        const result = await service.authenticate(mockRouter as any);

        expect(result.success).toBe(true);
        expect(mockRouter.replace).not.toHaveBeenCalled();
    });

    it('认证失败时，不清除 localStorage 中的 auth_redirect', async () => {
        localStorage.setItem('auth_redirect', '/credits');
        vi.mocked(tokenManager.getHashToken).mockReturnValue(null);
        vi.mocked(tokenManager.getToken).mockReturnValue(undefined);

        const mockRouter = { replace: vi.fn() };
        const result = await service.authenticate(mockRouter as any);

        expect(result.success).toBe(false);
        expect(localStorage.getItem('auth_redirect')).toBe('/credits');
        expect(mockRouter.replace).not.toHaveBeenCalled();
    });
});

describe('authService singleton', () => {
    it('should be an instance of AuthService', () => {
        expect(authService).toBeInstanceOf(AuthService);
    });
});
