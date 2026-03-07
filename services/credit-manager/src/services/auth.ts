import { tokenManager } from '@/utils/token';
import { userService, type UserInfo } from './user';
import { getUserToken } from '@/api/mods/quota.mod';

export interface AuthResult {
    success: boolean;
    user?: UserInfo;
    error?: string;
}

export class AuthService {
    private isAuthenticating = false;
    private userStore: ReturnType<typeof import('@/store/user').useUserStore> | null = null;

    // 延迟获取用户store，确保Pinia已经初始化
    private async getUserStore() {
        if (!this.userStore) {
            const { useUserStore } = await import('@/store/user');
            this.userStore = useUserStore();
        }
        return this.userStore;
    }

    async authenticate(): Promise<AuthResult> {
        // 避免重复认证
        if (this.isAuthenticating) {
            return { success: false, error: 'Authentication in progress' };
        }

        this.isAuthenticating = true;

        try {
            // 更新store中的认证状态
            const userStore = await this.getUserStore();
            userStore.setAuthenticating(true);
            userStore.setAuthError(null);

            // 1. 检查 hashToken
            const hashToken = tokenManager.getHashToken();

            if (hashToken) {
                // 处理 hashToken 认证流程
                return await this.handleHashTokenAuth(hashToken);
            } else {
                // 处理常规认证流程
                return await this.handleRegularAuth();
            }
        } catch (error) {
            console.error('Authentication failed:', error);
            const userStore = await this.getUserStore();
            userStore.setAuthError('Authentication failed');
            return { success: false, error: 'Authentication failed' };
        } finally {
            this.isAuthenticating = false;
            const userStore = await this.getUserStore();
            userStore.setAuthenticating(false);
        }
    }

    private async handleHashTokenAuth(hashToken: string): Promise<AuthResult> {
        try {
            // 设置 hashToken
            tokenManager.setToken(hashToken);

            // 获取用户 token
            const tokenResponse = await getUserToken();

            if (tokenResponse.data?.access_token) {
                // 设置新的 access_token
                tokenManager.setToken(tokenResponse.data.access_token);

                // 获取用户信息
                const userInfo = await userService.fetchUserInfo();

                // 更新用户状态
                const userStore = await this.getUserStore();
                userStore.updateUserInfo(userInfo);
                userStore.updateTokenInitialized(true);

                // 清理 URL 中的 state 参数（在用户状态更新后再清理，确保所有异步操作完成）
                setTimeout(() => {
                    tokenManager.cleanUrlState();
                }, 0);

                return { success: true, user: userInfo };
            }

            return { success: false, error: 'Failed to get access token' };
        } catch (error) {
            console.error('Hash token authentication failed:', error);
            const userStore = await this.getUserStore();
            userStore.setAuthError('Hash token authentication failed');
            return { success: false, error: 'Hash token authentication failed' };
        }
    }

    private async handleRegularAuth(): Promise<AuthResult> {
        try {
            // 检查现有 token
            const existingToken = tokenManager.getToken();

            if (!existingToken) {
                return { success: false, error: 'No token available' };
            }

            // 获取用户信息
            const userInfo = await userService.fetchUserInfo();

            // 更新用户状态
            const userStore = await this.getUserStore();
            userStore.updateUserInfo(userInfo);
            userStore.updateTokenInitialized(true);

            return { success: true, user: userInfo };
        } catch {
            const userStore = await this.getUserStore();
            userStore.setAuthError('Regular authentication failed');
            return { success: false, error: 'Regular authentication failed' };
        }
    }

    async isAuthenticated(): Promise<boolean> {
        try {
            const userStore = await this.getUserStore();
            return userStore.isTokenInitialized && !!tokenManager.getToken();
        } catch {
            return false;
        }
    }

    async logout(): Promise<void> {
        try {
            tokenManager.clearToken();
            const userStore = await this.getUserStore();
            userStore.updateTokenInitialized(false);
        } catch {
            console.error('Logout error');
        }
    }
}

export const authService = new AuthService();
