import axios, {
    type AxiosInstance,
    type AxiosRequestConfig,
    type AxiosResponse,
    type InternalAxiosRequestConfig,
} from 'axios';
import type { MessageApi } from 'naive-ui';
import { getT } from './i18n';
import { getToken } from './token';
import router from '@/router';

let messageInstance: MessageApi | null = null;

export const setMessageInstance = (instance: MessageApi) => {
    messageInstance = instance;
};

const service: AxiosInstance = axios.create({
    timeout: 10000,
});

service.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        const token = getToken();

        config.headers.Authorization = `Bearer ${token}`;

        return config;
    },
    (error) => {
        return Promise.reject(error);
    },
);

service.interceptors.response.use(
    (response: AxiosResponse) => {
        if (!messageInstance) {
            console.warn(
                'Message instance is not available. Please ensure it is set up in App.vue.',
            );
            return response;
        }

        const t = getT();

        const res = response.data;
        if (res.code !== 0 && response.status !== 200) {
            messageInstance.error(res.message || t('common.request.requestFailed'), {
                duration: 5000,
            });
            return Promise.reject(new Error(res.message || 'Error'));
        }
        return res;
    },
    (error) => {
        if (!messageInstance) {
            console.warn('Message instance is not available for error handling.');
            return Promise.reject(error);
        }

        const t = getT();

        let messageText = '';
        if (error.response) {
            switch (error.response.status) {
                case 401:
                    messageText = t('common.request.unauthorized');
                    // 跳转到登录页
                    if (router.currentRoute.value.path !== '/login') {
                        router.push('/login');
                    }
                    break;
            }
        } else if (error.message.includes('timeout')) {
            messageText = t('common.request.requestTimeout');
        } else if (error.message === 'Network Error') {
            messageText = t('common.request.networkConnectionError');
        }
        messageInstance.error(
            messageText ||
            error.response?.data?.message ||
            error.message ||
            t('common.request.networkError'),
            {
                duration: 5000,
            },
        );
        return Promise.reject(new Error(messageText || error.message));
    },
);

export const request = <T = unknown>(config: AxiosRequestConfig): Promise<T> => {
    return service(config);
};

export const get = <T = unknown>(url: string, params?: unknown): Promise<T> => {
    return request({ method: 'GET', url, params });
};

export const post = <T = unknown>(url: string, data?: unknown): Promise<T> => {
    return request({ method: 'POST', url, data });
};

export const put = <T = unknown>(url: string, data?: unknown): Promise<T> => {
    return request({ method: 'PUT', url, data });
};

export const del = <T = unknown>(url: string, params?: unknown): Promise<T> => {
    return request({ method: 'DELETE', url, params });
};
