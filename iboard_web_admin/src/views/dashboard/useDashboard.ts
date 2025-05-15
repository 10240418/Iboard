import { ref } from 'vue';

export function useDashboard() {
    const data = ref<string[]>([]);
    const loading = ref(false);
    const pagination = ref({
        currentPage: 1,
        pageSize: 10,
        total: 0,
    });
    const searchKeyword = ref('');

    const fetchData = async () => {
        loading.value = true;
        try {
            // Simulate an API call
            const response = await new Promise(resolve => setTimeout(() => resolve({
                data: ['item1', 'item2', 'item3'],
                pagination: { total: 30 }
            }), 1000));
            data.value = (response as { data: string[] }).data;
            pagination.value.total = (response as { pagination: { total: number } }).pagination.total;
        } catch (error) {
            console.error('Failed to fetch data', error);
        } finally {
            loading.value = false;
        }
    };

    return {
        data,
        loading,
        pagination,
        searchKeyword,
        fetchData,
    };
}
