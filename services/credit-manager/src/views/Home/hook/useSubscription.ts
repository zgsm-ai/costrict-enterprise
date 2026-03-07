/**
 * @description home页面订阅模块相关内容
 */
import { ref } from 'vue';
import { getOrders } from '@/api/mods/quota.mod';
import type { Order } from '@/api/bos/quota.bo';

export function useSubscription() {
    // 订单相关状态
    const ordersData = ref<Order[]>([]);
    const ordersTotal = ref<number>(0);
    const ordersPage = ref<number>(1);
    const ordersPageSize = ref<number>(10);
    const ordersLoading = ref<boolean>(false);

    // 获取订单数据
    const fetchOrders = async () => {
        ordersLoading.value = true;
        try {
            const { data } = await getOrders({
                page: ordersPage.value,
                page_size: ordersPageSize.value,
            });

            if (!data) {
                return;
            }

            ordersData.value = data.orders || [];
            ordersTotal.value = data.total || 0;
        } finally {
            ordersLoading.value = false;
        }
    };

    // 订单分页事件处理
    const handleOrdersPageChange = (page: number) => {
        ordersPage.value = page;
        fetchOrders();
    };

    const handleOrdersPageSizeChange = (pageSize: number) => {
        ordersPageSize.value = pageSize;
        ordersPage.value = 1; // 重置到第一页
        fetchOrders();
    };

    return {
        // 状态
        ordersData,
        ordersTotal,
        ordersPage,
        ordersPageSize,
        ordersLoading,

        // 方法
        fetchOrders,
        handleOrdersPageChange,
        handleOrdersPageSizeChange,
    };
}
