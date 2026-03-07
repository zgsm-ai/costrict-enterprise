import { getT } from './i18n';
import copy from 'copy-to-clipboard';

/**
 * copy util
 * @param text
 * @param messageHandler
 * @returns
 */
export const copyToClipboard = async (
    text: string,
    messageHandler?: {
        success?: (content: string) => void;
        error?: (content: string) => void;
    },
): Promise<boolean> => {
    const t = getT();

    try {
        const success = copy(text);

        if (success) {
            messageHandler?.success?.(t('utils.copySuccess'));
        } else {
            messageHandler?.error?.(t('utils.copyFailed'));
        }
        return success;
    } catch (err) {
        console.error('copy error:', err);
        messageHandler?.error?.(t('utils.copyFailed'));
        return false;
    }
};
