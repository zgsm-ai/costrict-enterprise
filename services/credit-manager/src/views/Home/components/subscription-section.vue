<template>
    <div class="subscription-section relative">
        <img
            class="absolute top-36 left-[-10px] w-55"
            src="../../../assets/subscription/bg_1.png"
            alt=""
        />
        <img
            class="absolute top-60 left-124 w-55"
            src="../../../assets/subscription/bg_2.png"
            alt=""
        />
        <img
            class="absolute top-37 left-155 w-28"
            src="../../../assets/subscription/bg_3.png"
            alt=""
        />
        <img
            class="absolute top-60 left-233 w-55"
            src="../../../assets/subscription/bg_4.png"
            alt=""
        />
        <img
            class="absolute top-37 left-264 w-28"
            src="../../../assets/subscription/bg_5.png"
            alt=""
        />
        <div v-if="false" class="subscription-survey">
            <div class="subscription-survey__copy flex items-center">
                <img
                    src="../../../assets/subscription/gift.webp"
                    alt=""
                />
                <div class="flex flex-col">
                    <div class="subscription-survey__title">
                        {{ t('subscriptionSection.survey.prefix') }}
                        <span class="subscription-survey__credits">{{
                            t('subscriptionSection.survey.credits')
                        }}</span>
                    </div>
                    <div class="subscription-survey__meta">
                        <span class="subscription-survey__text">
                            {{ t('subscriptionSection.survey.extraPrefix') }}
                            <span class="subscription-survey__highlight">{{
                                t('subscriptionSection.survey.extraAmount')
                            }}</span>
                            {{ t('subscriptionSection.survey.extraSuffix') }}
                        </span>
                        <span
                            class="subscription-survey__divider"
                            aria-hidden="true"
                        ></span>
                        <span class="subscription-survey__text">
                            {{ t('subscriptionSection.survey.deadlinePrefix') }}
                            <span class="subscription-survey__highlight">{{
                                t('subscriptionSection.survey.deadlineDate')
                            }}</span>
                        </span>
                    </div>
                </div>
            </div>
            <div class="subscription-survey__actions">
                <button
                    class="subscription-survey__button"
                    type="button"
                    @click="toSurvey"
                >
                    {{ t('subscriptionSection.survey.feedbackButton') }}
                </button>
                <button
                    class="subscription-survey__link"
                    type="button"
                    @click="toSurveyDetail"
                >
                    {{ t('subscriptionSection.survey.detailLink') }}
                    <span aria-hidden="true">›</span>
                </button>
            </div>
        </div>
        <!-- 订阅信息 -->
        <div class="subscription-title mt-4 flex items-center">
            <span class="text-base">{{ t('subscriptionSection.subscriptionTitle') }}</span>
            <span
                class="billing-btn cursor-pointer ml-2.5"
                @click="toBillingDocs"
                >{{ t('subscriptionSection.billingDescription') }}</span
            >
        </div>
        <section
            class="subscription-notice"
            aria-labelledby="subscription-notice-title"
        >
            <h2 id="subscription-notice-title">
                {{ t('subscriptionSection.notice.title') }}
            </h2>
            <p>{{ t('subscriptionSection.notice.greeting') }}</p>
            <p>{{ t('subscriptionSection.notice.thanks') }}</p>
            <I18nT
                keypath="subscriptionSection.notice.body"
                tag="p"
            >
                <template #cutoff>
                    <strong>{{ t('subscriptionSection.notice.cutoff') }}</strong>
                </template>
                <template #apiLink>
                    <a
                        href="https://docs.costrict.ai/plugin/guide/api-integration"
                        target="_blank"
                        rel="noopener noreferrer"
                    >{{ t('subscriptionSection.notice.apiLink') }}</a>
                </template>
                <template #enterpriseLink>
                    <a
                        href="https://costrict.ai/enterprise#enterprise-lead"
                        target="_blank"
                        rel="noopener noreferrer"
                    >{{ t('subscriptionSection.notice.enterpriseLink') }}</a>
                </template>
            </I18nT>
        </section>
        <p
            v-if="isCreditsOrderDebug"
            class="mt-4 text-amber-400"
            role="status"
        >
            {{ t('subscriptionSection.orderDebugNotice') }}
        </p>
        <div class="subscription-content mt-4">
            <div
                v-if="quotaTypesLoaded"
                class="content-version grid grid-cols-4 gap-5"
            >
                <div
                    v-for="(plan, index) in pricingPlans"
                    :key="index"
                    class="content-version__item min-88 px-5 py-6 relative"
                >
                    <div
                        v-if="plan.showTrafficLabel"
                        class="absolute right-0 top-[-4px]"
                    >
                        <img
                            src="../../../assets/label-bg.webp"
                            alt=""
                        />
                        <span class="absolute top-1 left-7 label-text">{{
                            t('subscriptionSection.trafficLabel')
                        }}</span>
                    </div>
                    <div class="content-version__item-title text-base">
                        {{ plan.title }}
                    </div>
                    <div
                        class="content-version__item-price flex items-center justify-between text-3xl mt-3"
                    >
                        <div class="price-summary flex items-center">
                            <span class="price-unit ml-[-8px]">￥</span>
                            <span class="price">{{ plan.price }}</span>
                            <span
                                v-if="plan.isFirstPurchase"
                                class="first-recharge-tip text-xs ml-2 mt-2 original-price__tips"
                                >{{ t('subscriptionSection.firstRechargeDiscount') }}</span
                            >
                            <span
                                v-if="plan.originalPrice"
                                class="original-price text-line-through text-base mt-2"
                                :class="plan.isFirstPurchase ? 'ml-1' : 'ml-2.5'"
                            >
                                ￥{{ plan.originalPrice }}
                            </span>
                        </div>
                        <span
                            v-if="plan.creditsTip"
                            class="content-version__item-credits text-base original-price__tips"
                        >
                            {{ plan.creditsTip }}
                        </span>
                    </div>
                    <div
                        class="content-version__item-btn h-10 text-center leading-10 mt-5 rounded-sm"
                        :class="{
                            'btn-purchase':
                                plan.buttonType === 'purchase' && !isCreditsServiceEnded,
                            'btn-download': plan.buttonType !== 'purchase',
                            'btn-disabled': plan.buttonType === 'purchase' && isCreditsServiceEnded,
                        }"
                        :aria-disabled="plan.buttonType === 'purchase' && isCreditsServiceEnded"
                        @click="handlePlanClick(plan)"
                    >
                        {{
                            plan.buttonType === 'purchase' && isCreditsServiceEnded
                                ? t('pricingPlans.purchaseUnavailable')
                                : plan.buttonText
                        }}
                    </div>
                    <ul class="content-version__item-features text-xs mt-4">
                        <li
                            v-for="(feature, featureIndex) in plan.features"
                            :key="featureIndex"
                            class="mb-4 last-of-type:mb-0"
                        >
                            <div class="flex items-start">
                                <img
                                    v-if="feature.available"
                                    src="../../../assets/icon/y.svg"
                                    alt="available"
                                    class="mt-0.5"
                                />
                                <img
                                    v-else
                                    src="../../../assets/icon/x.svg"
                                    alt="unavailable"
                                />
                                <p class="ml-2">{{ feature.text }}</p>
                            </div>
                        </li>
                    </ul>
                </div>
            </div>
        </div>
        <!-- 开通记录 -->
        <div class="subscription-records mt-9">
            <common-card :title="t('subscriptionSection.recordsTitle')">
                <n-data-table
                    :columns="recordColumns"
                    :data="recordsData"
                    :bordered="false"
                    :max-height="240"
                    size="small"
                    :loading="loading"
                    :pagination="paginationConfig"
                    remote
                    :scroll-x="scrollX"
                />
            </common-card>
        </div>

        <!-- 发票弹窗 -->
        <invoice-modal
            v-model:show="showInvoiceModal"
            :order-no="currentOrder?.order_id"
            :amount="currentOrder?.amount"
            @invoice-submitted="handleInvoiceSubmitted"
        />
    </div>
