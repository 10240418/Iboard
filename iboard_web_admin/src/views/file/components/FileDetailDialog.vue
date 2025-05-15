<script setup lang="ts">
import { ref, defineProps, watch, defineEmits } from 'vue';
import { File } from '@/model/fileModel';

const props = defineProps({
  visible: Boolean,
  fileData: Object as () => File | null,
});
const emit = defineEmits(['update:visible', 'cancel']);

const formData = ref<File>(new File());

watch(() => props.fileData, (newData) => {
  if (newData) {
    formData.value = new File(newData);
  }
}, { immediate: true });

const handleCancel = () => {
  emit('update:visible', false);
  emit('cancel');
};
</script>

<template>
  <a-modal :open="props.visible" @cancel="handleCancel" title="File Details" :mask-closable="false">
    <a-form :model="formData" :label-col="{ style: { width: '100px' } }" disabled>
      <a-form-item label="ID" name="id">
        <a-input v-model:value="formData.id" disabled />
      </a-form-item>
      <a-form-item label="MD5" name="md5">
        <a-input v-model:value="formData.md5" disabled />
      </a-form-item>
      <a-form-item label="MIME Type" name="mimeType">
        <a-input v-model:value="formData.mimeType" disabled />
      </a-form-item>
      <a-form-item label="Size" name="size">
        <a-input-number v-model:value="formData.size" disabled />
      </a-form-item>
      <a-form-item label="URL" name="url">
        <a-input v-model:value="formData.path" disabled />
      </a-form-item>
      <a-form-item label="OSS Region" name="ossRegion">
        <a-input v-model:value="formData.ossRegion" disabled />
      </a-form-item>
      <a-form-item label="OSS Bucket" name="ossBucket">
        <a-input v-model:value="formData.ossBucket" disabled />
      </a-form-item>
      <a-form-item label="Uploader Type" name="uploaderType">
        <a-input v-model:value="formData.uploaderType" disabled />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<style scoped>
/* Add any necessary styles here */
</style>
