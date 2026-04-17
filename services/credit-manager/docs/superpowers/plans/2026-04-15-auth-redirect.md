# Auth Redirect Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 用户访问需要登录的页面时，完成 OIDC 登录后自动跳转回原始目标页面，而不是固定跳回首页。

**Architecture:** 认证守卫在重定向到 `/login` 前，将目标路径（`to.fullPath`）写入 `localStorage['auth_redirect']`。认证服务 `authenticate()` 成功后，读取该 key，若存在则清除并跳转，若不存在则保持现有行为。

**Tech Stack:** Vue 3, Vue Router 4, TypeScript, Vitest, happy-dom

---

## 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `src/router/guards/auth.ts` | Modify | 认证失败时写入 `localStorage['auth_redirect']` |
| `src/services/auth.ts` | Modify | 认证成功后读取、清除 `localStorage['auth_redirect']` 并跳转 |
| `src/__tests__/router.guards.auth.spec.ts` | Create | 守卫的单元测试 |
| `src/__tests__/services.auth.spec.ts` | Modify | 补充 redirect 相关测试用例 |

---

## Task 1: 为守卫新建测试文件并写失败测试

**Files:**
- Create: `src/__tests__/router.guards.auth.spec.ts`

- [ ] **Step 1: 创建测试文件，写入失败测试**

```typescript
// src/__tests__/router.guards.auth.spec.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { Router, RouteLocationNormalized } from 'vue-router';
import { setupAuthGuard } from '@/router/guards/auth';
import { authService } from '@/services/auth';
import { PUBLIC_ROUTES } from '@/router';

vi.mock('@/services/auth', () => ({
    authService: {
        isAuthenticated: vi.fn(),
        authenticate: vi.fn(),
    },
}));

vi.mock('@/router', () => ({
    PUBLIC_ROUTES: ['/credit-reward-plan', '/credit-md-preview'],
}));

vi.mock('@/utils/token', () => ({
    tokenManager: {
        cleanUrlState: vi.fn(),
    },
}));

function makeRoute(path: string, query: Record<string, string> = {}): RouteLocationNormalized {
    const search = Object.keys(query).length
        ? '?' + new URLSearchParams(query).toString()
        : '';
    return {
        path,
        fullPath: path + search,
        query,
        hash: '',
        name: undefined,
        params: {},
        matched: [],
        meta: {},
        redirectedFrom: undefined,
    } as unknown as RouteLocationNormalized;
}

function makeRouter(): Router {
    const guards: ((to: any, from: any, next: any) => void)[] = [];
    const afterGuards: ((to: any) => void)[] = [];
    return {
        beforeEach: vi.fn((fn) => guards.push(fn)),
        afterEach: vi.fn((fn) => afterGuards.push(fn)),
        replace: vi.fn(),
        _guards: guards,
        _afterGuards: afterGuards,
    } as unknown as Router;
}

async function runGuard(
    router: Router,
    to: RouteLocationNormalized,
    from: RouteLocationNormalized = makeRoute('/')
) {
    const next = vi.fn();
    await (router as any)._guards[0](to, from, next);
    return next;
}

describe('setupAuthGuard - auth_redirect', () => {
    let router: Router;

    beforeEach(() => {
        router = makeRouter();
        setupAuthGuard(router);
        localStorage.clear();
        vi.clearAllMocks();
    });

    it('认证失败时，将目标 fullPath 写入 localStorage["auth_redirect"]', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(false);
        vi.mocked(authService.authenticate).mockResolvedValue({ success: false });

        await runGuard(router, makeRoute('/credits'));

        expect(localStorage.getItem('auth_redirect')).toBe('/credits');
    });

    it('认证失败时，带 query 的 fullPath 完整写入', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(false);
        vi.mocked(authService.authenticate).mockResolvedValue({ success: false });

        await runGuard(router, makeRoute('/credits', { tab: 'history' }));

        expect(localStorage.getItem('auth_redirect')).toBe('/credits?tab=history');
    });

    it('访问 /login 时不写入 auth_redirect', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(false);

        await runGuard(router, makeRoute('/login'));

        expect(localStorage.getItem('auth_redirect')).toBeNull();
    });

    it('访问公开路由时不写入 auth_redirect', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(false);

        await runGuard(router, makeRoute('/credit-reward-plan'));

        expect(localStorage.getItem('auth_redirect')).toBeNull();
    });

    it('已登录时不写入 auth_redirect', async () => {
        vi.mocked(authService.isAuthenticated).mockResolvedValue(true);

        await runGuard(router, makeRoute('/credits'));

        expect(localStorage.getItem('auth_redirect')).toBeNull();
    });
});
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd D:/fontend/zgsm-admin-system
npx vitest run src/__tests__/router.guards.auth.spec.ts
```

