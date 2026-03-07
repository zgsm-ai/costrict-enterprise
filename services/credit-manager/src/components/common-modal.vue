<template>
    <n-modal
        preset="card"
        :show="show"
        @update:show="(v: boolean) => emit('update:show', v)"
        :style="style"
        :bordered="bordered"
        :mask-closable="false"
    >
        <template #header>
            <div class="modal-title">{{ title }}</div>
        </template>

        <n-spin :show="showSpin">
            <slot></slot>
        </n-spin>

        <template #action>
            <slot name="action"></slot>
        </template>
    </n-modal>
</template>

<script setup lang="ts">
import { NModal, NSpin } from 'naive-ui';
import type { CSSProperties } from 'vue';

withDefaults(
    defineProps<{
        title: string;
        show: boolean;
        style?: CSSProperties | string;
        bordered?: boolean;
        showSpin?: boolean;
    }>(),
    {
        title: '',
        show: false,
        style: 'width: 490px;max-height: 750px;overflow: auto',
        bordered: true,
        showSpin: false,
    },
);

const emit = defineEmits<{
    (e: 'update:show', v: boolean): void;
}>();
</script>
