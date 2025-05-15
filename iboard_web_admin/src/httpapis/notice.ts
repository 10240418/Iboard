import axiosInstance from './index';
import { Notice } from '@/model/noticeModel';

export const getNotice = (id: number) => {
    return axiosInstance.get(`/admin/notice/${id}`);
};

export const getNotices = (query?: object) => {
    return axiosInstance.get('/admin/notice', { params: query });
};

export const createNotice = (data: Notice, fileUrl: string) => {
    return axiosInstance.post('/admin/notice', { path: fileUrl, ...data });
};

export const updateNotice = (data: Notice, fileUrl: string | undefined) => {
    return axiosInstance.put(`/admin/notice`, { path: fileUrl, ...data });
};

export const deleteNotices = (ids: number[]) => {
    return axiosInstance.delete('/admin/notice', { data: { ids } });
};

export const bindNoticesToBuilding = (buildingIds: number[], noticeIds: number[]) => {
    return axiosInstance.post(`/admin/notice_building/bind`, { noticeIds, buildingIds });
}

export const unbindNoticeFromBuilding = (buildingIds: number[], noticeId: number) => {
    return axiosInstance.post(`/admin/notice_building/unbind`, { noticeId, buildingIds });
}
