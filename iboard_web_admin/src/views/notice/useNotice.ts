import { ref } from 'vue';
import { message } from 'ant-design-vue';
import { getNotices, createNotice, updateNotice, deleteNotices, getNotice } from '@/httpapis/notice';
import { Notice } from '@/model/noticeModel';

export const useNoticeData = () => {
    const isLoading = ref(false);
    const data = ref<Notice[]>([]);
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
            const response = await getNotices(query);
            data.value = response.data.data;
            pagination.value.total = response.data.pagination.total;
        } catch (error: any) {
            message.error(`Failed to list notices: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const fetch = async (id: number) => {
        isLoading.value = true;
        try {
            const response = await getNotice(id);
            return response.data.data;
        } catch (error: any) {
            message.error(`Failed to fetch notice: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    }

    const create = async (noticeData: Notice, fileUrl: string) => {
        isLoading.value = true;
        try {
            // Handle file upload if necessary
            if (noticeData.file) {
                // Upload file logic here
            }
            await createNotice(noticeData, fileUrl);
            message.success('Notice created successfully');
        } catch (error: any) {
            message.error(`Failed to create notice: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const update = async (noticeData: Notice, fileUrl: string | undefined) => {
        isLoading.value = true;
        try {
            // Handle file upload if necessary
            if (noticeData.file) {
                // Upload file logic here
            }
            await updateNotice(noticeData, fileUrl);
            message.success('Notice updated successfully');
        } catch (error: any) {
            message.error(`Failed to update notice: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const remove = async (id: number[]) => {
        isLoading.value = true;
        try {
            await deleteNotices(id);
            message.success('Notice deleted successfully');
        } catch (error: any) {
            message.error(`Failed to delete notice: ${error.response.data.error}`);
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
