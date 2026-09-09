import type { MockMethod } from 'vite-plugin-mock';
import type { CreditMaintenance } from '../src/api/bos/announcement.bo';

const announcement: CreditMaintenance = {
    maintain: true,
    title: '配额模块维护中...',
    start_time: '2026-09-09 10:30:00',
    end_time: '2026-09-09 17:00:00',
    impacts: {
        quota: {
            affected: false,
            description: '用量统计暂时不可用，请在维护结束后查看。',
        },
        order: {
            affected: false,
            description: '订阅页面及套餐购买暂时不可用，请在维护结束后操作。',
        },
    },
};

export default [
    {
        url: '/costrict-static/announcement/credit-maintain.json',
        method: 'get',
        response: () => announcement,
    },
] as MockMethod[];
