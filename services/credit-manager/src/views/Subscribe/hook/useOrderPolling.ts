/**
 * @file 订单状态轮询 hook
 */
import { ref, onUnmounted } from 'vue';
import { getOrderById } from '@/api/mods/quota.mod';
import { ORDER_STATUS, POLLING_INTERVAL } from '../const';
import type { Order } from '@/api/bos/quota.bo';

// 全局轮询状态，防止重复轮询同一个订单
const globalPollingState = ref<{
    orderId: string | null;
    instanceCount: number;
}>({
    orderId: null,
    instanceCount: 0,
});

export const useOrderPolling = (callbacks?: {
    onSuccess?: (data: Order) => void;
    onFailure?: (data: Order) => void;
    onCancelled?: (data: Order) => void;
}) => {
    // 轮询相关状态
    const pollingTimer = ref<number | null>(null);
    const isPolling = ref<boolean>(false);

    // 轮询订单状态
    const pollOrderStatus = async (orderId: string) => {
        if (!orderId || !isPolling.value) {
            return;
        }

        const response = await getOrderById(orderId);

        if (response.code === 200 && response.data) {
            const { status } = response.data;

            // 根据状态进行处理
            switch (status) {
                case ORDER_STATUS.COMPLETED:
                    stopPolling();
                    callbacks?.onSuccess?.(response.data);
                    break;
                case ORDER_STATUS.FAILED:
                    stopPolling();
                    callbacks?.onFailure?.(response.data);
                    break;
                case ORDER_STATUS.CANCELLED:
                    stopPolling();
                    callbacks?.onCancelled?.(response.data);
                    break;
                default:
                    break;
            }
        }
    };

    // 开始轮询
    const startPolling = (orderId: string) => {
        if (isPolling.value || !orderId) {
            return;
        }

        // 检查是否已经在轮询同一个订单
        if (
            globalPollingState.value.orderId === orderId &&
            globalPollingState.value.instanceCount > 0
        ) {
            return;
        }

        isPolling.value = true;
        globalPollingState.value.orderId = orderId;
        globalPollingState.value.instanceCount += 1;

        // 立即执行第一次轮询
        pollOrderStatus(orderId);

        // 设置定时器，在间隔时间后执行后续轮询
        pollingTimer.value = window.setInterval(() => {
            if (isPolling.value) {
                pollOrderStatus(orderId);
            }
        }, POLLING_INTERVAL);
    };

    // 停止轮询
    const stopPolling = () => {
        if (pollingTimer.value) {
            clearInterval(pollingTimer.value);
            pollingTimer.value = null;
        }
        isPolling.value = false;

        // 更新全局状态
        if (globalPollingState.value.instanceCount > 0) {
            globalPollingState.value.instanceCount -= 1;
            if (globalPollingState.value.instanceCount === 0) {
                globalPollingState.value.orderId = null;
            }
        }
    };

    // 组件卸载时清理定时器
    onUnmounted(() => {
        stopPolling();
    });

    return {
        isPolling,
        startPolling,
        stopPolling,
    };
};
