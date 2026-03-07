<template>
    <div
        class="home-page mt-12"
        :class="responsiveClass"
    >
        <n-layout
            class="layout-container"
            :has-sider="!isMobileLayout"
        >
            <!-- 移动端横向菜单 -->
            <n-layout-header
                v-if="isMobileLayout"
                class="mobile-menu-header"
            >
                <n-menu
                    class="horizontal-menu"
                    mode="horizontal"
                    :options="menuOptions"
                    :value="activeMenuKey"
                    responsive
                    @update:value="handleMenuSelect"
                />
            </n-layout-header>

            <!-- 桌面端侧边栏菜单 -->
            <n-layout-sider
                v-if="!isMobileLayout"
                class="menu-sider"
                width="15%"
                :native-scrollbar="false"
            >
                <n-menu
                    class="side-menu"
                    :options="menuOptions"
                    :value="activeMenuKey"
                    @update:value="handleMenuSelect"
                />
            </n-layout-sider>

            <n-layout-content class="content-area ml-2">
                <div class="page-content">
                    <!-- 个人信息 -->
                    <section
                        class="info"
                        v-if="activeMenuKey === 'profile'"
                    >
                        <div class="info-title text-white text-xl mb-9 flex">
                            {{ t('homePageUi.personalInfo') }}
                            <img
                                v-if="isZh"
                                src="../../assets/summary_icon.webp"
                                alt="summary_icon"
                                class="w-18 absolute right-0 cursor-pointer"
                                @click="toAnnualSummary"
                            />
                        </div>
                        <!-- 基本信息 -->
                        <common-card :title="t('homePage.basicInfo')">
                            <template #default>
                                <profile-section
                                    :github-name="githubName"
                                    :phone-number="phoneNumber"
                                    :user-id="userId"
                                    :invite-code="inviteCode"
                                    :is-private="isPrivate"
                                    :is-zh="isZh"
                                    @bind-github="bindGithub"
                                    @bind-phone="bindPhone"
                                    @copy-user-id="copyCode"
                                    @copy-invite-code="copyInviteCode"
                                />
                            </template>
                        </common-card>

                        <!-- 登录管理 -->
                        <common-card :title="t('homePageUi.loginManagement')">
                            <template #default>
                                <div class="flex items-center">
                                    <span>{{ t('homePageUi.accountLogout') }}</span>
                                    <div
                                        class="cursor-pointer ml-auto text-xs btn-logout"
                                        @click="logout"
                                    >
                                        {{ t('homePageUi.logout') }}
                                    </div>
                                </div>
                            </template>
                        </common-card>
                    </section>

                    <!-- 订阅 -->
                    <section
                        class="subscription"
                        v-if="activeMenuKey === 'subscription'"
                    >
                        <div class="info-title text-white text-xl mb-9">
                            {{ t('homePageUi.subscription') }}
                        </div>
                        <subscription-section
                            :records-data="ordersData"
                            :total="ordersTotal"
                            :page="ordersPage"
                            :page-size="ordersPageSize"
                            :loading="ordersLoading"
                            @update:page="handleOrdersPageChange"
                            @update:pageSize="handleOrdersPageSizeChange"
                            @refresh-orders="fetchOrders"
                        />
                    </section>

                    <!-- 用量统计 -->
                    <section
                        class="usage"
                        v-if="activeMenuKey === 'usage'"
                    >
                        <div class="info-title text-white text-xl mb-9">
                            {{ t('homePageUi.usageStatistics') }}
                        </div>
                        <common-card :title="t('homePage.usageTitle')">
                            <template #header-extra>
                                <span
                                    class="credit-action"
                                    @click="transferCredit"
                                    >{{ t('homePage.creditTransfer') }}</span
                                >
                                <span
                                    class="usage-credit cursor-pointer"
                                    @click="toCredits"
                                    >{{ t('homePage.creditRecords') }}</span
                                >
                            </template>
                            <template #default>
                                <usage-section
                                    :used-quota="usedQuota"
                                    :total-quota="totalQuota"
                                    :columns-data="columnsData"
                                    :is-star="isStar"
                                    @transfer-credit="transferCredit"
                                    @to-credits="toCredits"
                                />
                            </template>
                        </common-card>

                        <!-- 用量消耗统计 -->
                        <section class="usage-consumption mt-7">
                            <common-card>
                                <template #header>
                                    <!-- 大屏幕布局 -->
                                    <div
                                        v-if="!isSmallScreen"
                                        class="flex items-center"
                                    >
                                        <span class="text-base">{{
                                            t('homePage.usageConsumptionTitle')
                                        }}</span>
                                        <div class="select-time__group flex ml-2.5 items-center">
                                            <div
                                                class="select-time__group-item text-xs cursor-pointer mr-1 opacity-70"
                                                :class="{
                                                    active: selectedTimeRange === 'today',
                                                }"
                                                @click="handleTimeRangeSelectWithReset('today')"
                                            >
                                                {{ t('homePageUi.today') }}
                                            </div>
                                            <div
                                                class="select-time__group-item text-xs cursor-pointer mr-1 opacity-70"
                                                :class="{
                                                    active: selectedTimeRange === '7days',
                                                }"
                                                @click="handleTimeRangeSelectWithReset('7days')"
                                            >
                                                {{ t('homePageUi.within7Days') }}
                                            </div>
                                            <div
                                                class="select-time__group-item text-xs cursor-pointer mr-1 opacity-70"
                                                :class="{
                                                    active: selectedTimeRange === '30days',
                                                }"
                                                @click="handleTimeRangeSelectWithReset('30days')"
                                            >
                                                {{ t('homePageUi.within30Days') }}
                                            </div>
                                            <custom-date-picker
                                                v-model="customDateRangeFormatted"
                                                :is-selected="selectedTimeRange === 'custom'"
                                                @update:model-value="
                                                    handleCustomDateRangeChangeWithReset
                                                "
                                            />
                                        </div>
                                    </div>

                                    <!-- 小屏幕布局 -->
                                    <div v-else>
                                        <span class="text-base">{{
                                            t('homePage.usageConsumptionTitle')
                                        }}</span>
                                        <div class="select-time__group mt-2">
                                            <div
                                                class="select-time__group-item text-xs cursor-pointer mr-1 opacity-70"
                                                :class="{
                                                    active: selectedTimeRange === 'today',
                                                }"
                                                @click="handleTimeRangeSelectWithReset('today')"
                                            >
                                                {{ t('homePageUi.today') }}
                                            </div>
                                            <div
                                                class="select-time__group-item text-xs cursor-pointer mr-1 opacity-70"
                                                :class="{
                                                    active: selectedTimeRange === '7days',
                                                }"
                                                @click="handleTimeRangeSelectWithReset('7days')"
                                            >
                                                {{ t('homePageUi.within7Days') }}
                                            </div>
                                            <div
                                                class="select-time__group-item text-xs cursor-pointer mr-1 opacity-70"
                                                :class="{
                                                    active: selectedTimeRange === '30days',
                                                }"
                                                @click="handleTimeRangeSelectWithReset('30days')"
                                            >
                                                {{ t('homePageUi.within30Days') }}
                                            </div>
                                            <custom-date-picker
                                                v-model="customDateRangeFormatted"
                                                :is-selected="selectedTimeRange === 'custom'"
                                                @update:model-value="
                                                    handleCustomDateRangeChangeWithReset
                                                "
                                            />
                                        </div>
                                    </div>
                                </template>
                                <template #header-extra>
                                    <span
                                        class="billing-btn cursor-pointer ml-auto whitespace-nowrap"
                                        @click="toBillingDocs"
                                        >{{ t('homePageUi.packageBillingDescription') }}</span
                                    >
                                </template>
                                <template #default>
                                    <usage-consumption-section
                                        :table-data="usageConsumptionData"
                                        :loading="usageConsumptionLoading"
                                        :current-page="usageConsumptionPage"
                                        :page-size="usageConsumptionPageSize"
                                        :total-count="usageConsumptionTotal"
                                        @update:page="handleUsageConsumptionPageChange"
                                        @update:pageSize="handleUsageConsumptionPageSizeChange"
                                    />
                                </template>
                            </common-card>
                        </section>
                    </section>

                    <!-- 运营活动 -->
                    <section
                        class="activity"
                        v-if="activeMenuKey === 'activity'"
                    >
                        <activity-section />
                    </section>
                </div>
            </n-layout-content>
        </n-layout>
        <credit-transfer-modal
            :user-quota-data="transferData"
            :show="showCreditTransferModal"
            :is-star="isStar"
            @update:show="updateCreditModalShow"
            @update:submit="openCreditCodeModal"
            @update:transferIn="transferInCallBack"
        />
        <credit-code-modal
            :show="showCreditCodeModal"
            @update:show="updateCreditCodeModalShow"
            :voucher-code="voucherCode"
        />
    </div>
