import axiosInstance from './index';
import { Advertisement } from '@/model/advertisementModel';

export const getAdvertisement = (id: number) => {
    return axiosInstance.get(`/admin/advertisement/${id}`);
};

export const getAdvertisements = (query?: object) => {
    return axiosInstance.get('/admin/advertisement', { params: query });
};

export const createAdvertisement = (data: Advertisement, fileUrl: string) => {
    return axiosInstance.post('/admin/advertisement', { path: fileUrl, ...data });
};

export const updateAdvertisement = (data: Advertisement, fileUrl: string | undefined) => {
    return axiosInstance.put(`/admin/advertisement`, { path: fileUrl, ...data });
};

export const deleteAdvertisements = (ids: number[]) => {
    return axiosInstance.delete('/admin/advertisement', { data: { ids } });
};

export const bindAdvertisementsToBuilding = (buildingIds: number[], advertisementIds: number[]) => {
    return axiosInstance.post(`/admin/advertisement_building/bind`, { advertisementIds, buildingIds });
}

export const unbindAdvertisementFromBuilding = (buildingIds: number[], advertisementId: number) => {
    return axiosInstance.post(`/admin/advertisement_building/unbind`, { advertisementId, buildingIds });
}