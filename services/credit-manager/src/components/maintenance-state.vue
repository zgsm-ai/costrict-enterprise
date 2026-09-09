<template>
    <NResult
        class="maintenance-state"
        status="warning"
        :title="t('maintenance.title')"
        :description="t('maintenance.unavailable', { end: maintenance.announcement?.end_time })"
        :theme-overrides="resultTheme"
        role="status"
    >
        <template #icon>
            <span
                class="maintenance-state-icon"
                aria-hidden="true"
            >
                <NIcon :size="28"><BuildOutline /></NIcon>
            </span>
        </template>
        <template
            v-if="$slots.footer"
            #footer
        >
            <slot name="footer" />
        </template>
    </NResult>
</template>

<script setup lang="ts">
import { BuildOutline } from '@vicons/ionicons5';
import { NIcon, NResult, type GlobalThemeOverrides } from 'naive-ui';
import { useI18n } from 'vue-i18n';
import { useMaintenanceStore } from '@/store/maintenance';

const { t } = useI18n();
const maintenance = useMaintenanceStore();
const resultTheme: GlobalThemeOverrides['Result'] = {
    titleTextColor: '#fff',
    textColor: 'rgba(255, 255, 255, 0.7)',
    titleFontSizeMedium: '20px',
    titleFontWeight: '500',
};
defineSlots<{ footer?: () => unknown }>();
</script>

<style scoped lang="less">
.maintenance-state {
    padding: 48px 24px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.04);
    overflow-wrap: anywhere;

    &-icon {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 64px;
        height: 64px;
        border: 1px solid rgba(255, 255, 255, 0.15);
        border-radius: 50%;
        background: rgba(255, 255, 255, 0.06);
        color: rgba(255, 255, 255, 0.7);
    }

    :deep(.n-result-content) {
        max-width: 560px;
        margin: 12px auto 0;
        line-height: 1.8;
    }

    :deep(.n-button) {
        min-height: 44px;
    }

    @media (max-width: 480px) {
        padding: 32px 20px;
    }
}
</style>
