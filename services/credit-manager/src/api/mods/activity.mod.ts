/**
 * @file activity api
 */

import { get, post } from '@/utils/request';
import type { GetUserMeRes, RecordShareRes } from '../bos/activity.bo';

export interface ApiResponse<T = unknown> {
    code: number;
    message: string;
    data: T;
}

/**
 * 获取当前用户个人信息
 * @returns 用户个人信息
 */
export const getUserMe = (): Promise<ApiResponse<GetUserMeRes>> => {
    return get('/operational_activities/api/v1/user/me');
};

/**
 * 记录分享埋点
 * @returns 分享记录数据
 */
export const recordShare = (): Promise<ApiResponse<RecordShareRes>> => {
    return post('/operational_activities/api/v1/share/record', {});
};
