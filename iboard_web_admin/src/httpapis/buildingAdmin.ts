import axiosInstance from './index';
import { BuildingAdmin } from '@/model/buildingAdminModel';

export const getBuildingAdmin = (id: number) => {
    return axiosInstance.get(`/admin/building_admin/${id}`);
};

export const getBuildingAdmins = (query?: object) => {
    return axiosInstance.get('/admin/building_admin', { params: query });
};

export const createBuildingAdmin = (data: BuildingAdmin) => {
    return axiosInstance.post('/admin/building_admin', data);
};

export const updateBuildingAdmin = (data: BuildingAdmin) => {
    return axiosInstance.put(`/admin/building_admin`, data);
};

export const deleteBuildingAdmins = (ids: number[]) => {
    return axiosInstance.delete('/admin/building_admin', { data: { ids } });
};
