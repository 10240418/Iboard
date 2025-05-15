<script setup lang="ts">
import { onMounted } from 'vue';
import { useDashboard } from './useDashboard';

const { data, pagination, loading, fetchData } = useDashboard();

onMounted(() => {
    fetchData();
});

const onPageChange = (page: number, pageSize: number) => {
    pagination.value.currentPage = page;
    pagination.value.pageSize = pageSize;
    fetchData();
};
</script>

<template>
    <div>
        <div v-if="loading">Loading...</div>
        <ul v-else>
            <li v-for="item in data" :key="item">{{ item }}</li>
        </ul>
        <a-table :data-source="data" :pagination="false">
            <!-- ...existing code... -->
        </a-table>
        <a-pagination class="mt15" v-model:current="pagination.currentPage" v-model:pageSize="pagination.pageSize"
            :total="pagination.total" show-size-changer @change="onPageChange" />
    </div>
</template>

<style lang="sass">
/* Add your styles here */
</style>