</template>

<script setup lang="ts">
/**
 * @file subscription-section.vue
 * @description 订阅管理组件 - 显示订阅管理的基本信息
 */
import { computed, h, ref, onMounted } from 'vue';
import { getInvoiceConstants, generatePricingPlansFromAPI } from '../const';
import { NDataTable } from 'naive-ui';
import { formatDate } from '@/utils/date';
import CommonCard from '@/components/common-card.vue';
import InvoiceModal from './invoice-modal.vue';
import type { Order, QuotaTypeWithMarketingRules } from '@/api/bos/quota.bo';
import { I18nT, useI18n } from 'vue-i18n';
import { formatAmount, withDefaultRender } from '../hook/useTableRender';
import { useRouter } from 'vue-router';
import newOrderIcon from '@/assets/price/new_order_icon.webp';
import { getQuotaTypes } from '@/api/mods/quota.mod';
import { useCreditsServiceCutoff } from '@/composables/useCreditsServiceCutoff';
import type { PricingPlan } from '../interface';

// 国际化
const { t } = useI18n();

// 控制发票弹窗显示
const showInvoiceModal = ref(false);
// 当前选中的订单信息
const currentOrder = ref<{ order_id: string; amount: number } | null>(null);

const router = useRouter();
const { isCreditsServiceEnded, isCreditsOrderDebug } = useCreditsServiceCutoff();

