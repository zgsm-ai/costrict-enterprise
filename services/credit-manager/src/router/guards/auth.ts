import type { Router } from 'vue-router';
import { authService } from '@/services/auth';
import { PUBLIC_ROUTES } from '@/router';
import { saveLoginReturnTo } from '@/utils/login';

export function setupAuthGuard(router: Router): void {
    router.beforeEach(async (to, from) => {
        const authenticated = await authService.isAuthenticated();
        if (PUBLIC_ROUTES.includes(to.path) || to.path === '/login' || authenticated) {
            return true;
        }

        // 首次进入保留原始地址（包括 state、tab 和 hash），站内导航使用目标地址。
        const returnTo =
            from.matched.length === 0
                ? window.location.href
                : new URL(router.resolve(to.fullPath).href, window.location.origin).href;
        saveLoginReturnTo(returnTo);
        return { path: '/login' };
    });
}
