/**
 * @file useMetaTags.ts
 * 动态meta标签管理composable
 */

import { watch, onMounted, type Ref } from 'vue';
import { useRoute } from 'vue-router';
import { useHead } from '@vueuse/head';
import {
    DEFAULT_SOCIAL_META,
    PAGE_META_CONFIG,
    generateMetaTags,
    type SocialMetaConfig,
} from '@/config/meta';

/**
 * 动态meta标签管理composable
 * @param customMeta 自定义meta配置（可选）
 * @returns meta标签管理对象
 */
export const useMetaTags = (customMeta?: Ref<SocialMetaConfig>) => {
    const route = useRoute();

    /**
     * 获取当前页面的meta配置
     * @returns 当前页面的meta配置
     */
    const getCurrentPageMeta = (): SocialMetaConfig => {
        // 如果有自定义meta配置，优先使用
        if (customMeta?.value) {
            return customMeta.value;
        }

        // 根据路由名称获取对应的meta配置
        const routeName = route.name as string;
        return PAGE_META_CONFIG[routeName] || DEFAULT_SOCIAL_META;
    };

    /**
     * 更新页面的meta标签
     */
    const updateMetaTags = () => {
        const currentUrl = window.location.href;
        const metaConfig = getCurrentPageMeta();
        const headTags = generateMetaTags(metaConfig, currentUrl);

        // 转换为@vueuse/head兼容的格式
        const headObject = {
            title: headTags.title,
            meta: headTags.meta.map((tag) => {
                if (tag.property) {
                    return { property: tag.property, content: tag.content };
                }
                return { name: tag.name, content: tag.content };
            }),
        };

        useHead(headObject);
    };

    /**
     * 设置自定义meta配置
     * @param meta 自定义meta配置
     */
    const setCustomMeta = (meta: SocialMetaConfig) => {
        if (customMeta) {
            customMeta.value = meta;
            updateMetaTags();
        }
    };

    // 监听路由变化，自动更新meta标签
    watch(
        () => route.name,
        () => {
            updateMetaTags();
        },
        { immediate: true },
    );

    // 监听自定义meta配置变化
    if (customMeta) {
        watch(
            customMeta,
            () => {
                updateMetaTags();
            },
            { deep: true },
        );
    }

    // 组件挂载时初始化meta标签
    onMounted(() => {
        updateMetaTags();
    });

    return {
        updateMetaTags,
        setCustomMeta,
        getCurrentPageMeta,
    };
};

/**
 * 为特定页面设置meta标签的便捷函数
 * @param pageName 页面名称
 * @param additionalConfig 额外的meta配置（可选）
 */
export const setPageMeta = (pageName: string, additionalConfig?: Partial<SocialMetaConfig>) => {
    const baseConfig = PAGE_META_CONFIG[pageName] || DEFAULT_SOCIAL_META;
    const mergedConfig = {
        og: { ...baseConfig.og, ...additionalConfig?.og },
        twitter: { ...baseConfig.twitter, ...additionalConfig?.twitter },
    };

    const currentUrl = window.location.href;
    const headTags = generateMetaTags(mergedConfig, currentUrl);

    // 转换为@vueuse/head兼容的格式
    const headObject = {
        title: headTags.title,
        meta: headTags.meta.map((tag) => {
            if (tag.property) {
                return { property: tag.property, content: tag.content };
            }
            return { name: tag.name, content: tag.content };
        }),
    };

    useHead(headObject);
};
