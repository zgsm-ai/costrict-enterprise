import type { Composer } from 'vue-i18n';

let i18nComposer: Composer | null = null;

export const setI18nComposer = (composer: Composer) => {
    i18nComposer = composer;
};

export const getT = () => {
    if (!i18nComposer) {
        console.warn('i18n Composer is not available. Please ensure it is set up in App.vue.');
        return (key: string) => key;
    }
    return i18nComposer.t;
};
