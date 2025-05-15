import DashboardView from '@/views/dashboard/DashboardView.vue'
import DeviceView from '@/views/device/DeviceView.vue'
import BuildingAdminView from '@/views/buildingadmin/BuildingAdminView.vue'
import SuperAdminView from '@/views/superadmin/SuperAdminView.vue'
import NoticeView from '@/views/notice/NoticeView.vue'
import FileView from '@/views/file/FileView.vue'
import BuildingView from '@/views/building/BuildingView.vue'
import APIView from '@/views/api/APIView.vue'
import AdvertisementView from '@/views/advertisement/AdvertisementView.vue'
import LoginView from '@/views/auth/LoginView.vue'

const routes = [
    {
        name: 'Dashboard',
        path: '/dashboard',
        component: DashboardView,
        meta: { requiresAuth: true }
    },
    {
        name: 'Device',
        path: '/device',
        component: DeviceView,
        meta: { requiresAuth: true }
    },
    {
        name: 'BuildingAdmin',
        path: '/building-admin',
        component: BuildingAdminView,
        meta: { requiresAuth: true }
    },
    {
        name: 'SuperAdmin',
        path: '/super-admin',
        component: SuperAdminView,
        meta: { requiresAuth: true }
    },
    {
        name: 'Notice',
        path: '/notice',
        component: NoticeView,
        meta: { requiresAuth: true }
    },
    {
        name: 'File',
        path: '/file',
        component: FileView,
        meta: { requiresAuth: true }
    },
    {
        name: 'Building',
        path: '/building',
        component: BuildingView,
        meta: { requiresAuth: true }
    },
    {
        name: 'API',
        path: '/api',
        component: APIView,
        meta: { requiresAuth: true }
    },
    {
        name: 'Advertisement',
        path: '/advertisement',
        component: AdvertisementView,
        meta: { requiresAuth: true }
    },
    {
        name: 'Login',
        path: '/login',
        component: LoginView,
        meta: { requiresAuth: false }
    },
]

export default routes