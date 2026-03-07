<template>
    <n-modal
        :show="modelValue"
        :mask-closable="false"
        :close-on-esc="false"
        transform-origin="center"
        @update:show="(value: boolean) => emit('update:modelValue', value)"
    >
        <div
            class="payment-modal flex flex-col"
            role="dialog"
            aria-modal="true"
        >
            <!-- 弹窗头部 -->
            <div class="payment-modal-header">
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

            <div class="payment-title text-white text-4xl mb-10 font-semibold text-center">
                {{ $t('subscribePage.paymentModal.title') }}
                <span class="payment-price">{{ displayPrice }}</span>
                {{ $t('subscribePage.paymentModal.unit') }}
            </div>

            <!-- 弹窗内容 -->
            <div class="payment-modal-content">
                <!-- 二维码占位区域 -->
                <div class="qr-code-placeholder mb-10">
                    <template v-if="isLoadingQrcode">
                        <div class="qr-placeholder-text">
                            {{ $t('subscribePage.paymentModal.loading') }}
                        </div>
                    </template>
                    <template v-else-if="qrCodeData">
                        <canvas
                            ref="qrCanvas"
                            class="qr-code-canvas"
                            width="230"
                            height="230"
                        ></canvas>
                    </template>
                    <template v-else>
                        <div class="qr-placeholder-text">
                            {{ $t('subscribePage.paymentModal.qrCodeError') }}
                        </div>
                    </template>
                </div>
                <p class="flex items-center text-white opacity-70 text-2xl font-semibold">
                    {{ $t('subscribePage.paymentModal.instruction') }}
                    <img
                        :src="paymentMethodIcon"
                        :alt="paymentMethodName"
                        class="text-xl ml-2"
                    />
                </p>
            </div>

            <!-- 支付提示 -->
            <div class="payment-tips mt-auto opacity-40">
                <p class="service-info text-xl font-semibold">
                    {{ $t('subscribePage.paymentModal.agreementText') }}
                </p>
            </div>
        </div>
    </n-modal>
</template>

<script setup lang="ts">
/**
 * @file 支付弹窗组件
 */
import { NModal, NButton, NIcon } from 'naive-ui';
import { CloseOutline } from '@vicons/ionicons5';
import { computed, ref, watch, nextTick } from 'vue';
import { postPaymentQrcode } from '@/api/mods/quota.mod';
import type { PostPaymentQrcodeReq } from '@/api/bos/quota.bo';
import QRCode from 'qrcode';
import { useOrderPolling } from '../hook/useOrderPolling';
import type { PaymentModalEmits, PaymentModalProps } from '../interface';
import { CMB } from '../const';

const props = defineProps<PaymentModalProps>();
const emit = defineEmits<PaymentModalEmits>();

// 使用订单轮询 hook，传递支付成功和失败的回调函数
const { startPolling, stopPolling } = useOrderPolling();

// 计算支付方式图标和名称
const paymentMethodIcon = computed(() => {
    if (!props.selectedPaymentMethod) return '';
    return (
        props.paymentMethods.find((method) => method.id === props.selectedPaymentMethod)?.icon || ''
    );
});

const paymentMethodName = computed(() => {
    if (!props.selectedPaymentMethod) return '';
    return (
        props.paymentMethods.find((method) => method.id === props.selectedPaymentMethod)?.name || ''
    );
});

// 从 orderInfo 中获取价格
const displayPrice = computed(() => {
    return props.orderInfo?.amount || 0;
});

// 二维码相关状态
const qrCodeData = ref<string>('');
const qrCanvas = ref<HTMLCanvasElement | null>(null);
const isLoadingQrcode = ref<boolean>(false);

// 生成二维码
const generateQRCode = async (data: string) => {
    if (!qrCanvas.value || !data) {
        return;
    }

    const ctx = qrCanvas.value.getContext('2d');
    if (ctx) {
        ctx.clearRect(0, 0, qrCanvas.value.width, qrCanvas.value.height);
    }

    await QRCode.toCanvas(qrCanvas.value, data, {
        width: 230,
        margin: 2,
        color: {
            dark: '#000000',
            light: '#FFFFFF',
        },
        errorCorrectionLevel: 'M',
    });
};

// 获取支付二维码
const fetchPaymentQrcode = async () => {
    if (!props.modelValue || !props.orderInfo?.order_id) {
        return;
    }

    isLoadingQrcode.value = true;
    const request: PostPaymentQrcodeReq = {
        order_id: props.orderInfo.order_id,
        payment_channel: CMB,
    };

    const response = await postPaymentQrcode(request).finally(() => {
        isLoadingQrcode.value = false;
    });

    if (response.code === 200 && response.data.qrCode) {
        qrCodeData.value = response.data.qrCode;
    } else {
        qrCodeData.value = '';
        stopPolling();
    }
};

// 监听弹窗打开和订单信息变化
watch(
    () => [props.modelValue, props.orderInfo?.order_id],
    ([isVisible, orderId]) => {
        if (isVisible && orderId) {
            // 重置二维码状态
            qrCodeData.value = '';
            fetchPaymentQrcode();
        } else if (!isVisible) {
            // 弹窗关闭时清理状态
            qrCodeData.value = '';
            isLoadingQrcode.value = false;
            // 停止轮询
            stopPolling();
        }
    },
    { immediate: true },
);

// 监听二维码数据变化，自动生成二维码并开始轮询
watch(qrCodeData, async (newData) => {
    if (newData && props.modelValue && props.orderInfo?.order_id) {
        await nextTick();
        generateQRCode(newData);
        // 二维码生成成功后开始轮询订单状态
        startPolling(props.orderInfo.order_id);
    }
});

// 关闭弹窗
const handleClose = () => {
    // 关闭弹窗时停止轮询
    stopPolling();
    emit('update:modelValue', false);
    emit('close');
};
</script>

<style lang="less" scoped>
.payment-modal {
    width: 650px;
    height: 650px;
    border-radius: 7px;
    background: #1b1d22;
    border-radius: 12px;
    overflow: hidden;
    padding: 36px 63px 93px;
}

.payment-modal-header {
    display: flex;
    margin-bottom: 42px;
}

.close-btn {
    margin-left: auto;
}

.payment-price {
    background: linear-gradient(101deg, #00ffb7 2%, #ffffff 68%, #c5dbff 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    text-fill-color: transparent;
}

.payment-modal-content {
    display: flex;
    flex-direction: column;
    align-items: center;
}

.qr-code-placeholder {
    width: 250px;
    height: 250px;
    background: #fff;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid #e0e0e0;
}

.qr-placeholder-text {
    color: #666;
    text-align: center;
    font-size: 14px;
    line-height: 1.5;
}

.qr-code-image {
    width: 100%;
    height: 100%;
    object-fit: contain;
}

.qr-code-canvas {
    width: 230px;
    height: 230px;
    border-radius: 6px;
    display: block;
    margin: 0 auto;
}

.payment-tips {
    text-align: center;
    color: #fff;
}

.payment-icon {
    width: 26px;
    height: 26px;
    object-fit: contain;
}

@media (max-width: 768px) {
    .payment-modal {
        width: 90%;
        height: 90%;
        padding: 15px 15px 30px;
    }

    .payment-modal-header {
        padding: 10px 0;
        margin-bottom: 10px;
    }
}
</style>
