<template>
    <n-modal
        v-model:show="visible"
        :mask-closable="false"
        class="invoice-modal"
        :on-esc="closeModal"
    >
        <div class="modal-container flex flex-col">
            <!-- 弹窗头部 -->
            <div class="mb-6 flex items-center justify-between">
                <span class="text-white">{{ t('invoiceModal.title') }}</span>
                <n-button
                    quaternary
                    @click="closeModal"
                >
                    <template #icon>
                        <n-icon>
                            <CloseOutline />
                        </n-icon>
                    </template>
                </n-button>
            </div>

            <!-- 弹窗内容 -->
            <div class="modal-content">
                <div class="mb-4 text-xs tip-text">
                    {{ t('invoiceModal.tip') }}
                </div>

                <n-form
                    :model="formData"
                    :rules="formRules"
                    ref="formRef"
                    label-placement="left"
                    label-width="auto"
                    class="invoice-form"
                    require-mark-placement="left"
                >
                    <!-- 抬头类型 -->
                    <n-form-item
                        path="headerType"
                        :label="t('invoiceModal.headerType.label')"
                        class="form-item"
                        :show-feedback="false"
                    >
                        <n-radio-group
                            v-model:value="formData.headerType"
                            class="radio-group"
                            size="small"
                        >
                            <n-radio
                                :value="invoiceConstants.HEADER_TYPE.COMPANY"
                                class="radio-item"
                                >{{ t('invoiceModal.headerType.company') }}</n-radio
                            >
                            <n-radio
                                :value="invoiceConstants.HEADER_TYPE.PERSONAL"
                                class="radio-item"
                                >{{ t('invoiceModal.headerType.personal') }}</n-radio
                            >
                        </n-radio-group>
                    </n-form-item>

                    <!-- 发票类型 -->
                    <n-form-item
                        path="invoiceType"
                        :label="t('invoiceModal.invoiceType.label')"
                        :show-feedback="false"
                        class="form-item"
                    >
                        <n-radio-group
                            v-model:value="formData.invoiceType"
                            class="radio-group"
                            size="small"
                        >
                            <n-radio
                                :value="invoiceConstants.INVOICE_TYPE.VAT"
                                class="radio-item"
                                >{{ t('invoiceModal.invoiceType.vat') }}</n-radio
                            >
                            <n-radio
                                v-if="isCompany"
                                :value="invoiceConstants.INVOICE_TYPE.SPECIAL"
                                class="radio-item"
                                >{{ t('invoiceModal.invoiceType.special') }}</n-radio
                            >
                        </n-radio-group>
                    </n-form-item>

                    <!-- 发票抬头 -->
                    <n-form-item
                        path="title"
                        :label="t('invoiceModal.invoiceTitle.label')"
                        :show-feedback="showTitleFeedback"
                        class="form-item"
                        ref="invoiceTitleRef"
                    >
                        <n-input
                            v-model:value="formData.title"
                            :placeholder="
                                isCompany
                                    ? t('invoiceModal.invoiceTitle.companyPlaceholder')
                                    : t('invoiceModal.invoiceTitle.personalPlaceholder')
                            "
                            class="dark-input"
                        />
                    </n-form-item>

                    <!-- 纳税人识别号 - 仅企业类型显示 -->
                    <n-form-item
                        v-if="isCompany"
                        path="taxNumber"
                        :label="t('invoiceModal.taxNumber.label')"
                        class="form-item"
                        :show-feedback="showTaxNumberFeedback"
                    >
                        <n-input
                            v-model:value="formData.taxNumber"
                            :placeholder="t('invoiceModal.taxNumber.placeholder')"
                            class="dark-input"
                        />
                    </n-form-item>

                    <!-- 企业注册地址 - 仅企业类型显示 -->
                    <n-form-item
                        v-if="isCompany"
                        path="address"
                        :label="t('invoiceModal.address.label')"
                        :show-feedback="showAddressFeedback"
                        class="form-item"
                    >
                        <n-input
                            v-model:value="formData.address"
                            :placeholder="t('invoiceModal.address.placeholder')"
                            class="dark-input"
                        />
                    </n-form-item>

                    <!-- 企业注册电话 - 仅企业类型显示 -->
                    <n-form-item
                        v-if="formData.headerType === invoiceConstants.HEADER_TYPE.COMPANY"
                        path="companyPhone"
                        :label="t('invoiceModal.phone.label')"
                        :show-feedback="showPhoneFeedback"
                        class="form-item"
                    >
                        <n-input
                            v-model:value="formData.companyPhone"
                            :placeholder="t('invoiceModal.phone.placeholder')"
                            class="dark-input"
                        />
                    </n-form-item>

                    <!-- 开户银行 - 仅企业类型显示 -->
                    <n-form-item
                        v-if="formData.headerType === invoiceConstants.HEADER_TYPE.COMPANY"
                        path="bank"
                        :label="t('invoiceModal.bank.label')"
                        :show-feedback="showBankFeedback"
                        class="form-item"
                    >
                        <n-input
                            v-model:value="formData.bank"
                            :placeholder="t('invoiceModal.bank.placeholder')"
                            class="dark-input"
                        />
                    </n-form-item>

                    <!-- 银行账号 - 仅企业类型显示 -->
                    <n-form-item
                        v-if="formData.headerType === invoiceConstants.HEADER_TYPE.COMPANY"
                        path="bankAccount"
                        :label="t('invoiceModal.bankAccount.label')"
                        :show-feedback="showBankAccountFeedback"
                        class="form-item"
                    >
                        <n-input
                            v-model:value="formData.bankAccount"
                            :placeholder="t('invoiceModal.bankAccount.placeholder')"
                            class="dark-input"
                        />
                    </n-form-item>

                    <!-- 手机号 - 仅个人类型显示 -->
                    <n-form-item
                        v-if="!isCompany"
                        path="phone"
                        :label="t('invoiceModal.mobile.label')"
                        :show-feedback="showMobileFeedback"
                        class="form-item"
                    >
                        <n-input
                            v-model:value="formData.phone"
                            :placeholder="t('invoiceModal.mobile.placeholder')"
                            class="dark-input"
                        />
                    </n-form-item>

                    <!-- 收票邮箱 -->
                    <n-form-item
                        path="email"
                        :label="t('invoiceModal.email.label')"
                        :show-feedback="showEmailFeedback"
                        class="form-item"
                    >
                        <n-input
                            v-model:value="formData.email"
                            :placeholder="t('invoiceModal.email.placeholder')"
                            class="dark-input"
                        />
                    </n-form-item>
                </n-form>
            </div>

            <n-divider />

            <!-- 订单信息 -->
            <div class="text-white text-xs flex self-center">
                <div>
                    {{ t('invoiceModal.orderInfo.orderNumber')
                    }}<span class="opacity-70">{{ props.orderNo || '123456789' }}</span>
                </div>
                <div class="ml-8">
                    {{ t('invoiceModal.orderInfo.amount')
                    }}<span class="opacity-70">￥{{ props.amount || '10.00' }}</span>
                </div>
                <div class="ml-8">
                    {{ t('invoiceModal.orderInfo.content')
                    }}<span class="opacity-70">{{ t('invoiceModal.orderInfo.serviceFee') }}</span>
                </div>
            </div>

            <!-- 弹窗底部 -->

            <div class="mt-13 self-end flex">
                <div
                    @click="submitForm"
                    class="text-white cursor-pointer submit-btn"
                >
                    {{ t('invoiceModal.buttons.submit') }}
                </div>
                <div
                    @click="closeModal"
                    class="ml-2 text-white cursor-pointer cancel-btn"
                >
                    {{ t('invoiceModal.buttons.cancel') }}
                </div>
            </div>
        </div>
    </n-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import {
    NModal,
    NButton,
    NIcon,
    NForm,
    NFormItem,
    NInput,
    NRadio,
    NRadioGroup,
    NDivider,
    useMessage,
} from 'naive-ui';
import type { FormInst, FormItemInst, FormItemRule } from 'naive-ui';
import { CloseOutline } from '@vicons/ionicons5';
import { getInvoiceConstants } from '../const';
import { useI18n } from 'vue-i18n';
import { postCreateInvoice } from '@/api/mods/quota.mod';
import type { PostCreateInvoiceReq } from '@/api/bos/quota.bo';
import type { InvoiceModalFormData, InvoiceModalProps } from '../interface';

