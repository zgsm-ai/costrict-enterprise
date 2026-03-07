/**
 * @description home页面菜单hook
 */
import { ref, watch, h, computed, type ComputedRef } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { type MenuOption } from 'naive-ui';
import MenuIcons from '../components/menu-icons.vue';
import newIcons from '@/assets/new.svg';

// 定义菜单键的类型
export type MenuKey = 'profile' | 'subscription' | 'usage' | 'activity';

export function useMenu() {
    const { t, locale } = useI18n();
    const route = useRoute();
    const router = useRouter();

    // 菜单相关数据
    const activeMenuKey = ref<MenuKey>('profile');

    // 为每个菜单项添加加载状态
    const menuLoadingStates = ref<Record<MenuKey, boolean>>({
        profile: false,
        subscription: false,
        usage: false,
        activity: false,
    });

    // 存储每个菜单项是否已加载过，避免重复请求
    const menuLoadedStates = ref<Record<MenuKey, boolean>>({
        profile: false,
        subscription: false,
        usage: false,
        activity: false,
    });

    // 从URL参数初始化菜单状态
    const initMenuFromRoute = () => {
        const tab = route.query.tab as string;
        const availableTabs =
            locale.value === 'zh'
                ? ['profile', 'subscription', 'usage', 'usage-consumption', 'activity']
                : ['profile', 'usage', 'usage-consumption', 'activity'];

        if (tab && availableTabs.includes(tab)) {
            activeMenuKey.value = tab as MenuKey;
        }
    };

    // 监听路由变化
    watch(
        () => route.query.tab,
        (newTab) => {
            const availableTabs =
                locale.value === 'zh'
                    ? ['profile', 'subscription', 'usage', 'usage-consumption', 'activity']
                    : ['profile', 'usage', 'usage-consumption', 'activity'];

            if (newTab && availableTabs.includes(newTab as string)) {
                activeMenuKey.value = newTab as MenuKey;
            }
        },
    );

    // 监听语言变化，当从英文切换到中文时，如果当前没有选中菜单项，则初始化菜单状态
    watch(locale, () => {
        initMenuFromRoute();
    });

    // 监听语言和路由变化，如果用户在英文环境下尝试访问订阅页面，重定向到profile页面
    watch(
        [locale, () => route.query.tab],
        () => {
            if (locale.value === 'en' && route.query.tab === 'subscription') {
                // 使用 history API 重定向到 profile 页面
                const newUrl = `${window.location.origin}${window.location.pathname}?tab=profile`;
                window.history.replaceState({}, '', newUrl);

                // 手动更新 activeMenuKey，确保组件状态正确
                activeMenuKey.value = 'profile';
            }
        },
        { immediate: true },
    );

    // 渲染菜单图标
    const renderMenuIcon = (name: MenuKey) => {
        return () => h(MenuIcons, { name });
    };

    // 菜单选项
    const menuOptions: ComputedRef<MenuOption[]> = computed(() => {
        const baseOptions = [
            {
                label: t('homePage.menu.profile'),
                key: 'profile',
                icon: renderMenuIcon('profile'),
            },
            {
                label: t('homePage.menu.usage'),
                key: 'usage',
                icon: renderMenuIcon('usage'),
            },
            {
                label: () =>
                    h('div', { class: 'menu-label-with-new-icon flex' }, [
                        h('span', t('homePage.menu.activity')),
                        h('img', {
                            src: newIcons,
                            alt: 'new',
                            class: 'new-icon ml-3 w-8',
                        }),
                    ]),
                key: 'activity',
                icon: renderMenuIcon('activity'),
            },
        ];

        // 只有在中文版时才显示订阅菜单项
        if (locale.value === 'zh') {
            baseOptions.splice(1, 0, {
                label: t('homePage.menu.subscription'),
                key: 'subscription',
                icon: renderMenuIcon('subscription'),
            });
        }

        return baseOptions;
    });

    // 处理菜单选择
    const handleMenuSelect = (key: string) => {
        activeMenuKey.value = key as MenuKey;
        // 更新URL参数，同时清除state参数
        const filteredQuery = Object.fromEntries(
            Object.entries(route.query).filter(([param]) => param !== 'state'),
        );
        router.replace({
            query: { ...filteredQuery, tab: key },
        });
    };

    // 根据菜单设置加载状态
    const setMenuLoading = (menuKey: MenuKey, loading: boolean) => {
        menuLoadingStates.value[menuKey] = loading;
    };

    // 标记菜单为已加载
    const markMenuAsLoaded = (menuKey: MenuKey) => {
        menuLoadedStates.value[menuKey] = true;
    };

    // 检查菜单是否已加载
    const isMenuLoaded = (menuKey: MenuKey) => {
        return menuLoadedStates.value[menuKey];
    };

    // 重置usage菜单的加载状态（允许根据时间范围重新加载）
    const resetUsageMenuLoadedState = () => {
        menuLoadedStates.value.usage = false;
    };

    // 初始化菜单状态
    initMenuFromRoute();

    return {
        activeMenuKey,
        menuOptions,
        handleMenuSelect,
        menuLoadingStates,
        menuLoadedStates,
        setMenuLoading,
        markMenuAsLoaded,
        isMenuLoaded,
        resetUsageMenuLoadedState,
    };
}
