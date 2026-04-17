import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useProfile } from '@/views/Home/hook/useProfile';
import { setActivePinia, createPinia } from 'pinia';
import { useUserStore } from '@/store/user';

// Mock the router
const mockPush = vi.fn();
const mockOpen = vi.fn();
vi.mock('vue-router', () => ({
    useRouter: () => ({
        push: mockPush,
    }),
}));

// Mock naive-ui
const mockSuccess = vi.fn();
const mockError = vi.fn();
vi.mock('naive-ui', () => ({
    useMessage: () => ({
        success: mockSuccess,
        error: mockError,
    }),
}));

// Mock API calls
vi.mock('@/api/mods/quota.mod', () => ({
    getUserQuota: vi.fn(),
    getBindAccount: vi.fn(),
    getInviteCode: vi.fn(),
}));

// Mock copy utility
vi.mock('@/utils/copy', () => ({
    copyToClipboard: vi.fn(),
}));

// Mock auth service
vi.mock('@/services/auth', () => ({
    authService: {
        logout: vi.fn(),
    },
}));

import { getUserQuota, getBindAccount, getInviteCode } from '@/api/mods/quota.mod';
import { copyToClipboard } from '@/utils/copy';
import { authService } from '@/services/auth';

const mockedGetUserQuota = vi.mocked(getUserQuota);
const mockedGetBindAccount = vi.mocked(getBindAccount);
const mockedGetInviteCode = vi.mocked(getInviteCode);
const mockedCopyToClipboard = vi.mocked(copyToClipboard);
const mockedLogout = vi.mocked(authService.logout);