</template>

<script setup lang="ts">
/**
 * @file home-page.vue
 */
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { NLayout, NLayoutSider, NLayoutContent, NLayoutHeader, NMenu } from 'naive-ui';
import CommonCard from '@/components/common-card.vue';
import CreditTransferModal from '@/views/Home/credit-transfer-modal.vue';
import CreditCodeModal from './credit-code-modal.vue';
import ProfileSection from './components/profile-section.vue';
import SubscriptionSection from './components/subscription-section.vue';
import UsageSection from './components/usage-section.vue';
import ActivitySection from './components/activity-section.vue';
import CustomDatePicker from './components/custom-date-picker.vue';
import { useMenu, type MenuKey } from './hook/useMenu';
import { useProfile } from './hook/useProfile';
import { useSubscription } from './hook/useSubscription';
import { useUsageStatistics } from './hook/useUsageStatistics';
import { useResponsive } from './hook/useResponsive';
import usageConsumptionSection from './components/usage-consumption-section.vue';

const { t, locale } = useI18n();
const router = useRouter();
const isZh = computed(() => locale.value === 'zh');

// 使用响应式布局hook
const { isMobileLayout, isSmallScreen, responsiveClass } = useResponsive();

// 使用菜单hook
const {
    activeMenuKey,
    menuOptions,
    handleMenuSelect,
    setMenuLoading,
    markMenuAsLoaded,
    resetUsageMenuLoadedState,
} = useMenu();