期望：测试失败，错误类似 `localStorage.getItem('auth_redirect') to be '/credits'` received `null`

---

## Task 2: 修改认证守卫，使测试通过

**Files:**
- Modify: `src/router/guards/auth.ts`

- [ ] **Step 1: 修改 auth.ts，在认证失败时写入 localStorage**

将 `src/router/guards/auth.ts` 完整替换为：

```typescript
import type { Router } from 'vue-router';
import { authService } from '@/services/auth';
import { PUBLIC_ROUTES } from '@/router';
import { tokenManager } from '@/utils/token';

const AUTH_REDIRECT_KEY = 'auth_redirect';

export function setupAuthGuard(router: Router) {
    router.beforeEach(async (to, from, next) => {
        try {
            // 处理年度总结封面页的特殊逻辑
            if (to.path === '/annual-summary-cover') {
                const authResult = await authService.authenticate();
                if (authResult.success) {
                    next('/annual-summary');
                } else {
                    next();
                }
                return;
            }

            // 检查是否为公开路由或登录页面
            if (PUBLIC_ROUTES.includes(to.path) || to.path === '/login') {
                next();
                return;
            }

            // 检查是否已经认证过
            const isAuthenticated = await authService.isAuthenticated();

            if (isAuthenticated) {
                next();
                return;
            }

            // 对于非公开路由，先放行让页面渲染，然后在后台进行认证
            next();

            // 在后台进行认证，不阻塞页面渲染
            authService
                .authenticate()
                .then((authResult) => {
                    if (!authResult.success) {
                        // 记录目标路径，登录后跳回
                        localStorage.setItem(AUTH_REDIRECT_KEY, to.fullPath);
                        router.replace('/login');
                    }
                })
                .catch((error) => {
                    console.error('Background authentication error:', error);
                    localStorage.setItem(AUTH_REDIRECT_KEY, to.fullPath);
                    router.replace('/login');
                });
        } catch (error) {
            console.error('Authentication error:', error);
            localStorage.setItem(AUTH_REDIRECT_KEY, to.fullPath);
            next('/login');
        }
    });

    router.afterEach((to) => {
        if (to.query.state !== undefined) {
            tokenManager.cleanUrlState();
        }
    });
}
```

- [ ] **Step 2: 运行测试，确认通过**

```bash
cd D:/fontend/zgsm-admin-system
npx vitest run src/__tests__/router.guards.auth.spec.ts
```

期望：所有 5 个测试 PASS

- [ ] **Step 3: Commit**

```bash
cd D:/fontend/zgsm-admin-system
git add src/router/guards/auth.ts src/__tests__/router.guards.auth.spec.ts
git commit -m "feat: save auth redirect path to localStorage on auth failure"
```

---

## Task 3: 为认证服务补充失败测试用例

**Files:**
- Modify: `src/__tests__/services.auth.spec.ts`

- [ ] **Step 1: 在现有测试文件末尾，在 `authService singleton` describe 块之前，添加新的 describe 块**

在 `src/__tests__/services.auth.spec.ts` 的 `describe('authService singleton'` 之前插入：

```typescript
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
```

- [ ] **Step 2: 运行新增测试，确认失败**

```bash
cd D:/fontend/zgsm-admin-system
npx vitest run src/__tests__/services.auth.spec.ts
```

期望：新增的 3 个测试失败（`authenticate` 不接受 router 参数，也没有 redirect 逻辑）

---

## Task 4: 修改认证服务，使测试通过

**Files:**
- Modify: `src/services/auth.ts`

- [ ] **Step 1: 修改 auth.ts，authenticate 接受可选 router 参数，成功后处理 redirect**

将 `src/services/auth.ts` 完整替换为：

