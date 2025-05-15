import { ref } from 'vue';
import { message } from 'ant-design-vue';
import { getDevices, createDevice, updateDevice, deleteDevices, getDevice } from '@/httpapis/device';
import { Device } from '@/model/deviceModel';

export const useDeviceData = () => {
    const isLoading = ref(false);
    const data = ref<Device[]>([]);
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
            title: 'Device ID',
            dataIndex: 'deviceId',
            key: 'deviceId',
        },
        {
            title: 'Building',
            dataIndex: 'building',
            key: 'building',
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
            const response = await getDevices(query);
            data.value = response.data.data;
            pagination.value.total = response.data.pagination.total;
        } catch (error: any) {
            message.error(`Failed to list devices: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const fetch = async (id: number) => {
        isLoading.value = true;
        try {
            const response = await getDevice(id);
            return response.data.data;
        } catch (error: any) {
            message.error(`Failed to fetch device: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    }

    const create = async (deviceData: Device) => {
        isLoading.value = true;
        try {
            const res = await createDevice(deviceData);
            message.success('Device created successfully');
            return res.data.data;
        } catch (error: any) {
            message.error(`Failed to create device: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const update = async (deviceData: Device) => {
        isLoading.value = true;
        try {
            await updateDevice(deviceData);
            message.success('Device updated successfully');
        } catch (error: any) {
            message.error(`Failed to update device: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const remove = async (id: number[]) => {
        isLoading.value = true;
        try {
            await deleteDevices(id);
            message.success('Device deleted successfully');
        } catch (error: any) {
            message.error(`Failed to delete device: ${error.response.data.error}`);
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
