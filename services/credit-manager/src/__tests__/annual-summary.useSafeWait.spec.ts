import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useSafeWait } from '@/views/AnnualSummary/hooks/useSafeWait';
import { ref, nextTick } from 'vue';

// Mock @vueuse/core
const mockIsMounted = ref(true);
vi.mock('@vueuse/core', () => ({
    promiseTimeout: (ms: number) => new Promise((resolve) => setTimeout(resolve, ms)),
    useMounted: () => mockIsMounted,
}));

describe('useSafeWait', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        vi.useFakeTimers();
        mockIsMounted.value = true;
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    describe('safeWait', () => {
        it('should wait for specified milliseconds when mounted', async () => {
            const { safeWait } = useSafeWait();
            const timeoutMs = 100;

            // Use real timers for this test
            vi.useRealTimers();

            const startTime = Date.now();
            await safeWait(timeoutMs);
            const endTime = Date.now();

            // Should wait at least timeoutMs
            expect(endTime - startTime).toBeGreaterThanOrEqual(timeoutMs - 10);

            // Restore fake timers
            vi.useFakeTimers();
        });

        it('should throw UNMOUNTED_ERROR when component is unmounted', async () => {
            const { safeWait } = useSafeWait();

            const waitPromise = safeWait(100);

            // Simulate unmount
            mockIsMounted.value = false;

            vi.advanceTimersByTime(100);

            await expect(waitPromise).rejects.toThrow();
        });

        it('should complete successfully when still mounted', async () => {
            const { safeWait } = useSafeWait();

            const waitPromise = safeWait(500);
            vi.advanceTimersByTime(500);

            await expect(waitPromise).resolves.not.toThrow();
        });
    });

    describe('runSafe', () => {
        it('should execute function successfully when mounted', async () => {
            const { runSafe } = useSafeWait();
            const mockFn = vi.fn().mockResolvedValue(undefined);

            await runSafe(mockFn);

            expect(mockFn).toHaveBeenCalledTimes(1);
        });

        it('should not throw when function succeeds', async () => {
            const { runSafe } = useSafeWait();
            const mockFn = vi.fn().mockResolvedValue(undefined);

            await expect(runSafe(mockFn)).resolves.not.toThrow();
        });

        it('should handle function errors gracefully', async () => {
            const { runSafe } = useSafeWait();
            const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
            const mockError = new Error('Test error');
            const mockFn = vi.fn().mockRejectedValue(mockError);

            await runSafe(mockFn);

            expect(consoleErrorSpy).toHaveBeenCalledWith(mockError);
            consoleErrorSpy.mockRestore();
        });

        it('should silently return when UNMOUNTED_ERROR is thrown', async () => {
            const { runSafe, safeWait } = useSafeWait();
            const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

            // Create a function that throws UNMOUNTED_ERROR
            const mockFn = async () => {
                await safeWait(100);
            };

            const runPromise = runSafe(mockFn);

            // Simulate unmount during execution
            mockIsMounted.value = false;
            vi.advanceTimersByTime(100);

            await runPromise;

            // Should not log error for UNMOUNTED_ERROR
            expect(consoleErrorSpy).not.toHaveBeenCalled();
            consoleErrorSpy.mockRestore();
        });
    });

    describe('integration', () => {
        it('should handle race condition with unmount', async () => {
            const { runSafe, safeWait } = useSafeWait();
            const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

            const operations: string[] = [];

            const mockFn = async () => {
                operations.push('start');
                await safeWait(100);
                if (mockIsMounted.value) {
                    operations.push('end');
                }
            };

            const promise = runSafe(mockFn);

            // Simulate unmount right after delay
            vi.advanceTimersByTime(50);
            mockIsMounted.value = false;
            vi.advanceTimersByTime(50);

            await promise;

            // Operation should have started but not completed
            expect(operations).toContain('start');
            expect(operations).not.toContain('end');

            consoleErrorSpy.mockRestore();
        });
    });
});
