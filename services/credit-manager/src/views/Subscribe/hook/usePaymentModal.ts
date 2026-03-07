/**
 * @file 支付弹窗逻辑Hook
 */
import { ref } from 'vue';

export function usePaymentModal() {
    // 弹窗显示状态
    const showPaymentModal = ref(false);
    const showSuccessModal = ref(false);
    const showFailureModal = ref(false);

    // 显示支付弹窗
    const openPaymentModal = () => {
        showPaymentModal.value = true;
    };

    // 关闭支付弹窗
    const closePaymentModal = () => {
        showPaymentModal.value = false;
    };

    // 显示支付成功弹窗
    const showPaymentSuccess = () => {
        closePaymentModal();
        showSuccessModal.value = true;
    };

    // 显示支付失败弹窗
    const showPaymentFailure = () => {
        closePaymentModal();
        showFailureModal.value = true;
    };

    // 关闭成功弹窗
    const closeSuccessModal = () => {
        showSuccessModal.value = false;
    };

    // 关闭失败弹窗
    const closeFailureModal = () => {
        showFailureModal.value = false;
    };

    return {
        // 状态
        showPaymentModal,
        showSuccessModal,
        showFailureModal,

        // 方法
        openPaymentModal,
        closePaymentModal,
        showPaymentSuccess,
        showPaymentFailure,
        closeSuccessModal,
        closeFailureModal,
    };
}
