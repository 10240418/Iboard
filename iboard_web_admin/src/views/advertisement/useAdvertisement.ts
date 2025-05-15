import { ref } from 'vue';
import { message } from 'ant-design-vue';
import { getAdvertisements, createAdvertisement, updateAdvertisement, deleteAdvertisements, getAdvertisement } from '@/httpapis/advertisement';
import { Advertisement } from '@/model/advertisementModel';

export const useAdvertisementData = () => {
    const isLoading = ref(false);
    const data = ref<Advertisement[]>([]);
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
            title: 'Title',
            dataIndex: 'title',
            key: 'title',
        },
        {
            title: 'Description',
            dataIndex: 'description',
            key: 'description',
        },
        {
            title: 'Type',
            dataIndex: 'type',
            key: 'type',
        },
        {
            title: 'Duration',
            dataIndex: 'duration',
            key: 'duration',
        },
        {
            title: 'Status',
            dataIndex: 'status',
            key: 'status',
        },
        {
            title: 'Action',
            key: 'action',
        }
    ];

    const list = async () => {
        isLoading.value = true;
        const query = {
            pageNum: pagination.value.currentPage,
            pageSize: pagination.value.pageSize,
            keyword: searchKeyword.value,
        };
        try {
            const response = await getAdvertisements(query);
            data.value = response.data.data;
            pagination.value.total = response.data.pagination.total;
        } catch (error: any) {
            message.error(`Failed to list advertisements: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const fetch = async (id: number) => {
        isLoading.value = true;
        try {
            const response = await getAdvertisement(id);
            return response.data.data;
        } catch (error: any) {
            message.error(`Failed to fetch advertisement: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    }

    const create = async (adData: Advertisement, fileUrl: string) => {
        isLoading.value = true;
        try {
            await createAdvertisement(adData, fileUrl);
            message.success('Advertisement created successfully');
        } catch (error: any) {
            message.error(`Failed to create advertisement: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const update = async (adData: Advertisement, fileUrl: string | undefined) => {
        isLoading.value = true;
        try {
            await updateAdvertisement(adData, fileUrl);
            message.success('Advertisement updated successfully');
        } catch (error: any) {
            message.error(`Failed to update advertisement: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const remove = async (id: number[]) => {
        isLoading.value = true;
        try {
            await deleteAdvertisements(id);
            message.success('Advertisement deleted successfully');
        } catch (error: any) {
            message.error(`Failed to delete advertisement: ${error.response.data.error}`);
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
