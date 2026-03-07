import type { Router } from 'vue-router';
import { authService } from '@/services/auth';
import { PUBLIC_ROUTES } from '@/router';
import { tokenManager } from '@/utils/token';

export function setupAuthGuard(router: Router) {
    router.beforeEach(async (to, from, next) => {
        try {
            // 处理年度总结封面页的特殊逻辑
            if (to.path === '/annual-summary-cover') {
                // 先尝试认证（包括从 URL 获取 token 等逻辑）
                const authResult = await authService.authenticate();

                if (authResult.success) {
                    // 认证成功，重定向到年度总结页面
                    next('/annual-summary');
                } else {
                    // 认证失败，让页面正常加载（静态封面页）
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
                // 已经认证，直接放行
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
                        // 认证失败，重定向到登录页
                        router.replace('/login');
                    }
                })
                .catch((error) => {
                    console.error('Background authentication error:', error);
                    router.replace('/login');
                });
        } catch (error) {
            console.error('Authentication error:', error);
            next('/login');
        }
    });

    router.afterEach((to) => {
        // 在路由导航完成后，检查并清理URL中的state参数
        // 这确保了即使在其他地方有导航操作，state参数也会被清理
        if (to.query.state !== undefined) {
            tokenManager.cleanUrlState();
        }
    });
}
