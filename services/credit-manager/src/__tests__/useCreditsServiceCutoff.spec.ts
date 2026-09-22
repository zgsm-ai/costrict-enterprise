import { describe, expect, it } from 'vitest';
import {
    CREDITS_SERVICE_CUTOFF_AT,
    hasCreditsServiceEnded,
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
});
