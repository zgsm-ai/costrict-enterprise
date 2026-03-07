/**
 * @file activity types
 */

/**
 * 用户个人信息数据
 */
export interface UserMeData {
    /** 用户ID */
    userId: string;
    /** 用户名 */
    username: string;
    /** 显示名称 */
    displayName: string;
    /** Casdoor ID */
    casdoorId: string;
    /** 创建时间 */
    createdTime: string;
    /** 注册顺序 */
    registerOrder: number;
    /** 注册顺序百分比 */
    registerOrderPercent: number;
    /** 最新活动时间 */
    latestActivity: string;
    /** 总延迟时间（毫秒） */
    totalLatencyMs: number;
    /** 总 Token 数 */
    totalTokens: number;
    /** 总请求数 */
    totalRequests: number;
    /** 总使用天数 */
    totalUsageDays: number;
    /** 模型统计（JSON 字符串） */
    modelStats: string;
    /** 模式统计（JSON 字符串） */
    modeStats: string;
    /** 更新时间 */
    updatedAt: string;
    /** 积分使用量 */
    creditUsage: number;
    /** 积分使用排名 */
    creditUsageOrder: number;
    /** 积分使用排名百分比 */
    creditUsageOrderPercent: number;
    /** 访问ID */
    accessId: string;
    /** 是否内部用户 */
    isInner: number;
    /** 模型统计数据（JSON 对象） */
    modelStatsJson: Record<string, number> | null;
    /** 模式统计数据（JSON 对象） */
    modeStatsJson: Record<string, number>;
    /** 用户身份 */
    identity: string;
}

/**
 * 获取个人信息响应
 */
export type GetUserMeRes = UserMeData;

/**
 * 分享记录数据
 */
export interface ShareRecordData {
    /** 用户ID */
    userId: string;
    /** 访问ID */
    accessId: string;
    /** 用户名 */
    username: string;
    /** 分享次数 */
    shareNum: number;
    /** 邀请码 */
    inviteCode: string;
    /** 更新时间 */
    updatedAt: string;
}

/**
 * 记录分享响应
 */
export type RecordShareRes = ShareRecordData;
