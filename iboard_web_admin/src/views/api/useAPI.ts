import { ref } from 'vue';
import { message } from 'ant-design-vue';
import { getApis } from '@/httpapis/api';
import { Api } from '@/model/apiModel';

export const useAPIData = () => {
    const isLoading = ref(false);
    const data = ref<Api[]>([]);
    const pagination = ref({
        currentPage: 1,
        pageSize: 10,
    });
    const searchKeyword = ref('');

    const columns = [
        {
            title: 'Name',
            dataIndex: 'name',
            key: 'name',
        },
        {
            title: 'Endpoint',
            dataIndex: 'endpoint',
            key: 'endpoint',
        },
        {
            title: 'Method',
            dataIndex: 'method',
            key: 'method',
        },
        {
            title: 'Action',
            key: 'action',
        },
    ];

    const list = async () => {
        isLoading.value = true;
        const query = {
            page: pagination.value.currentPage,
            size: pagination.value.pageSize,
            keyword: searchKeyword.value,
        };
        try {
            const response = await getApis(query);
            data.value = response.data;
        } catch (error: unknown) {
            message.error(`Failed to list APIs: ${error}`);
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
    };
};
