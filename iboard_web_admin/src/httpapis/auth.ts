import axiosInstance from './index'
import { LoginModel } from '@/model/authModel'

export const login = (data: LoginModel) => {
    return axiosInstance.post('/admin/login', data)
}

export const logout = () => {
    return axiosInstance.post('/admin/logout')
}

export const getCurrentUser = () => {
    return axiosInstance.get('/admin/current_user')
}