const SURVEY_URL = 'https://v.wjx.cn/vm/t7BdP0M.aspx';
const SURVEY_DETAIL_URL = 'https://mp.weixin.qq.com/s/kX8zt50Yu01a4NB6zwbn-g';

// 存储API返回的配额类型数据
const quotaTypes = ref<QuotaTypeWithMarketingRules[]>([]);
// 标记是否已加载API数据
const quotaTypesLoaded = ref(false);

// 计算国际化的价格计划
const pricingPlans = computed(() => {
    // 只有在请求完成后才展示所有套餐
    if (quotaTypesLoaded.value) {
        return generatePricingPlansFromAPI(t, router, quotaTypes.value);
    }
    // 否则返回空数组，不展示任何套餐
    return [];
});

// 计算国际化的发票常量
const invoiceConstants = computed(() => getInvoiceConstants(t));

// Props 定义
const props = withDefaults(
    defineProps<{
        recordsData: Order[];
        total?: number;
        page?: number;
        pageSize?: number;
        loading?: boolean;
    }>(),
    {
        recordsData: () => [],
        total: 0,
        page: 1,
        pageSize: 10,
        loading: false,
    },
);

// 分页相关状态
const currentPage = ref(props.page);
const pageSizeRef = ref(props.pageSize);

// 定义事件
const emit = defineEmits<{
    'update:page': [page: number];
    'update:pageSize': [pageSize: number];
    'refresh-orders': [];
}>();

// 分页配置
const paginationConfig = computed(() => ({
    page: currentPage.value,
    pageSize: pageSizeRef.value,
    pageCount: Math.ceil(props.total / pageSizeRef.value),
    showSizePicker: true,
    pageSizes: [10, 20, 50],
    prefix: () => t('subscriptionSection.totalItems', { total: props.total }),
    onChange: (page: number) => {
        currentPage.value = page;
        emit('update:page', page);
    },
    onUpdatePageSize: (pageSize: number) => {
        pageSizeRef.value = pageSize;
        currentPage.value = 1; // 重置到第一页
        emit('update:pageSize', pageSize);
    },
    simple: true,
}));

// 创建带有默认渲染函数的处理器
const withDefaultColumnRender = withDefaultRender<Order>();

const isSmallScreen = computed(() => window.innerWidth <= 1366);

const scrollX = computed(() => (isSmallScreen.value ? '1000' : '90%'));

// 开通记录表格列配置
const recordColumns = computed(() =>
    withDefaultColumnRender([
        {
            title: t('subscriptionSection.orderNumber'),
            key: 'order_id',
            render: (row: Order, index: number) => {
                // 判断是否是第一条数据且第一页
                const isFirstItemAndFirstPage = index === 0 && currentPage.value === 1;

                return h('div', { class: 'flex items-end' }, [
                    h('span', { class: 'text-nowrap' }, row.order_id),
                    isFirstItemAndFirstPage
                        ? h('img', {
                              src: newOrderIcon,
                              class: 'mb-0.5',
                              alt: 'new order',
                          })
                        : null,
                ]);
            },
            width: 220,
        },
        {
            title: t('subscriptionSection.orderType'),
            key: 'quota_type',
        },
        {
            title: t('subscriptionSection.purchaseCount'),
            key: 'credit_count',
        },
        {
            title: t('subscriptionSection.orderAmount'),
            key: 'amount',
            render: (row: Order) => formatAmount(row.amount, '￥'),
        },
        {
            title: t('subscriptionSection.orderTime'),
            key: 'created_at',
            render: (row: Order) => {
                const value = formatDate(row.created_at);
                return value || '-';
            },
        },
        {
            title: t('subscriptionSection.validityPeriod'),
            key: 'credit_expire_date',
            render: (row: Order) => {
                const value = formatDate(row.credit_expire_date);
                return value || '-';
            },
        },
        {
            title: t('subscriptionSection.invoiceStatus'),
            key: 'invoice_status',
            render: (row: Order) => {
                const statusText = invoiceConstants.value.STATUS_TEXT[row.invoice_status] || '-';
                const statusColor =
                    invoiceConstants.value.STATUS_COLORS[row.invoice_status] || '#FF4D4F';

                return h(
                    'span',
                    {
                        style: {
                            color: statusColor,
                            fontWeight: '500',
                        },
                    },
                    statusText,
                );
            },
        },
        {
            title: t('subscriptionComments.operation'),
            key: 'operation',
            render: (row: Order) => {
                // 只有待开票状态才能进行开票操作
                const canInvoice = row.invoice_status === invoiceConstants.value.STATUS.PENDING;

                return h(
                    'div',
                    {
                        class: canInvoice ? 'cursor-pointer text-[#4D99FF]' : 'text-[#999999]',
                        onClick: canInvoice
                            ? () => {
                                  currentOrder.value = {
                                      order_id: row.order_id,
                                      amount: row.amount,
                                  };
                                  showInvoiceModal.value = true;
                              }
                            : undefined,
                    },
                    t('subscriptionComments.goToInvoice'),
                );
            },
        },
    ]),
);

