import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';

export const CREDITS_SERVICE_CUTOFF_AT = new Date('2026-09-23T23:59:00+08:00').getTime();

export const hasCreditsServiceEnded = (now = Date.now()): boolean => {
    return now >= CREDITS_SERVICE_CUTOFF_AT;
};

export const useCreditsServiceCutoff = () => {
    const route = useRoute();
    const currentTime = ref(Date.now());
    let cutoffTimer: number | undefined;

    const isCreditsServiceEnded = computed(
        () => route.query.preview === 'disabled' || hasCreditsServiceEnded(currentTime.value),
    );

    onMounted(() => {
        const delay = CREDITS_SERVICE_CUTOFF_AT - Date.now();
        if (delay <= 0) {
            currentTime.value = Date.now();
            return;
        }

        cutoffTimer = window.setTimeout(() => {
            currentTime.value = Date.now();
        }, delay);
    });

    onBeforeUnmount(() => {
        if (cutoffTimer !== undefined) {
            window.clearTimeout(cutoffTimer);
        }
    });

    return { isCreditsServiceEnded };
};
