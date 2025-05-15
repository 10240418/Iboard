<script setup lang="ts">
import { ref, defineProps, watch, defineEmits } from 'vue';
import { SuperAdmin } from '@/model/superAdminModel';
import { useSuperAdminData } from '../useSuperAdmin';

const props = defineProps({
    visible: Boolean,
    mode: String, // 'create' or 'edit'
    adminData: Object as () => SuperAdmin | null,
});
const emit = defineEmits(['update:visible', 'cancel', 'created', 'updated']);

const { create, updatePassword } = useSuperAdminData();

const formRef = ref();
const formData = ref<SuperAdmin>({
    id: null,
    email: '',
    password: '',
});

watch(() => props.visible, (newData) => {
    if (props.mode === 'edit' && newData) {
        formData.value = props.adminData!;
    } else {
        formData.value = {
            id: null,
            email: '',
            password: '',
        };
    }
}, { immediate: true });

const handleOk = () => {
    formRef.value.validate().then(() => {
        if (props.mode === 'edit') {
            if (!props.adminData) return;
            if (typeof props.adminData.id !== 'number' || typeof props.adminData.password !== 'string') return;

            updatePassword({ id: props.adminData.id, newPassword: props.adminData.password }).then((response) => {
                emit('update:visible', false);
                emit('updated', response);
            }).catch((error) => {
                console.log('Admin update failed:', error);
            });
            return;
        } else {
            create(formData.value).then((response) => {
                emit('update:visible', false);
                emit('created', response);
            }).catch((error) => {
                console.log('Admin creation failed:', error);
            });
        }
    }).catch((e: object) => {
        console.log('Validation failed:', e);
    });
};

const handleCancel = () => {
    // Handle cancel action
    emit('update:visible', false);
    emit('cancel');
};

</script>

<template>
    <a-modal :open="props.visible" @ok="handleOk" @cancel="handleCancel" title="Set Password" :mask-closable="false">
        <a-form ref="formRef" :model="formData" :label-col="{ style: { width: '80px' } }">
            <a-form-item label="Email" name="email" :rules="[{ required: true, message: 'Please input email!' }]">
                <a-input v-model:value="formData.email" :disabled="props.mode === 'edit'" />
            </a-form-item>
            <a-form-item label="Password" name="password"
                :rules="[{ required: props.mode === 'create', message: 'Please input password!' }]">
                <a-input type="password" v-model:value="formData.password" />
            </a-form-item>
        </a-form>
    </a-modal>
</template>

<style scoped>
/* Add any necessary styles here */
</style>
