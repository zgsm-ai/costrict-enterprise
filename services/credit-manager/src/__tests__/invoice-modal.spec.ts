import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { nextTick } from 'vue';
import InvoiceModal from '@/views/Home/components/invoice-modal.vue';

const mockSuccess = vi.fn();
const mockError = vi.fn();

// Mock naive-ui
vi.mock('naive-ui', async () => {
    const actual = await vi.importActual('naive-ui');
    return {
        ...actual,
        useMessage: () => ({
            success: mockSuccess,
            error: mockError,
        }),
    };
});

// Mock API
const mockPostCreateInvoice = vi.fn();
vi.mock('@/api/mods/quota.mod', () => ({
    postCreateInvoice: (data: any) => mockPostCreateInvoice(data),
}));

// Mock i18n
vi.mock('vue-i18n', () => ({
    useI18n: () => ({
        t: (key: string) => key,
    }),
}));

const createWrapper = (props = {}) => {
    return mount(InvoiceModal, {
        props: {
            show: true,
            orderNo: 'ORDER123',
            amount: 100,
            ...props,
        },
        attachTo: document.body,
    });
};

describe('InvoiceModal', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    describe('initial state', () => {
        it('should default to company header type and VAT invoice type', () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            expect(vm.formData.headerType).toBe('company');
            expect(vm.formData.invoiceType).toBe('vat');
            expect(vm.isCompany).toBe(true);
            expect(vm.isSpecialInvoice).toBe(false);
        });

        it('should show all company fields by default', () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            expect(vm.isCompany).toBe(true);
            expect(vm.formData.taxNumber).toBe('');
            expect(vm.formData.address).toBe('');
            expect(vm.formData.companyPhone).toBe('');
            expect(vm.formData.bank).toBe('');
            expect(vm.formData.bankAccount).toBe('');
        });

        it('should not show personal phone field for company header type', () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            expect(vm.isCompany).toBe(true);
            expect(vm.formData.phone).toBe('');
        });
    });

    describe('header type switching', () => {
        it('should switch to personal and force VAT invoice type', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.headerType = 'personal';
            await nextTick();

            expect(vm.formData.invoiceType).toBe('vat');
            expect(vm.isCompany).toBe(false);
        });

        it('should hide special invoice option for personal header type', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.headerType = 'personal';
            await nextTick();

            expect(vm.isCompany).toBe(false);
            expect(vm.formData.invoiceType).toBe('vat');
        });
    });

    describe('invoice type switching', () => {
        it('should allow special invoice for company header type', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.invoiceType = 'special';
            await nextTick();

            expect(vm.isSpecialInvoice).toBe(true);
        });

        it('should not allow special invoice for personal header type', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.headerType = 'personal';
            await nextTick();

            expect(vm.formData.invoiceType).toBe('vat');
            expect(vm.isSpecialInvoice).toBe(false);
        });
    });

    describe('form validation rules', () => {
        it('should require tax number for company header type', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            const rules = vm.formRules;
            expect(rules.taxNumber).toBeDefined();
        });

        it('should require address for special invoice', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.invoiceType = 'special';
            await nextTick();

            const rules = vm.formRules;
            expect(rules.address).toBeDefined();
            expect(rules.address.required).toBe(true);
        });

        it('should not require address for VAT invoice', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.invoiceType = 'vat';
            await nextTick();

            const rules = vm.formRules;
            expect(rules.address).toBeDefined();
            expect(rules.address.required).toBe(false);
        });

        it('should require companyPhone for special invoice', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.invoiceType = 'special';
            await nextTick();

            const rules = vm.formRules;
            expect(rules.companyPhone.required).toBe(true);
        });

        it('should not require companyPhone for VAT invoice', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.invoiceType = 'vat';
            await nextTick();

            const rules = vm.formRules;
            expect(rules.companyPhone.required).toBe(false);
        });

        it('should validate mobile phone format for personal phone', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.headerType = 'personal';
            await nextTick();

            const rules = vm.formRules;
            const phoneRule = rules.phone;

            // Valid mobile number
            let result = phoneRule.validator({}, '13800138000');
            expect(result).toBe(true);

            // Valid landline with area code
            result = phoneRule.validator({}, '010-12345678');
            expect(result).toBe(true);

            // Valid landline without area code
            result = phoneRule.validator({}, '12345678');
            expect(result).toBe(true);

            // Invalid format
            result = phoneRule.validator({}, 'invalid');
            expect(result).toBeInstanceOf(Error);
        });

        it('should validate landline format for company phone', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            const rules = vm.formRules;
            const companyPhoneRule = rules.companyPhone;

            // Valid mobile number
            let result = companyPhoneRule.validator({}, '13800138000');
            expect(result).toBe(true);

            // Valid landline with area code
            result = companyPhoneRule.validator({}, '0755-12345678');
            expect(result).toBe(true);

            // Valid landline without area code
            result = companyPhoneRule.validator({}, '87654321');
            expect(result).toBe(true);
        });
    });

    describe('form submission', () => {
        it('should assemble request data correctly for company VAT type', async () => {
            mockPostCreateInvoice.mockResolvedValue({ code: 200 });
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.title = 'Test Company';
            vm.formData.taxNumber = '91110000123456789X';
            vm.formData.address = 'Beijing';
            vm.formData.companyPhone = '010-12345678';
            vm.formData.bank = 'Bank of China';
            vm.formData.bankAccount = '6222021234567890123';
            vm.formData.email = 'test@example.com';
            await nextTick();

            await vm.submitForm();
            await flushPromises();

            expect(mockPostCreateInvoice).toHaveBeenCalled();
            const callData = mockPostCreateInvoice.mock.calls[0][0];
            expect(callData.title_type).toBe(1);
            expect(callData.invoice_type).toBe(1);
            expect(callData.invoice_title).toBe('Test Company');
            expect(callData.taxpayer_id).toBe('91110000123456789X');
            expect(callData.company_address).toBe('Beijing');
            expect(callData.company_phone).toBe('010-12345678');
            expect(callData.bank_name).toBe('Bank of China');
            expect(callData.bank_account).toBe('6222021234567890123');
            expect(callData.receive_email).toBe('test@example.com');
        });

        it('should assemble request data correctly for company special invoice type', async () => {
            mockPostCreateInvoice.mockResolvedValue({ code: 200 });
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.invoiceType = 'special';
            vm.formData.title = 'Test Company';
            vm.formData.taxNumber = '91110000123456789X';
            vm.formData.address = 'Beijing';
            vm.formData.companyPhone = '0755-87654321';
            vm.formData.bank = 'Bank of China';
            vm.formData.bankAccount = '6222021234567890123';
            vm.formData.email = 'test@example.com';
            await nextTick();

            await vm.submitForm();
            await flushPromises();

            expect(mockPostCreateInvoice).toHaveBeenCalled();
            const callData = mockPostCreateInvoice.mock.calls[0][0];
            expect(callData.title_type).toBe(1);
            expect(callData.invoice_type).toBe(2);
            expect(callData.invoice_title).toBe('Test Company');
            expect(callData.taxpayer_id).toBe('91110000123456789X');
            expect(callData.company_address).toBe('Beijing');
            expect(callData.company_phone).toBe('0755-87654321');
            expect(callData.bank_name).toBe('Bank of China');
            expect(callData.bank_account).toBe('6222021234567890123');
            expect(callData.receive_email).toBe('test@example.com');
        });

        it('should assemble request data correctly for personal type', async () => {
            mockPostCreateInvoice.mockResolvedValue({ code: 200 });
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.headerType = 'personal';
            await nextTick();

            vm.formData.title = '张三';
            vm.formData.phone = '13800138000';
            vm.formData.email = 'zhangsan@example.com';
            await nextTick();

            await vm.submitForm();
            await flushPromises();

            expect(mockPostCreateInvoice).toHaveBeenCalled();
            const callData = mockPostCreateInvoice.mock.calls[0][0];
            expect(callData.title_type).toBe(2);
            expect(callData.invoice_type).toBe(1);
            expect(callData.invoice_title).toBe('张三');
            expect(callData.phone).toBe('13800138000');
            expect(callData.receive_email).toBe('zhangsan@example.com');
        });

        it('should support landline number for personal type', async () => {
            mockPostCreateInvoice.mockResolvedValue({ code: 200 });
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.headerType = 'personal';
            await nextTick();

            vm.formData.title = '张三';
            vm.formData.phone = '010-12345678';
            vm.formData.email = 'zhangsan@example.com';
            await nextTick();

            await vm.submitForm();
            await flushPromises();

            expect(mockPostCreateInvoice).toHaveBeenCalled();
            const callData = mockPostCreateInvoice.mock.calls[0][0];
            expect(callData.phone).toBe('010-12345678');
        });
    });

    describe('close modal', () => {
        it('should emit update:show event when closing', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.closeModal();
            await nextTick();

            expect(wrapper.emitted('update:show')).toBeTruthy();
            expect(wrapper.emitted('update:show')?.[0]).toEqual([false]);
        });

        it('should reset form data when closing', async () => {
            const wrapper = createWrapper();
            const vm = wrapper.vm as any;

            vm.formData.title = 'Test';
            vm.formData.email = 'test@test.com';
            await nextTick();

            vm.closeModal();
            await nextTick();

            expect(vm.formData.title).toBe('');
            expect(vm.formData.email).toBe('');
        });
    });
});
