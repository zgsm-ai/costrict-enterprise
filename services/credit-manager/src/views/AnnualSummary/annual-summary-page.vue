<template>
    <div class="annual-summary-page">
        <div class="container">
            <Transition
                :name="transitionName"
                :mode="transitionMode"
                @after-enter="handleAfterEnter"
            >
                <AnnualSummaryStep1
                    v-if="currentStep === 1"
                    :step-id="1"
                    @next="handleNext"
                />
                <AnnualSummaryStep2
                    v-else-if="currentStep === 2"
                    :step-id="2"
                    :user-data="userData"
                    @next="handleNext"
                />
                <AnnualSummaryStep3
                    v-else-if="currentStep === 3"
                    :step-id="3"
                    :user-data="userData"
                    @next="handleNext"
                />
                <AnnualSummaryStep4
                    v-else-if="currentStep === 4"
                    :step-id="4"
                    :user-data="userData"
                    @next="handleNext"
                />
                <AnnualSummaryStep5
                    v-else-if="currentStep === 5"
                    :step-id="5"
                    :user-data="userData"
                    @next="handleNext"
                />
                <AnnualSummaryStep6
                    v-else-if="currentStep === 6"
                    :step-id="6"
                    @next="handleNext"
                />
                <AnnualSummaryStep7
                    v-else-if="currentStep === 7"
                    :step-id="7"
                    :type="resultType"
                    @next="handleNext"
                />
                <AnnualSummaryResult
                    v-else
                    :step-id="8"
                    :type="resultType"
                    :invite-code="inviteCode"
                    :login-url="loginUrl"
                    @reset="handleReset"
                />
            </Transition>
        </div>
    </div>
</template>

<script setup lang="ts">
/**
 * @file 年度总结页面
 */
import { ref, watch, computed } from 'vue';
import AnnualSummaryStep1 from './components/annual-summary-step1.vue';
import AnnualSummaryStep2 from './components/annual-summary-step2.vue';
import AnnualSummaryStep3 from './components/annual-summary-step3.vue';
import AnnualSummaryStep4 from './components/annual-summary-step4.vue';
import AnnualSummaryStep5 from './components/annual-summary-step5.vue';
import AnnualSummaryStep6 from './components/annual-summary-step6.vue';
import AnnualSummaryStep7 from './components/annual-summary-step7.vue';
// 导出结果步骤（默认注释，暂不使用）
// import AnnualSummaryResultExport from './components/annual-summary-result-export.vue';
import AnnualSummaryResult from './components/annual-summary-result.vue';
import { getInviteCode } from '@/api/mods/quota.mod';
import { getUserMe } from '@/api/mods/activity.mod';
import type { UserMeData } from '@/api/bos/activity.bo';
import { DEFAULT_RESULT_TYPE, BACKEND_TYPE_MAP } from './const';
import { useMessage } from 'naive-ui';
import { getT } from '@/utils/i18n';
import { useUserStore } from '@/store/user';
import { storeToRefs } from 'pinia';

const message = useMessage();
const t = getT();
const currentStep = ref(1);
const inviteCode = ref('');
const loginUrl = ref('');
const resultType = ref(DEFAULT_RESULT_TYPE);
const userData = ref<UserMeData | null>(null);
const isDataLoaded = ref(false);
const hasFetchError = ref(false);

// 获取用户状态
const userStore = useUserStore();
const { isTokenInitialized } = storeToRefs(userStore);

// 1. 动态计算 Transition Mode
// 当进入结果页(Step 8)时，去掉 'out-in'，实现“重叠过渡”。
// 其他步骤保持 'out-in'，保持界面整洁。
const transitionMode = computed(() => {
    return currentStep.value === 8 ? undefined : 'out-in';
});

// 2. 动态计算 Transition Name
// 结果页使用 'step-overlay' 动画，其他使用默认 'step-fade'
const transitionName = computed(() => {
    return currentStep.value === 8 ? 'step-overlay' : 'step-fade';
});