// 使用个人信息hook
const {
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
    transferData,
    fetchUserQuota,
    fetchInviteCode,
    bindGithub,
    bindPhone,
    copyCode,
    copyInviteCode,
    logout,
    toCredits,
    transferInCallBack,
} = useProfile();

// 使用订阅hook
const {
    ordersData,
    ordersTotal,
    ordersPage,
    ordersPageSize,
    ordersLoading,
    fetchOrders,
    handleOrdersPageChange,
    handleOrdersPageSizeChange,
} = useSubscription();

// 使用用量统计hook
const {
    usageConsumptionData,
    usageConsumptionLoading,
    usageConsumptionPage,
    usageConsumptionPageSize,
    usageConsumptionTotal,
    selectedTimeRange,
    customDateRangeFormatted,
    fetchUsageConsumptionData,
    handleUsageConsumptionPageChange,
    handleUsageConsumptionPageSizeChange,
    handleTimeRangeSelect,
    handleCustomDateRangeChange,
} = useUsageStatistics();

// 修改时间范围选择处理函数，以便在时间范围改变时重置usage菜单的加载状态
const handleTimeRangeSelectWithReset = (range: string) => {
    resetUsageMenuLoadedState();
    handleTimeRangeSelect(range);
};

// 修改自定义日期范围变化处理函数
const handleCustomDateRangeChangeWithReset = (value: [string, string] | null) => {
    resetUsageMenuLoadedState();
    handleCustomDateRangeChange(value);
};

