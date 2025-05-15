<script setup lang="ts">
import { ref, defineProps, watch, defineEmits } from 'vue';
import { Advertisement, AdvertisementStatus, AdvertisementTypes, DisplayTypes } from '@/model/advertisementModel';
import { uploadFile } from '@/httpapis/file';
import { message, Upload, UploadFile, UploadProps, type UploadChangeParam } from 'ant-design-vue';
import dayjs, { Dayjs } from 'dayjs';
import { useAdvertisementData } from '../useAdvertisement';
import FilePreview from '@/components/FilePreview.vue';

const props = defineProps({
  visible: Boolean,
  mode: String, // 'create' or 'edit'
  adData: Object as () => Advertisement | undefined,
});
const emit = defineEmits(['update:visible', 'created', 'updated', 'cancel']);

const { update, create } = useAdvertisementData();

const formRef = ref();
const formData = ref<Advertisement>(new Advertisement());
const fileList = ref<UploadFile[]>([]);
let fileChanged = false;
const dateRange = ref<[Dayjs, Dayjs]>([dayjs(), dayjs()]);

const previewVisible = ref(false);
const previewFile = ref<{ file: string | File | undefined, name: string }>({ file: '', name: '' });

const handlePreview = (file: UploadFile) => {
  previewFile.value.file = file.originFileObj || file.url || file.thumbUrl;
  previewFile.value.name = file.name;
  previewVisible.value = true;
};

// Initialize form data when dialog is opened
watch(() => props.visible, () => {
  if (!props.visible) {
    return
  }
  fileChanged = false;

  if (props.mode === 'edit' && props.adData) {
    formData.value = new Advertisement({ ...props.adData })
    formData.value.startTime = dayjs(props.adData.startTime);
    formData.value.endTime = dayjs(props.adData.endTime);
    dateRange.value = [formData.value.startTime, formData.value.endTime];
    const fname = formData.value.file?.path!.split('?')[0].split('/').pop();
    fileList.value = [{
      uid: '-1',
      name: fname!,
      status: 'done',
      url: formData.value.file?.path,
    }];
  } else if (props.mode === 'create') {
    formData.value = new Advertisement({ status: AdvertisementStatus.Active });
    fileList.value = [];
    dateRange.value = [dayjs(), dayjs()];
  }
}, { immediate: true });


const handleOk = () => {
  formRef.value.validate().then(() => {
    formData.value.file = undefined;
    formData.value.startTime = dateRange.value[0];
    formData.value.endTime = dateRange.value[1];
    if (props.mode === 'edit') {
      if (fileChanged) {
        uploadFile(fileList.value[0].originFileObj!).then((file) => {
          update(formData.value, file!.path!).then((response) => {
            emit('update:visible', false);
            emit('updated', response);
          })
        }).catch((error) => {
          message.error(`Failed to upload file: ${error}`);
        });
      } else {
        update(formData.value, undefined).then((response) => {
          emit('update:visible', false);
          emit('updated', response);
        })
      }
    } else if (props.mode === 'create') {
      uploadFile(fileList.value[0].originFileObj!).then((file) => {
        create(formData.value, file!.path!).then((response) => {
          emit('update:visible', false);
          emit('created', response);
        })
      }).catch((error) => {
        message.error(`Failed to upload file: ${error}`);
      });
    }
  })
};

const handleCancel = () => {
  emit('update:visible', false);
  emit('cancel');
};

const handleChange = ({ file /* fileList */ }: UploadChangeParam) => {
  if (!file) return;
  if (file!.type!.startsWith('image/')) {
    formData.value.type = AdvertisementTypes.Image;
  } else if (file.type === 'video/mp4') {
    formData.value.type = AdvertisementTypes.Video;
  }
  fileChanged = true;
}

const beforeUpload: UploadProps['beforeUpload'] = file => {
  const isAllowedType = ['image/png', 'image/jpeg', 'image/gif', 'video/mp4'].includes(file.type);
  if (!isAllowedType) {
    message.error(`${file.name} is not a supported file type`);
  }
  return isAllowedType ? false : (isAllowedType || Upload.LIST_IGNORE);
};

