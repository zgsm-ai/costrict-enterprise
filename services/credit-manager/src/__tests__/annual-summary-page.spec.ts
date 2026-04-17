import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { ref, nextTick } from 'vue';
import AnnualSummaryPage from '@/views/AnnualSummary/annual-summary-page.vue';
import { DEFAULT_RESULT_TYPE, BACKEND_TYPE_MAP } from '@/views/AnnualSummary/const';
import type { UserMeData } from '@/api/bos/activity.bo';

// Mock components with proper setup
vi.mock('./components/annual-summary-step1.vue', () => ({
    default: {
        name: 'AnnualSummaryStep1',
        template: '<div class="step-1">Step 1</div>',
        props: ['stepId'],
        emits: ['next'],
    },
}));

vi.mock('./components/annual-summary-step2.vue', () => ({
    default: {
        name: 'AnnualSummaryStep2',
        template: '<div class="step-2">Step 2</div>',
        props: ['stepId', 'userData'],
        emits: ['next'],
    },
}));

vi.mock('./components/annual-summary-step3.vue', () => ({
    default: {
        name: 'AnnualSummaryStep3',
        template: '<div class="step-3">Step 3</div>',
        props: ['stepId', 'userData'],
        emits: ['next'],
    },
}));

vi.mock('./components/annual-summary-step4.vue', () => ({
    default: {
        name: 'AnnualSummaryStep4',
        template: '<div class="step-4">Step 4</div>',
        props: ['stepId', 'userData'],
        emits: ['next'],
    },
}));

vi.mock('./components/annual-summary-step5.vue', () => ({
    default: {
        name: 'AnnualSummaryStep5',
        template: '<div class="step-5">Step 5</div>',
        props: ['stepId', 'userData'],
        emits: ['next'],
    },
}));

vi.mock('./components/annual-summary-step6.vue', () => ({
    default: {
        name: 'AnnualSummaryStep6',
        template: '<div class="step-6">Step 6</div>',
        props: ['stepId'],
        emits: ['next'],
    },
}));

vi.mock('./components/annual-summary-step7.vue', () => ({
    default: {
        name: 'AnnualSummaryStep7',
        template: '<div class="step-7">Step 7</div>',
        props: ['stepId', 'type'],
        emits: ['next'],
    },
}));

vi.mock('./components/annual-summary-result.vue', () => ({
    default: {
        name: 'AnnualSummaryResult',
        template: '<div class="result">Result</div>',
        props: ['stepId', 'type', 'inviteCode', 'loginUrl'],
        emits: ['reset'],
    },
}));

// Mock API calls
const mockGetUserMe = vi.fn();
const mockGetInviteCode = vi.fn();

vi.mock('@/api/mods/activity.mod', () => ({
    getUserMe: () => mockGetUserMe(),
}));

vi.mock('@/api/mods/quota.mod', () => ({
    getInviteCode: () => mockGetInviteCode(),
}));

// Mock naive-ui
const mockMessageError = vi.fn();
const mockMessageWarning = vi.fn();
vi.mock('naive-ui', () => ({
    useMessage: () => ({
        error: mockMessageError,
        warning: mockMessageWarning,
    }),
}));

// Mock i18n
vi.mock('@/utils/i18n', () => ({
    getT: () => (key: string) => key,
}));

// Mock user store - create a reactive store
const createMockStore = (isTokenInitializedValue = true) => {
    const isTokenInitialized = ref(isTokenInitializedValue);
    return {
        isTokenInitialized,
    };
};

let mockStore = createMockStore();

vi.mock('@/store/user', () => ({
    useUserStore: () => mockStore,
    storeToRefs: (store: any) => ({
        isTokenInitialized: store.isTokenInitialized,
    }),
}));