const handleNext = () => {
    if (!isDataLoaded.value || hasFetchError.value) {
        if (hasFetchError.value) {
            message.error(t('annualSummary.dataLoadFailed'));
        } else {
            message.warning(t('annualSummary.dataLoading'));
        }
        return;
    }
    if (currentStep.value < 8) {
        currentStep.value++;
    }
};

const fetchInviteCode = async () => {
    try {
        const {
            data: { invite_code = '' },
        } = await getInviteCode();
        inviteCode.value = invite_code;
    } catch (error) {
        // 邀请码获取失败不影响下一步操作，仅记录空值
        inviteCode.value = '';
        console.warn('获取邀请码失败，但不影响流程:', error);
    }
};

const fetchLoginUrl = async () => {
    if (!inviteCode.value) return;
    try {
        // 直接构建年度总结封面页的 URL，带上分享标识和邀请码
        loginUrl.value = `https://zgsm.sangfor.com/credit/manager/annual-summary-cover?isShare=true&inviteCode=${inviteCode.value}`;
        isDataLoaded.value = true;
    } catch (error) {
        hasFetchError.value = true;
        throw error;
    }
};

const fetchUserMe = async () => {
    try {
        const { data } = await getUserMe();
        userData.value = data;
        if (data.identity) {
            resultType.value = BACKEND_TYPE_MAP[data.identity] || DEFAULT_RESULT_TYPE;
        }
    } catch (error) {
        hasFetchError.value = true;
        throw error;
    }
};

const handleReset = () => {
    currentStep.value = 1;
};

const handleAfterEnter = () => {};

// 初始化时加载页面数据
const loadData = async () => {
    try {
        await fetchUserMe();
        await fetchInviteCode();
        await fetchLoginUrl();
        isDataLoaded.value = true;
    } catch {
        hasFetchError.value = true;
        isDataLoaded.value = false;
    }
};

// 等待 token 初始化完成后加载数据
watch(
    isTokenInitialized,
    (val) => {
        if (val) {
            loadData();
        }
    },
    {
        immediate: true,
    },
);
</script>

<style scoped lang="less">
.annual-summary-page {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;

    .container {
        position: relative; // 必须是 relative，为了让内部的 absolute 元素定位
        width: 100%;
        max-width: 375px;
    }
}

/* --- 普通步骤切换动画 (Step 1-6) --- */
.step-fade-enter-active,
.step-fade-leave-active {
    transition: opacity 800ms ease-in-out;
}
.step-fade-enter-from,
.step-fade-leave-to {
    opacity: 0;
}

/* --- 无缝衔接动画 (Step 7 -> Result) --- */

/* 
   Result 页进入：
   正常淡入，并在最上层 (z-index 控制)
*/
.step-overlay-enter-active {
    transition: opacity 800ms ease-in-out;
    position: relative;
    z-index: 2;
}
.step-overlay-enter-from {
    opacity: 0;
}

/* 
   Step 7 离开：
   1. position: absolute -> 让它脱离文档流，浮在 container 顶部，防止挤压新进来的 Result 页。
   2. top: 0, width: 100% -> 确保位置和尺寸不变。
   3. transition: opacity 800ms -> 让它与 Result 页同时淡出（交叉淡入淡出）。
      如果你希望 Step 7 保持不透明直到 Result 完全盖住它，可以把 opacity 改为 1 或者去掉 leave-to。
      但通常交叉淡入淡出（Cross-fade）视觉效果最柔和。
*/
.step-overlay-leave-active {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    z-index: 1;
    transition: opacity 800ms ease-in-out;
}

.step-overlay-leave-to {
    opacity: 0; /* 如果想让 Step 7 慢慢变淡，保留这个。如果想让它一直显示直到被完全覆盖，可以删掉这行 */
}
</style>
