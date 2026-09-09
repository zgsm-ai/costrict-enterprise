import type { CreditMaintenance, MaintenanceImpact } from '@/api/bos/announcement.bo';

export function parseMaintenanceTime(value: unknown): number {
    if (typeof value !== 'string' || !/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(value)) {
        return NaN;
    }
    const timestamp = Date.parse(value.replace(' ', 'T') + '+08:00');
    if (!Number.isFinite(timestamp)) return NaN;
    const normalized = new Date(timestamp + 8 * 60 * 60 * 1000)
        .toISOString()
        .slice(0, 19)
        .replace('T', ' ');
    return normalized === value ? timestamp : NaN;
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isImpact(value: unknown): value is MaintenanceImpact {
    return (
        isRecord(value) &&
        typeof value.affected === 'boolean' &&
        typeof value.description === 'string' &&
        (!value.affected || value.description.trim().length > 0)
    );
}

export function isCreditMaintenance(value: unknown): value is CreditMaintenance {
    if (
        !isRecord(value) ||
        typeof value.maintain !== 'boolean' ||
        typeof value.title !== 'string' ||
        !value.title.trim() ||
        !isRecord(value.impacts) ||
        !isImpact(value.impacts.quota) ||
        !isImpact(value.impacts.order)
    ) {
        return false;
    }
    const start = parseMaintenanceTime(value.start_time);
    const end = parseMaintenanceTime(value.end_time);
    return Number.isFinite(start) && Number.isFinite(end) && start < end;
}
