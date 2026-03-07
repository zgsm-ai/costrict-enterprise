<template>
    <common-modal
        :title="t('creditTransferModal.title')"
        :show="show"
        @update:show="(v: boolean) => emit('update:show', v)"
        :show-spin="isLoading"
    >
        <div class="modal-tab">
            <ul class="tab-list">
                <li
                    :class="{ 'tab-item': true, 'active-tab': tab.value === activeTab }"
                    v-for="tab in tabs"
                    :key="tab.value"
                    @click="activeTab = tab.value"
                >
                    {{ tab.label }}
                </li>
            </ul>
        </div>
        <n-form
            v-if="isCreditOut"
            ref="formRef"
            :model="formModel"
            :rules="rules"
        >
            <n-form-item
                show-require-mark
                require-mark-placement="left"
                path="receiverId"
            >
                <template #label>
                    <div class="user-label">
                        <span>{{ t('creditTransferModal.receiverIdLabel') }}</span>
                        <p>{{ t('creditTransferModal.receiverIdTip') }}</p>
                    </div>
                </template>
                <n-input
                    v-model:value="formModel.receiverId"
                    clearable
                    :placeholder="t('creditTransferModal.receiverIdPlaceholder')"
                />
            </n-form-item>
            <n-form-item
                show-require-mark
                require-mark-placement="left"
                path="checkedRowKeys"
            >
                <template #label>
                    <div class="transfer-label">
                        <span>{{ t('creditTransferModal.transferCreditLabel') }}</span>
                        <p
                            v-if="!starred"
                            class="transfer-label__text"
                        >
                            {{ t('creditTransferModal.onlySupportCreditTransfer') }}
                            <a
                                href="https://github.com/zgsm-ai/costrict"
                                target="_blank"
                                >git star一下</a
                            >，
                            {{ t('creditTransferModal.effectiveIn') }}
                        </p>
                    </div>
                </template>
                <n-data-table
                    :row-key="rowKey"
                    v-model:checked-row-keys="formModel.checkedRowKeys"
                    :columns="columns"
                    :data="quotaData"
                    :bordered="false"
                    size="small"
                />
            </n-form-item>
        </n-form>
        <template v-else>
            <n-form
                ref="formRef"
                :model="formModel"
            >
                <n-form-item
                    :validation-status="inputValidationStatus"
                    :feedback="inputFeedback"
                >
                    <template #label>
                        <div class="redeem-code-label">
                            <span>{{ t('creditTransferModal.redeemCodeLabel') }}</span>
                            <p v-if="transferSuccess">
                                {{ t('creditTransferModal.redeemSuccessStatus') }}
                            </p>
                        </div>
                    </template>
                    <n-input
                        v-model:value="formModel.redeemCode"
                        clearable
                        :placeholder="t('creditTransferModal.redeemCodePlaceholder')"
                    />
                    <n-button
                        type="info"
                        style="margin-left: 10px"
                        :disabled="!formModel.redeemCode"
                        @click="confirmTransferIn"
                    >
                        {{ t('creditTransferModal.confirmButton') }}
                    </n-button>
                </n-form-item>
            </n-form>
            <template v-if="transferSuccess">
                <div class="credit-receive-detail">
                    <span>{{ t('creditTransferModal.creditReceiveDetailTitle') }}</span>
                    <div>
                        {{ t('creditTransferModal.viewReceiveRecordLeft') }}
                        <span
                            class="detail-tag"
                            @click="toCredit"
                            >Credit</span
                        >
                        {{ t('creditTransferModal.viewReceiveRecordRight') }}
                    </div>
                </div>
                <n-data-table
                    :row-key="rowKey"
                    v-model:checked-row-keys="formModel.checkedRowKeys"
                    :columns="transferInQuotaColumns"
                    :data="transferInQuota"
                    :bordered="false"
                    size="small"
                    style="margin-bottom: 10px"
                    :row-class-name="rowClassName"
                />
            </template>
        </template>
        <template v-if="isCreditOut">
            <div class="btn-group">
                <n-button
                    type="info"
                    @click="handleSubmit"
                    >{{ t('creditTransferModal.confirmButton') }}</n-button
                >
                <n-button @click="emit('update:show', false)">{{
                    t('creditTransferModal.cancelButton')
                }}</n-button>
            </div>
        </template>
    </common-modal>
</template>

<script setup lang="ts">
/**
 * @file credit-transfer-modal.vue
 */
import CommonModal from '@/components/common-modal.vue';
import { computed, h, watch, ref } from 'vue';
import {
    NButton,
    NForm,
    NFormItem,
    NInput,
    NDataTable,
    NInputNumber,
    useMessage,
    NTag,
    type FormItemRule,
} from 'naive-ui';
import type { QuotaList, QuotaTransferInQuotaList } from '@/api/bos/quota.bo';
import dayjs from 'dayjs';
import { postQuotaIn, postQuotaOut } from '@/api/mods/quota.mod';
import { useI18n } from 'vue-i18n';
import type { FormModel, RowData } from './interface';
import { STATUS, TRANSFER_IN_STATUS, TRANSFER_TYPE } from './const';