const toBillingDocs = () => {
    window.open('https://docs.costrict.ai/billing/purchase');
};

const handlePlanClick = (plan: PricingPlan) => {
    if (plan.buttonType === 'purchase' && isCreditsServiceEnded.value) {
        return;
    }

    plan.clickEvent();
};

const toSurvey = () => {
    window.open(SURVEY_URL, '_blank', 'noopener');
};

const toSurveyDetail = () => {
    window.open(SURVEY_DETAIL_URL, '_blank', 'noopener');
};

// 加载配额类型数据
const loadQuotaTypes = async () => {
    const { code, data } = await getQuotaTypes();
    if (code !== 200) {
        return;
    }
    quotaTypes.value = data || [];
    quotaTypesLoaded.value = true;
};

// 组件挂载时加载配额类型数据
onMounted(() => {
    loadQuotaTypes();
});

// 处理发票提交成功事件
const handleInvoiceSubmitted = () => {
    // 通知父组件刷新订单数据
    emit('refresh-orders');
};
</script>

<style scoped lang="less">
.subscription-section {
    color: white;
    .subscription-notice {
        position: relative;
        margin-top: 24px;
        padding: 24px;
        border: 1px solid #263445;
        border-radius: 8px;
        background: #0b111a;
        color: #b8c4d2;
        font-size: 14px;
        line-height: 1.9;
        overflow-wrap: anywhere;

        h2 {
            margin: 0 0 20px;
            color: #f0f4f8;
            font-size: 20px;
            font-weight: 600;
            text-align: center;
        }

        p {
            margin: 0 0 12px;

            &:last-child {
                margin-bottom: 0;
            }
        }

        strong {
            color: #f0f4f8;
        }

        a {
            color: #73c0ff;
            text-decoration: underline;
            text-underline-offset: 3px;

            &:hover {
                color: #a8d9ff;
            }

            &:focus-visible {
                outline: 2px solid currentColor;
                outline-offset: 3px;
            }
        }

        @media (max-width: 768px) {
            padding: 16px;
        }
    }

    .subscription-title {
        .billing-btn {
            background: linear-gradient(95deg, #2a7fff 5%, #ffffff 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            text-fill-color: transparent;
        }
    }

    .subscription-survey {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 16px 24px;
        padding: 16px 18px;
        border-radius: 4px;
        background: radial-gradient(
                36.88% 213.56% at 34.25% -0.79%,
                rgba(28, 149, 255, 0.46) 0%,
                rgba(28, 149, 255, 0.18) 42%,
                rgba(28, 149, 255, 0) 100%
            )
            /* warning: gradient uses a rotation that is not supported by CSS and may not behave as expected */;
        color: rgba(255, 255, 255, 0.7);

        @media (max-width: 960px) {
            align-items: flex-start;
            flex-direction: column;
            justify-content: center;
            padding: 0 14px;
        }

        &__copy {
            min-width: 0;
        }

        &__title {
            color: rgb(255, 255, 255);
            font-size: 18px;
            font-weight: 600;

            @media (max-width: 960px) {
                font-size: 16px;
            }
        }

        &__credits {
            color: rgb(255, 255, 255);
        }

        &__highlight {
            color: rgba(34, 255, 208, 1);
        }

        &__meta,
        &__actions {
            display: flex;
            align-items: center;
            flex-wrap: wrap;
            gap: 10px 24px;
        }

        &__meta {
            margin-top: 8px;

            @media (max-width: 960px) {
                gap: 8px 12px;
                margin-top: 6px;
            }
        }

        &__actions {
            flex-shrink: 0;

            @media (max-width: 960px) {
                gap: 8px 12px;
            }
        }

        &__text {
            min-width: 0;
            font-size: 14px;
            color: rgba(255, 255, 255, 0.65);

            @media (max-width: 960px) {
                font-size: 13px;
            }
        }

        &__divider {
            width: 1px;
            height: 19px;
            background: rgba(255, 255, 255, 0.45);

            @media (max-width: 960px) {
                display: none;
            }
        }

        &__button {
            min-width: 126px;
            height: 36px;
            cursor: pointer;
            border: 0;
            border-radius: 2px;
            background: linear-gradient(90deg, #2a7fff 0%, #00d8b7 100%);
            color: #fff;
            font-size: 14px;
            font-weight: 600;

            @media (max-width: 960px) {
                min-width: 120px;
            }
        }

        &__link {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            cursor: pointer;
            border: 0;
            background: transparent;
            color: #2a7fff;
            font-size: 14px;
            font-weight: 600;
            white-space: nowrap;
        }
    }

    .content-version {
        @media (max-width: 1200px) {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }

        @media (max-width: 768px) {
            grid-template-columns: repeat(1, minmax(0, 1fr));
        }

        &__item {
            position: relative;
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            -webkit-backdrop-filter: blur(10px);
            border-radius: 10px;
            z-index: 0;
            padding: 24px 20px;

            &-price {
                min-height: 64px;
                column-gap: 8px;
                row-gap: 4px;
                flex-wrap: wrap;
                align-content: center;

                .price-summary {
                    flex: 0 0 auto;
                    white-space: nowrap;
                }

                .first-recharge-tip,
                .content-version__item-credits {
                    white-space: nowrap;
                }

                .content-version__item-credits {
                    flex: 0 0 auto;
                    margin-left: auto;
                    text-align: right;
                }
            }

            &-btn {
                background: rgba(255, 255, 255, 0.2);
                cursor: pointer;
                transition: all 0.3s ease;

                &.btn-download {
                    border: 1px solid;
                    border-image: linear-gradient(
                            107deg,
                            #0066ff 38%,
                            #00ffb7 52%,
                            rgba(247, 255, 253, 0.51) 88%,
                            rgba(0, 94, 255, 0.09) 100%
                        )
                        1;
                    background: rgba(255, 255, 255, 0.2);

                    &:hover {
                        box-shadow: 0 0 10px rgba(0, 102, 255, 0.3);
                    }
                }

                &.btn-purchase {
                    background: linear-gradient(91deg, #005eff -9%, #00ffb7 104%);
                }

                &.btn-disabled {
                    cursor: not-allowed;
                    color: rgba(255, 255, 255, 0.55);
                    background: rgba(255, 255, 255, 0.12);
                }
            }

            .original-price {
                text-decoration: line-through;
                color: rgba(255, 255, 255, 0.7);

                &__tips {
                    background: linear-gradient(91deg, #00ffb7, #fff 31%, #c5dbff 150%);
                    -webkit-background-clip: text;
                    -webkit-text-fill-color: transparent;
                }
            }

            .label-text {
                background: linear-gradient(185deg, #005eff 35%, #00ffb7 93%);
                -webkit-background-clip: text;
                -webkit-text-fill-color: transparent;
                background-clip: text;
            }

        }

        &__item::before {
            content: '';
            position: absolute;
            inset: 0;
            padding: 1px;
            border-radius: 10px;
            background: linear-gradient(
                176deg,
                #0066ff -7%,
                #00ffb7 16%,
                rgba(247, 255, 253, 0.51) 51%,
                rgba(0, 94, 255, 0.09) 99%
            );
            -webkit-mask:
                linear-gradient(#fff 0 0) content-box,
                linear-gradient(#fff 0 0);
            -webkit-mask-composite: xor;
            z-index: -1;
        }
    }
}

.subscription-operation {
    .operation-content {
        .timeline-icon {
            display: flex;
            align-items: center;
            justify-content: center;
            width: 20px;
            height: 20px;
            border-radius: 10px;
            font-size: 12px;
            color: #fff;
            background: #4083e8;
        }

        .step-wrapper {
            .step-content-area {
                width: 100%;
            }

            .step-image {
                width: 100%;
                height: auto;
            }

            .image-text-pairs {
                display: flex;
                flex-direction: column;
                gap: 16px;
                margin-top: 16px;
            }
        }

        :deep(.n-timeline-item) {
            .n-timeline-item-content__title {
                color: #f4f8ff;
                font-size: 12px;
                margin-top: 2px;
                line-height: 20px;
            }

            .n-timeline-item-timeline__line {
                --n-color-start: #4083e8;
            }

            .n-timeline-item-timeline__circle {
                background: #4083e8;
            }
        }
    }
}

.subscription-records {
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
</style>
