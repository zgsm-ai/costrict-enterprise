/**
 * @file meta.ts
 * 社交平台分享预览配置文件
 */

export interface MetaConfig {
    title: string;
    description: string;
    image: string;
    url?: string;
    type?: string;
    siteName?: string;
    locale?: string;
}

export interface SocialMetaConfig {
    og: MetaConfig;
    twitter: {
        card: 'summary' | 'summary_large_image';
        title: string;
        description: string;
        image: string;
    };
}

/**
 * 通用的社交平台分享配置
 */
export const DEFAULT_SOCIAL_META: SocialMetaConfig = {
    og: {
        title: 'CoStrict-开源免费 AI 编程工具',
        description: '为企业严肃编程量身打造的AI编程助手，支持多种编程语言和框架',
        image: '/social-share.png',
        type: 'website',
        siteName: 'CoStrict',
        locale: 'zh_CN',
    },
    twitter: {
        card: 'summary_large_image',
        title: 'CoStrict-开源免费 AI 编程工具',
        description: '为企业严肃编程量身打造的AI编程助手，支持多种编程语言和框架',
        image: '/social-share.png',
    },
};

/**
 * 积分奖励计划页面的社交平台分享配置
 */
export const CREDIT_REWARD_PLAN_SOCIAL_META: SocialMetaConfig = {
    og: {
        title: 'CoStrict 喊你来薅羊毛啦（送 Credits）！',
        description:
            'CoStrict-开源免费 AI 编程工具，为企业严肃编程量身打造。邀请好友注册，双方均可获得Credits奖励！',
        image: '/social-share.png',
        type: 'website',
        siteName: 'CoStrict',
        locale: 'zh_CN',
    },
    twitter: {
        card: 'summary_large_image',
        title: 'CoStrict 喊你来薅羊毛啦（送 Credits）！',
        description:
            'CoStrict-开源免费 AI 编程工具，为企业严肃编程量身打造。邀请好友注册，双方均可获得Credits奖励！',
        image: '/social-share.png',
    },
};

/**
 * 页面meta配置映射
 */
export const PAGE_META_CONFIG: Record<string, SocialMetaConfig> = {
    'credit-reward-plan': CREDIT_REWARD_PLAN_SOCIAL_META,
};

/**
 * Meta标签接口定义
 */
export interface MetaTag {
    name?: string;
    property?: string;
    content: string;
}

/**
 * Head标签接口定义
 */
export interface HeadTags {
    title?: string;
    meta: MetaTag[];
}

/**
 * 生成完整的meta标签对象
 * @param config 社交平台配置
 * @param currentUrl 当前页面URL
 * @returns 完整的meta标签对象
 */
export const generateMetaTags = (config: SocialMetaConfig, currentUrl: string): HeadTags => {
    const fullConfig = {
        ...config,
        og: {
            ...config.og,
            url: currentUrl,
        },
    };

    return {
        // title: fullConfig.og.title,
        meta: [
            // Open Graph 标签
            { property: 'og:title', content: fullConfig.og.title },
            { property: 'og:description', content: fullConfig.og.description },
            { property: 'og:image', content: fullConfig.og.image },
            { property: 'og:url', content: fullConfig.og.url },
            ...(fullConfig.og.type ? [{ property: 'og:type', content: fullConfig.og.type }] : []),
            ...(fullConfig.og.siteName
                ? [{ property: 'og:site_name', content: fullConfig.og.siteName }]
                : []),
            ...(fullConfig.og.locale
                ? [{ property: 'og:locale', content: fullConfig.og.locale }]
                : []),

            // Twitter Card 标签
            { name: 'twitter:card', content: fullConfig.twitter.card },
            { name: 'twitter:title', content: fullConfig.twitter.title },
            { name: 'twitter:description', content: fullConfig.twitter.description },
            { name: 'twitter:image', content: fullConfig.twitter.image },

            // 其他优化标签
            { name: 'description', content: fullConfig.og.description },
            { name: 'keywords', content: 'AI编程,代码生成,编程助手,开源工具,Credits' },
        ],
    };
};
