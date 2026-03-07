<template>
    <n-modal
        :show="modelValue"
        :mask-closable="false"
        :close-on-esc="false"
        transform-origin="center"
        @update:show="(value: boolean) => emit('update:modelValue', value)"
    >
        <div
            class="payment-result-modal flex flex-col items-center"
            role="dialog"
            aria-modal="true"
        >
            <!-- 弹窗头部 -->
            <div class="payment-modal-header self-end flex mb-5">
                <!-- 关闭按钮 -->
                <n-button
                    quaternary
                    circle
                    class="close-btn"
                    @click="handleClose"
                >
                    <template #icon>
                        <n-icon>
                            <CloseOutline />
                        </n-icon>
                    </template>
                </n-button>
            </div>

            <!-- 失败图标 -->
            <div class="mb-8">
                <img
                    src="../../../assets/subscribe/failure.webp"
                    alt=""
                />
            </div>

            <!-- 失败标题 -->
            <div class="failure-title text-white text-4xl mb-4 font-semibold text-center">
                {{ $t('paymentModals.paymentFailure.title') }}
            </div>

            <!-- 提示文案 -->
            <div class="payment-tip text-center mb-8">
                <p class="text-white opacity-40 text-xl">
                    {{ $t('paymentModals.paymentFailure.tip') }}
                </p>
            </div>

            <!-- 返回按钮 -->
            <div
                class="failure-btn h-12 w-22.5 text-xl text-center leading-12 rounded-[3px] text-white cursor-pointer"
                @click="handleClose"
            >
                {{ $t('paymentModals.paymentFailure.returnButton') }}
            </div>
        </div>
    </n-modal>
</template>

<script setup lang="ts">
/**
 * @file 支付失败弹窗组件
 */
import { NModal, NButton, NIcon } from 'naive-ui';
import { CloseOutline } from '@vicons/ionicons5';

interface Props {
    modelValue: boolean;
}

interface Emits {
    (event: 'update:modelValue', value: boolean): void;
    (event: 'close'): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();

// 关闭弹窗
const handleClose = () => {
    emit('update:modelValue', false);
    emit('close');
};
</script>

<style lang="less" scoped>
.payment-result-modal {
    width: 72%;
    height: 734px;
    border-radius: 7px;
    background: #1b1d22;
    overflow: hidden;
    padding: 66px 80px 110px;
}

.failure-btn {
    background: #1770e6;
}

@media (max-width: 768px) {
    .payment-result-modal {
        width: 90%;
        height: auto;
        padding: 20px 20px 100px;
    }

    .failure-title {
        font-size: 24px;
        margin-bottom: 16px;
    }

    .payment-tip p {
        font-size: 16px;
    }
}
</style>