const message = useMessage();

const { t } = useI18n();

const props = withDefaults(
    defineProps<{
        show: boolean;
        userQuotaData: QuotaList[];
        isStar?: string;
    }>(),
    {
        show: false,
        userQuotaData: () => [],
    },
);

const emit = defineEmits<{
    (e: 'update:show', v: boolean): void;
    (e: 'update:submit', v: string): void;
    (e: 'update:transferIn'): void;
}>();

const quotaData = ref<RowData[]>([]);

const inputValidationStatus = computed(() =>
    tranferStatus.value === STATUS.error ? STATUS.error : STATUS.undefined,
);

const inputFeedback = ref('');

const starred = computed(() => props.isStar === 'true' || props.isStar === undefined);

watch(
    () => props.userQuotaData,
    (val) => {
        if (!val.length) {
            return;
        }

        quotaData.value = val.map((item, index) => {
            return {
                ...item,
                id: index,
                transferAmount: item.amount,
            };
        });

        formModel.value.checkedRowKeys = quotaData.value.map((item) => item.id);
    },
    {
        immediate: true,
    },
);

const formRef = ref();

const formModel = ref<FormModel>({
    receiverId: '',
    redeemCode: '',
    checkedRowKeys: [],
});

const resetForm = () => {
    if (!formRef.value) {
        return;
    }

    formRef.value?.restoreValidation();
    formModel.value = {
        receiverId: '',
        redeemCode: '',
        checkedRowKeys: quotaData.value.map((item) => item.id),
    };

    quotaData.value = props.userQuotaData.map((item, index) => {
        return {
            ...item,
            id: index,
            transferAmount: item.amount,
        };
    });

    tranferStatus.value = STATUS.undefined;
    inputFeedback.value = '';
};
const activeTab = ref<keyof typeof TRANSFER_TYPE>(TRANSFER_TYPE.out);

watch(activeTab, () => {
    resetForm();
});

watch(
    () => props.show,
    () => {
        resetForm();
        activeTab.value = TRANSFER_TYPE.out;
    },
);

const tabs = computed(() => [
    {
        label: t('creditTransferModal.tabCreditOut'),
        value: TRANSFER_TYPE.out,
    },
    {
        label: t('creditTransferModal.tabCreditIn'),
        value: TRANSFER_TYPE.in,
    },
]);

const columns = computed(() => [
    {
        type: 'selection' as const,
        key: 'selection',
    },
    {
        title: t('creditTransferModal.expiryDateColumn'),
        key: 'expiry_date',
        render: (row: QuotaList) => dayjs(row.expiry_date).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
        title: t('creditTransferModal.totalCreditColumn'),
        key: 'amount',
    },
    {
        title: t('creditTransferModal.transferOutCreditColumn'),
        key: 'transferAmount',
        render: (row: RowData) => {
            return h(NInputNumber, {
                value: row.transferAmount,
                onUpdateValue: (val: number | null) => {
                    if (val !== null) {
                        row.transferAmount = val;
                    }
                },
                min: 1,
                max: row.amount,
                showButton: false,
                placeholder: '',
                style: 'width: 60px',
                status: 'success',
                size: 'small',
                precision: 0,
            });
        },
    },
]);

const transferInQuotaColumns = computed(() => [
    {
        title: t('creditTransferModal.expiryDateColumn'),
        key: 'expiry_date',
        render: (row: QuotaTransferInQuotaList) => {
            return h(
                'div',
                {
                    style: row.is_expired ? 'opacity: 0.5' : '',
                },
                dayjs(row.expiry_date).format('YYYY-MM-DD HH:mm:ss'),
            );
        },
    },
    {
        title: t('creditTransferModal.totalCreditColumn'),
        key: 'amount',
        render: (row: QuotaTransferInQuotaList) => {
            return h('div', { style: 'display: flex; align-items: center;' }, [
                h(
                    'span',
                    {
                        style: row.is_expired ? 'opacity: 0.5' : '',
                    },
                    row.amount,
                ),
                row.is_expired
                    ? h(
                          NTag,
                          {
                              type: 'warning',
                              size: 'small',
                              style: 'margin-left: 12px; border-radius: 2px;',
                          },
                          {
                              default: () => t('creditTransferModal.expiredTag'),
                          },
                      )
                    : '',
            ]);
        },
    },
]);

const rowKey = (row: RowData) => row.id;

const transferInQuota = ref<QuotaTransferInQuotaList[]>([]);

