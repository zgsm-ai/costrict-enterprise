import { defineStore } from 'pinia';

interface UserInfo {
    phoneNumber: string;
    userId: string;
    githubName: string;
    employeeNumber: string;
    userName: string;
    isPrivate: boolean;
}

export const useUserStore = defineStore('counter', {
    state: () => ({
        phoneNumber: '',
        userId: '',
        githubName: '',
        employeeNumber: '',
        userName: '',
        isPrivate: false,
        isTokenInitialized: false,
        isAuthenticating: false,
        authError: null as string | null,
    }),

    actions: {
        updateUserInfo({
            phoneNumber,
            userId,
            githubName,
            employeeNumber,
            userName,
            isPrivate,
        }: UserInfo) {
            this.$patch({
                phoneNumber,
                userId,
                githubName,
                employeeNumber,
                userName,
                isPrivate,
            });
        },

        updateTokenInitialized(isTokenInitialized: boolean) {
            this.$patch({
                isTokenInitialized,
            });
        },

        setAuthenticating(isAuthenticating: boolean) {
            this.$patch({
                isAuthenticating,
            });
        },

        setAuthError(error: string | null) {
            this.$patch({
                authError: error,
            });
        },

        resetAuth() {
            this.$patch({
                isTokenInitialized: false,
                isAuthenticating: false,
                authError: null,
            });
        },
    },
});
