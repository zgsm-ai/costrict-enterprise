export interface MaintenanceImpact {
    affected: boolean;
    description: string;
}

export interface CreditMaintenance {
    maintain: boolean;
    title: string;
    /** 北京时间，格式 YYYY-MM-DD HH:mm:ss */
    start_time: string;
    end_time: string;
    impacts: {
        quota: MaintenanceImpact;
        order: MaintenanceImpact;
    };
}
