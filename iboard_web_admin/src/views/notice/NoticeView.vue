<script setup lang="ts">
import { ref } from 'vue';
import { useNoticeData } from './useNotice';
import { Notice } from '@/model/noticeModel';
import NoticeEditDialog from './components/NoticeEditDialog.vue';
import { Modal } from 'ant-design-vue';

const { data, columns, pagination, searchKeyword, isLoading, remove, fetch, list } = useNoticeData();

const isEditDialogVisible = ref(false);
const editDialogMode = ref<'create' | 'edit'>('create');
const selectedNoticeData = ref<Notice | undefined>(undefined);

const showAddNoticeDialog = () => {
  editDialogMode.value = 'create';
  selectedNoticeData.value = undefined;
  isEditDialogVisible.value = true;
};

const editNotice = async (notice: Notice) => {
  editDialogMode.value = 'edit';
  selectedNoticeData.value = await fetch(notice.id!);
  isEditDialogVisible.value = true;
};

const deleteNotice = async (id: number) => {
  await remove([id]);
  list();
};

const confirmDeleteNotice = (id: number) => {
  Modal.confirm({
    title: 'Are you sure you want to delete this notice?',
    onOk: () => deleteNotice(id),
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
  <NoticeEditDialog :visible="isEditDialogVisible" :mode="editDialogMode" :noticeData="selectedNoticeData"
    @update:visible="isEditDialogVisible = $event" @created="handleCreated" @updated="handleUpdated" />

  <div style="display: flex; justify-content: space-between; margin-bottom: 16px;">
    <a-input v-model:value="searchKeyword" placeholder="Search by title" style="width: 33%;" />
    <a-button type="primary" @click="showAddNoticeDialog">Add Notice</a-button>
  </div>

  <a-table :columns="columns" :data-source="data" :pagination="false" :loading="isLoading">
    <template #bodyCell="{ column, record }">
      <template v-if="column.key === 'id'">
        <a>{{ record.id }}</a>
      </template>
      <template v-if="column.key === 'type'">
        <a-tag>{{ record.type }}</a-tag>
      </template>
      <template v-else-if="column.key === 'action'">
        <span>
          <a @click="editNotice(record)">Edit</a>
          <a-divider type="vertical" />
          <a style="color: lightcoral;" @click="confirmDeleteNotice(record.id)">Delete</a>
        </span>
      </template>
    </template>
  </a-table>
  <a-pagination class="mt15" v-model:current="pagination.currentPage" v-model:pageSize="pagination.pageSize"
    :total="pagination.total" show-size-changer @change="onPageChange" />
</template>

<style lang="sass"></style>