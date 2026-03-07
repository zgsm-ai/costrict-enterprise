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

            <!-- 成功图标 -->
            <div class="mb-8">
                <img
                    src="../../../assets/subscribe/success.webp"
                    alt=""
                />
            </div>

            <!-- 成功标题 -->
            <div class="success-title text-white text-4xl mb-4 font-semibold text-center">
                {{ $t('paymentModals.paymentSuccess.title') }}
            </div>

            <!-- 提示文案 -->
            <div class="payment-tip text-center mb-4">
                <p class="text-white opacity-40 text-xl">
                    {{ $t('paymentModals.paymentSuccess.tip') }}
                </p>
            </div>

            <!-- 二维码区域 -->
            <div class="text-center mt-auto">
                <img
                    class="w-65"
                    src="../../../assets/subscribe/qrCode.webp"
                    alt=""
                />
            </div>
        </div>
    </n-modal>
</template>

<script setup lang="ts">
/**
 * @file 支付成功弹窗组件
 */
import { NModal, NButton, NIcon } from 'naive-ui';
import { CloseOutline } from '@vicons/ionicons5';
import { useRouter } from 'vue-router';

interface Props {
    modelValue: boolean;
}

interface Emits {
    (event: 'update:modelValue', value: boolean): void;
    (event: 'close'): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();
const router = useRouter();

// 关闭弹窗
const handleClose = () => {
    emit('update:modelValue', false);
    // emit('close');
    router.push({
        path: '/',
        query: {
            tab: 'subscription',
        },
    });
};
</script>

<style lang="less" scoped>
.payment-result-modal {
    width: 72%;
    height: 734px;
    border-radius: 7px;
    background: #1b1d22;
    overflow: hidden;
    padding: 56px 80px 110px;
}

@media (max-width: 768px) {
    .payment-result-modal {
        width: 90%;
        padding: 20px 20px 100px;
    }

    .success-title {
        font-size: 24px;
        margin-bottom: 16px;
    }

    .payment-tip p {
        font-size: 16px;
    }
}
</style>