const rowClassName = (row: QuotaTransferInQuotaList) => {
    return row.is_expired ? 'is-expired' : '';
};

const uuidValidatorWithMessage = (_: FormItemRule, value: string) => {
    if (!value) {
        return new Error(t('creditTransferModal.validationReceiverIdRequired'));
    }

    const uuidRegex =
        /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/;

    if (!uuidRegex.test(value.trim())) {
        return new Error(t('creditTransferModal.invalidUUID'));
    }

    return true;
};

const rules = {
    receiverId: [
        {
            validator: uuidValidatorWithMessage,
            trigger: ['input', 'blur'],
        },
    ],
    checkedRowKeys: [
        {
            validator: () => formModel.value.checkedRowKeys.length > 0,
            message: t('creditTransferModal.validationSelectCreditRequired'),
            trigger: 'change',
        },
    ],
};

const isLoading = ref(false);

const handleSubmit = () => {
    formRef.value?.validate(async (errors: Array<{ message: string }> | null) => {
        if (!errors) {
            const quota_list = quotaData.value
                .filter((item) => {
                    return formModel.value.checkedRowKeys.includes(item.id);
                })
                .map((filterItem) => {
                    return {
                        amount: filterItem.transferAmount,
                        expiry_date: filterItem.expiry_date,
                    };
                });

            const params = {
                receiver_id: formModel.value.receiverId.trim(),
                quota_list: quota_list,
            };

            isLoading.value = true;

            const { data } = await postQuotaOut(params).finally(() => {
                isLoading.value = false;
            });

            if (!data.voucher_code) {
                message.error(t('creditTransferModal.generateVoucherCodeFailed'));

                return;
            }

            if (data.voucher_code) {
                message.success(t('creditTransferModal.generateVoucherCodeSuccess'));
                emit('update:submit', data.voucher_code);
            }
        }
    });
};

const confirmTransferIn = async () => {
    isLoading.value = true;

    const { data } = await postQuotaIn({
        voucher_code: formModel.value.redeemCode.trim(),
    }).finally(() => {
        isLoading.value = false;
    });

    // success
    if (
        data.status === TRANSFER_IN_STATUS.success ||
        data.status === TRANSFER_IN_STATUS.partialSuccess
    ) {
        tranferStatus.value = STATUS.success;

        transferInQuota.value = data.quota_list;

        message.success(t('creditTransferModal.redeemSuccessMessage'));

        emit('update:transferIn');
    } else {
        tranferStatus.value = STATUS.error;

        if (data.status === TRANSFER_IN_STATUS.alreadyRedeemed) {
            inputFeedback.value = t('creditTransferModal.redeemCodeAlreadyRedeemed');
        } else {
            inputFeedback.value = data.voucher_code
                ? t('creditTransferModal.redeemCodeExpiredReminder')
                : t('creditTransferModal.redeemCodeError');
        }
    }
};

const isCreditOut = computed(() => activeTab.value === TRANSFER_TYPE.out);

const tranferStatus = ref<undefined | string>(STATUS.undefined);

const transferSuccess = computed(() => tranferStatus.value === STATUS.success);

const toCredit = () => window.open('/credit/manager/credits');
</script>

<style lang="less" scoped>
.btn-group {
    display: flex;
    justify-content: end;

    & > button:first-of-type {
        margin-right: 8px;
    }
}

.modal-tab {
    display: flex;
    justify-content: center;
    margin: 10px 0 24px;

    .tab-list {
        display: flex;
        padding: 0;

        .tab-item {
            width: 80px;
            height: 28px;
            line-height: 28px;
            text-align: center;
            color: #ffffff;
            border-radius: 2px;
            background: rgba(255, 255, 255, 0.2);
            cursor: pointer;
            font-size: 12px;
        }

        .active-tab {
            color: #1876f2;
            background: #f5f9ff;
            font-weight: 600;
        }
    }
}

.user-label {
    display: flex;

    p {
        color: rgba(255, 255, 255, 0.7);
    }
}

.transfer-label {
    &__text {
        margin-top: 4px;
        line-height: 18px;
        color: #fed400;

        a {
            text-decoration: none;
            color: #1876f2;
            font-weight: bold;
        }
    }
}

.redeem-code-label {
    display: flex;
    align-items: center;
    font-size: 12px;

    p {
        letter-spacing: normal;
        background: linear-gradient(
            103deg,
            #0066ff 1%,
            #00ffb7 28%,
            #f7fffd 55%,
            #ffffff 70%,
            #005eff 101%
        );
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
        text-fill-color: transparent;
    }
}

.credit-receive-detail {
    display: flex;
    align-items: center;
    font-size: 12px;
    color: #fff;
    margin-bottom: 8px;

    .detail-tag {
        color: #1770e6;
        cursor: pointer;
    }
}
</style>
