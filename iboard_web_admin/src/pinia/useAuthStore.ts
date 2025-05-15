import router from '@/router/router';
import { defineStore } from 'pinia';

export const useAuthStore = defineStore('auth', {
    state: () => (<{
        token: string | undefined | null;
        email: string | undefined;
    }>{
            token: undefined,
            email: undefined,
        }),
    getters: {
        isAuthenticated(state) {
            if (!state.token) {
                state.token = localStorage.getItem('admin/authenticationToken');
            }
            return !!state.token;
        },
        authenticationHeader(state) {
            if (!state.token) {
                state.token = localStorage.getItem('admin/authenticationToken');
            }
            return state.token ? `Bearer ${state.token}` : undefined;
        }
    },
    actions: {
        setToken(newToken: string) {
            this.token = newToken;
            localStorage.setItem('admin/authenticationToken', newToken);
        },
        setEmail(newEmail: string) {
            this.email = newEmail;
        },
        clearAuth() {
            this.token = undefined;
            this.email = undefined;
            localStorage.removeItem('admin/authenticationToken');
        },
        handleUnauthorized() {
            this.clearAuth();
            router.push({ name: 'Login' });
        },
    },
});
