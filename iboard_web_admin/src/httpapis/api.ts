import axiosInstance from './index';
import { Api } from '@/model/apiModel';

export const getApi = (id: number) => {
    return axiosInstance.get(`/api/${id}`);
};

export const getApis = (query?: object) => {
    return axiosInstance.get('/api', { params: query });
};

export const createApi = (data: Api) => {
    return axiosInstance.post('/api', data);
};

export const updateApi = (id: number, data: Api) => {
    return axiosInstance.put(`/api/${id}`, data);
};

export const deleteApi = (id: number) => {
    return axiosInstance.delete(`/api/${id}`);
};

export const deleteApis = (ids: string[]) => {
    return axiosInstance.delete('/api', { data: { ids } });
};
