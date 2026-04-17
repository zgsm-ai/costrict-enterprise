import { describe, it, expect, vi } from 'vitest';
import { formatDate } from '@/utils/date';

describe('formatDate', () => {
    it('should format valid date string with default format', () => {
        const result = formatDate('2024-01-15 10:30:00');

        expect(result).toBe('2024-01-15 10:30:00');
    });

    it('should format valid date string with custom format', () => {
        const result = formatDate('2024-01-15 10:30:00', 'YYYY-MM-DD');

        expect(result).toBe('2024-01-15');
    });

    it('should return fallback for null input', () => {
        const result = formatDate(null as unknown as string);

        expect(result).toBe('');
    });

    it('should return fallback for undefined input', () => {
        const result = formatDate(undefined as unknown as string);

        expect(result).toBe('');
    });

    it('should return fallback for empty string input', () => {
        const result = formatDate('');

        expect(result).toBe('');
    });

    it('should return fallback for invalid date string', () => {
        const result = formatDate('invalid-date', 'YYYY-MM-DD', 'Invalid Date');

        expect(result).toBe('Invalid Date');
    });

    it('should use custom fallback value', () => {
        const result = formatDate('', 'YYYY-MM-DD', 'N/A');

        expect(result).toBe('N/A');
    });

    it('should format ISO date string', () => {
        const result = formatDate('2024-01-15T10:30:00.000Z', 'YYYY-MM-DD HH:mm:ss');

        // Note: This might vary based on timezone
        expect(typeof result).toBe('string');
        expect(result).not.toBe('');
    });

    it('should format date-only string', () => {
        const result = formatDate('2024-01-15', 'YYYY-MM-DD');

        expect(result).toBe('2024-01-15');
    });

    it('should handle leap year date', () => {
        const result = formatDate('2024-02-29', 'YYYY-MM-DD');

        expect(result).toBe('2024-02-29');
    });
});
