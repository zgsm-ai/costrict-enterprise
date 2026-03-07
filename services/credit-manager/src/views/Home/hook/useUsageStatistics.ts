/**
 * @description home页面用量统计模块
 */
import { ref } from 'vue';
import { getUsageStatistics } from '@/api/mods/quota.mod';
import type { UsageConsumptionRecord, GetUsageStatisticsReq } from '@/api/bos/quota.bo';

export function useUsageStatistics() {
    // 用量消耗相关状态
    const usageConsumptionData = ref<UsageConsumptionRecord[]>([]);
    const usageConsumptionLoading = ref<boolean>(false);
    const usageConsumptionPage = ref<number>(1);
    const usageConsumptionPageSize = ref<number>(10);
    const usageConsumptionTotal = ref<number>(0);

    // 时间选择相关状态
    const selectedTimeRange = ref<string>('today');
    const customDateRange = ref<[string, string] | null>(null);
    const customDateRangeFormatted = ref<[string, string] | null>(null);

    // 获取用量消耗数据
    const fetchUsageConsumptionData = async () => {
        usageConsumptionLoading.value = true;
        try {
            const params: GetUsageStatisticsReq = {
                page: usageConsumptionPage.value,
                page_size: usageConsumptionPageSize.value,
            };

            // 根据时间范围选择参数
            if (selectedTimeRange.value && selectedTimeRange.value !== 'custom') {
                params.time_range = selectedTimeRange.value;
            } else if (selectedTimeRange.value === 'custom' && customDateRangeFormatted.value) {
                params.start_time = `${customDateRangeFormatted.value[0]} 00:00:00`;
                params.end_time = `${customDateRangeFormatted.value[1]} 00:00:00`;
            }

            const { data } = await getUsageStatistics(params);

            if (data) {
                usageConsumptionData.value = data.records || [];
                usageConsumptionTotal.value = data.total || 0;

                // 如果当前页没有数据且不是第一页，回到第一页
                if (usageConsumptionData.value.length === 0 && usageConsumptionPage.value > 1) {
                    usageConsumptionPage.value = 1;
                }
            }
        } catch (error) {
            console.error('获取用量消耗统计数据失败:', error);
            usageConsumptionData.value = [];
            usageConsumptionTotal.value = 0;
        } finally {
            usageConsumptionLoading.value = false;
        }
    };

    // 用量消耗统计分页事件处理
    const handleUsageConsumptionPageChange = (page: number) => {
        usageConsumptionPage.value = page;
        fetchUsageConsumptionData();
    };

    const handleUsageConsumptionPageSizeChange = (pageSize: number) => {
        usageConsumptionPageSize.value = pageSize;
        usageConsumptionPage.value = 1; // 重置到第一页
        fetchUsageConsumptionData();
    };

    // 时间选择相关方法
    const handleTimeRangeSelect = (range: string) => {
        selectedTimeRange.value = range;
        customDateRange.value = null;
        usageConsumptionPage.value = 1; // 重置到第一页
        fetchUsageConsumptionData();
    };

    const handleCustomDateRangeChange = (value: [string, string] | null) => {
        if (value && value.length === 2) {
            selectedTimeRange.value = 'custom';
            customDateRangeFormatted.value = value;
            usageConsumptionPage.value = 1; // 重置到第一页

            // 触发数据获取
            fetchUsageConsumptionData();
        }
    };

    return {
        // 状态
        usageConsumptionData,
        usageConsumptionLoading,
        usageConsumptionPage,
        usageConsumptionPageSize,
        usageConsumptionTotal,
        selectedTimeRange,
        customDateRange,
        customDateRangeFormatted,

        // 方法
        fetchUsageConsumptionData,
        handleUsageConsumptionPageChange,
        handleUsageConsumptionPageSizeChange,
        handleTimeRangeSelect,
        handleCustomDateRangeChange,
    };
}
