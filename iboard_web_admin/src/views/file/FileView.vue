<script setup lang="ts">
import { ref } from 'vue';
import { useFileData } from './useFile';
import { File } from '@/model/fileModel';
import FileDetailDialog from './components/FileDetailDialog.vue';
import { Modal } from 'ant-design-vue';

const { columns, pagination, searchKeyword, data, list, remove } = useFileData();

const selectedFileData = ref<File | null>(null);

const isDetailDialogVisible = ref(false);

const showDetailDialog = (record: File) => {
  selectedFileData.value = record;
  isDetailDialogVisible.value = true;
};

const handleDelete = (id: number) => {
  remove([id]).then(() => {
    list();
  });
};

const confirmDeleteFile = (id: number) => {
  Modal.confirm({
    title: 'Are you sure you want to delete this file?',
    onOk: () => handleDelete(id),
  });
};

const onPageChange = () => {
  list();
};

list();
</script>
<template>
  <div>
    <div style="display: flex; justify-content: space-between; margin-bottom: 16px;">
      <a-input v-model:value="searchKeyword" placeholder="Search by name" style="width: 33%;" />
    </div>
    <a-table :columns="columns" :data-source="data" :pagination="false">
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'id'">
          <a>{{ record.id }}</a>
        </template>
        <template v-else-if="column.key === 'path'">
          <a :href="record.path" class="url-column" target="_blank">
            {{ record.path }}
          </a>
        </template>
        <template v-else-if="column.key === 'action'">
          <span>
            <a @click="showDetailDialog(record)">Detail</a>
            <a-divider type="vertical" />
            <a style="color: lightcoral;" @click="confirmDeleteFile(record.id)">Delete</a>
          </span>
        </template>
      </template>
    </a-table>
    <a-pagination class="mt15" v-model:current="pagination.currentPage" v-model:pageSize="pagination.pageSize"
      :total="pagination.total" show-size-changer @change="onPageChange" />
    <FileDetailDialog :visible="isDetailDialogVisible" :fileData="selectedFileData"
      @update:visible="isDetailDialogVisible = $event" />
  </div>
</template>

<style lang="css">
.url-column {
  display: block;
  max-width: 160px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>