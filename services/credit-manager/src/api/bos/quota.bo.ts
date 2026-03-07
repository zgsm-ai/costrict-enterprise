/**
 * @file quota types
 */

export interface QuotaList {
    amount: number;
    expiry_date: string;
}

export interface UsageConsumptionRecord {
    id: number;
    user_id: string;
    model: string;
    mode: string;
    tokens: number;
    credits_used: number;
    package: string;
    record_time: string;
    create_time: string;
    update_time: string;
}

export interface GetUsageStatisticsReq {
    page: number;
    page_size: number;
    start_time?: string;
    end_time?: string;
    time_range?: string;
}

export interface GetUsageStatisticsRes {
    records: Array<UsageConsumptionRecord>;
    total: number;
    page: number;
    page_size: number;
}

export interface GetUserQuotaRes {
    total_quota: number;
    used_quota: number;
    quota_list: Array<QuotaList>;
    is_star?: string;
}

export interface GetQuotaAuditRecordsReq {
    page: number;
    page_size: number;
}

export interface DetailItems {
    amount: number;
    expiry_date: string;
    status: string;
}

interface RecordDetail {
    operation: string;
    items: Array<DetailItems>;
}

export interface QuotaAuditRecord {
    amount: number;
    operation: string;
    voucher_code: string;
    related_user: string;
    strategy_name: string;
    expiry_date: string;
    create_time: string;
    details: RecordDetail;
}

export interface GetQuotaAuditRecordsRes {
    total: number;
    records: Array<QuotaAuditRecord>;
}

export interface PostQuotaTransferOutReq {
    receiver_id: string;
    quota_list: Array<QuotaList>;
}

export interface PostQuotaTransferOutRes {
    voucher_code: string;
    related_user: string;
    operation: string;
    quota_list: Array<QuotaList>;
}

export interface PostQuotaTransferInReq {
    voucher_code: string;
}

export interface QuotaTransferInQuotaList extends QuotaList {
    is_expired: boolean;
    success: boolean;
}

export interface PostQuotaTransferInRes {
    giver_id: string;
    giver_name: string;
    giver_phone: string;
    giver_github: string;
    receiver_id: string;
    quota_list: Array<QuotaTransferInQuotaList>;
    voucher_code: string;
    operation: string;
    status: string;
    message: string;
}

export interface GetUserTokenRes {
    access_token: string;
    message: string;
    state: string;
    success: boolean;
}

export interface UserInfoData {
    email: string;
    githubID: string;
    githubName: string;
    phone: string;
    username: string;
    uuid: string;
    employee_number?: string;
    isPrivate?: boolean;
}

export type GetUserInfoRes = UserInfoData;

export interface GetBindAccountReq {
    bindType: string;
    state: string;
}

export interface GetBindAccountRes {
    url: string;
    message: string;
    success: boolean;
}

export interface GetInviteCodeRes {
    invite_code: string;
}

export interface GetLoginUrlReq {
    inviter_code?: string;
    redirect_service?: string;
}

export interface GetLoginUrlRes {
    url: string;
}

export interface GetOrdersReq {
    page: number;
    page_size: number;
}

export interface Order {
    id: number;
    order_id: string;
    user_id: string;
    amount: number;
    status: string;
    quota_type: string;
    order_source: string;
    credit_count: number;
    credit_expire_date: string;
    created_at: string;
    updated_at: string;
    invoice_status: number;
    description: string;
    jd_order_id: string;
}

export interface GetOrdersRes {
    orders: Array<Order>;
    total: number;
    limit: number;
    offset: number;
}

export interface PostCreateOrderReq {
    quota_type: string;
    quantity?: number;
}

export interface PostCreateOrderRes {
    id: number;
    order_id: string;
    user_id: string;
    amount: number;
    credit_count: number;
    credit_expire_date: string;
    status: string;
    quota_type: string;
    description: string;
    order_source: string;
    jd_order_id: string;
    invoice_status: number;
    created_at: string;
    updated_at: string;
}

export type TPaymentChannel = 'CMB' | 'WECHAT' | 'ALIPAY';

export interface PostPaymentQrcodeReq {
    order_id: string;
    payment_channel: TPaymentChannel;
}

export interface PostPaymentQrcodeRes {
    version: string;
    encoding: string;
    sign: string;
    signMethod: string;
    returnCode: string;
    respCode: string;
    qrCode: string;
    cmbOrderId: string;
    txnTime: string;
    biz_content: {
        merId: string;
        orderId: string;
        cmbOrderId: string;
        qrCode: string;
        txnTime: string;
    };
}

export interface PostCreateInvoiceReq {
    title_type: 1 | 2;
    invoice_type: 1 | 2;
    invoice_title: string;
    taxpayer_id?: string;
    company_address?: string;
    company_phone?: string;
    bank_name?: string;
    bank_account?: string;
    receive_email: string;
    order_id: string;
    invoice_content: string;
}

export interface PostCreateInvoiceRes {
    invoice_id: number;
    title_type: number;
    invoice_type: number;
    invoice_title: string;
    taxpayer_id?: string;
    company_address?: string;
    company_phone?: string;
    bank_name?: string;
    bank_account?: string;
    receive_email: string;
    order_id: string;
    amount: number;
    invoice_content: string;
    invoice_status: number;
    apply_time: string;
    issue_time?: string;
}

// 配额类型相关接口类型定义
export interface QuotaMarketingRules {
    id: number;
    rule_type: number; // 规则类型：1=首次充值特惠，2=限时特惠，3=渠道专属，4=阶梯充值
    pay_amount: number; // 优惠后支付金额
    pay_discount: number; // 优惠支付折扣
    status: number; // 规则状态：0=未生效，1=生效中，2=已停用
    apply_user_type: number; // 适用用户类型：1=新用户，2=老用户，3=所有用户
    created_at: string;
    updated_at: string;
}

export interface QuotaTypeWithMarketingRules {
    id: number;
    display_name: string;
    credit_count: number;
    amount: number;
    valid_days: number | null;
    quota_marketing_rules_id: number | null;
    original_amount: number;
    equivalent_credits: number;
    bonus_credits: number;
    estimated_requests: number;
    created_at: string;
    updated_at: string;
    marketing_rules: QuotaMarketingRules | null;
    isFirstPurchase?: boolean;
}

export type GetQuotaTypesRes = Array<QuotaTypeWithMarketingRules>;

export interface GetQuotaTypeByIdReq {
    id: number;
    quantity?: number;
}

export type GetQuotaTypeByIdRes = QuotaTypeWithMarketingRules;
