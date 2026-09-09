import { computed, ref } from 'vue';
import { isAxiosError } from 'axios';
import { defineStore } from 'pinia';
import { getCreditMaintenance } from '@/api/mods/announcement.mod';
import type { CreditMaintenance } from '@/api/bos/announcement.bo';
import { parseMaintenanceTime, isCreditMaintenance } from '@/utils/maintenance';

export function isMaintenanceActive(config: CreditMaintenance | null, now: number): boolean {
    if (config?.maintain !== true) return false;
    const start = parseMaintenanceTime(config.start_time);
    const end = parseMaintenanceTime(config.end_time);
    return now >= start && now < end;
}

export const useMaintenanceStore = defineStore('maintenance', () => {
    const announcement = ref<CreditMaintenance | null>(null);
    const now = ref(Date.now());
    const ready = ref(false);
    let pending: Promise<void> | null = null;
    const active = computed(() => isMaintenanceActive(announcement.value, now.value));
    const quotaAffected = computed(
        () => active.value && announcement.value?.impacts?.quota?.affected === true,
    );
    const orderAffected = computed(
        () => active.value && announcement.value?.impacts?.order?.affected === true,
    );

    function isTabBlocked(tab: unknown): boolean {
        return (
            (quotaAffected.value && (tab === 'usage' || tab === 'usage-consumption')) ||
            (orderAffected.value && tab === 'subscription')
        );
    }

    async function refresh(): Promise<void> {
        if (pending) return pending;
        pending = (async () => {
            try {
                const data = await getCreditMaintenance();
                if (isCreditMaintenance(data)) {
                    announcement.value = data;
                } else {
                    announcement.value = null;
                    console.error('Invalid maintenance announcement; ignoring it.');
                }
            } catch (error) {
                if (
                    error instanceof SyntaxError ||
                    (isAxiosError(error) &&
                        (error.response?.status === 404 || error.name === 'SyntaxError'))
                ) {
                    // 文件不存在或 JSON 无法解析时，清除之前的维护限制。
                    announcement.value = null;
                } else {
                    // 首次失败不阻断系统，其他刷新错误保留上次公告并继续按时间判断。
                    console.error('Failed to load maintenance announcement:', error);
                }
            } finally {
                now.value = Date.now();
                ready.value = true;
            }
        })();
        try {
            await pending;
        } finally {
            pending = null;
        }
    }

    return {
        announcement,
        now,
        ready,
        active,
        quotaAffected,
        orderAffected,
        isTabBlocked,
        refresh,
    };
});
