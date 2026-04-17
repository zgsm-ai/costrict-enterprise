import { describe, it, expect, beforeEach } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useUserStore } from '@/store/user';

describe('useUserStore', () => {
    beforeEach(() => {
        setActivePinia(createPinia());
    });

    describe('initial state', () => {
        it('should have correct initial state', () => {
            const store = useUserStore();

            expect(store.phoneNumber).toBe('');
            expect(store.userId).toBe('');
            expect(store.githubName).toBe('');
            expect(store.employeeNumber).toBe('');
            expect(store.userName).toBe('');
            expect(store.isPrivate).toBe(false);
            expect(store.isTokenInitialized).toBe(false);
            expect(store.isAuthenticating).toBe(false);
            expect(store.authError).toBeNull();
        });
    });

    describe('updateUserInfo', () => {
        it('should update all user info fields', () => {
            const store = useUserStore();
            const userInfo = {
                phoneNumber: '13800138000',
                userId: 'user-123',
                githubName: 'testuser',
                employeeNumber: 'EMP001',
                userName: 'Test User',
                isPrivate: true,
            };

            store.updateUserInfo(userInfo);

            expect(store.phoneNumber).toBe('13800138000');
            expect(store.userId).toBe('user-123');
            expect(store.githubName).toBe('testuser');
            expect(store.employeeNumber).toBe('EMP001');
            expect(store.userName).toBe('Test User');
            expect(store.isPrivate).toBe(true);
        });

        it('should handle partial updates', () => {
            const store = useUserStore();

            store.updateUserInfo({
                phoneNumber: '13800138000',
                userId: 'user-123',
                githubName: '',
                employeeNumber: '',
                userName: '',
                isPrivate: false,
            });

            expect(store.phoneNumber).toBe('13800138000');
            expect(store.userId).toBe('user-123');
        });
    });

    describe('updateTokenInitialized', () => {
        it('should update token initialized status to true', () => {
            const store = useUserStore();

            store.updateTokenInitialized(true);

            expect(store.isTokenInitialized).toBe(true);
        });

        it('should update token initialized status to false', () => {
            const store = useUserStore();
            store.updateTokenInitialized(true);

            store.updateTokenInitialized(false);

            expect(store.isTokenInitialized).toBe(false);
        });
    });

    describe('setAuthenticating', () => {
        it('should set authenticating status to true', () => {
            const store = useUserStore();

            store.setAuthenticating(true);

            expect(store.isAuthenticating).toBe(true);
        });

        it('should set authenticating status to false', () => {
            const store = useUserStore();
            store.setAuthenticating(true);

            store.setAuthenticating(false);

            expect(store.isAuthenticating).toBe(false);
        });
    });

    describe('setAuthError', () => {
        it('should set auth error message', () => {
            const store = useUserStore();

            store.setAuthError('Authentication failed');

            expect(store.authError).toBe('Authentication failed');
        });

        it('should clear auth error when set to null', () => {
            const store = useUserStore();
            store.setAuthError('Some error');

            store.setAuthError(null);

            expect(store.authError).toBeNull();
        });
    });

    describe('resetAuth', () => {
        it('should reset all auth-related state', () => {
            const store = useUserStore();
            store.updateTokenInitialized(true);
            store.setAuthenticating(true);
            store.setAuthError('Some error');

            store.resetAuth();

            expect(store.isTokenInitialized).toBe(false);
            expect(store.isAuthenticating).toBe(false);
            expect(store.authError).toBeNull();
        });

        it('should not affect user info fields', () => {
            const store = useUserStore();
            store.updateUserInfo({
                phoneNumber: '13800138000',
                userId: 'user-123',
                githubName: 'testuser',
                employeeNumber: 'EMP001',
                userName: 'Test User',
                isPrivate: false,
            });

            store.resetAuth();

            expect(store.phoneNumber).toBe('13800138000');
            expect(store.userId).toBe('user-123');
        });
    });
});
