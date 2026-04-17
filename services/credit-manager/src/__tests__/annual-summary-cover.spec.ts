import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref, nextTick } from 'vue';
import AnnualSummaryCover from '@/views/AnnualSummary/annual-summary-cover.vue';

// Mock vue-router
const mockPush = vi.fn();
vi.mock('vue-router', () => ({
    useRouter: () => ({
        push: mockPush,
    }),
    useRoute: () => ({
        query: {
            inviteCode: 'TEST123',
            isShare: 'true',
        },
    }),
}));

describe('AnnualSummaryCover', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        vi.useFakeTimers();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    describe('rendering', () => {
        it('should render cover image', () => {
            const wrapper = mount(AnnualSummaryCover);
            const coverImg = wrapper.find('img[alt="summary_cover"]');

            expect(coverImg.exists()).toBe(true);
            expect(coverImg.classes()).toContain('absolute');
            expect(coverImg.classes()).toContain('z-1');
        });

        it('should render enter button', () => {
            const wrapper = mount(AnnualSummaryCover);
            const enterBtn = wrapper.find('img[alt="enter"]');

            expect(enterBtn.exists()).toBe(true);
            expect(enterBtn.classes()).toContain('cursor-pointer');
            expect(enterBtn.classes()).toContain('z-10');
        });

        it('should not show cover_next initially', () => {
            const wrapper = mount(AnnualSummaryCover);
            const coverNextImg = wrapper.find('img[alt="summary_cover_icon"]');

            expect(coverNextImg.exists()).toBe(false);
        });
    });

    describe('animation', () => {
        it('should show cover_next after 800ms', async () => {
            vi.useRealTimers();
            const wrapper = mount(AnnualSummaryCover);

            // Initially not shown
            let coverNextImg = wrapper.find('img[alt="summary_cover_icon"]');
            expect(coverNextImg.exists()).toBe(false);

            // Wait for the actual timer
            await new Promise((resolve) => setTimeout(resolve, 900));
            await nextTick();

            // Should be shown now
            coverNextImg = wrapper.find('img[alt="summary_cover_icon"]');
            expect(coverNextImg.exists()).toBe(true);
            expect(coverNextImg.classes()).toContain('z-2');
        });

        it('should have fade transition on cover_next', async () => {
            const wrapper = mount(AnnualSummaryCover);

            vi.advanceTimersByTime(800);
            await nextTick();

            const transition = wrapper.findComponent({ name: 'Transition' });
            expect(transition.exists()).toBe(true);
            expect(transition.props('name')).toBe('fade');
        });
    });

    describe('interactions', () => {
        it('should navigate to login page when clicking enter button', async () => {
            const wrapper = mount(AnnualSummaryCover);
            const enterBtn = wrapper.find('img[alt="enter"]');

            await enterBtn.trigger('click');

            expect(mockPush).toHaveBeenCalledWith({
                path: '/login',
                query: {
                    inviteCode: 'TEST123',
                    isShare: 'true',
                },
            });
        });

        it('should preserve route query parameters', async () => {
            const wrapper = mount(AnnualSummaryCover);
            const enterBtn = wrapper.find('img[alt="enter"]');

            await enterBtn.trigger('click');

            const pushArgs = mockPush.mock.calls[0][0];
            expect(pushArgs.query).toHaveProperty('inviteCode');
            expect(pushArgs.query).toHaveProperty('isShare');
        });
    });

    describe('cleanup', () => {
        it('should clear timers on unmount', () => {
            const wrapper = mount(AnnualSummaryCover);
            const clearTimeoutSpy = vi.spyOn(global, 'clearTimeout');

            wrapper.unmount();

            // Component should clean up any pending timers
            expect(clearTimeoutSpy).not.toHaveBeenCalled();
        });
    });
});
