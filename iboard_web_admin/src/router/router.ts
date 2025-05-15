import { createRouter, createWebHistory } from "vue-router";
import routes from "./routes";
import { useAuthStore } from '@/pinia/useAuthStore';

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes,
})

// Add a global navigation guard
router.beforeEach((to, _from, next) => {
    const authStore = useAuthStore();
    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
        next({ name: 'Login' });
    } else {
        next();
    }
});

export default router