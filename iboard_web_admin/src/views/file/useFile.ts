import { ref } from 'vue';
import { message } from 'ant-design-vue';
import { getFiles, deleteFiles } from '@/httpapis/file';
import { File } from '@/model/fileModel';

export const useFileData = () => {
    const isLoading = ref(false);
    const data = ref<File[]>([]);
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
            title: 'MIME Type',
            dataIndex: 'mimeType',
            key: 'mimeType',
        },
        {
            title: 'Path',
            dataIndex: 'path',
            key: 'path',
        },
        {
            title: 'Uploader',
            dataIndex: 'uploader',
            key: 'uploader',
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
            const response = await getFiles(query);
            data.value = response.data.data;
            pagination.value.total = response.data.pagination.total;
        } catch (error: any) {
            message.error(`Failed to list files: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const remove = async (id: number[]) => {
        isLoading.value = true;
        try {
            await deleteFiles(id);
            message.success('File deleted successfully');
        } catch (error: any) {
            message.error(`Failed to delete file: ${error.response.data.error}`);
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
        remove,
    };
};
