<template>
  <a-modal :visible="visible" @cancel="handleCancel" :footer="null" :title="fileName">
    <template v-if="isImage">
      <img :src="fileUrl" alt="fileName" style="width: 100%;" />
    </template>
    <template v-else-if="isVideo">
      <video :src="fileUrl" controls style="width: 100%;"></video>
    </template>
    <template v-else-if="isPdf">
      <iframe :src="fileUrl" type="application/pdf" style="width: 100%; height: 700px;" />
      {{ fileUrl }}
    </template>
    {{ fileUrl }}
  </a-modal>
</template>

<script setup lang="ts">
import { defineProps, defineEmits, computed } from 'vue';

const props = defineProps({
  file: [String, File],
  fileName: String,
  visible: Boolean,
});

const emit = defineEmits(['update:visible']);

const handleCancel = () => {
  emit('update:visible', false);
};

const fileUrl = computed(() => {
  if (typeof props.file === 'string') {
    return props.file;
  } else if (props.file instanceof File) {
    return URL.createObjectURL(props.file);
  }
  return '';
});

const isImage = computed(() => {
  return ['jpg', 'jpeg', 'png', 'gif'].some(ext => props.fileName!.toLowerCase().endsWith(ext));
});

const isVideo = computed(() => {
  return ['mp4'].some(ext => props.fileName!.toLowerCase().endsWith(ext));
});

const isPdf = computed(() => {
  return props.fileName!.toLowerCase().endsWith('pdf');
});
</script>

<style scoped>
/* Add any necessary styles here */
</style>