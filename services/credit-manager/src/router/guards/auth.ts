import type { Router } from 'vue-router';
import { authService } from '@/services/auth';
import { PUBLIC_ROUTES } from '@/router';
import { tokenManager } from '@/utils/token';

const AUTH_REDIRECT_KEY = 'auth_redirect';

export function setupAuthGuard(router: Router) {
    router.beforeEach(async (to, from, next) => {
        try {
            // 处理年度总结封面页的特殊逻辑
            // if (to.path === '/annual-summary-cover') {
            //     const authResult = await authService.authenticate(router);
            //     if (authResult.success) {
            //         next('/annual-summary');
            //     } else {
            //         next();
            //     }
            //     return;
            // }

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
