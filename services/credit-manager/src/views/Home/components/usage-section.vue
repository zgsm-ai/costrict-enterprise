<template>
    <div class="usage-section">
        <div class="usage-content">
            <div class="content-title">
                {{ t('homePage.userCredits') }}
            </div>
            <div class="content-desc">
                {{ t('homePage.creditDesc') }}
            </div>
            <div class="content-progress">
                <n-progress
                    type="line"
                    :percentage="percentage"
                    :show-indicator="false"
                    :color="{
                        stops: [
                            'rgba(0, 102, 255, 0.7)',
                            'rgba(0, 255, 183, 0.7)',
                            'rgba(247, 255, 253, 0.7)',
                            'rgba(0, 94, 255, 0.7)',
                        ],
                    }"
                    :offset="[0, 47, 65, 99]"
                    :height="6"
                    rail-color="rgba(255, 255, 255, 0.2)"
                />
            </div>
            <div class="usage-static">
                <div class="static-use">
                    {{
                        t('homePage.useCredit', {
                            used: formatNumber(usedQuota),
                            total: formatNumber(totalQuota),
                        })
                    }}
                </div>
                <div class="static-total">
                    {{
                        t('homePage.restCredit', {
                            rest: formatNumber(remainingQuota),
                        })
                    }}
                </div>
            </div>
            <div class="credit-validity">
                <n-collapse
                    arrow-placement="right"
                    :default-expanded-names="['1']"
                >
                    <n-collapse-item
                        :title="t('homePage.creditValidity')"
                        name="1"
                    >
                        <n-data-table
                            :columns="columns"
                            :data="columnsData"
                            :bordered="false"
                            :max-height="120"
                            size="small"
                        />
                    </n-collapse-item>
                </n-collapse>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
/**
 * @file usage-section.vue
 * @description 用量统计组件 - 显示用户积分使用情况和有效期信息
 */
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { NProgress, NCollapse, NCollapseItem, NDataTable } from 'naive-ui';
import type { QuotaList } from '@/api/bos/quota.bo';
import { formatDate } from '@/utils/date';
import { withDefaultRender } from '../hook/useTableRender';

const { t } = useI18n();

// Props 定义
const props = withDefaults(
    defineProps<{
        usedQuota: number;
        totalQuota: number;
        columnsData: QuotaList[];
        isStar?: string;
    }>(),
    {
        usedQuota: 0,
        totalQuota: 0,
        columnsData: () => [],
        isStar: undefined,
    },
);

// Emits 定义
defineEmits<{
    'transfer-credit': [];
    'to-credits': [];
}>();

// 计算属性
const percentage = computed(() => {
    if (props.totalQuota === 0) return 0;
    return Number(((props.usedQuota / props.totalQuota) * 100).toFixed(0));
});

const remainingQuota = computed(() => {
    return props.totalQuota - props.usedQuota;
});

const withDefaultColumnRender = withDefaultRender<QuotaList>();
// 表格列配置
const columns = computed(() =>
    withDefaultColumnRender([
        {
            title: t('homePage.expiryDate'),
            key: 'expiry_date',
            render: (row: QuotaList) => formatDate(row.expiry_date),
            width: 120,
        },
        {
            title: t('homePage.creditNum'),
            key: 'amount',
        },
        {
            title: t('homePage.source'),
            key: 'source',
        },
    ]),
);

// 方法
const formatNumber = (num: number): string => {
    return num.toFixed(2);
};
</script>

<style scoped lang="less">
.usage-section {
    .usage-content {
        .content-title {
            font-weight: 600;
            margin-bottom: 12px;
            color: #fff;
        }

        .content-desc {
            font-size: 14px;
            opacity: 0.7;
            margin-bottom: 20px;
            color: #fff;
        }

        .content-progress {
            margin-bottom: 16px;
        }

        .usage-static {
            display: flex;
            justify-content: space-between;
            margin-bottom: 20px;

            .static-use {
                font-size: 14px;
                color: #fff;
                opacity: 0.8;
            }

            .static-total {
                font-size: 14px;
                color: #fff;
                opacity: 0.8;
            }
        }

        .credit-validity {
            margin-bottom: 20px;

            :deep(.n-collapse-item-header) {
                color: #fff;
                opacity: 0.9;
            }

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

                // 覆盖表格内部的 Spin 样式
                .n-spin {
                    .n-spin-icon {
                        color: #1876f2;
                    }
                }
            }
        }
    }
}
</style>
