<template>
    <!-- 用户资料图标 -->
    <img
        v-if="name === 'profile'"
        src="../../../assets/icon/info.svg"
        alt=""
    />

    <!-- 订阅管理图标 -->
    <img
        v-else-if="name === 'subscription'"
        src="../../../assets/icon/subscribe.svg"
        alt=""
    />

    <!-- 用量统计图标 -->
    <img
        v-else-if="name === 'usage'"
        src="../../../assets/icon/usage.svg"
        alt=""
    />

    <!-- 活动记录图标 -->
    <img
        v-else-if="name === 'activity'"
        src="../../../assets/icon/activity.svg"
        alt=""
    />
</template>

<script setup lang="ts">
/**
 * @file menu-icons.vue
 * @description 菜单图标组件 - 提供自定义的菜单图标
 */

interface Props {
    name: 'profile' | 'subscription' | 'usage' | 'activity';
}

defineProps<Props>();
</script>

<style scoped lang="less">
svg {
    display: inline-block;
    vertical-align: middle;
}
</style>
