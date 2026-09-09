import { fileURLToPath, URL } from 'node:url';

import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import viteCompression from 'vite-plugin-compression';
import vueDevTools from 'vite-plugin-vue-devtools';
import tailwindcss from '@tailwindcss/vite';
import { viteMockServe } from 'vite-plugin-mock';

// https://vite.dev/config/
export default defineConfig(({ command, mode }) => {
    // 开发环境 API 代理目标，默认指向共享测试环境。
    // 可通过环境变量或 .env.local 覆盖（变量名见 .env.example），例如：
    //   PROXY_TARGET_OIDC=http://127.0.0.1:8084 npm run dev
    const env = loadEnv(mode, process.cwd(), '');
    const proxyTarget = (key: string, fallback: string) => env[key] || fallback;

    return {
        plugins: [
            vue(),
            viteCompression(),
            vueDevTools(),
            tailwindcss(),
            viteMockServe({
                mockPath: 'mock',
                enable: command === 'serve' && mode === 'mock',
                watchFiles: true,
            }),
        ],
        base: '/credit/manager',
        resolve: {
            alias: {
                '@': fileURLToPath(new URL('./src', import.meta.url)),
            },
        },
        server: {
            port: 9527,
            host: true,
            proxy: {
                '/costrict-static': {
                    target: proxyTarget('PROXY_TARGET_STATIC', 'https://zgsm.sangfor.com'),
                    changeOrigin: true,
                    secure: false,
                },
                '/quota-manager': {
                    target: proxyTarget('PROXY_TARGET_QUOTA', 'https://zgsmtest.cn:30443/'),
                    changeOrigin: true,
                    secure: false,
                },
                '/oidc-auth': {
                    target: proxyTarget('PROXY_TARGET_OIDC', 'https://zgsmtest.cn:30443/'),
                    changeOrigin: true,
                    secure: false,
                },
                '/quota-order-manager': {
                    target: proxyTarget('PROXY_TARGET_ORDER', 'https://zgsmtest.cn:30443/'),
                    changeOrigin: true,
                    secure: false,
                },
                '/operational_activities': {
                    target: proxyTarget('PROXY_TARGET_ACTIVITY', 'https://zgsm.sangfor.com'),
                    changeOrigin: true,
                    secure: false,
                },
            },
        },
    };
});
