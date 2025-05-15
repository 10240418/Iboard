import { ref } from 'vue';
import { message } from 'ant-design-vue';
import { getSuperAdmins, createSuperAdmin, deleteSuperAdmins, getSuperAdmin, updateSuperAdminPassword } from '@/httpapis/superAdmin';
import { SuperAdmin } from '@/model/superAdminModel';

export const useSuperAdminData = () => {
    const isLoading = ref(false);
    const data = ref([]);
    const pagination = ref({
        currentPage: 1,
        pageSize: 10,
        total: 0,
    });
    const searchKeyword = ref('');

    const columns = [
        {
            title: 'ID',
            dataIndex: 'id',
            key: 'id',
            width: 40,
        },
        {
            title: 'Email',
            dataIndex: 'email',
            key: 'email',
        },
        {
            title: 'Action',
            key: 'action',
        },
    ];

    const list = async () => {
        isLoading.value = true;
        const query = {
            pageNum: pagination.value.currentPage,
            pageSize: pagination.value.pageSize,
            keyword: searchKeyword.value,
        };
        try {
            const response = await getSuperAdmins(query);
            data.value = response.data.data;
            pagination.value.total = response.data.pagination.total; // Add this line
        } catch (error: any) {
            message.error(`Failed to list super admins: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const fetch = async (id: number) => {
        isLoading.value = true;
        try {
            const response = await getSuperAdmin(id);
            return response.data.data;
        } catch (error: any) {
            message.error(`Failed to fetch super admins: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    }

    const create = async (adminData: SuperAdmin) => {
        isLoading.value = true;
        try {
            await createSuperAdmin(adminData);
            message.success('Super admin created successfully');
        } catch (error: any) {
            message.error(`Failed to create super admin: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const updatePassword = async (data: { newPassword: string, id: number }) => {
        isLoading.value = true;
        try {
            await updateSuperAdminPassword(data);
            message.success('Super admin updated successfully');
        } catch (error: any) {
            message.error(`Failed to update super admin: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const remove = async (id: number[]) => {
        isLoading.value = true;
        try {
            await deleteSuperAdmins(id);
            message.success('Super admin deleted successfully');
        } catch (error: any) {
            message.error(`Failed to delete super admin: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    return {
        isLoading,
        data,
        columns,
        pagination,
        searchKeyword,
        list,
        fetch,
        create,
        updatePassword,
        remove,
    };
};