const props = withDefaults(defineProps<InvoiceModalProps>(), {
    show: false,
    orderNo: '',
    amount: null,
});

const emit = defineEmits<{
    'update:show': [value: boolean];
    'invoice-submitted': [];
}>();

// 国际化
const { t } = useI18n();

// 计算国际化的发票常量
const invoiceConstants = computed(() => getInvoiceConstants(t));

// 消息提示
const message = useMessage();

const visible = computed({
    get: () => props.show,
    set: (value) => emit('update:show', value),
});

// 表单引用
const formRef = ref<FormInst | null>(null);
const invoiceTitleRef = ref<FormItemInst | null>(null);

// 控制显示反馈
const showTitleFeedback = ref(false);
const showEmailFeedback = ref(false);
const showTaxNumberFeedback = ref(false);
const showMobileFeedback = ref(false);
const showAddressFeedback = ref(false);
const showPhoneFeedback = ref(false);
const showBankFeedback = ref(false);
const showBankAccountFeedback = ref(false);

// 初始表单数据
const initialFormData: InvoiceModalFormData = {
    headerType: invoiceConstants.value.HEADER_TYPE.COMPANY, // 抬头类型
    invoiceType: invoiceConstants.value.INVOICE_TYPE.VAT, // 发票类型
    title: '', // 发票抬头
    taxNumber: '', // 纳税人识别号
    address: '', // 企业注册地址
    companyPhone: '', // 企业注册电话
    bank: '', // 开户银行
    bankAccount: '', // 银行账号
    email: '', // 收票邮箱
    phone: '', // 手机号
};