```typescript
import { tokenManager } from '@/utils/token';
import { userService, type UserInfo } from './user';
import { getUserToken } from '@/api/mods/quota.mod';
import type { Router } from 'vue-router';

const AUTH_REDIRECT_KEY = 'auth_redirect';

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

    async authenticate(router?: Router): Promise<AuthResult> {
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

            let result: AuthResult;
            if (hashToken) {
                result = await this.handleHashTokenAuth(hashToken);
            } else {
                result = await this.handleRegularAuth();
            }

            // 认证成功后处理跳转
            if (result.success && router) {
                const redirectPath = localStorage.getItem(AUTH_REDIRECT_KEY);
                if (redirectPath) {
                    localStorage.removeItem(AUTH_REDIRECT_KEY);
                    router.replace(redirectPath);
                }
            }

            return result;
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
```

- [ ] **Step 2: 运行全部认证服务测试，确认通过**

```bash
cd D:/fontend/zgsm-admin-system
npx vitest run src/__tests__/services.auth.spec.ts
```

期望：全部测试 PASS（包含原有测试 + 新增 3 个）

- [ ] **Step 3: Commit**

```bash
cd D:/fontend/zgsm-admin-system
git add src/services/auth.ts src/__tests__/services.auth.spec.ts
git commit -m "feat: redirect to original path after successful authentication"
```

---

## Task 5: 更新守卫以传入 router 给 authenticate

**Files:**
- Modify: `src/router/guards/auth.ts`

守卫在调用 `authService.authenticate()` 时需要传入 `router` 实例，让认证服务执行跳转。

- [ ] **Step 1: 更新守卫中的 authenticate 调用，传入 router**

将 `src/router/guards/auth.ts` 完整替换为：

```typescript
import type { Router } from 'vue-router';
import { authService } from '@/services/auth';
import { PUBLIC_ROUTES } from '@/router';
import { tokenManager } from '@/utils/token';

const AUTH_REDIRECT_KEY = 'auth_redirect';

export function setupAuthGuard(router: Router) {
    router.beforeEach(async (to, from, next) => {
        try {
            // 处理年度总结封面页的特殊逻辑
            if (to.path === '/annual-summary-cover') {
                const authResult = await authService.authenticate(router);
                if (authResult.success) {
                    next('/annual-summary');
                } else {
                    next();
                }
                return;
            }

            // 检查是否为公开路由或登录页面
            if (PUBLIC_ROUTES.includes(to.path) || to.path === '/login') {
                next();
                return;
            }

            // 检查是否已经认证过
            const isAuthenticated = await authService.isAuthenticated();

            if (isAuthenticated) {
                next();
                return;
            }

            // 对于非公开路由，先放行让页面渲染，然后在后台进行认证
            next();

            // 在后台进行认证，不阻塞页面渲染
            authService
                .authenticate(router)
                .then((authResult) => {
                    if (!authResult.success) {
                        localStorage.setItem(AUTH_REDIRECT_KEY, to.fullPath);
                        router.replace('/login');
                    }
                })
                .catch((error) => {
                    console.error('Background authentication error:', error);
                    localStorage.setItem(AUTH_REDIRECT_KEY, to.fullPath);
                    router.replace('/login');
                });
        } catch (error) {
            console.error('Authentication error:', error);
            localStorage.setItem(AUTH_REDIRECT_KEY, to.fullPath);
            next('/login');
        }
    });

    router.afterEach((to) => {
        if (to.query.state !== undefined) {
            tokenManager.cleanUrlState();
        }
    });
}
```

- [ ] **Step 2: 运行全部相关测试**

```bash
cd D:/fontend/zgsm-admin-system
npx vitest run src/__tests__/router.guards.auth.spec.ts src/__tests__/services.auth.spec.ts
```

期望：全部 PASS

- [ ] **Step 3: 运行完整测试套件，确认无回归**

```bash
cd D:/fontend/zgsm-admin-system
npx vitest run
```

期望：全部 PASS

- [ ] **Step 4: Commit**

```bash
cd D:/fontend/zgsm-admin-system
git add src/router/guards/auth.ts
git commit -m "feat: pass router to authenticate for post-login redirect"
```

---

## 自检结果

- **Spec 覆盖：** ✓ 守卫写入 / 认证服务读取清除跳转 / 边界情况（/login、PUBLIC_ROUTES、已登录）均有测试
- **Placeholder：** 无
- **类型一致性：** `AUTH_REDIRECT_KEY` 常量在两个文件中独立声明（避免循环依赖）；`authenticate(router?: Router)` 签名在 Task 3 测试和 Task 4 实现中一致
- **范围：** 聚焦两个文件，无额外重构
