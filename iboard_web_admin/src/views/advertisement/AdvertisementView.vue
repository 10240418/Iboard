<script setup lang="ts">
import { ref } from 'vue';
import { useAdvertisementData } from './useAdvertisement';
import { Advertisement } from '@/model/advertisementModel';
import AdvertisementEditDialog from './components/AdvertisementEditDialog.vue';
import { Modal } from 'ant-design-vue';

const { data, columns, pagination, searchKeyword, remove, fetch, list } = useAdvertisementData();

const isEditDialogVisible = ref(false);
const editDialogMode = ref<'create' | 'edit'>('create');
const selectedAdData = ref<Advertisement | undefined>(undefined);

const showAddAdDialog = () => {
  editDialogMode.value = 'create';
  selectedAdData.value = undefined;
  isEditDialogVisible.value = true;
};

const editAd = async (ad: Advertisement) => {
  editDialogMode.value = 'edit';
  selectedAdData.value = await fetch(ad.id!);
  isEditDialogVisible.value = true;
};

const deleteAd = async (id: number) => {
  await remove([id]);
  list();
};

const confirmDeleteAd = (id: number) => {
  Modal.confirm({
    title: 'Are you sure you want to delete this advertisement?',
    onOk: () => deleteAd(id),
  });
};

const onPageChange = () => {
  list();
};

const handleCreated = () => {
  list();
};

const handleUpdated = () => {
  list();
};

list();
</script>

<template>
  <div>
    <AdvertisementEditDialog :visible="isEditDialogVisible" :mode="editDialogMode" :adData="selectedAdData"
      @update:visible="isEditDialogVisible = $event" @created="handleCreated" @updated="handleUpdated" />

    <div style="display: flex; justify-content: space-between; margin-bottom: 16px;">
      <a-input v-model:value="searchKeyword" placeholder="Search by title" style="width: 33%;" />
      <a-button type="primary" @click="showAddAdDialog">Add Advertisement</a-button>
    </div>

    <a-table :columns="columns" :data-source="data" :pagination="false">
      <template #headerCell="{ column }">
        <template v-if="column.key === 'title'">
          <span>
            Title
          </span>
        </template>
      </template>

      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'id'">
          <a>{{ record.id }}</a>
        </template>
        <template v-if="column.key === 'type'">
          <a-tag>{{ record.type }}</a-tag>
        </template>
        <template v-else-if="column.key === 'action'">
          <span>
            <a @click="editAd(record)">Edit</a>
            <a-divider type="vertical" />
            <a style="color: lightcoral;" @click="confirmDeleteAd(record.id)">Delete</a>
          </span>
        </template>
      </template>
    </a-table>
    <a-pagination class="mt15" v-model:current="pagination.currentPage" v-model:pageSize="pagination.pageSize"
      :total="pagination.total" show-size-changer @change="onPageChange" />
  </div>
</template>

<style lang="sass"></style>