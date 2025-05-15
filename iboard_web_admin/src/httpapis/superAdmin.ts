import { SuperAdmin } from '@/model/superAdminModel';
import axiosInstance from './index';

export const getSuperAdmin = (id: number) => {
    return axiosInstance.get(`/admin/super_admin/${id}`);
};

export const getSuperAdmins = (query?: object) => {
    return axiosInstance.get('/admin/super_admin', { params: query });
};

export const createSuperAdmin = (data: SuperAdmin) => {
    return axiosInstance.post('/admin/super_admin', data);
};

export const updateSuperAdminPassword = (data: { newPassword: string, id: number }) => {
    return axiosInstance.post(`/admin/super_admin/update_password`, data);
};

export const deleteSuperAdmins = (ids: number[]) => {
    return axiosInstance.delete('/admin/super_admin', { data: { ids } });
};
