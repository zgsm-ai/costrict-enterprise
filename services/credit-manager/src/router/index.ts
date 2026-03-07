import { createRouter, createWebHistory } from 'vue-router';

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    scrollBehavior(to, from, savedPosition) {
        // 如果有保存的位置（如浏览器前进/后退），则恢复该位置
        if (savedPosition) {
            return savedPosition;
        }
        // 否则，总是滚动到顶部
        return new Promise((resolve) => {
            // 使用更强制的方式滚动到顶部
            setTimeout(() => {
                // 尝试滚动到内容区域的顶部
                const contentEl = document.querySelector('.content');
                if (contentEl) {
                    contentEl.scrollTop = 0;
                }

                // 同时也尝试窗口滚动（以防某些页面使用窗口滚动）
                window.scrollTo(0, 0);
                document.documentElement.scrollTop = 0;
                document.body.scrollTop = 0;

                resolve({ top: 0, left: 0 });
            }, 50); // 增加延迟以确保DOM已更新
        });
    },
    routes: [
        {
            path: '/',
            name: 'home',
            component: () => import('@/views/Home/home-page.vue'),
        },
        {
            path: '/login',
            name: 'login',
            component: () => import('@/views/Login/login-page.vue'),
        },
        {
            path: '/credits',
            name: 'credits',
            component: () => import('@/views/Credits/credits-page.vue'),
        },
        {
            path: '/md-preview',
            name: 'md-preview',
            component: () => import('@/views/Preview/MarkdownPreview.vue'),
        },
        {
            path: '/credit-md-preview',
            name: 'credit-md-preview',
            component: () => import('@/views/Credits/get-credit-md.vue'),
        },
        {
            path: '/credit-reward-plan',
            name: 'credit-reward-plan',
            component: () => import('@/views/Reward/credit-reward-plan.vue'),
        },
        {
            path: '/subscribe',
            name: 'subscribe',
            component: () => import('@/views/Subscribe/subscribe-page.vue'),
        },
        {
            path: '/annual-summary',
            name: 'annual-summary',
            component: () => import('@/views/AnnualSummary/annual-summary-page.vue'),
        },
        {
            path: '/annual-summary-cover',
            name: 'annual-summary-cover',
            component: () => import('@/views/AnnualSummary/annual-summary-cover.vue'),
        },
        {
            path: '/:pathMatch(.*)*',
            name: 'not-found',
            redirect: '/',
        },
    ],
});

export const PUBLIC_ROUTES = [
    '/credit-reward-plan',
    '/credit-md-preview',
    '/annual-summary-cover',
];

// 不显示 footer 的路由
export const NO_FOOTER_ROUTES = ['/annual-summary', '/annual-summary-cover'];

// 仅在中文版下可访问的路由
export const ZH_ONLY_ROUTES = ['/subscribe', '/annual-summary', '/annual-summary-cover'];

// 添加路由守卫，在英文版环境下屏蔽仅中文路由
router.beforeEach((to, from, next) => {
    // 获取当前语言设置
    let currentLocale = 'zh'; // 默认中文

    // 尝试从 localStorage 获取语言设置
    try {
        const savedLanguage = localStorage.getItem('app-language');
        if (savedLanguage === 'zh' || savedLanguage === 'en') {
            currentLocale = savedLanguage;
        }
    } catch (error) {
        console.warn('Failed to read language from localStorage:', error);
    }

    // 如果是英文版且访问仅中文路由，则重定向到首页
    if (currentLocale === 'en' && ZH_ONLY_ROUTES.includes(to.path)) {
        next('/');
    } else {
        next();
    }
});

export default router;