// 模态框相关状态
const voucherCode = ref();
const showCreditCodeModal = ref(false);
const showCreditTransferModal = ref(false);

const openCreditCodeModal = (code: string) => {
    fetchUserQuota();

    updateCreditModalShow(false);

    showCreditCodeModal.value = true;

    voucherCode.value = code;
};

const updateCreditCodeModalShow = (status: boolean) => {
    showCreditCodeModal.value = status;
};

const transferCredit = () => {
    showCreditTransferModal.value = true;
};

const updateCreditModalShow = (status: boolean) => {
    showCreditTransferModal.value = status;
};

// 根据菜单加载相应的数据
const loadMenuData = async (menuKey: MenuKey) => {
    if (!isTokenInitialized.value) return;

    setMenuLoading(menuKey, true);

    try {
        switch (menuKey) {
            case 'profile':
                await Promise.all([fetchUserQuota(), fetchInviteCode()]);
                break;
            case 'subscription':
                // 只有在中文环境下才加载订阅数据
                if (isZh.value) {
                    await fetchOrders();
                }
                break;
            case 'usage':
                await Promise.all([fetchUserQuota(), fetchUsageConsumptionData()]);
                break;
            case 'activity':
                break;
        }

        markMenuAsLoaded(menuKey);
    } finally {
        setMenuLoading(menuKey, false);
    }
};

const toBillingDocs = () => {
    window.open('https://docs.costrict.ai/billing/purchase');
};

const toAnnualSummary = () => {
    router.push('/annual-summary');
};

// 初始化时加载默认菜单数据
watch(
    isTokenInitialized,
    (val) => {
        if (val) {
            loadMenuData(activeMenuKey.value);
        }
    },
    {
        immediate: true,
    },
);

// 监听菜单切换，加载相应数据
watch(activeMenuKey, (newKey) => {
    if (newKey) {
        loadMenuData(newKey);
    }
});
</script>

