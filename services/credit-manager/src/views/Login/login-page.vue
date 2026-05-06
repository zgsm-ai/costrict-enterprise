<template>
    <div class="login-page h-full flex flex-col justify-center">
        <n-empty size="huge">
            <template #icon>
                <img
                    src="../../assets/empty.svg"
                    alt=""
                />
            </template>
            <template #default>
                <p class="text-center text-xs mt-2.5">{{ t('login.notLogged') }}</p>
                <p class="text-center text-xs mt-1 mb-3">{{ t('login.getUserInfoBeforeLogin') }}</p>
            </template>
            <template #extra>
                <n-button
                    type="info"
                    size="medium"
                    :disabled="isLoggingIn"
                    :loading="isLoggingIn"
                    @click="toLogin"
                >
                    {{ t('login.loginImmediatly') }}
                </n-button>
            </template>
        </n-empty>
    </div>
</template>

<script setup lang="ts">
/**
 * @file 登录页
 */
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { getLoginUrl } from '@/api/mods/quota.mod';
import { authService } from '@/services/auth';
import { tokenManager } from '@/utils/token';
import { NEmpty, NButton } from 'naive-ui';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const isLoggingIn = ref(false);

const toLogin = async () => {
    if (isLoggingIn.value) return;

    // 仅在 mock 模式下直接本地模拟登录，无需跳转外部 OAuth
    if (import.meta.env.MODE === 'mock') {
        isLoggingIn.value = true;
        try {
            tokenManager.setToken('mock_access_token_12345');
            const result = await authService.authenticate(router);
            if (result.success) {
                const redirectPath = localStorage.getItem('auth_redirect');
                if (!redirectPath) {
                    router.replace('/');
                }
                return;
            }
            console.warn('Mock authentication failed, falling back to normal login flow');
        } catch (error) {
            console.error('Mock login failed:', error);
        } finally {
            isLoggingIn.value = false;
        }
    }

    isLoggingIn.value = true;
    try {
        const inviteCode = route.query.inviteCode as string;
        const isShare = route.query.isShare as string;

        const params: {
            inviter_code?: string;
            redirect_service?: string;
        } = {};
        if (inviteCode) {
            params.inviter_code = inviteCode;
        }
        if (isShare === 'true') {
            params.redirect_service = 'act_2026';
        }

        const {
            data: { url },
        } = await getLoginUrl(params);

        let loginUrl = url;
        if (isShare === 'true') {
            loginUrl = url + '&isShare=true';
        }

        // console.log(loginUrl);
        window.location.href = loginUrl;
    } catch (error) {
        console.error('登录失败:', error);
        isLoggingIn.value = false;
    }
};
</script>