// 表单数据
const formData = ref<InvoiceModalFormData>({ ...initialFormData });

// 判断是否为企业类型
const isCompany = computed(
    () => formData.value.headerType === invoiceConstants.value.HEADER_TYPE.COMPANY,
);

// 判断是否为专用发票
const isSpecialInvoice = computed(
    () => formData.value.invoiceType === invoiceConstants.value.INVOICE_TYPE.SPECIAL,
);

// 当抬头类型切换时，清空校验并强制个人类型为普通发票
watch(
    () => formData.value.headerType,
    (newVal) => {
        formRef.value?.restoreValidation();
        showTitleFeedback.value = false;
        showTaxNumberFeedback.value = false;
        showAddressFeedback.value = false;
        showPhoneFeedback.value = false;
        showBankFeedback.value = false;
        showBankAccountFeedback.value = false;
        showMobileFeedback.value = false;
        showEmailFeedback.value = false;
        if (newVal === invoiceConstants.value.HEADER_TYPE.PERSONAL) {
            formData.value.invoiceType = invoiceConstants.value.INVOICE_TYPE.VAT;
        }
    },
);

// 当发票类型切换时，清空校验
watch(
    () => formData.value.invoiceType,
    () => {
        formRef.value?.restoreValidation();
        showTitleFeedback.value = false;
        showTaxNumberFeedback.value = false;
        showAddressFeedback.value = false;
        showPhoneFeedback.value = false;
        showBankFeedback.value = false;
        showBankAccountFeedback.value = false;
        showMobileFeedback.value = false;
        showEmailFeedback.value = false;
    },
);

