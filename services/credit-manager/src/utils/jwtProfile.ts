export interface JwtProfile {
    userName: string;
    userId: string;
    universalId: string;
    employeeNumber: string;
}

// 只解码展示字段；JWT 的可信性及接口权限仍由服务端验证。
export function readJwtProfile(token: string | undefined): JwtProfile {
    const empty: JwtProfile = { userName: '', userId: '', universalId: '', employeeNumber: '' };
    if (!token) return empty;

    try {
        const parts = token.split('.');
        if (parts.length !== 3) return empty;
        const payload = parts[1].replace(/-/g, '+').replace(/_/g, '/');
        const binary = atob(payload.padEnd(Math.ceil(payload.length / 4) * 4, '='));
        const bytes = Uint8Array.from(binary, (character) => character.charCodeAt(0));
        const claims: unknown = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(bytes));
        if (!claims || typeof claims !== 'object' || Array.isArray(claims)) return empty;
        const fields = claims as Record<string, unknown>;
        const text = (key: string): string => (typeof fields[key] === 'string' ? fields[key] : '');
        return {
            userName: text('display_name'),
            userId: text('sub'),
            universalId: text('universal_id'),
            employeeNumber: text('employee_number'),
        };
    } catch {
        // 缺失或无法解码的载荷不影响基于 Cookie 存在性的登录判断。
        return empty;
    }
}
