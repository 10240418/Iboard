<script setup lang="ts">
import { ref } from 'vue';
import EditDialog from './components/PasswordEditDialog.vue';
import { SuperAdmin } from '@/model/superAdminModel';
import { useSuperAdminData } from './useSuperAdmin';
import { Modal } from 'ant-design-vue';

const { data, columns, pagination, searchKeyword, remove, fetch, list } = useSuperAdminData();

const isEditDialogVisible = ref(false);
const editDialogMode = ref<'create' | 'edit'>('create');
const selectedAdminData = ref<SuperAdmin | null>(null);

const showAddAdminDialog = () => {
  editDialogMode.value = 'create';
  selectedAdminData.value = null;
  isEditDialogVisible.value = true;
};

const editAdmin = async (admin: SuperAdmin) => {
  editDialogMode.value = 'edit';
  selectedAdminData.value = await fetch(admin.id!);
  isEditDialogVisible.value = true;
};

const deleteAdmin = async (id: number) => {
  await remove([id]);
  list();
};

const confirmDeleteAdmin = (id: number) => {
  Modal.confirm({
    title: 'Are you sure you want to delete this admin?',
    onOk: () => deleteAdmin(id),
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
  <EditDialog :visible="isEditDialogVisible" :mode="editDialogMode" :adminData="selectedAdminData"
    @update:visible="isEditDialogVisible = $event" @created="handleCreated" @updated="handleUpdated" />
  <div>
    <div style="display: flex; justify-content: space-between; margin-bottom: 16px;">
      <a-input v-model:value="searchKeyword" placeholder="Search by email" style="width: 33%;" />
      <a-button type="primary" @click="showAddAdminDialog">Add Admin</a-button>
    </div>
    <a-table :columns="columns" :data-source="data" :pagination="false">
      <template #headerCell="{ column }">
        <template v-if="column.key === 'name'">
        </template>
      </template>

      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'id'">
          <a>{{ record.id }}</a>
        </template>
        <template v-else-if="column.key === 'action'">
          <span>
            <a @click="editAdmin(record)">Set Password</a>
            <a-divider type="vertical" />
            <a style="color: lightcoral;" @click="confirmDeleteAdmin(record.id)">Delete</a>
          </span>
        </template>
      </template>
    </a-table>
    <a-pagination class="mt15" v-model:current="pagination.currentPage" v-model:pageSize="pagination.pageSize"
      :total="pagination.total" show-size-changer @change="onPageChange" />
  </div>
</template>

<style lang="sass"></style>