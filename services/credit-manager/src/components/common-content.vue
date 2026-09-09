<template>
    <div class="common-content">
        <div class="common-content-view">
            <MaintenanceNotice v-if="!isLoginPage" />
            <p
                v-if="!isLoginPage && !maintenance.ready"
                role="status"
            >
                {{ t('maintenance.loading') }}
            </p>
            <MaintenanceState
                v-else-if="purchaseBlocked"
                class="purchase-maintenance"
            >
                <template #footer>
                    <NButton @click="router.push('/?tab=profile')">{{
                        t('maintenance.back')
                    }}</NButton>
                </template>
            </MaintenanceState>
            <RouterView v-else />
        </div>
        <common-footer v-if="!isNoFooterPage" />
    </div>
</template>

<script setup lang="ts">
/**
 * @file common-content.vue
 */
import { computed, onMounted, onUnmounted } from 'vue';
import { NButton, useMessage } from 'naive-ui';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import CommonFooter from '@/components/common-footer.vue';
import MaintenanceNotice from '@/components/maintenance-notice.vue';
import MaintenanceState from '@/components/maintenance-state.vue';
import { NO_FOOTER_ROUTES } from '@/router';
import { useMaintenanceStore } from '@/store/maintenance';
import { setI18nComposer } from '@/utils/i18n';
import { setMessageInstance } from '@/utils/request';

const route = useRoute();
const isLoginPage = computed(() => route.name === 'login');

const isNoFooterPage = computed(() => {
    return NO_FOOTER_ROUTES.includes(route.path);
});

const message = useMessage();
setMessageInstance(message);
const composer = useI18n();
const { t } = composer;
setI18nComposer(composer);
const router = useRouter();
const maintenance = useMaintenanceStore();
const purchaseBlocked = computed(() => route.path === '/subscribe' && maintenance.orderAffected);
const removeGuard = router.beforeEach((to) => {
    maintenance.now = Date.now();
    if (
        (to.path === '/subscribe' && maintenance.orderAffected) ||
        (to.path === '/' && maintenance.isTabBlocked(to.query.tab))
    ) {
        message.warning(t('maintenance.unavailable', { end: maintenance.announcement?.end_time }));
        return false;
    }
});
let clockTimer: ReturnType<typeof setInterval>;
let refreshTimer: ReturnType<typeof setInterval>;
const refreshOnFocus = (): void => {
    void maintenance.refresh();
};
onMounted(() => {
    void maintenance.refresh();
    clockTimer = setInterval(() => {
        maintenance.now = Date.now();
    }, 1000);
    refreshTimer = setInterval(() => {
        void maintenance.refresh();
    }, 30000);
    window.addEventListener('focus', refreshOnFocus);
});
onUnmounted(() => {
    clearInterval(clockTimer);
    clearInterval(refreshTimer);
    window.removeEventListener('focus', refreshOnFocus);
    removeGuard();
});
</script>

<style lang="less" scoped>
.purchase-maintenance {
    width: 90%;
    margin: 32px auto;

    @media (max-width: 1440px) {
        width: calc(100% - 40px);
        margin: 20px auto;
    }
}

.common-content {
    height: calc(100vh - 68px);
    display: flex;
    flex-direction: column;
    justify-content: space-between;

    &-view {
        flex: 1;
        position: relative;
    }
}
</style>