// 动态计算表单验证规则
const formRules = computed(() => {
    const rules: Record<string, FormItemRule> = {
        headerType: {
            required: true,
            message: t('invoiceModal.validation.selectHeaderType'),
            trigger: ['blur', 'change'],
        } as FormItemRule,
        invoiceType: {
            required: true,
            message: t('invoiceModal.validation.selectInvoiceType'),
            trigger: ['blur', 'change'],
        } as FormItemRule,
        email: {
            required: true,
            message: t('invoiceModal.validation.enterEmail'),
            trigger: ['blur', 'input'],
            validator: (rule: FormItemRule, value: string) => {
                if (!value) {
                    showEmailFeedback.value = true;
                    return new Error(t('invoiceModal.validation.enterEmail'));
                }
                const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
                if (!emailRegex.test(value)) {
                    showEmailFeedback.value = true;
                    return new Error(t('invoiceModal.validation.emailFormatInvalid'));
                }
                showEmailFeedback.value = false;
                return true;
            },
        } as FormItemRule,
    };

    // 企业/单位类型的必填字段
    if (isCompany.value) {
        rules.title = {
            required: true,
            message: t('invoiceModal.validation.enterInvoiceTitle'),
            trigger: ['blur', 'input'],
            validator: (rule: FormItemRule, value: string) => {
                if (!value) {
                    showTitleFeedback.value = true;
                    return new Error(t('invoiceModal.validation.enterInvoiceTitle'));
                }
                showTitleFeedback.value = false;
                return true;
            },
        } as FormItemRule;

        rules.taxNumber = {
            required: true,
            trigger: ['blur', 'input'],
            validator: (rule: FormItemRule, value: string) => {
                if (!value) {
                    showTaxNumberFeedback.value = true;
                    return new Error(t('invoiceModal.validation.enterTaxNumber'));
                }
                if (!/^[0-9A-Z]{18}$/.test(value)) {
                    showTaxNumberFeedback.value = true;
                    return new Error(t('invoiceModal.validation.taxNumberFormatInvalid'));
                }
                showTaxNumberFeedback.value = false;
                return true;
            },
        } as FormItemRule;

        // 专用发票下必填，普通发票下选填
        rules.address = {
            required: isSpecialInvoice.value,
            trigger: ['blur', 'input'],
            validator: (rule: FormItemRule, value: string) => {
                if (isSpecialInvoice.value && !value) {
                    showAddressFeedback.value = true;
                    return new Error(t('invoiceModal.validation.enterAddress'));
                }
                showAddressFeedback.value = false;
                return true;
            },
        } as FormItemRule;

        rules.companyPhone = {
            required: isSpecialInvoice.value,
            trigger: ['blur', 'input'],
            validator: (rule: FormItemRule, value: string) => {
                if (isSpecialInvoice.value && !value) {
                    showPhoneFeedback.value = true;
                    return new Error(t('invoiceModal.validation.enterCompanyPhone'));
                }
                if (value && !/^(1[3-9]\d{9}|0\d{2,3}-?\d{7,8}|\d{7,8})$/.test(value)) {
                    showPhoneFeedback.value = true;
                    return new Error(t('invoiceModal.validation.mobileFormatInvalid'));
                }
                showPhoneFeedback.value = false;
                return true;
            },
        } as FormItemRule;

        rules.bank = {
            required: isSpecialInvoice.value,
            trigger: ['blur', 'input'],
            validator: (rule: FormItemRule, value: string) => {
                if (isSpecialInvoice.value && !value) {
                    showBankFeedback.value = true;
                    return new Error(t('invoiceModal.validation.enterBank'));
                }
                showBankFeedback.value = false;
                return true;
            },
        } as FormItemRule;

        rules.bankAccount = {
            required: isSpecialInvoice.value,
            trigger: ['blur', 'input'],
            validator: (rule: FormItemRule, value: string) => {
                if (isSpecialInvoice.value && !value) {
                    showBankAccountFeedback.value = true;
                    return new Error(t('invoiceModal.validation.enterBankAccount'));
                }
                showBankAccountFeedback.value = false;
                return true;
            },
        } as FormItemRule;
    }
    // 个人类型的必填字段
    else {
        rules.title = {
            required: true,
            message: t('invoiceModal.validation.invoiceTitleEmpty'),
            trigger: ['blur', 'input'],
            validator: (rule: FormItemRule, value: string) => {
                if (!value) {
                    showTitleFeedback.value = true;
                    return new Error(t('invoiceModal.validation.invoiceTitleEmpty'));
                }
                if (!/^[^0-9].{1,}$/.test(value)) {
                    showTitleFeedback.value = true;
                    return new Error(t('invoiceModal.validation.invoiceTitleInvalid'));
                }
                showTitleFeedback.value = false;
                return true;
            },
        } as FormItemRule;

        rules.phone = {
            required: true,
            trigger: ['blur', 'input'],
            validator: (rule: FormItemRule, value: string) => {
                if (!value) {
                    showMobileFeedback.value = true;
                    return new Error(t('invoiceModal.validation.enterMobile'));
                }
                const phoneRegex = /^(1[3-9]\d{9}|0\d{2,3}-?\d{7,8}|\d{7,8})$/;
                if (!phoneRegex.test(value)) {
                    showMobileFeedback.value = true;
                    return new Error(t('invoiceModal.validation.mobileFormatInvalid'));
                }
                showMobileFeedback.value = false;
                return true;
            },
        } as FormItemRule;
    }

    return rules;
});

