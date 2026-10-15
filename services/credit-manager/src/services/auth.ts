import { tokenManager } from '@/utils/token';
import { useUserStore } from '@/store/user';
import { readJwtProfile } from '@/utils/jwtProfile';

export interface AuthResult {
    success: boolean;
    error?: string;
}

export class AuthService {
    async authenticate(): Promise<AuthResult> {
        const userStore = useUserStore();
        // 登录态只由共享 Cookie 是否存在决定，不兑换 state，也不请求用户信息。
        const token = tokenManager.getToken();
        const success = !!token;
        userStore.$patch({
            ...readJwtProfile(token),
            isTokenInitialized: success,
            isAuthenticating: false,
            authError: null,
        });
        return success ? { success: true } : { success: false, error: 'No token available' };
    }

    async isAuthenticated(): Promise<boolean> {
        return (await this.authenticate()).success;
    }

    async logout(): Promise<void> {
        tokenManager.clearToken();
        useUserStore().$reset();
    }
}

export const authService = new AuthService();
