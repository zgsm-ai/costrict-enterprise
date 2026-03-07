<template>
    <n-spin :show="isLoading">
        <div class="credits-page">
            <div class="credits-title">{{ t('creditsPage.title') }}</div>
            <div class="credits-usage-left">
                <div class="credits-usage-left__label">{{ t('creditsPage.remainingCredit') }}</div>
                <div class="credits-usage-left__num">{{ restQuota }}</div>
            </div>
            <n-data-table
                :columns="columns"
                :data="columnData"
                :bordered="false"
                :max-height="650"
                size="small"
            />
        </div>
    </n-spin>
</template>

<script setup lang="ts">
/**
 * @file credits-page.vue
 */
import { NDataTable, NPopover, NSpin } from 'naive-ui';
import { h, ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { getQuotaAuditRecords, getUserQuota } from '@/api/mods/quota.mod';
import type { GetUserQuotaRes, QuotaAuditRecord } from '@/api/bos/quota.bo';
import dayjs from 'dayjs';
import type { GroupedItem } from './interface';
import { OPERATION_TYPE, PAGE_PARAMS, POPOVER_SPAN_STYLE, RECHARGE_STRATEGY_NAME } from './const';
import { formatDate } from '@/utils/date';
import { storeToRefs } from 'pinia';
import { useUserStore } from '@/store/user';

const { t, locale } = useI18n();

const isZh = computed(() => locale.value === 'zh');

const popoverRender = (voucherCode: string) => {
    return h(
        NPopover,
        {
            trigger: 'hover',
            placement: 'top',
            style: 'max-width: 300px',
        },
        {
            trigger: () =>
                h(
                    'span',
                    {
                        style: POPOVER_SPAN_STYLE,
                    },
                    t('creditsPage.popoverTriggerText'),
                ),
            default: () => h('span', null, voucherCode),
        },
    );
};

const columns = computed(() => [
    {
        title: t('creditsPage.creditChange'),
        key: 'amount',
        render: (row: QuotaAuditRecord) => {
            const color = row.amount >= 0 ? '#30C461' : '#FF4A4A';
            return h('span', { style: { color } }, row.amount > 0 ? `+${row.amount}` : row.amount);
        },
        width: isZh.value ? 80 : 120,
    },
    {
        title: t('creditsPage.effectiveDate'),
        key: 'create_time',
        render: (row: QuotaAuditRecord) => formatDate(row.create_time),
        width: 180,
    },
    {
        title: t('creditsPage.description'),
        key: 'description',
        render: (row: QuotaAuditRecord) => {
            const { operation, related_user, voucher_code, amount, details, strategy_name } = row;

            switch (operation) {
                case OPERATION_TYPE.transferOut:
                    const transferOutBaseDescription = t('creditsPage.transferOutDesc', {
                        relatedUser: related_user,
                        amount: Math.abs(amount),
                    });

                    return h('div', { class: 'description-cell' }, [
                        h('span', null, transferOutBaseDescription),
                        voucher_code ? popoverRender(voucher_code) : null,
                    ]);
                case OPERATION_TYPE.transferIn:
                    const now = dayjs();
                    let validAmount = 0;
                    let expiredAmount = 0;
                    const validDetails: string[] = [];
                    const expiredDetails: string[] = [];

                    const groupedItems =
                        details?.items?.reduce<Record<string, GroupedItem>>((acc, detail) => {
                            const key = formatDate(detail.expiry_date, 'YYYY-MM-DD');
                            const absAmount = Math.abs(detail.amount);

                            if (!acc[key]) {
                                acc[key] = {
                                    amount: 0,
                                    expiry_date: detail.expiry_date,
                                    isExpired: now.isAfter(dayjs(detail.expiry_date)),
                                };
                            }
                            acc[key].amount += absAmount;
                            return acc;
                        }, {}) || {};

                    Object.values<GroupedItem>(groupedItems).forEach((item) => {
                        const absAmount = Math.abs(item.amount);
                        const expiryDate = formatDate(item.expiry_date, 'YYYY-MM-DD');

                        if (item.isExpired) {
                            expiredAmount += absAmount;
                            expiredDetails.push(
                                t('creditsPage.expiredDetailItem', {
                                    amount: absAmount,
                                    expiryDate: expiryDate,
                                }),
                            );
                        } else {
                            validAmount += absAmount;
                            validDetails.push(
                                t('creditsPage.validDetailItem', {
                                    amount: absAmount,
                                    expiryDate: expiryDate,
                                }),
                            );
                        }
                    });

                    const validText =
                        validDetails.length > 0
                            ? t('creditsPage.validCreditDetails', {
                                  validAmount: validAmount,
                                  validDetails: validDetails.join(t('common.comma')),
                              })
                            : '';

                    const expiredText =
                        expiredDetails.length > 0
                            ? t('creditsPage.expiredCreditDetails', {
                                  expiredAmount: expiredAmount,
                                  expiredDetails: expiredDetails.join(t('common.comma')),
                              })
                            : '';

                    const baseDescription =
                        t('creditsPage.transferInBaseDesc', {
                            relatedUser: related_user,
                            amount: amount,
                        }) +
                        (validText ? `${t('common.comma')}${validText}` : '') +
                        (expiredText ? `${t('common.comma')}${expiredText}` : '') +
                        t('creditsPage.voucherCodeSuffix');

                    return h('div', { class: 'description-cell' }, [
                        h('span', null, baseDescription),
                        voucher_code ? popoverRender(voucher_code) : null,
                    ]);
                case OPERATION_TYPE.reCharge:
                    if (Object.values(RECHARGE_STRATEGY_NAME).includes(strategy_name)) {
                        switch (strategy_name) {
                            case RECHARGE_STRATEGY_NAME.inviteeStarReward:
                                return t('creditsPage.inviteeStarReward');
                            case RECHARGE_STRATEGY_NAME.inviterRegisterReward:
                                return t('creditsPage.inviterRegisterReward', {
                                    relatedUser: related_user,
                                });
                            case RECHARGE_STRATEGY_NAME.inviterStarReward:
                                return t('creditsPage.inviterStarReward', {
                                    relatedUser: related_user,
                                });
                            default:
                                return strategy_name;
                        }
                    }
                    return strategy_name;
                case OPERATION_TYPE.mergeIn:
                    return t('creditsPage.mergeInDesc', {
                        relatedUser: related_user,
                    });
                default:
                    return strategy_name;
            }
        },
    },
    {
        title: t('creditsPage.expiryDate'),
        key: 'expiry_date',
        render: (row: QuotaAuditRecord) => formatDate(row.expiry_date),
        width: 180,
    },
]);

const columnData = ref<GetUserQuotaRes['quota_list']>([]);

const restQuota = ref(0);

const isLoading = ref(false);

const fetchQuotaAuditRecords = async () => {
    const { data } = await getQuotaAuditRecords(PAGE_PARAMS);

    columnData.value = data.records || [];
};

const fetchUserQuota = async () => {
    const { data } = await getUserQuota();

    if (!data) {
        return;
    }

    restQuota.value = Number((data.total_quota - data.used_quota).toFixed(2)) || 0;
};

const userStore = useUserStore();

const { isTokenInitialized } = storeToRefs(userStore);

watch(
    isTokenInitialized,
    (val) => {
        if (val) {
            isLoading.value = true;
            Promise.all([fetchQuotaAuditRecords(), fetchUserQuota()]).finally(() => {
                isLoading.value = false;
            });
        }
    },
    {
        immediate: true,
    },
);
</script>

<style lang="less" scoped>
.credits-page {
    width: 65%;
    margin: auto;
    padding-top: 24px;

    .credits-title {
        font-size: 20px;
        margin-left: 12px;
        font-weight: 600;
    }

    .credits-usage-left {
        margin: 24px 0 40px 12px;
        display: flex;
        align-items: center;
        font-weight: 600;
        font-size: 16px;

        &__num {
            background: linear-gradient(
                126deg,
                #0066ff 0%,
                #00ffb7 42%,
                #f7fffd 54%,
                #ffffff 61%,
                #005eff 101%
            );
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            text-fill-color: transparent;
        }
    }
}
</style>
