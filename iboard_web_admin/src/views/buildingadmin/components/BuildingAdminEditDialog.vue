<script setup lang="ts">
import { ref, defineProps, watch, defineEmits } from 'vue';
import { BuildingAdmin, BuildingAdminStatus } from '@/model/buildingAdminModel';
import { createBuildingAdmin, updateBuildingAdmin } from '@/httpapis/buildingAdmin';

const props = defineProps({
    visible: Boolean,
    mode: String, // 'create' or 'edit'
    adminData: Object as () => BuildingAdmin | null,
});
const emit = defineEmits(['update:visible', 'created', 'updated', 'cancel']);

const formRef = ref();
const formData = ref<BuildingAdmin>({
    id: null,
    email: '',
    password: '',
    status: '',
});

watch(() => props.visible, () => {
    if (!props.visible) {
        return
    }

    if (props.mode === 'edit' && props.adminData) {
        formData.value = { ...props.adminData };
    } else {
        formData.value = {
            id: null,
            email: '',
            password: '',
            status: '',
        };
    }
}, { immediate: true });

const handleOk = () => {
    formRef.value.validate().then(() => {
        if (props.mode === 'create') {
            createBuildingAdmin(formData.value).then((response) => {
                emit('update:visible', false);
                emit('created', response);
            }).catch((error) => {
                console.log('Admin creation failed:', error);
            });
        } else {
            updateBuildingAdmin(formData.value).then((response) => {
                emit('update:visible', false);
                emit('updated', response);
            }).catch((error) => {
                console.log('Admin update failed:', error);
            });
        }
    }).catch((e: object) => {
        console.log('Validation failed:', e);
    });
};

const handleCancel = () => {
    emit('update:visible', false);
    emit('cancel');
};
</script>

<template>
    <a-modal :open="props.visible" @ok="handleOk" @cancel="handleCancel"
        :title="props.mode === 'create' ? 'Add Admin' : 'Edit Admin'" :mask-closable="false">
        <a-form ref="formRef" :model="formData" :label-col="{ style: { width: '80px' } }">
            <a-form-item label="Email" name="email" :rules="[{ required: true, message: 'Please input email!' }]">
                <a-input v-model:value="formData.email" />
            </a-form-item>
            <a-form-item label="Password" name="password"
                :rules="[{ required: props.mode === 'create', message: 'Please input password!' }]">
                <a-input type="password" v-model:value="formData.password" />
            </a-form-item>
            <a-form-item label="Status" name="status" :rules="[{ required: true, message: 'Please input status!' }]">
                <a-select v-model:value="formData.status">
                    <a-select-option v-for="status in BuildingAdminStatus" :key="status" :value="status">{{ status
                        }}</a-select-option>
                </a-select>
            </a-form-item>
        </a-form>
    </a-modal>
</template>

<style scoped>
/* Add any necessary styles here */
</style>
