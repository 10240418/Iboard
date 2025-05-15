<script setup lang="ts">
import { ref, defineProps, watch, defineEmits } from 'vue';
import { Building } from '@/model/buildingModel'; // Assuming Building interface includes optional location: string
import { useBuildingData } from '../useBuilding';
import { locationOptions } from '@/stores/locations'; // Import location options

const props = defineProps({
  visible: Boolean,
  mode: String, // 'create' or 'edit'
  buildingData: Object as () => Building | undefined,
});
const emit = defineEmits(['update:visible', 'cancel', 'created', 'updated']);

const { update, create } = useBuildingData();

const formRef = ref();
const formData = ref<Building>({
  id: undefined,
  name: '',
  ismartId: '',
  location: undefined, // Added location
  remark: '',
});

watch(() => props.buildingData, (newData) => {
  if (props.mode === 'edit' && newData) {
    formData.value = { ...newData };
  } else {
    // Reset form for create mode
    formData.value = {
      id: undefined,
      name: '',
      ismartId: '',
      location: undefined, // Added location
      remark: '',
    };
  }
}, { immediate: true });

const handleOk = () => {
  formRef.value.validate().then(() => {
    if (props.mode === 'edit') {
      update(formData.value).then((response) => {
        emit('update:visible', false);
        emit('updated', response);
      })
      return;
    } else {
      create(formData.value).then((response) => {
        emit('update:visible', false);
        emit('created', response as Building);
      })
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
    :title="props.mode === 'create' ? 'Add Building' : 'Edit Building'" :mask-closable="false">
    <a-form ref="formRef" :model="formData" :label-col="{ style: { width: '100px' } }">
      <a-form-item label="Name" name="name" :rules="[{ required: true, message: 'Please input name!' }]">
        <a-input v-model:value="formData.name" />
      </a-form-item>
      <a-form-item label="Ismart ID" name="ismartId" :rules="[{ required: true, message: 'Please input Ismart ID!' }]">
        <a-input v-model:value="formData.ismartId" />
      </a-form-item>
      <a-form-item label="Location" name="location">
        <a-select v-model:value="formData.location" :options="locationOptions" placeholder="Select location" />
      </a-form-item>
      <a-form-item label="Remark" name="remark">
        <a-textarea v-model:value="formData.remark" />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<style scoped>
/* Add any necessary styles here */
</style>
