<template>
    <div class="subscribe-page text-white">
        <!-- 面包屑导航 -->
        <div class="breadcrumb-container">
            <n-breadcrumb separator=">">
                <n-breadcrumb-item href="/credit/manager/?tab=subscription">{{
                    $t('subscribePage.breadcrumb.price')
                }}</n-breadcrumb-item>
                <n-breadcrumb-item>{{ $t('subscribePage.breadcrumb.purchase') }}</n-breadcrumb-item>
            </n-breadcrumb>
        </div>

        <div class="subscribe-content">
            <!-- 流量套餐标题和数量选择器 -->
            <div class="mb-1">
                <div class="package-info flex items-center flex-wrap">
                    <span class="text-xl font-semibold mr-5">{{ currentPlan?.title }}</span>
                    <p class="text-[#a7abb4] leading-6">
                        {{ $t('subscribePage.package.description') }}
                    </p>
                </div>
            </div>

            <!-- 价格显示 -->
            <div class="package-price flex mb-4 items-center">
                <div class="price-wrapper">
                    <span class="currency text-3xl">¥</span>
                    <span class="price text-3xl ml-1">{{ unitPrice }}</span>
                    <span class="unit text-base text-[#A7ABB4] self-end mb-0.5 ml-1">{{
                        $t('subscribePage.package.unit')
                    }}</span>
                    <span
                        v-if="currentPlan?.isFirstPurchase"
                        class="text-xs price__tips"
                        >{{ $t('subscribePage.package.tips') }}</span
                    >
                </div>

                <div class="quantity-selector ml-auto">
                    <n-form
                        ref="formRef"
                        :model="formData"
                        :rules="formRules"
                    >
                        <n-form-item path="quantity">
                            <n-input-number
                                :value="quantity"
                                @update:value="handleQuantityChange"
                                :min="1"
                                :show-button="true"
                                button-placement="both"
                                size="large"
                                :placeholder="$t('subscribePage.quantityPlaceholder')"
                                :precision="0"
                            />
                        </n-form-item>
                    </n-form>
                </div>
            </div>

            <n-divider />

            <!-- 价格汇总 -->
            <div class="flex items-end justify-end mb-10">
                <span class="text-base font-semibold">{{ $t('subscribePage.total') }}</span>
                <div class="flex items-end ml-1">
                    <span class="text-base">¥</span>
                    <span class="text-3xl ml-1 total-price">{{ totalPrice }}</span>
                </div>
            </div>

            <!-- 支付方式 -->
            <div class="payment-section mb-13">
                <h2 class="text-xl font-semibold leading-8 mb-4">
                    {{ $t('subscribePage.paymentMethod') }}
                </h2>

                <n-radio-group
                    v-model:value="selectedPaymentMethod"
                    class="payment-radio-group"
                >
                    <div
                        v-for="method in paymentMethods"
                        :key="method.id"
                        class="payment-option"
                    >
                        <n-radio
                            :value="method.id"
                            class="payment-radio"
                            size="large"
                        >
                            <div class="payment-content flex items-center">
                                <img
                                    class="mr-3 ml-3"
                                    :src="method.icon"
                                    :alt="method.name"
                                />
                                <span class="text-white">{{ method.name }}</span>
                            </div>
                        </n-radio>
                    </div>
                </n-radio-group>
            </div>

            <!-- 协议同意 -->
            <div class="mb-4">
                <n-checkbox v-model:checked="agreeToTerms">
                    <div>
                        <span class="text-[#999]">{{ $t('subscribePage.agreement.prefix') }}</span>
                        <a
                            href="https://docs.costrict.ai/billing/purchase"
                            target="_blank"
                            class="text-[#197DFF]"
                            @click.stop
                            >{{ $t('subscribePage.agreement.linkText') }}</a
                        >
                    </div>
                </n-checkbox>
            </div>

            <!-- 购买按钮 -->
            <n-button
                type="primary"
                size="large"
                class="purchase-btn"
                :disabled="!selectedPaymentMethod || quantityError || isPurchasing"
                :loading="isPurchasing"
                @click="handlePurchaseWithModal"
                block
            >
                {{ $t('subscribePage.purchase') }}
            </n-button>
        </div>

        <!-- 支付弹窗 -->
        <PaymentModal
            v-model="showPaymentModal"
            :selected-payment-method="selectedPaymentMethod"
            :payment-methods="paymentMethods"
            :order-info="orderInfo"
            @close="closePaymentModal"
            @payment-success="showPaymentSuccess"
            @payment-failure="showPaymentFailure"
        />

        <!-- 支付成功弹窗 -->
        <PaymentSuccessModal
            v-model="showSuccessModal"
            @close="closeSuccessModal"
        />

        <!-- 支付失败弹窗 -->
        <PaymentFailureModal
            v-model="showFailureModal"
            @close="closeFailureModal"
        />

        <!-- 协议确认弹窗 -->
        <n-modal
            v-model:show="showAgreementModal"
            preset="dialog"
            :mask-closable="false"
            :show-icon="false"
            :title="$t('subscribePage.agreement.modalTitle')"
        >
            <template #header>
                {{ $t('subscribePage.agreement.modalTitle') }}
            </template>
            {{ $t('subscribePage.agreement.modalContent') }}
            <a
                href="https://docs.costrict.ai/billing/purchase"
                target="_blank"
                class="text-[#197DFF]"
                @click.stop
                >{{ $t('subscribePage.agreement.linkText') }}</a
            >
            <template #action>
                <div class="flex justify-end gap-3">
                    <n-button
                        type="info"
                        size="medium"
                        @click="handleConfirmAgreementAndProceed"
                    >
                        {{ $t('subscribePage.agreement.confirm') }}
                    </n-button>
                    <n-button
                        size="medium"
                        class="cancel-btn"
                        @click="handleCancelAgreement"
                    >
                        {{ $t('subscribePage.agreement.cancel') }}
                    </n-button>
                </div>
            </template>
        </n-modal>
    </div>
