import { createI18n } from 'vue-i18n';
import zh from './zh.json';
import en from './en.json';

// 从 localStorage 读取保存的语言设置，如果没有则使用默认值 'zh'
const getSavedLanguage = () => {
    try {
        const savedLanguage = localStorage.getItem('app-language');
        // 只支持 'zh' 和 'en' 两种语言，如果保存的值不是这两种，则使用默认值
        if (savedLanguage === 'zh' || savedLanguage === 'en') {
            return savedLanguage;
        }
    } catch (error) {
        console.warn('Failed to read language from localStorage:', error);
    }
    return 'zh'; // 默认语言
};

const i18n = createI18n({
    legacy: false,
    locale: getSavedLanguage(),
    fallbackLocale: 'en',
    messages: {
        zh,
        en,
    },
});

export default i18n;
