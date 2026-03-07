import Cookies from 'js-cookie';

const TOKEN_KEY = 'zgsmAdminToken';

export const getHashToken = () => {
    const url = new URL(window.location.href);
    const params = new URLSearchParams(url.search);
    const hasToken = params.get('state');

    return hasToken;
};

export const getToken = () => {
    return Cookies.get(TOKEN_KEY);
};

export const setToken = (token: string) => {
    return Cookies.set(TOKEN_KEY, token);
};

export const clearToken = () => {
    return Cookies.remove(TOKEN_KEY);
};

export const cleanUrlState = () => {
    const url = new URL(window.location.href);
    const params = new URLSearchParams(url.search);

    if (params.has('state')) {
        params.delete('state');
        const newUrl =
            url.origin +
            url.pathname +
            (params.toString() ? '?' + params.toString() : '') +
            url.hash;
        history.replaceState({}, document.title, newUrl);
    }
};

export const tokenManager = {
    getHashToken,
    getToken,
    setToken,
    clearToken,
    cleanUrlState,
    validateToken: (token: string) => {
        return !!token && token.length > 0;
    },
};