<style scoped lang="less">
.home-page {
    width: 90%;
    display: flex;
    flex-direction: column;
    margin-left: auto;
    margin-right: auto;

    // 移动端布局调整
    @media (max-width: 1440px) {
        width: 100%;
        max-width: 100%;
        padding: 0 20px;
        box-sizing: border-box;
        margin-top: 20px;
    }

    .layout-container {
        flex: 1;
        display: flex;
        background: transparent;
        min-height: 0;

        // 移动端上下布局
        @media (max-width: 1440px) {
            flex-direction: column;
        }
    }

    // 移动端横向菜单样式
    .mobile-menu-header {
        background: rgba(0, 0, 0);
        padding: 0;
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);

        :deep(.n-menu-item) {
            margin-right: 0;
        }

        :deep(.n-menu-item-content) {
            opacity: 0.7;
            border-radius: 0;
            transition: all 0.3s ease;

            &:hover {
                opacity: 1;
            }

            &--selected {
                opacity: 1;
            }
        }
    }

    .select-time__group {
        .select-time__group-item {
            padding: 4px 8px;
            border-radius: 4px;
            transition: all 0.3s ease;
            height: 28px;
            display: flex;
            align-items: center;

            &:hover {
                opacity: 1;
                background-color: rgba(255, 255, 255, 0.1);
            }

            &.active {
                opacity: 1;
                background-color: rgba(255, 255, 255, 0.2);
            }
        }

        // 小屏幕布局调整
        @media (max-width: 768px) {
            display: flex;
            flex-wrap: wrap;
            gap: 8px;

            .select-time__group-item {
                margin-right: 0;
                flex: 1;
                min-width: 60px;
                justify-content: center;
            }
        }
    }

    .menu-sider {
        background: rgba(0, 0, 0);

        .side-menu {
            background: transparent;
            color: #fff;
            height: 100%;
        }

        :deep(.n-menu-item-content) {
            opacity: 0.7;
        }

        :deep(.n-menu-item-content--selected) {
            opacity: 1;
        }
    }

    .content-area {
        flex: 1;
        background: transparent;
        overflow-y: auto;
        min-width: 0;
        box-sizing: border-box;

        // 移动端内容区域调整
        @media (max-width: 1440px) {
            padding: 20px 0;
        }
    }

    .page-content {
        width: 100%;
        max-width: 1300px;
        margin: 0 auto;

        // 移动端页面内容调整
        @media (max-width: 1440px) {
            max-width: 100%;
            padding: 0;
        }
    }

    .usage,
    .activity,
    .subscription {
        .activity-title {
            font-size: 16px;
            font-weight: 600;
        }
    }

    .info {
        margin-top: 0;

        .btn-logout {
            background: rgba(255, 255, 255, 0.15);
            padding: 5.5px 16px;
        }
    }

    .subscription {
        .subscription-content {
            padding: 20px 0;

            .subscription-info {
                text-align: center;
                color: rgba(255, 255, 255, 0.7);
                font-size: 14px;

                p {
                    margin: 0;
                    line-height: 1.5;
                }
            }
        }
    }

    .info {
        .info-item {
            display: flex;
            align-items: center;
            flex-wrap: wrap;

            .item-account,
            .item-phone {
                display: flex;
                align-items: center;
                margin-right: 40px;

                .label {
                    opacity: 0.7;
                }

                span {
                    font-weight: 600;
                    color: #1876f2;
                }
            }

            .item-userId {
                display: flex;
                align-items: center;
                margin-right: 40px;

                .label {
                    opacity: 0.7;
                }
            }
        }
    }

    .usage {
        .usage-credit {
            background: linear-gradient(95deg, #2a7fff 5%, #ffffff 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            text-fill-color: transparent;
        }

        .credit-action {
            width: 130px;
            height: 24px;
            text-align: center;
            line-height: 24px;
            border-radius: 20px;
            background: rgba(255, 255, 255, 0.2);
            box-sizing: border-box;
            color: rgba(255, 255, 255, 0.9);
            font-size: 14px;
            margin-right: 12px;
            cursor: pointer;
            display: inline-block;
        }

        .credit-action::before {
            content: '';
            position: absolute;
            inset: 0;
            border-radius: inherit;
            padding: 1px;
            mask:
                linear-gradient(#fff 0 0) content-box,
                linear-gradient(#fff 0 0);
            mask-composite: exclude;
            -webkit-mask-composite: destination-out;
            pointer-events: none;
            opacity: 0.5;
        }

        .usage-content {
            display: flex;
            flex-direction: column;

            .content-title {
                font-size: 14px;
            }

            .content-desc {
                opacity: 0.7;
                margin-top: 8px;
            }
        }

        .content-progress {
            margin-top: 24px;
        }

        .usage-static {
            margin-top: 10px;
            display: flex;
            justify-content: space-between;
        }

        .credit-validity {
            margin-top: 30px;
        }

        .billing-btn {
            background: linear-gradient(95deg, #2a7fff 5%, #ffffff 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            text-fill-color: transparent;
        }
    }

    .usage-consumption {
        .select-time__group {
            &-item {
                padding: 2px 8px;
                background: rgba(255, 255, 255, 0.15);
            }
        }
    }
}

.modal-title {
    font-size: 14px;
    text-align: center;
}

:deep(.card-header-extra) {
    @media (max-width: 640px) {
        align-self: baseline;
        position: absolute;
        right: 0;
    }
}

// 英文版本的响应式样式
&.responsive-en {
    :deep(.card-header-extra) {
        @media (max-width: 830px) {
            align-self: baseline;
            position: absolute;
            right: 0;
        }
    }
    .select-time__group {
        @media (max-width: 830px) {
            display: flex;
            flex-wrap: wrap;
            gap: 8px;

            .select-time__group-item {
                margin-right: 0;
                flex: 1;
                min-width: 60px;
                justify-content: center;
                white-space: nowrap;
            }
        }
    }
}

// 菜单项new图标样式
.menu-label-with-new-icon {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;

    .new-icon {
        flex-shrink: 0;
    }
}
</style>
