import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
    getHashToken,
    getToken,
    setToken,
    clearToken,
    cleanUrlState,
    tokenManager,
} from '@/utils/token';

// Mock js-cookie
vi.mock('js-cookie', () => ({
    default: {
        get: vi.fn(),
        set: vi.fn(),
        remove: vi.fn(),
    },
}));

import Cookies from 'js-cookie';

describe('token utils', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    describe('getHashToken', () => {
        it('should return state parameter from URL', () => {
            // Mock window.location using Object.defineProperty
            const originalHref = window.location.href;
            Object.defineProperty(window, 'location', {
                value: { href: 'https://example.com?state=test-token-123' },
                writable: true,
            });

            const result = getHashToken();

            expect(result).toBe('test-token-123');

            // Restore original location
            Object.defineProperty(window, 'location', {
                value: { href: originalHref },
                writable: true,
            });
        });

        it('should return null when state parameter is not present', () => {
            Object.defineProperty(window, 'location', {
                value: { href: 'https://example.com?other=value' },
                writable: true,
            });

            const result = getHashToken();

            expect(result).toBeNull();
        });
    });

    describe('getToken', () => {
        it('should get token from cookies', () => {
            const mockToken = 'mock-token-123';
            (Cookies.get as ReturnType<typeof vi.fn>).mockReturnValue(mockToken);

            const result = getToken();

            expect(Cookies.get).toHaveBeenCalledWith('zgsmAdminToken');
            expect(result).toBe(mockToken);
        });

        it('should return undefined when token does not exist', () => {
            (Cookies.get as ReturnType<typeof vi.fn>).mockReturnValue(undefined);

            const result = getToken();

            expect(result).toBeUndefined();
        });
    });

    describe('setToken', () => {
        it('should set token in cookies', () => {
            const token = 'new-token-456';

            setToken(token);

            expect(Cookies.set).toHaveBeenCalledWith('zgsmAdminToken', token);
        });
    });

    describe('clearToken', () => {
        it('should remove token from cookies', () => {
            clearToken();

            expect(Cookies.remove).toHaveBeenCalledWith('zgsmAdminToken');
        });
    });

    describe('cleanUrlState', () => {
        it('should remove state parameter from URL', () => {
            const replaceStateMock = vi.fn();

            Object.defineProperty(window, 'location', {
                value: {
                    href: 'https://example.com?state=test&other=value',
                    origin: 'https://example.com',
                    pathname: '/',
                    search: '?state=test&other=value',
                    hash: '',
                },
                writable: true,
            });

            Object.defineProperty(window, 'history', {
                value: {
                    replaceState: replaceStateMock,
                },
                writable: true,
            });

            cleanUrlState();

            expect(replaceStateMock).toHaveBeenCalled();
        });

        it('should not modify URL when state parameter is not present', () => {
            const replaceStateMock = vi.fn();

            Object.defineProperty(window, 'location', {
                value: {
                    href: 'https://example.com?other=value',
                    origin: 'https://example.com',
                    pathname: '/',
                    search: '?other=value',
                    hash: '',
                },
                writable: true,
            });

            Object.defineProperty(window, 'history', {
                value: {
                    replaceState: replaceStateMock,
                },
                writable: true,
            });

            cleanUrlState();

            expect(replaceStateMock).not.toHaveBeenCalled();
        });
    });

    describe('tokenManager', () => {
        describe('validateToken', () => {
            it('should return true for non-empty string token', () => {
                expect(tokenManager.validateToken('valid-token')).toBe(true);
            });

            it('should return false for empty string', () => {
                expect(tokenManager.validateToken('')).toBe(false);
            });

            it('should return false for null', () => {
                expect(tokenManager.validateToken(null as unknown as string)).toBe(false);
            });

            it('should return false for undefined', () => {
                expect(tokenManager.validateToken(undefined as unknown as string)).toBe(false);
            });
        });
    });
});