const validateFile = (_rule: never, _value: never, callback: (err?: Error) => Promise<Error>) => {
  if (props.mode === 'edit' && !fileChanged) {
    callback();
  } else if (fileList.value.length === 0) {
    callback(new Error('Please upload a file!'));
  } else {
    const file = fileList.value[0];
    const isAllowedType = ['image/png', 'image/jpeg', 'image/gif', 'video/mp4'].includes(file.type!.toString());
    if (!isAllowedType) {
      callback(new Error('Unsupported file type!'));
    } else {
      callback();
    }
  }
};

const validateDataRange = (_rule: never, _value: never, callback: (err?: Error) => Promise<Error>) => {
  console.log(dateRange.value)
  if (!dateRange.value[0] || !dateRange.value[1]) {
    callback(new Error('Please select start and end date!'));
  } else if (dateRange.value[0].isAfter(dateRange.value[1])) {
    callback(new Error('End date must be after start date!'));
  } else {
    callback();
  }
};

</script>

<template>
  <a-modal :open="props.visible" @ok="handleOk" @cancel="handleCancel"
    :title="props.mode === 'create' ? 'Add Advertisement' : 'Edit Advertisement'" :mask-closable="false">
    <a-form ref="formRef" :model="formData" :label-col="{ style: { width: '100px' } }">
      <a-form-item label="Title" name="title" :rules="[{ required: true, message: 'Please input title!' }]">
        <a-input v-model:value="formData.title" />
      </a-form-item>

      <a-form-item label="Description" name="description"
        :rules="[{ required: true, message: 'Please input description!' }]">
        <a-textarea v-model:value="formData.description" />
      </a-form-item>

      <a-form-item label="Display" name="display" :rules="[{ required: true, message: 'Please select display!' }]">
        <a-select v-model:value="formData.display">
          <a-select-option v-for="display in DisplayTypes" :key="display" :value="display">
            {{ display }}
          </a-select-option>
        </a-select>
      </a-form-item>

      <a-form-item label="Date Range" name="endDate" :rules="[{ validator: validateDataRange }]">
        <a-range-picker v-model:value="dateRange" />
      </a-form-item>

      <a-form-item label="Status" name="status" :rules="[{ required: true, message: 'Please select status!' }]">
        <a-select v-model:value="formData.status">
          <a-select-option v-for="status in AdvertisementStatus" :key="status" :value="status">
            {{ status }}
          </a-select-option>
        </a-select>
      </a-form-item>

      <a-form-item label="File" name="file" :rules="[{ validator: validateFile, required: true }]">
        <a-upload-dragger v-model:fileList="fileList" name="file" @change="handleChange" :max-count="1"
          :before-upload="beforeUpload">
          <p class="ant-upload-text">Click or drag file to this area</p>
          <p class="ant-upload-hint">
            Support format are mp4, jpg, png, gif.
          </p>

          <template #itemRender="{ file, actions }">
            <a-space>
              {{ file.name }}
              <a :style="file.status === 'error' ? 'color: red' : ''" @click="handlePreview(file)">Preview</a>
              <a style="color: lightcoral;" href="javascript:;" @click="actions.remove">Delete</a>
            </a-space>
          </template>
        </a-upload-dragger>
      </a-form-item>

      <a-form-item label="Type" name="type" :rules="[{ required: true, message: 'Please select type!' }]">
        <a-select v-model:value="formData.type">
          <a-select-option v-for="type in AdvertisementTypes" :key="type" :value="type">{{ type
            }}</a-select-option>
        </a-select>
      </a-form-item>

      <a-form-item label="Duration(s)" name="duration" :rules="[{ required: true, message: 'Please input duration!' }]">
        <a-input-number v-model:value="formData.duration" />
      </a-form-item>

      <a-form-item label="Public" name="isPublic">
        <a-switch v-model:checked="formData.isPublic" />
      </a-form-item>
    </a-form>
  </a-modal>
  <FilePreview :file="previewFile.file" :fileName="previewFile.name" v-model:visible="previewVisible" />
</template>

<style scoped>
/* Add any necessary styles here */
</style>
