import { describe, expect, it, vi } from 'vitest';
import { defineComponent, h } from 'vue';
import { mount } from '@vue/test-utils';
import { createMemoryHistory, createRouter } from 'vue-router';
import {
    CREDITS_SERVICE_CUTOFF_AT,
    hasCreditsServiceEnded,
    useCreditsServiceCutoff,
} from '@/composables/useCreditsServiceCutoff';

describe('Credits service cutoff', () => {
    it('keeps purchase available immediately before the Beijing-time cutoff', () => {
        expect(hasCreditsServiceEnded(CREDITS_SERVICE_CUTOFF_AT - 1)).toBe(false);
    });

    it('ends purchase at the exact Beijing-time cutoff', () => {
        expect(hasCreditsServiceEnded(CREDITS_SERVICE_CUTOFF_AT)).toBe(true);
    });

    it('uses 2026-09-23 23:59:00 in Beijing time', () => {
        expect(CREDITS_SERVICE_CUTOFF_AT).toBe(new Date('2026-09-23T23:59:00+08:00').getTime());
    });

    it('only bypasses the cutoff for explicit debug mode and restores it when removed', async () => {
        const page = defineComponent({
            setup() {
                const { isCreditsServiceEnded, isCreditsOrderDebug } = useCreditsServiceCutoff();
                return () => h('div', `${isCreditsServiceEnded.value}:${isCreditsOrderDebug.value}`);
            },
        });
        const router = createRouter({
            history: createMemoryHistory(),
            routes: [{ path: '/', component: page }],
        });
        const now = vi.spyOn(Date, 'now').mockReturnValue(CREDITS_SERVICE_CUTOFF_AT + 60_000);
        await router.push('/');
        const wrapper = mount(page, { global: { plugins: [router] } });
        try {
            expect(wrapper.text()).toBe('true:false');
            await router.push('/?preview=enabled');
            expect(wrapper.text()).toBe('false:true');
            await router.push('/');
            expect(wrapper.text()).toBe('true:false');
            await router.push('/?preview=unknown');
            expect(wrapper.text()).toBe('true:false');
            await router.push('/?preview=enabled&preview=disabled');
            expect(wrapper.text()).toBe('true:false');
            await router.push('/?preview=disabled');
            expect(wrapper.text()).toBe('true:false');
        } finally {
            wrapper.unmount();
            now.mockRestore();
        }
    });
});