const closeModal = () => {
    visible.value = false;
    // 重置表单数据到初始值
    formData.value = { ...initialFormData };
    // 重置反馈显示状态
    showTitleFeedback.value = false;
    showEmailFeedback.value = false;
    showTaxNumberFeedback.value = false;
    showMobileFeedback.value = false;
    showAddressFeedback.value = false;
    showPhoneFeedback.value = false;
    showBankFeedback.value = false;
    showBankAccountFeedback.value = false;
};

const submitForm = async () => {
    try {
        await formRef.value?.validate();
        // 验证通过，准备提交数据
        const requestData: PostCreateInvoiceReq = {
            title_type:
                formData.value.headerType === invoiceConstants.value.HEADER_TYPE.COMPANY ? 1 : 2,
            invoice_type:
                formData.value.invoiceType === invoiceConstants.value.INVOICE_TYPE.VAT ? 1 : 2,
            invoice_title: formData.value.title.trim(),
            receive_email: formData.value.email.trim(),
            order_id: props.orderNo || '',
            invoice_content: '软件服务费',
        };

        // 如果是企业类型，添加企业相关字段
        if (isCompany.value) {
            requestData.taxpayer_id = formData.value.taxNumber.trim();
            requestData.company_address = formData.value.address.trim();
            requestData.company_phone = formData.value.companyPhone.trim();
            requestData.bank_name = formData.value.bank.trim();
            requestData.bank_account = formData.value.bankAccount.trim();
        } else {
            requestData.phone = formData.value.phone.trim();
        }

        const response = await postCreateInvoice(requestData);

        if (response.code === 200) {
            message.success(t('invoiceModal.submitSuccess'));
            // 触发发票提交成功事件，以便父组件刷新表格数据
            emit('invoice-submitted');
            closeModal();
        }
    } catch (error) {
        console.log(error);
    }
};
</script>

<style scoped lang="less">
.modal-container {
    width: 555px;
    background: #1b1d22;
    border-radius: 6px;
    overflow-y: scroll;
    max-height: 655px;
    padding: 16px 20px 20px;
}

.modal-content {
    color: #fff;

    .radio-item {
        display: flex;
        align-items: center;

        @media (max-width: 414px) {
            align-items: baseline;
        }
    }

    .invoice-form {
        .form-item {
            margin-bottom: 8px;
            :deep(.n-form-item-label) {
                font-size: 12px;
                min-width: 108px; // 确保所有标签宽度一致，对齐表单项
            }
        }

        .radio-group {
            display: flex;
            gap: 12px;
            :deep(.n-radio .n-radio__label) {
                color: #fff;
                font-size: 12px;
            }
        }
    }

    :deep(.n-input-wrapper) {
        background: #1b1d22;
    }
}

.submit-btn {
    background: #1770e6;
    padding: 5px 20px;
    border-radius: 2px;
    border: 1px solid #156ad9;
}

.cancel-btn {
    padding: 5px 20px;
    border-radius: 2px;
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.1);
}

.tip-text::before {
    content: '*';
    color: #ff6b6b;
    margin-right: 4px;
}
</style>
