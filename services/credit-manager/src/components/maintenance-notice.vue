<template>
    <NAlert
        v-if="maintenance.active && announcement"
        type="warning"
        :title="announcement.title || t('maintenance.title')"
        class="maintenance-notice"
        role="status"
    >
        <template #icon>
            <NIcon aria-hidden="true"><BuildOutline /></NIcon>
        </template>
        <p class="maintenance-time">
            {{
                t('maintenance.time', {
                    start: announcement.start_time,
                    end: announcement.end_time,
                })
            }}
        </p>
        <p v-if="maintenance.quotaAffected">
            {{ t('maintenance.quota', { description: announcement.impacts.quota.description }) }}
        </p>
        <p v-if="maintenance.orderAffected">
            {{ t('maintenance.order', { description: announcement.impacts.order.description }) }}
        </p>
        <p v-if="!maintenance.quotaAffected && !maintenance.orderAffected">
            {{ t('maintenance.noImpact') }}
        </p>
    </NAlert>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { BuildOutline } from '@vicons/ionicons5';
import { NAlert, NIcon } from 'naive-ui';
import { useI18n } from 'vue-i18n';
import { useMaintenanceStore } from '@/store/maintenance';

const { t } = useI18n();
const maintenance = useMaintenanceStore();
const { announcement } = storeToRefs(maintenance);
</script>

<style scoped lang="less">
.maintenance-notice {
    width: 90%;
    margin: 24px auto 0;
    padding: 20px 24px;
    box-sizing: border-box;

    :deep(.n-alert__title) {
        font-size: 16px;
        font-weight: 500;
        line-height: 1.5;
    }

    .maintenance-time {
        margin: 8px 0 12px;
        color: rgba(255, 255, 255, 0.7);
        font-variant-numeric: tabular-nums;
    }

    @media (max-width: 1440px) {
        width: calc(100% - 40px);
        margin-top: 20px;
    }

    @media (max-width: 480px) {
        padding: 16px;
    }

    p {
        margin: 6px 0;
        overflow-wrap: anywhere;
    }
}
</style>
