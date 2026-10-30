<template>
    <div class="profile-section">
        <div class="info-item">
            <div class="profile-field">
                <div class="label">{{ t('homePage.userNameLabel') }}</div>
                <span :class="{ 'ml-1': !isZh }">{{ userName || '-' }}</span>
            </div>
            <div class="profile-field max-w-[100%]">
                <div class="label whitespace-nowrap">{{ t('homePage.userIdLabel') }}</div>
                <span
                    v-hover="userId"
                    class="truncate flex-1"
                    :class="{ 'ml-1': !isZh }"
                    >{{ userId || '-' }}</span
                >
                <NIcon
                    v-if="userId"
                    size="14"
                    class="copy-icon"
                    color="#197DFF"
                    @click="emit('copy-user-id', userId)"
                >
                    <CopyOutline />
                </NIcon>
            </div>
            <div class="profile-field max-w-[100%]">
                <div class="label whitespace-nowrap line-through">
                    {{ t('homePage.universalIdLabel') }}
                </div>
                <span
                    v-hover="universalId"
                    class="truncate flex-1 line-through"
                    :class="{ 'ml-1': !isZh }"
                    >{{ universalId || '-' }}</span
                >
                <NIcon
                    v-if="universalId"
                    size="14"
                    class="copy-icon"
                    color="#197DFF"
                    @click="emit('copy-universal-id', universalId)"
                >
                    <CopyOutline />
                </NIcon>
            </div>
            <div class="profile-field">
                <div class="label">{{ t('homePage.employeeNumberLabel') }}</div>
                <span :class="{ 'ml-1': !isZh }">{{ employeeNumber || '-' }}</span>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { NIcon } from 'naive-ui';
import { CopyOutline } from '@vicons/ionicons5';

const { t } = useI18n();

interface Props {
    userName?: string;
    userId?: string;
    universalId?: string;
    employeeNumber?: string;
    isZh?: boolean;
}

withDefaults(defineProps<Props>(), {
    userName: '',
    userId: '',
    universalId: '',
    employeeNumber: '',
    isZh: false,
});

const emit = defineEmits<{
    'copy-user-id': [value: string];
    'copy-universal-id': [value: string];
}>();
</script>

<style scoped lang="less">
.profile-section {
    .info-item {
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        column-gap: 40px;
        row-gap: 10px;
    }

    .profile-field {
        display: flex;
        align-items: center;
        min-width: 0;

        .label {
            opacity: 0.7;
        }
    }

    .copy-icon {
        margin-left: 8px;
        cursor: pointer;
        flex-shrink: 0;
    }
}
</style>
