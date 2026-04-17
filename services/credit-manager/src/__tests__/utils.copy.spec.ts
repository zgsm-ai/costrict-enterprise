import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { copyToClipboard } from '@/utils/copy';

// Mock copy-to-clipboard
vi.mock('copy-to-clipboard', () => ({
    default: vi.fn(),
}));

// Mock i18n
vi.mock('@/utils/i18n', () => ({
    getT: () => (key: string) => {
        const translations: Record<string, string> = {
            'utils.copySuccess': 'Copy successful',
            'utils.copyFailed': 'Copy failed',
        };
        return translations[key] || key;
    },
}));

import copy from 'copy-to-clipboard';

describe('copyToClipboard', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('should copy text successfully and call success handler', async () => {
        const successHandler = vi.fn();
        const errorHandler = vi.fn();
        (copy as ReturnType<typeof vi.fn>).mockReturnValue(true);

        const result = await copyToClipboard('test text', {
            success: successHandler,
            error: errorHandler,
        });

        expect(copy).toHaveBeenCalledWith('test text');
        expect(successHandler).toHaveBeenCalledWith('Copy successful');
        expect(errorHandler).not.toHaveBeenCalled();
        expect(result).toBe(true);
    });

    it('should handle copy failure and call error handler', async () => {
        const successHandler = vi.fn();
        const errorHandler = vi.fn();
        (copy as ReturnType<typeof vi.fn>).mockReturnValue(false);

        const result = await copyToClipboard('test text', {
            success: successHandler,
            error: errorHandler,
        });

        expect(copy).toHaveBeenCalledWith('test text');
        expect(successHandler).not.toHaveBeenCalled();
        expect(errorHandler).toHaveBeenCalledWith('Copy failed');
        expect(result).toBe(false);
    });

    it('should handle copy error and call error handler', async () => {
        const successHandler = vi.fn();
        const errorHandler = vi.fn();
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        (copy as ReturnType<typeof vi.fn>).mockImplementation(() => {
            throw new Error('Copy error');
        });

        const result = await copyToClipboard('test text', {
            success: successHandler,
            error: errorHandler,
        });

        expect(copy).toHaveBeenCalledWith('test text');
        expect(successHandler).not.toHaveBeenCalled();
        expect(errorHandler).toHaveBeenCalledWith('Copy failed');
        expect(consoleErrorSpy).toHaveBeenCalledWith('copy error:', expect.any(Error));
        expect(result).toBe(false);

        consoleErrorSpy.mockRestore();
    });

    it('should work without message handlers', async () => {
        (copy as ReturnType<typeof vi.fn>).mockReturnValue(true);

        const result = await copyToClipboard('test text');

        expect(copy).toHaveBeenCalledWith('test text');
        expect(result).toBe(true);
    });

    it('should copy empty string', async () => {
        (copy as ReturnType<typeof vi.fn>).mockReturnValue(true);

        const result = await copyToClipboard('');

        expect(copy).toHaveBeenCalledWith('');
        expect(result).toBe(true);
    });
});
