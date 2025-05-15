import axios from 'axios';
import axiosInstance from './index';
import { File as FileModel } from '@/model/fileModel';


export const getFiles = (query?: object) => {
    return axiosInstance.get('/admin/file', { params: query });
};

export const deleteFile = (id: number) => {
    return axiosInstance.delete(`/admin/file/${id}`);
};

export const deleteFiles = (ids: number[]) => {
    return axiosInstance.delete('/admin/file', { data: { ids } });
};

export const getUploadPolicy = (data: { 'filename': string }) => {
    return axiosInstance.post('/admin/upload/params', data);
}

export const uploadFile = async (file: File): Promise<FileModel | undefined> => {
    // get policy from out server
    let res = await getUploadPolicy({ "filename": file.name }).catch(err => {
        return Promise.reject(err);
    });
    if (!res) Promise.reject('Failed to get upload policy');

    // make aliyun upload request
    const policy = res.data.data;
    const formData = new FormData();
    formData.append('key', policy.dir);
    formData.append('policy', policy.policy);
    formData.append('OSSAccessKeyId', policy.accessid);
    formData.append('success_action_status', '200');
    formData.append('callback', policy.callback);
    formData.append('signature', policy.signature);
    formData.append('file', file);

    // upload file directly to oss
    res = await axios.post(policy.host, formData).catch(err => {
        return Promise.reject(err);
    });
    if (!res || res.status !== 200) return Promise.reject('Failed to upload file');

    const fileUploaded = policy.host + '/' + policy.dir;
    return new FileModel({
        path: fileUploaded,
        id: undefined,
        md5: undefined,
        mimeType: file.type,
        size: file.size,
        ossRegion: policy.host,
        ossBucket: undefined,
        uploader: undefined,
        uploaderType: undefined,
    });
}