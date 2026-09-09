import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { getCreditMaintenance } from '@/api/mods/announcement.mod';
import type { CreditMaintenance } from '@/api/bos/announcement.bo';
import { isMaintenanceActive, useMaintenanceStore } from '@/store/maintenance';

vi.mock('@/api/mods/announcement.mod', () => ({ getCreditMaintenance: vi.fn() }));

const config: CreditMaintenance = {
    maintain: true,
    title: '维护中',
    start_time: '2026-09-09 16:30:00',
    end_time: '2026-09-09 17:00:00',
    impacts: {
        quota: { affected: true, description: '配额维护' },
        order: { affected: false, description: '' },
    },
};
const start = Date.parse('2026-09-09T08:30:00Z');
const end = Date.parse('2026-09-09T09:00:00Z');

beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
});

describe('maintenance window', () => {
    it('uses UTC+8 and includes the start but excludes the end', () => {
        expect(isMaintenanceActive(config, start - 1)).toBe(false);
        expect(isMaintenanceActive(config, start)).toBe(true);
        expect(isMaintenanceActive(config, end - 1)).toBe(true);
        expect(isMaintenanceActive(config, end)).toBe(false);
    });

    it('requires the master switch and a valid window', () => {
        expect(isMaintenanceActive({ ...config, maintain: false }, start)).toBe(false);
        expect(isMaintenanceActive(null, start)).toBe(false);
        expect(isMaintenanceActive({ ...config, start_time: 'invalid' }, start)).toBe(false);
        expect(isMaintenanceActive({ ...config, end_time: config.start_time }, start)).toBe(false);
        expect(isMaintenanceActive({ ...config, end_time: '2026-09-08 17:00:00' }, start)).toBe(
            false,
        );
    });

    it('blocks quota and order pages independently and restores access at the end', () => {
        const store = useMaintenanceStore();
        store.announcement = structuredClone(config);
        store.now = start;
        expect(store.isTabBlocked('usage')).toBe(true);
        expect(store.isTabBlocked('usage-consumption')).toBe(true);
        expect(store.isTabBlocked('subscription')).toBe(false);
        expect(store.isTabBlocked('profile')).toBe(false);
        store.announcement.impacts.quota.affected = false;
        store.announcement.impacts.order.affected = true;
        expect(store.isTabBlocked('usage')).toBe(false);
        expect(store.isTabBlocked('subscription')).toBe(true);
        store.now = end;
        expect(store.isTabBlocked('subscription')).toBe(false);
    });

    it('applies refreshed switches without a page reload', async () => {
        const store = useMaintenanceStore();
        vi.mocked(getCreditMaintenance).mockResolvedValue(structuredClone(config));
        await store.refresh();
        store.now = start;
        expect(store.quotaAffected).toBe(true);
        vi.mocked(getCreditMaintenance).mockResolvedValue({ ...config, maintain: false });
        await store.refresh();
        store.now = start;
        expect(store.quotaAffected).toBe(false);
    });

    it('allows initial failure and retains the previous announcement on refresh failure', async () => {
        const error = vi.spyOn(console, 'error').mockImplementation(() => {});
        const store = useMaintenanceStore();
        vi.mocked(getCreditMaintenance).mockRejectedValue(new Error('offline'));
        await store.refresh();
        expect(store.ready).toBe(true);
        expect(store.active).toBe(false);
        store.announcement = structuredClone(config);
        await store.refresh();
        store.now = start;
        expect(store.quotaAffected).toBe(true);
        store.now = end;
        expect(store.quotaAffected).toBe(false);
        error.mockRestore();
    });

    it('treats 404 as no maintenance, including after an active announcement', async () => {
        const error = vi.spyOn(console, 'error').mockImplementation(() => {});
        try {
            const store = useMaintenanceStore();
            vi.mocked(getCreditMaintenance).mockRejectedValue({
                isAxiosError: true,
                response: { status: 404 },
            });
            await store.refresh();
            expect(store.ready).toBe(true);
            expect(store.active).toBe(false);

            store.announcement = structuredClone(config);
            store.announcement.impacts.order.affected = true;
            store.now = start;
            expect(store.quotaAffected).toBe(true);
            expect(store.orderAffected).toBe(true);

            await store.refresh();
            store.now = start;
            expect(store.announcement).toBeNull();
            expect(store.active).toBe(false);
            expect(store.isTabBlocked('usage')).toBe(false);
            expect(store.isTabBlocked('subscription')).toBe(false);
            expect(store.orderAffected).toBe(false);
            expect(error).not.toHaveBeenCalled();
        } finally {
            error.mockRestore();
        }
    });
    it.each([
        null,
        [],
        '',
        '{"maintain": true,',
        '<html>Not found</html>',
        {},
        { ...config, maintain: 'true' },
        { ...config, title: {} },
        { ...config, title: ' ' },
        { ...config, start_time: 123 },
        { ...config, start_time: '2026-02-30 16:30:00' },
        { ...config, end_time: config.start_time },
        { ...config, end_time: '2026-09-08 17:00:00' },
        { ...config, impacts: null },
        { ...config, impacts: { quota: { affected: 'true', description: '维护' } } },
        { ...config, impacts: { ...config.impacts, order: { affected: true, description: {} } } },
    ])('ignores malformed content and clears previous restrictions: %j', async (data) => {
        const error = vi.spyOn(console, 'error').mockImplementation(() => {});
        try {
            const store = useMaintenanceStore();
            vi.mocked(getCreditMaintenance).mockResolvedValue(data);
            for (const previous of [null, config]) {
                store.announcement = previous;
                await store.refresh();
                store.now = start;
                expect(store.ready).toBe(true);
                expect(store.announcement).toBeNull();
                expect(store.active).toBe(false);
                expect(store.isTabBlocked('usage')).toBe(false);
                expect(store.isTabBlocked('subscription')).toBe(false);
            }
        } finally {
            error.mockRestore();
        }
    });

    it('clears maintenance if JSON parsing throws', async () => {
        const store = useMaintenanceStore();
        store.announcement = structuredClone(config);
        vi.mocked(getCreditMaintenance).mockRejectedValue(new SyntaxError('Invalid JSON'));
        await store.refresh();
        store.now = start;
        expect(store.ready).toBe(true);
        expect(store.active).toBe(false);
        expect(store.announcement).toBeNull();
    });
});
