import { promiseTimeout, useMounted } from '@vueuse/core';

const UNMOUNTED_ERROR = Symbol('UNMOUNTED_ERROR');

export const useSafeWait = () => {
    const isMounted = useMounted();

    const safeWait = async (ms: number) => {
        await promiseTimeout(ms);
        if (!isMounted.value) {
            throw UNMOUNTED_ERROR;
        }
    };

    const runSafe = async (fn: () => Promise<void>) => {
        try {
            await fn();
        } catch (e) {
            if (e === UNMOUNTED_ERROR) return;
            console.error(e);
        }
    };

    return {
        safeWait,
        runSafe,
    };
};
