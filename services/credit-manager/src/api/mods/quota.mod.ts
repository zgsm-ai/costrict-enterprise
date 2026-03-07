/**
 * @file quota api
 */

import { get, post } from '@/utils/request';
import type {
    GetBindAccountReq,
    GetBindAccountRes,
    GetInviteCodeRes,
    GetLoginUrlReq,
    GetLoginUrlRes,
    GetOrdersReq,
    GetOrdersRes,
    GetQuotaAuditRecordsReq,
    GetQuotaAuditRecordsRes,
    GetUsageStatisticsReq,
    GetUsageStatisticsRes,
    GetUserInfoRes,
    GetUserQuotaRes,
    GetUserTokenRes,
    PostCreateOrderReq,
    PostCreateOrderRes,
    PostPaymentQrcodeReq,
    PostPaymentQrcodeRes,
    PostCreateInvoiceReq,
    PostCreateInvoiceRes,
    PostQuotaTransferInReq,
    PostQuotaTransferInRes,
    PostQuotaTransferOutReq,
    PostQuotaTransferOutRes,
    Order,
    GetQuotaTypesRes,
    GetQuotaTypeByIdReq,
    GetQuotaTypeByIdRes,
} from '../bos/quota.bo';

export interface ApiResponse<T = unknown> {
    code: number;
    message: string;
    data: T;
}

export const getUserQuota = (): Promise<ApiResponse<GetUserQuotaRes>> => {
    return get('/quota-manager/api/v1/quota');
};

export const getQuotaAuditRecords = (
    params: GetQuotaAuditRecordsReq,
): Promise<ApiResponse<GetQuotaAuditRecordsRes>> => {
    return get('/quota-manager/api/v1/quota/audit', params);
};

export const postQuotaOut = (
    params: PostQuotaTransferOutReq,
): Promise<ApiResponse<PostQuotaTransferOutRes>> => {
    return post('/quota-manager/api/v1/quota/transfer-out', params);
};

export const postQuotaIn = (
    params: PostQuotaTransferInReq,
): Promise<ApiResponse<PostQuotaTransferInRes>> => {
    return post('/quota-manager/api/v1/quota/transfer-in', params);
};

export const getUserToken = (): Promise<ApiResponse<GetUserTokenRes>> => {
    return get('/oidc-auth/api/v1/manager/token');
};

export const getUserInfo = (): Promise<ApiResponse<GetUserInfoRes>> => {
    return get('/oidc-auth/api/v1/manager/userinfo');
};

export const getBindAccount = (
    params: GetBindAccountReq,
): Promise<ApiResponse<GetBindAccountRes>> => {
    return get('/oidc-auth/api/v1/manager/bind/account', params);
};

export const getInviteCode = (): Promise<ApiResponse<GetInviteCodeRes>> => {
    return get('/oidc-auth/api/v1/manager/invite-code');
};

export const getLoginUrl = (params?: GetLoginUrlReq): Promise<ApiResponse<GetLoginUrlRes>> => {
    return get('/oidc-auth/api/v1/manager/login', params);
};

export const getUsageStatistics = (
    params: GetUsageStatisticsReq,
): Promise<ApiResponse<GetUsageStatisticsRes>> => {
    return get('/quota-manager/api/v1/usage/statistics', params);
};

export const getOrders = (params: GetOrdersReq): Promise<ApiResponse<GetOrdersRes>> => {
    return get('/quota-order-manager/api/v1/orders', params);
};

export const postCreateOrder = (
    params: PostCreateOrderReq,
): Promise<ApiResponse<PostCreateOrderRes>> => {
    return post('/quota-order-manager/api/v1/orders', params);
};

export const postPaymentQrcode = (
    params: PostPaymentQrcodeReq,
): Promise<ApiResponse<PostPaymentQrcodeRes>> => {
    return post('/quota-order-manager/api/v1/payment/initiate', params);
};

export const postCreateInvoice = (
    params: PostCreateInvoiceReq,
): Promise<ApiResponse<PostCreateInvoiceRes>> => {
    return post('/quota-order-manager/api/v1/invoices', params);
};

export const getOrderById = (orderId: string): Promise<ApiResponse<Order>> => {
    return get(`/quota-order-manager/api/v1/orders/${orderId}`);
};

export const getQuotaTypes = (): Promise<ApiResponse<GetQuotaTypesRes>> => {
    return get('/quota-order-manager/api/v1/quotas/types');
};

export const getQuotaTypeById = (
    params: GetQuotaTypeByIdReq,
): Promise<ApiResponse<GetQuotaTypeByIdRes>> => {
    return get(`/quota-order-manager/api/v1/quotas/types/${params.id}`, {
        quantity: params.quantity,
    });
};