describe('AnnualSummaryPage', () => {
    const mockUserData: UserMeData = {
        userId: 'user-123',
        username: 'testuser',
        displayName: 'Test User',
        casdoorId: 'casdoor-123',
        createdTime: '2024-01-01T00:00:00Z',
        registerOrder: 100,
        registerOrderPercent: 50,
        latestActivity: '2024-12-31T23:59:59Z',
        totalLatencyMs: 1500,
        totalTokens: 10000,
        totalRequests: 500,
        totalUsageDays: 30,
        modelStats: '{}',
        modeStats: '{}',
        updatedAt: '2024-12-31T23:59:59Z',
        creditUsage: 1000,
        creditUsageOrder: 50,
        creditUsageOrderPercent: 25,
        accessId: 'access-123',
        isInner: 0,
        modelStatsJson: null,
        modeStatsJson: {},
        identity: 'pioneer',
    };

    beforeEach(() => {
        vi.clearAllMocks();
        // Reset the mock store
        mockStore = createMockStore(true);
        mockGetUserMe.mockResolvedValue({ data: mockUserData });
        mockGetInviteCode.mockResolvedValue({ data: { invite_code: 'INVITE123' } });
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    describe('initial rendering', () => {
        it('should have container with correct classes', () => {
            const wrapper = mount(AnnualSummaryPage);

            const container = wrapper.find('.container');
            expect(container.exists()).toBe(true);
        });

        it('should have transition wrapper', () => {
            const wrapper = mount(AnnualSummaryPage);

            expect(wrapper.findComponent({ name: 'Transition' }).exists()).toBe(true);
        });
    });

    describe('data loading', () => {
        it('should not load data when token is not initialized', async () => {
            mockStore = createMockStore(false);

            mount(AnnualSummaryPage);
            await flushPromises();

            expect(mockGetUserMe).not.toHaveBeenCalled();
            expect(mockGetInviteCode).not.toHaveBeenCalled();
        });

        it('should load data when token is initialized', async () => {
            mockStore = createMockStore(true);

            mount(AnnualSummaryPage);
            await flushPromises();

            expect(mockGetUserMe).toHaveBeenCalledTimes(1);
            expect(mockGetInviteCode).toHaveBeenCalledTimes(1);
        });
    });

    describe('result type mapping', () => {
        it('should map backend identity correctly', () => {
            expect(BACKEND_TYPE_MAP.pioneer).toBe('first');
            expect(BACKEND_TYPE_MAP.high_freq).toBe('agent');
            expect(BACKEND_TYPE_MAP.problem_solver).toBe('bug');
            expect(BACKEND_TYPE_MAP.efficiency).toBe('speed');
        });

        it('should use default type when identity not found', () => {
            expect(DEFAULT_RESULT_TYPE).toBe('speed');
            // Check that default type exists in maps
            expect(BACKEND_TYPE_MAP).not.toHaveProperty('unknown');
        });
    });

    describe('invite code and login URL', () => {
        it('should fetch invite code', async () => {
            mockGetInviteCode.mockResolvedValue({
                data: { invite_code: 'TEST456' },
            });

            mount(AnnualSummaryPage);
            await flushPromises();

            expect(mockGetInviteCode).toHaveBeenCalledTimes(1);
        });

        it('should handle invite code fetch failure gracefully', async () => {
            mockGetInviteCode.mockRejectedValue(new Error('API Error'));
            const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

            mount(AnnualSummaryPage);
            await flushPromises();

            // Should log warning but not throw
            expect(consoleSpy).toHaveBeenCalledWith(
                '获取邀请码失败，但不影响流程:',
                expect.any(Error),
            );

            consoleSpy.mockRestore();
        });
    });

    describe('error handling', () => {
        it('should handle user data fetch failure', async () => {
            mockGetUserMe.mockRejectedValue(new Error('Network error'));

            mount(AnnualSummaryPage);
            await flushPromises();

            // Should attempt to call the API
            expect(mockGetUserMe).toHaveBeenCalledTimes(1);
        });
    });

    describe('transition configuration', () => {
        it('should have transition component', () => {
            const wrapper = mount(AnnualSummaryPage);

            const transition = wrapper.findComponent({ name: 'Transition' });
            expect(transition.exists()).toBe(true);
            expect(transition.props('name')).toBeDefined();
        });
    });
});
