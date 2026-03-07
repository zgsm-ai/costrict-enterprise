/**
 * @description home页面个人信息模块hook
 */
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useMessage } from 'naive-ui';
import { useUserStore } from '@/store/user';
import { storeToRefs } from 'pinia';
import { getUserQuota, getBindAccount, getInviteCode } from '@/api/mods/quota.mod';
import type { QuotaList } from '@/api/bos/quota.bo';
import { BING_TYPE } from '../const';
import { copyToClipboard } from '@/utils/copy';
import { authService } from '@/services/auth';

export function useProfile() {
    const router = useRouter();
    const message = useMessage();
    const userStore = useUserStore();
    const { githubName, phoneNumber, userId, isPrivate, isTokenInitialized } =
        storeToRefs(userStore);

    // 配额相关状态
    const usedQuota = ref<number>(0);
    const totalQuota = ref<number>(0);
    const columnsData = ref<QuotaList[]>([]);
    const isStar = ref<string | undefined>();
    const inviteCode = ref('');

    // 计算属性
    const transferData = computed(() =>
        isStar.value === 'true' || isStar.value === undefined ? columnsData.value : [],
    );

    // 获取用户配额信息
    const fetchUserQuota = async () => {
        const { data } = await getUserQuota();

        if (!data) {
            return;
        }

        columnsData.value = data.quota_list || [];
        usedQuota.value = data.used_quota || 0;
        totalQuota.value = data.total_quota || 0;
        isStar.value = data.is_star;
    };

    // 获取邀请码
    const fetchInviteCode = async () => {
        const {
            data: { invite_code = '' },
        } = await getInviteCode();

        inviteCode.value = invite_code;
    };

    // 绑定账号操作
    const bindAction = async (bindType: keyof typeof BING_TYPE) => {
        const { data } = await getBindAccount({
            bindType,
            state: 'state',
        });

        if (!data.url) {
            return;
        }

        window.location.href = data.url;
    };

    // 绑定GitHub
    const bindGithub = () => bindAction(BING_TYPE.github);

    // 绑定手机号
    const bindPhone = () => bindAction(BING_TYPE.sms);

    // 复制用户ID
    const copyCode = () => {
        copyToClipboard(userId.value, {
            success: message.success,
            error: message.error,
        });
    };

    // 复制邀请码
    const copyInviteCode = () => {
        copyToClipboard(inviteCode.value, {
            success: message.success,
            error: message.error,
        });
    };

    // 退出登录
    const logout = async () => {
        try {
            // 调用退出登录服务
            await authService.logout();
            // 重置用户store状态
            userStore.resetAuth();
            // 跳转到登录页
            router.push('/login');
        } catch (error) {
            console.error('Logout failed:', error);
        }
    };

    // 跳转到积分管理页面
    const toCredits = () => window.open('/credit/manager/credits');

    // 转账回调
    const transferInCallBack = () => {
        fetchUserQuota();
    };

    return {
        // 状态
        usedQuota,
        totalQuota,
        columnsData,
        isStar,
        inviteCode,
        githubName,
        phoneNumber,
        userId,
        isPrivate,
        isTokenInitialized,

        // 计算属性
        transferData,

        // 方法
        fetchUserQuota,
        fetchInviteCode,
        bindGithub,
        bindPhone,
        copyCode,
        copyInviteCode,
        logout,
        toCredits,
        transferInCallBack,
    };
}
