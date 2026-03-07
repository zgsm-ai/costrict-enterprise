<template>
    <div class="app-container">
        <common-header />
        <div class="content">
            <n-config-provider
                :theme-overrides="themeOverrides"
                :locale="zhCN"
            >
                <n-message-provider>
                    <common-content />
                </n-message-provider>
            </n-config-provider>
        </div>
        <auth-loading />
    </div>
</template>

<script setup lang="ts">
import { NMessageProvider, NConfigProvider, zhCN } from 'naive-ui';
import CommonHeader from '@/components/common-header.vue';
import CommonContent from '@/components/common-content.vue';
import AuthLoading from '@/components/auth-loading.vue';
import { themeOverrides } from '@/theme-overrides';
</script>

<style lang="less" scoped>
.app-container {
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    box-sizing: border-box;
    background-color: #000;

    .content {
        flex: 1;
        overflow: auto;
        padding: 0 48px;
        box-sizing: border-box;

        @media (max-width: 480px) {
            padding: 0;
        }
    }
}
</style>
