import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AxiosError, AxiosHeaders, type AxiosResponse } from 'axios';
import type { MessageApi } from 'naive-ui';
import Cookies from 'js-cookie';
import { request, setMessageInstance } from '@/utils/request';

vi.mock('@/utils/i18n', () => ({ getT: () => () => '账号认证异常' }));

const errorMessage = vi.fn();

describe('Cookie request authentication', () => {
    beforeEach(() => {
        Cookies.remove('zgsmAdminToken');
        vi.clearAllMocks();
        setMessageInstance({ error: errorMessage } as unknown as MessageApi);
    });

    it('reads the current Cookie for every request and never sends Bearer undefined', async () => {
        const headers: unknown[] = [];
        const send = () =>
            request({
                url: '/quota-manager/test',
                adapter: async (config) => {
                    headers.push(config.headers.Authorization);
                    return {
                        config,
                        status: 200,
                        statusText: 'OK',
                        headers: {},
                        data: { code: 0 },
                    };
                },
            });
        await send();
        Cookies.set('zgsmAdminToken', 'first');
        await send();
        Cookies.set('zgsmAdminToken', 'second');
        await send();
        expect(headers).toEqual([undefined, 'Bearer first', 'Bearer second']);
    });

    it('reports 401 while preserving both Cookie and page URL', async () => {
        Cookies.set('zgsmAdminToken', 'expired');
        const original = window.location.href;
        await expect(
            request({
                url: '/quota-manager/test',
                adapter: async (config) => {
                    const response: AxiosResponse = {
                        config,
                        status: 401,
                        statusText: 'Unauthorized',
                        headers: new AxiosHeaders(),
                        data: {},
                    };
                    throw new AxiosError(
                        'Request failed',
                        'ERR_BAD_REQUEST',
                        config,
                        undefined,
                        response,
                    );
                },
            }),
        ).rejects.toThrow('账号认证异常');
        expect(errorMessage).toHaveBeenCalledWith('账号认证异常', { duration: 5000 });
        expect(Cookies.get('zgsmAdminToken')).toBe('expired');
        expect(window.location.href).toBe(original);
    });
});