describe('useProfile', () => {
    beforeEach(() => {
        setActivePinia(createPinia());
        vi.clearAllMocks();
        // Mock window.open
        // @ts-ignore
        Object.defineProperty(window, 'location', {
            value: {
                href: 'https://example.com',
                ancestorOrigins: {} as DOMStringList,
                hash: '',
                host: '',
                hostname: '',
                origin: '',
                pathname: '',
                port: '',
                protocol: '',
                search: '',
                assign: vi.fn(),
                reload: vi.fn(),
                replace: vi.fn(),
            } as Location,
            writable: true,
        });
        (window as Window).open = mockOpen as (
            url?: string | URL,
            target?: string,
            features?: string,
        ) => Window | null;
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    describe('initial state', () => {
        it('should have correct initial state', () => {
            const profile = useProfile();

            expect(profile.usedQuota.value).toBe(0);
            expect(profile.totalQuota.value).toBe(0);
            expect(profile.columnsData.value).toEqual([]);
            expect(profile.inviteCode.value).toBe('');
        });
    });

    describe('fetchUserQuota', () => {
        it('should fetch and update quota data successfully', async () => {
            const mockQuotaData = {
                data: {
                    quota_list: [
                        { id: 1, name: 'Quota 1' },
                        { id: 2, name: 'Quota 2' },
                    ],
                    used_quota: 100,
                    total_quota: 500,
                    is_star: 'true',
                },
            };
            mockedGetUserQuota.mockResolvedValue(mockQuotaData as any);

            const profile = useProfile();
            await profile.fetchUserQuota();

            expect(mockedGetUserQuota).toHaveBeenCalledTimes(1);
            expect(profile.columnsData.value).toEqual(mockQuotaData.data.quota_list);
            expect(profile.usedQuota.value).toBe(100);
            expect(profile.totalQuota.value).toBe(500);
            expect(profile.isStar.value).toBe('true');
        });

        it('should handle null data response', async () => {
            mockedGetUserQuota.mockResolvedValue({ data: null } as any);

            const profile = useProfile();
            await profile.fetchUserQuota();

            expect(profile.columnsData.value).toEqual([]);
            expect(profile.usedQuota.value).toBe(0);
            expect(profile.totalQuota.value).toBe(0);
        });

        it('should throw API error', async () => {
            mockedGetUserQuota.mockRejectedValue(new Error('API Error'));

            const profile = useProfile();
            // Should throw error when API fails
            await expect(profile.fetchUserQuota()).rejects.toThrow('API Error');
        });
    });

    describe('fetchInviteCode', () => {
        it('should fetch and update invite code', async () => {
            mockedGetInviteCode.mockResolvedValue({
                data: { invite_code: 'INVITE123' },
            } as any);

            const profile = useProfile();
            await profile.fetchInviteCode();

            expect(mockedGetInviteCode).toHaveBeenCalledTimes(1);
            expect(profile.inviteCode.value).toBe('INVITE123');
        });

        it('should use default value when invite code is missing', async () => {
            mockedGetInviteCode.mockResolvedValue({ data: {} } as any);

            const profile = useProfile();
            await profile.fetchInviteCode();

            expect(profile.inviteCode.value).toBe('');
        });
    });

    describe('bindAction', () => {
        it('should redirect to bind URL for github', async () => {
            mockedGetBindAccount.mockResolvedValue({
                data: { url: 'https://github.com/bind' },
            } as any);

            const profile = useProfile();
            await profile.bindGithub();

            expect(mockedGetBindAccount).toHaveBeenCalledWith({
                bindType: 'github',
                state: 'state',
            });
            expect(window.location.href).toBe('https://github.com/bind');
        });

        it('should redirect to bind URL for sms', async () => {
            mockedGetBindAccount.mockResolvedValue({
                data: { url: 'https://sms.com/bind' },
            } as any);

            const profile = useProfile();
            await profile.bindPhone();

            expect(mockedGetBindAccount).toHaveBeenCalledWith({
                bindType: 'sms',
                state: 'state',
            });
            expect(window.location.href).toBe('https://sms.com/bind');
        });

        it('should not redirect when URL is missing', async () => {
            const originalHref = window.location.href;
            mockedGetBindAccount.mockResolvedValue({ data: {} } as any);

            const profile = useProfile();
            await profile.bindGithub();

            expect(window.location.href).toBe(originalHref);
        });
    });

    describe('copyCode', () => {
        it('should copy user ID with success handler', async () => {
            mockedCopyToClipboard.mockResolvedValue(true);

            const profile = useProfile();
            profile.copyCode();

            expect(mockedCopyToClipboard).toHaveBeenCalledWith(profile.userId.value, {
                success: mockSuccess,
                error: mockError,
            });
        });
    });

    describe('copyInviteCode', () => {
        it('should copy invite code with success handler', async () => {
            mockedCopyToClipboard.mockResolvedValue(true);

            const profile = useProfile();
            profile.inviteCode.value = 'INVITE123';
            profile.copyInviteCode();

            expect(mockedCopyToClipboard).toHaveBeenCalledWith('INVITE123', {
                success: mockSuccess,
                error: mockError,
            });
        });
    });

    describe('logout', () => {
        it('should logout and redirect to login page', async () => {
            mockedLogout.mockResolvedValue(undefined);

            const profile = useProfile();
            await profile.logout();

            expect(mockedLogout).toHaveBeenCalledTimes(1);
            expect(mockPush).toHaveBeenCalledWith('/login');
        });

        it('should handle logout error gracefully', async () => {
            const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
            mockedLogout.mockRejectedValue(new Error('Logout failed'));

            const profile = useProfile();
            await profile.logout();

            expect(consoleErrorSpy).toHaveBeenCalledWith('Logout failed:', expect.any(Error));
            consoleErrorSpy.mockRestore();
        });
    });

    describe('toCredits', () => {
        it('should open credits page in new window', () => {
            const profile = useProfile();
            profile.toCredits();

            expect(mockOpen).toHaveBeenCalledWith('/credit/manager/credits');
        });
    });

    describe('transferInCallBack', () => {
        it('should refresh user quota', async () => {
            const mockQuotaData = {
                data: {
                    quota_list: [],
                    used_quota: 0,
                    total_quota: 0,
                    is_star: undefined,
                },
            };
            mockedGetUserQuota.mockResolvedValue(mockQuotaData as any);

            const profile = useProfile();
            await profile.transferInCallBack();

            expect(mockedGetUserQuota).toHaveBeenCalledTimes(1);
        });
    });

    describe('transferData computed', () => {
        it('should return columnsData when isStar is true', () => {
            const profile = useProfile();
            profile.isStar.value = 'true';
            profile.columnsData.value = [{ amount: 100, expiry_date: '2024-12-31' }];

            expect(profile.transferData.value).toEqual([
                { amount: 100, expiry_date: '2024-12-31' },
            ]);
        });

        it('should return columnsData when isStar is undefined', () => {
            const profile = useProfile();
            profile.isStar.value = undefined;
            profile.columnsData.value = [{ amount: 100, expiry_date: '2024-12-31' }];

            expect(profile.transferData.value).toEqual([
                { amount: 100, expiry_date: '2024-12-31' },
            ]);
        });

        it('should return empty array when isStar is false', () => {
            const profile = useProfile();
            profile.isStar.value = 'false';
            profile.columnsData.value = [{ amount: 100, expiry_date: '2024-12-31' }];

            expect(profile.transferData.value).toEqual([]);
        });
    });
});
