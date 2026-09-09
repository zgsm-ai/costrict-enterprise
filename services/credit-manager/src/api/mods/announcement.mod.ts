import { getStaticJson } from '@/utils/request';

export function getCreditMaintenance(): Promise<unknown> {
    return getStaticJson('/costrict-static/announcement/credit-maintain.json');
}
