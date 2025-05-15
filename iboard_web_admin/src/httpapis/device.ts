import axiosInstance from './index';
import { Device } from '@/model/deviceModel';

export const getDevice = (id: number) => {
    return axiosInstance.get(`/admin/device/${id}`);
};

export const getDevices = (query?: object) => {
    return axiosInstance.get('/admin/device', { params: query });
};

export const createDevice = (data: Device) => {
    return axiosInstance.post('/admin/device', data);
};

export const updateDevice = (data: Device) => {
    return axiosInstance.put(`/admin/device`, data);
};

export const deleteDevices = (ids: number[]) => {
    return axiosInstance.delete('/admin/device', { data: { ids } });
};
