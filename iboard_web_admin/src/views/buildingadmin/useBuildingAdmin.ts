import { ref } from 'vue';
import { message } from 'ant-design-vue';
import { getBuildingAdmins, createBuildingAdmin, updateBuildingAdmin, deleteBuildingAdmins, getBuildingAdmin } from '@/httpapis/buildingAdmin';
import { BuildingAdmin } from '@/model/buildingAdminModel';

export const useBuildingAdminData = () => {
    const isLoading = ref(false);
    const data = ref<BuildingAdmin[]>([]);
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
            title: 'Status',
            dataIndex: 'status',
            key: 'status',
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
            const response = await getBuildingAdmins(query);
            data.value = response.data.data;
            pagination.value.total = response.data.pagination.total;
        } catch (error: any) {
            message.error(`Failed to list building admins: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const fetch = async (id: number) => {
        isLoading.value = true;
        try {
            const response = await getBuildingAdmin(id);
            return response.data.data;
        } catch (error: any) {
            message.error(`Failed to fetch building admin: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    }

    const create = async (adminData: BuildingAdmin) => {
        isLoading.value = true;
        try {
            await createBuildingAdmin(adminData);
            message.success('Building admin created successfully');
        } catch (error: any) {
            message.error(`Failed to create building admin: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const update = async (adminData: BuildingAdmin) => {
        isLoading.value = true;
        try {
            await updateBuildingAdmin(adminData);
            message.success('Building admin updated successfully');
        } catch (error: any) {
            message.error(`Failed to update building admin: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const remove = async (id: number[]) => {
        isLoading.value = true;
        try {
            await deleteBuildingAdmins(id);
            message.success('Building admin deleted successfully');
        } catch (error: any) {
            message.error(`Failed to delete building admin: ${error.response.data.error}`);
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
        update,
        remove,
    };
};
