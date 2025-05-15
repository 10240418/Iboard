<script setup lang="ts">
import { ref } from 'vue';
import { useDeviceData } from './useDevice';
import { Device } from '@/model/deviceModel';
import DeviceEditDialog from './components/DeviceEditDialog.vue';

const { data, columns, isLoading, pagination, searchKeyword, remove, fetch, list } = useDeviceData();

const isEditDialogVisible = ref(false);
const editDialogMode = ref<'create' | 'edit'>('create');
const selectedDevice = ref<Device | undefined>(undefined);

const showAddBuildingDialog = () => {
  editDialogMode.value = 'create';
  selectedDevice.value = undefined;
  isEditDialogVisible.value = true;
};

const handleEdit = async (device: Device) => {
  editDialogMode.value = 'edit';
  selectedDevice.value = await fetch(device.id!);
  isEditDialogVisible.value = true;
};

const handleDelete = async (id: number) => {
  await remove([id]);
  list();
};

const handleChange = () => {
  list();
};

list();
</script>

<template>
  <DeviceEditDialog :visible="isEditDialogVisible" :mode="editDialogMode" :device-data="selectedDevice"
    @update:visible="isEditDialogVisible = $event" @created="handleChange" @updated="handleChange" />
  <div>
    <div style="display: flex; justify-content: space-between; margin-bottom: 16px;">
      <a-input v-model:value="searchKeyword" placeholder="Search by name" style="width: 33%;" />
      <a-button type="primary" @click="showAddBuildingDialog">Add Device</a-button>
    </div>
    <a-table :columns="columns" :data-source="data" :loading="isLoading" :pagination="false">
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'id'">
          <a>{{ record.id }}</a>
        </template>
        <template v-else-if="column.key === 'building'">
          <a v-if="record.building">{{ record.building.name }}</a>
        </template>
        <template v-else-if="column.key === 'action'">
          <span>
            <a @click="handleEdit(record)">Edit</a>
            <a-divider type="vertical" />
            <a style="color: lightcoral;" @click="handleDelete(record.id)">Delete</a>
          </span>
        </template>
      </template>
    </a-table>
    <a-pagination class="mt15" v-model:current="pagination.currentPage" v-model:pageSize="pagination.pageSize"
      :total="pagination.total" show-size-changer @change="handleChange" />
  </div>
</template>