</template>

<script setup lang="ts">
/**
 * @file 订阅页面
 */
import {
    NBreadcrumb,
    NBreadcrumbItem,
    NButton,
    NCheckbox,
    NInputNumber,
    NDivider,
    NRadioGroup,
    NRadio,
    NForm,
    NFormItem,
    NModal,
} from 'naive-ui';
import { useSubscribe } from './hook/useSubscribe';
import { usePaymentModal } from './hook/usePaymentModal';
import { useOrderPolling } from './hook/useOrderPolling';
import PaymentModal from './components/PaymentModal.vue';
import PaymentSuccessModal from './components/PaymentSuccessModal.vue';
import PaymentFailureModal from './components/PaymentFailureModal.vue';

// 使用支付弹窗Hook
const {
    showPaymentModal,
    showSuccessModal,
    showFailureModal,
    openPaymentModal,
    closePaymentModal,
    showPaymentSuccess,
    showPaymentFailure,
    closeSuccessModal,
    closeFailureModal,
} = usePaymentModal();

// 使用订单轮询Hook
const { startPolling } = useOrderPolling({
    onSuccess: () => showPaymentSuccess(),
    onFailure: () => showPaymentFailure(),
    onCancelled: () => showPaymentFailure(),
});

// 使用订阅页面Hook
const {
    // 表单相关
    formRef,
    formData,
    formRules,

    // 数据状态
    quantity,
    selectedPaymentMethod,
    agreeToTerms,
    quantityError,
    currentPlan,

    // 计算属性
    unitPrice,
    totalPrice,
    paymentMethods,

    // 方法
    handleQuantityChange,

    // 协议相关
    showAgreementModal,
    handleCancelAgreement,

    // 购买相关
    orderInfo,
    isPurchasing,
    handlePurchaseWithModal,
    handleConfirmAgreementAndProceed,
} = useSubscribe({
    openPaymentModal,
    startPolling,
});
</script>

<style lang="less" scoped>
.subscribe-page {
    width: 44%;
    margin: auto;
    color: #fff;
}

.breadcrumb-container {
    margin-bottom: 30px;

    :deep(.n-breadcrumb-item__link) {
        color: #888;
        text-decoration: none;

        &:hover {
            color: #fff;
        }
    }

    :deep(.n-breadcrumb-item__separator) {
        color: #666;
    }
}

.quantity-selector {
    width: 150px;

    :deep(.n-input-number) {
        width: 100%;
    }

    :deep(.n-input__input-el) {
        text-align: center;
        color: #fff;
    }
}

.payment-radio-group {
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.3);
    overflow: hidden;
    width: 100%;
}

.payment-option {
    padding: 27px 0 27px 29px;
    background: rgba(255, 255, 255, 0.05);

    &:not(:last-child) {
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    }
}

.payment-radio {
    width: 100%;
    align-items: center;

    :deep(.n-radio__dot::before) {
        left: 50%;
        top: 50%;
        transform: translate(-50%, -50%);
        width: calc(100%-10px);
        height: calc(100%-10px);
    }
}

.purchase-btn {
    background: linear-gradient(90deg, #005eff -9%, #00ffb7 80%);
    backdrop-filter: blur(10px);
    height: 50px;
}

.price__tips {
    background: linear-gradient(102deg, #00ffb7 3%, #ffffff 68%, #ffffff 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
}

:deep(.n-divider) {
    margin-top: 0;
}

:deep(.n-input__suffix) {
    .n-button {
        color: #fff;
    }
}

:deep(.n-input__prefix) {
    .n-button {
        color: #fff;
    }
}

:deep(.cancel-btn) {
    color: #000;
}

:deep(.cancel-btn):hover {
    color: inherit;
    background: rgba(0, 0, 0, 0.1);
}

@media (max-width: 768px) {
    .subscribe-page {
        width: 95%;
        padding: 15px;
    }

    .payment-title {
        font-size: 20px;
        margin-bottom: 12px;
    }

    .qr-code-placeholder {
        width: 160px;
        height: 160px;
        margin-bottom: 16px;
    }

    .scan-tip {
        font-size: 13px;
        margin-bottom: 6px;
    }

    .service-info {
        font-size: 12px;
        line-height: 1.4;
    }

    .payment-icon {
        width: 16px;
        height: 16px;
    }

    .payment-modal-content {
        padding: 10px;
        gap: 16px;
        p {
            font-size: 14px;
        }
        img {
            width: 14px;
        }
        .payment-test-buttons {
            padding-top: 16px;
            flex-direction: column;
            gap: 8px;

            .n-button {
                font-size: 12px;
                height: 36px;
            }
        }
    }
}
</style>
