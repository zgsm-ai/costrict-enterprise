import { getUserInfo } from '@/api/mods/quota.mod';

export interface UserInfo {
    phoneNumber: string;
    userId: string;
    employeeNumber: string;
    githubName: string;
    userName: string;
    isPrivate: boolean;
}

export class UserService {
    async fetchUserInfo(): Promise<UserInfo> {
        try {
            const response = await getUserInfo();

            if (response.data) {
                return {
                    phoneNumber: response.data.phone || '',
                    userId: response.data.uuid || '',
                    employeeNumber: response.data.employee_number || '',
                    githubName: response.data.githubName || '',
                    userName: response.data.username || '',
                    isPrivate: response.data.isPrivate || false,
                };
            }

            throw new Error('Failed to fetch user info');
        } catch (error) {
            console.error('Failed to fetch user info:', error);
            throw error;
        }
    }
}

export const userService = new UserService();
