<template>
    <div
        v-if="isAuthenticating"
        class="auth-loading-overlay"
    >
        <div class="auth-loading-content">
            <n-spin
                size="large"
                stroke="#1876F2"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { NSpin } from 'naive-ui';
import { storeToRefs } from 'pinia';
import { useUserStore } from '@/store/user';

const userStore = useUserStore();
const { isAuthenticating } = storeToRefs(userStore);
</script>

<style lang="less" scoped>
.auth-loading-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 9999;
}
</style>
