<template>
    <div class="common-card">
        <div
            class="card-header-wrapper mb-4"
            v-if="$slots.header || title"
        >
            <slot name="header">
                <div class="card-header">{{ title }}</div>
            </slot>
            <div
                class="card-header-extra"
                v-if="$slots['header-extra']"
            >
                <slot name="header-extra"></slot>
            </div>
        </div>

        <div
            class="card-content"
            :class="{ 'with-border': contentBorder }"
        >
            <slot></slot>
        </div>

        <div
            class="card-footer"
            v-if="$slots.footer"
        >
            <slot name="footer"></slot>
        </div>

        <div
            class="card-action"
            v-if="$slots.action"
        >
            <slot name="action"></slot>
        </div>
    </div>
</template>

<script lang="ts" setup>
withDefaults(
    defineProps<{
        title?: string;
        contentBorder?: boolean;
    }>(),
    {
        title: '',
        contentBorder: true,
    },
);

defineSlots<{
    default?: (props: Record<string, unknown>) => unknown;
    header?: (props: Record<string, unknown>) => unknown;
    'header-extra'?: (props: Record<string, unknown>) => unknown;
    footer?: (props: Record<string, unknown>) => unknown;
    action?: (props: Record<string, unknown>) => unknown;
}>();
</script>

<style lang="less" scoped>
.common-card {
    border-radius: 10px;
    box-sizing: border-box;
    color: #fff;
    backdrop-filter: blur(10px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

    .card-header-wrapper {
        display: flex;
        justify-content: space-between;
        align-items: center;

        .card-header {
            color: #fff;
            font-size: 16px;
        }

        .card-header-extra {
            margin-left: auto;
        }
    }

    .card-content {
        margin-bottom: 12px;

        &.with-border {
            padding: 20px 24px;
            border: 1px solid rgba(255, 255, 255, 0.3);
            border-radius: 10px;
        }
    }

    .card-footer {
        margin-top: 12px;
        padding-top: 12px;
        border-top: 1px solid rgba(255, 255, 255, 0.1);
    }

    .card-action {
        margin-top: 12px;
        padding-top: 12px;
        border-top: 1px solid rgba(255, 255, 255, 0.1);
    }
}
</style>
