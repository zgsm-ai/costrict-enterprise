const RETURN_TO_KEY = 'login_service_return_to';

const getBasePath = (): string =>
    (import.meta.env.BASE_URL || '/credit/manager').replace(/\/$/, '');

export function saveLoginReturnTo(url: string): void {
    sessionStorage.setItem(RETURN_TO_KEY, url);
}

export function getLoginReturnTo(): string {
    const fallback = new URL(getBasePath() || '/', window.location.origin).href;
    const saved = sessionStorage.getItem(RETURN_TO_KEY);
    if (!saved) return fallback;

    try {
        const url = new URL(saved);
        const base = getBasePath();
        if (
            url.origin === window.location.origin &&
            (url.pathname === base || url.pathname.startsWith(`${base}/`)) &&
            url.pathname !== `${base}/login`
        ) {
            return url.href;
        }
    } catch {
        // 无效或旧的回跳地址退回项目首页。
    }
    return fallback;
}

export function getLoginServiceUrl(returnTo: string = getLoginReturnTo()): string {
    const base = import.meta.env.VITE_LOGIN_SERVICE_BASE_URL || window.location.origin;
    const url = new URL('/login-service/api/v1/web/login', base);
    url.searchParams.set('return_to', returnTo);
    return url.href;
}
