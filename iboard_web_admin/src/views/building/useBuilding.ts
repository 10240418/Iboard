import { ref } from 'vue';
import { message } from 'ant-design-vue';
import { getBuildings, createBuilding, updateBuilding, deleteBuildings, getBuilding, syncBuildingNotice } from '@/httpapis/building';
import { bindAdvertisementsToBuilding, unbindAdvertisementFromBuilding } from '@/httpapis/advertisement';
import { bindNoticesToBuilding, unbindNoticeFromBuilding } from '@/httpapis/notice';
import { Building } from '@/model/buildingModel';

export const useBuildingData = () => {
    const isLoading = ref(false);
    const data = ref<Building[]>([]);
    const pagination = ref({
        currentPage: 1,
        pageSize: 10,
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
            title: 'Name',
            dataIndex: 'name',
            key: 'name',
        },
        {
            title: 'iSmart ID',
            dataIndex: 'ismartId',
            key: 'ismartId',
        },
        {
            title: 'Location',
            dataIndex: 'location',
            key: 'location',
        },
        {
            title: 'Remark',
            dataIndex: 'remark',
            key: 'remark',
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
            const response = await getBuildings(query);
            data.value = response.data.data;
        } catch (error: any) {
            message.error(`Failed to list buildings: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const fetch = async (id: number) => {
        isLoading.value = true;
        try {
            const response = await getBuilding(id);
            return response.data.data;
        } catch (error: any) {
            message.error(`Failed to fetch building: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    }

    const create = async (buildingData: Building) => {
        isLoading.value = true;
        try {
            const res = await createBuilding(buildingData);
            message.success('Building created successfully');
            return res.data.data
        } catch (error: any) {
            message.error(`Failed to create building: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const update = async (buildingData: Building) => {
        isLoading.value = true;
        try {
            await updateBuilding(buildingData);
            message.success('Building updated successfully');
        } catch (error: any) {
            message.error(`Failed to update building: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const remove = async (id: number[]) => {
        isLoading.value = true;
        try {
            await deleteBuildings(id);
            message.success('Building deleted successfully');
        } catch (error: any) {
            message.error(`Failed to delete building: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const bindAdvertisements = async (buildingId: number, advertisementIds: number[]) => {
        isLoading.value = true;
        try {
            await bindAdvertisementsToBuilding([buildingId], advertisementIds);
            message.success('Advertisements bound successfully');
        } catch (error: any) {
            message.error(`Failed to bind advertisements: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const unbindAdvertisements = async (buildingId: number, advertisementIds: number[]) => {
        isLoading.value = true;
        try {
            for (const advertisementId of advertisementIds) {
                await unbindAdvertisementFromBuilding([buildingId], advertisementId);
            }
            if (advertisementIds.length > 0) message.success('Advertisement unbound successfully');
        } catch (error: any) {
            message.error(`Failed to unbind advertisement: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    }

    const bindNotices = async (buildingId: number, noticeIds: number[]) => {
        isLoading.value = true;
        try {
            await bindNoticesToBuilding([buildingId], noticeIds);
            message.success('Notices bound successfully');
        } catch (error: any) {
            message.error(`Failed to bind notices: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    };

    const unbindNotices = async (buildingId: number, noticeIds: number[]) => {
        isLoading.value = true;
        try {
            for (const noticeId of noticeIds) {
                await unbindNoticeFromBuilding([buildingId], noticeId);
            }
            if (noticeIds.length > 0) message.success('Notices unbound successfully');
        } catch (error: any) {
            message.error(`Failed to unbind notices: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    }

    const syncNotice = async (id: number) => {
        isLoading.value = true;
        try {
            await syncBuildingNotice(id);
            message.success('Notices synced successfully');
        } catch (error: any) {
            message.error(`Failed to sync notices: ${error.response.data.error}`);
            return Promise.reject(error);
        } finally {
            isLoading.value = false;
        }
    }

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
        bindAdvertisements,
        unbindAdvertisements,
        bindNotices,
        unbindNotices,
        syncNotice,
    };
};
