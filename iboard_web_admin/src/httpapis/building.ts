import axiosInstance from './index';
import { Building } from '@/model/buildingModel';

export const getBuilding = (id: number) => {
    return axiosInstance.get(`/admin/building/${id}`);
};

export const getBuildings = (query?: object) => {
    return axiosInstance.get('/admin/building', { params: query });
};

export const createBuilding = (data: Building) => {
    return axiosInstance.post('/admin/building', data);
};

export const updateBuilding = (data: Building) => {
    return axiosInstance.put(`/admin/building`, data);
};

export const deleteBuildings = (ids: number[]) => {
    return axiosInstance.delete('/admin/building', { data: { ids } });
};

export const getBuildingsAdvertisement = (id: number) => {
    return axiosInstance.get(`/admin/advertisement_building/advertisements`, { params: { buildingId: id } });
}

export const getBuildingsNotice = (id: number) => {
    return axiosInstance.get(`/admin/notice_building/notices`, { params: { buildingId: id } });
}

export const syncBuildingNotice = (id: number) => {
    return axiosInstance.post(`/admin/building/${id}/sync_notice`, undefined, { timeout: 120 * 1000 });
}