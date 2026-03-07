<template>
    <common-modal
        :title="t('creditCodeModal.title')"
        :show="show"
        @update:show="(v: boolean) => emit('update:show', v)"
    >
        <n-form ref="formRef">
            <n-form-item
                path="receiverId"
                :label="t('creditCodeModal.codeLabel')"
                :show-feedback="false"
            >
                <n-input
                    :value="voucherCode"
                    disabled
                    :placeholder="t('creditCodeModal.codePlaceholder')"
                    :bordered="false"
                />
                <n-tooltip trigger="hover">
                    <template #trigger>
                        <n-icon
                            size="14"
                            style="margin-left: 8px; cursor: pointer"
                            color="#197DFF"
                            @click="copyCode"
                        >
                            <copy-outline />
                        </n-icon>
                    </template>
                    {{ t('creditCodeModal.copyTooltip') }}
                </n-tooltip>
            </n-form-item>
            <p class="tips">{{ t('creditCodeModal.tip1') }}</p>
            <p class="tips">{{ t('creditCodeModal.tip2') }}</p>
        </n-form>
    </common-modal>
</template>

<script setup lang="ts">
import CommonModal from '@/components/common-modal.vue';
import { NForm, NFormItem, NInput, NIcon, NTooltip } from 'naive-ui';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { CopyOutline } from '@vicons/ionicons5';
import { useMessage } from 'naive-ui';
import { copyToClipboard } from '@/utils/copy';

const { t } = useI18n();
const message = useMessage();

const props = withDefaults(
    defineProps<{
        show: boolean;
        voucherCode: string;
    }>(),
    {
        show: false,
        voucherCode: '',
    },
);

const emit = defineEmits<{
    (e: 'update:show', v: boolean): void;
}>();

const formRef = ref();

const copyCode = () => {
    copyToClipboard(props.voucherCode, {
        success: message.success,
        error: message.error,
    });
};
</script>

<style lang="less" scoped>
.tips {
    margin-top: 10px;
    font-size: 12px;
    line-height: 16px;
    color: rgba(255, 255, 255, 0.5);
    margin-bottom: 10px;
}

.tips:nth-of-type(1) {
    margin-bottom: 0;
}

.tips:nth-of-type(2) {
    margin-top: 0;
}
</style>
