<script setup lang="ts">
import { ref, computed, watch, WatchHandle } from 'vue';
import { useBuildingData } from './useBuilding';
import { Building } from '@/model/buildingModel'; // Assuming Building interface includes optional location: string
import { DownOutlined } from '@ant-design/icons-vue';
import { Modal } from 'ant-design-vue';
import BuildingEditDialog from './components/BuildingEditDialog.vue';
import BuildingAdvBindingDialog from './components/BuildingAdvBindingDialog.vue';
import BuildingNoticeBindingDialog from './components/BuildingNoticeBindingDialog.vue';
import BuildingDetailDialog from './components/BuildingDetailDialog.vue';

const { data, columns, isLoading, remove, fetch, list, syncNotice } = useBuildingData();

const searchQuery = ref('');
const currentPage = ref(1);
const pageSize = ref(10);

const filteredData = computed(() => {
  return data.value.filter(item => item.name!.toLowerCase().includes(searchQuery.value.toLowerCase()));
});

const isEditDialogVisible = ref(false);
const editDialogMode = ref<'create' | 'edit'>('create');
const selectedBuildingData = ref<Building | undefined>(undefined);

const showAddBuildingDialog = () => {
  editDialogMode.value = 'create';
  selectedBuildingData.value = undefined;
  isEditDialogVisible.value = true;
};

const editBuilding = async (building: Building) => {
  editDialogMode.value = 'edit';
  selectedBuildingData.value = await fetch(building.id!);
  isEditDialogVisible.value = true;
};

const deleteBuilding = async (id: number) => {
  await remove([id]);
};

const confirmDeleteBuilding = (id: number) => {
  Modal.confirm({
    title: 'Are you sure you want to delete this building?',
    onOk: () => deleteBuilding(id).then(() => list()),
  });
};

const isAdvBindingDialogVisible = ref(false);
const bindAdvertisementsToBuilding = (building: Building) => {
  selectedBuildingData.value = building;
  isAdvBindingDialogVisible.value = true;
};

const isNoticeBindingDialogVisible = ref(false);
const bindNoticesToBuilding = (building: Building) => {
  selectedBuildingData.value = building;
  isNoticeBindingDialogVisible.value = true;
};

const onSyncNotice = (building: Building) => {
  Modal.confirm({
    title: 'Syncing notices will remove all existing notices, the process may take a few minutes. Are you sure you want to sync?',
    onOk: () => syncNotice(building.id!),
  });
};

const onCreate = (building: Building) => {
  bindAdvertisementsToBuilding(building);
  bindNoticesToBuilding(building);
  let w1, w2: WatchHandle;
  // wait until the binding dialog is closed
  w1 = watch(isAdvBindingDialogVisible, (value) => {
    if (!value && !isNoticeBindingDialogVisible.value) {
      list().then(() => {
        Modal.confirm({
          title: 'Do you want to sync notices from iSmart now?',
          onOk: () => onSyncNotice(building),
        });
      }).then(() => {
        w1!.stop();
        w2!.stop();
      });
    }
  });
  w2 = watch(isNoticeBindingDialogVisible, (value) => {
    if (!value && !isAdvBindingDialogVisible.value) {
      list().then(() => {
        Modal.confirm({
          title: 'Do you want to sync notices from iSmart now?',
          onOk: () => onSyncNotice(building),
        });
      }).then(() => {
        w1!.stop();
        w2!.stop();
      });;
    }
  });
};

const onPageChange = () => {
  list();
};

const handleUpdated = () => {
  list();
};

const isDetailDialogVisible = ref(false);

const showBuildingDetail = async (building: Building) => {
  selectedBuildingData.value = await fetch(building.id!);
  isDetailDialogVisible.value = true;
};

list();
</script>

<template>
  <BuildingEditDialog :visible="isEditDialogVisible" :mode="editDialogMode" :buildingData="selectedBuildingData"
    @update:visible="isEditDialogVisible = $event" @created="onCreate" @updated="handleUpdated" />
  <BuildingAdvBindingDialog :visible="isAdvBindingDialogVisible" :building="selectedBuildingData"
    @update:visible="isAdvBindingDialogVisible = $event" />
  <BuildingNoticeBindingDialog :visible="isNoticeBindingDialogVisible" :building="selectedBuildingData"
    @update:visible="isNoticeBindingDialogVisible = $event" />
  <BuildingDetailDialog :visible="isDetailDialogVisible" :building="selectedBuildingData"
    @update:visible="isDetailDialogVisible = $event" />
  <div>
    <div style="display: flex; justify-content: space-between; margin-bottom: 16px;">
      <a-input v-model:value="searchQuery" placeholder="Search by name" style="width: 33%;" />
      <a-button type="primary" @click="showAddBuildingDialog">Add Building</a-button>
    </div>
    <a-table :columns="columns" :data-source="filteredData" :pagination="{ current: currentPage, pageSize: pageSize }"
      @change="onPageChange" :loading="isLoading" rowKey="id"> {/* Added rowKey */}
      <template #headerCell="{ column }">
        <template v-if="column.key === 'name'">
          <span>
            Name
          </span>
        </template>
      </template>

      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'id'">
          <a>{{ record.id }}</a>
        </template>
        <template v-else-if="column.key === 'location'">
          <span>{{ record.location || '-' }}</span>
        </template>
        <template v-else-if="column.key === 'action'">
          <span>
            <a @click="showBuildingDetail(record)">Detail</a>
            <a-divider type="vertical" />
            <a @click="editBuilding(record)">Edit</a>
            <a-divider type="vertical" />
            <a-dropdown>
              <a class="ant-dropdown-link" @click.prevent>
                More
                <DownOutlined />
              </a>
              <template #overlay>
                <a-menu>
                  <a-menu-item @click="bindAdvertisementsToBuilding(record)">
                    <a style="color: #646cff;">Bind Ad.</a>
                  </a-menu-item>
                  <a-menu-item @click="bindNoticesToBuilding(record)">
                    <a style="color: #646cff;">Bind Notice</a>
                  </a-menu-item>
                  <a-menu-item @click="onSyncNotice(record)">
                    <a style="color: #646cff;">Sync</a>
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
            <a-divider type="vertical" />
            <a style="color: lightcoral;" @click="confirmDeleteBuilding(record.id)">Delete</a>
          </span>
        </template>
        <!-- Implicitly handles other columns like name, ismartId, remark if they are in the columns array -->
      </template>
    </a-table>
  </div>
</template>

<style lang="sass"></style>