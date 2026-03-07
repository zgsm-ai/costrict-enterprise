<template>
    <div class="usage-consumption-section">
        <div class="consumption-table">
            <n-data-table
                :columns="columns"
                :data="tableData"
                :bordered="false"
                :max-height="300"
                size="small"
                :loading="loading"
            />
        </div>
        <div
            class="pagination-wrapper"
            v-if="totalCount && totalCount > 0"
        >
            <n-pagination
                :page="currentPage"
                :page-count="totalPages"
                :page-size="pageSize"
                show-size-picker
                :page-sizes="[10, 20, 50]"
                @update:page="handlePageChange"
                @update:page-size="handlePageSizeChange"
                simple
            />
        </div>
    </div>
</template>

<script setup lang="ts">
/**
 * @file usage-consumption-section.vue
 * @description 用量消耗统计组件 - 只负责展示，数据由父组件提供
 */
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { NDataTable, NPagination } from 'naive-ui';
import type { UsageConsumptionRecord } from '@/api/bos/quota.bo';
import { formatDate } from '@/utils/date';
import { withDefaultRender } from '../hook/useTableRender';

const { t } = useI18n();

// Props 定义 - 接收父组件传递的数据
const props = defineProps<{
    tableData: UsageConsumptionRecord[];
    loading?: boolean;
    currentPage?: number;
    pageSize?: number;
    totalCount?: number;
}>();

// 定义事件
const emit = defineEmits<{
    'update:page': [page: number];
    'update:pageSize': [pageSize: number];
}>();

// 计算总页数
const totalPages = computed(() => Math.ceil((props.totalCount || 0) / (props.pageSize || 10)));

// 处理页码变化
const handlePageChange = (page: number) => {
    emit('update:page', page);
};

// 处理每页条数变化
const handlePageSizeChange = (size: number) => {
    emit('update:pageSize', size);
};

const withDefaultColumnRender = withDefaultRender<UsageConsumptionRecord>();

// 表格列配置
const columns = computed(() =>
    withDefaultColumnRender([
        {
            title: t('homePage.startTime'),
            key: 'record_time',
            width: 120,
            render: (row: UsageConsumptionRecord) => formatDate(row.record_time),
        },
        {
            title: t('homePage.model'),
            key: 'model',
            width: 120,
        },
        {
            title: t('homePage.mode'),
            key: 'mode',
            width: 100,
        },
        {
            title: t('homePage.creditsUsed'),
            key: 'credits_used',
            width: 120,
        },
        {
            title: t('homePage.package'),
            key: 'package',
            width: 120,
        },
    ]),
);
</script>

<style scoped lang="less">
.usage-consumption-section {
    .consumption-table {
        :deep(.n-data-table) {
            background: transparent;

            .n-data-table-thead {
                background: rgba(255, 255, 255, 0.1);
            }

            .n-data-table-tbody {
                background: transparent;
            }

            .n-data-table-tr {
                background: transparent;
                color: #fff;

                &:hover {
                    background: rgba(255, 255, 255, 0.05);
                }
            }

            .n-data-table-td {
                color: #fff;
                opacity: 0.8;
            }

            .n-data-table-th {
                color: #fff;
                opacity: 0.9;
                font-weight: 500;
            }

            .n-base-loading__icon {
                color: #1876f2;
            }
        }
    }

    .pagination-wrapper {
        display: flex;
        justify-content: flex-end;
        margin-top: 16px;
    }
}
</style>
