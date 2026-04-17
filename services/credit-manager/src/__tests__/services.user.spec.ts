import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { userService, UserService } from '@/services/user';
import { getUserInfo } from '@/api/mods/quota.mod';

// Mock the API module
vi.mock('@/api/mods/quota.mod', () => ({
    getUserInfo: vi.fn(),
}));

const mockedGetUserInfo = vi.mocked(getUserInfo);

describe('UserService', () => {
    let service: UserService;

    beforeEach(() => {
        service = new UserService();
        vi.clearAllMocks();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    describe('fetchUserInfo', () => {
        it('should fetch and transform user info successfully', async () => {
            const mockResponse = {
                data: {
                    phone: '13800138000',
                    uuid: 'user-123',
                    employee_number: 'EMP001',
                    githubName: 'testuser',
                    username: 'Test User',
                    isPrivate: true,
                },
            };
            mockedGetUserInfo.mockResolvedValue(mockResponse as any);

            const result = await service.fetchUserInfo();

            expect(mockedGetUserInfo).toHaveBeenCalledTimes(1);
            expect(result).toEqual({
                phoneNumber: '13800138000',
                userId: 'user-123',
                employeeNumber: 'EMP001',
                githubName: 'testuser',
                userName: 'Test User',
                isPrivate: true,
            });
        });

        it('should handle missing fields with empty string defaults', async () => {
            const mockResponse = {
                data: {
                    phone: null,
                    uuid: null,
                    employee_number: null,
                    githubName: null,
                    username: null,
                    isPrivate: false,
                },
            };
            mockedGetUserInfo.mockResolvedValue(mockResponse as any);

            const result = await service.fetchUserInfo();

            expect(result).toEqual({
                phoneNumber: '',
                userId: '',
                employeeNumber: '',
                githubName: '',
                userName: '',
                isPrivate: false,
            });
        });

        it('should handle partial data', async () => {
            const mockResponse = {
                data: {
                    phone: '13800138000',
                    uuid: 'user-123',
                },
            };
            mockedGetUserInfo.mockResolvedValue(mockResponse as any);

            const result = await service.fetchUserInfo();

            expect(result).toEqual({
                phoneNumber: '13800138000',
                userId: 'user-123',
                employeeNumber: '',
                githubName: '',
                userName: '',
                isPrivate: false,
            });
        });

        it('should throw error when API returns no data', async () => {
            mockedGetUserInfo.mockResolvedValue({ data: null } as any);

            await expect(service.fetchUserInfo()).rejects.toThrow('Failed to fetch user info');
        });

        it('should throw error when API call fails', async () => {
            const error = new Error('Network error');
            mockedGetUserInfo.mockRejectedValue(error);

            await expect(service.fetchUserInfo()).rejects.toThrow('Network error');
        });
    });
});

describe('userService singleton', () => {
    it('should be an instance of UserService', () => {
        expect(userService).toBeInstanceOf(UserService);
    });
});
