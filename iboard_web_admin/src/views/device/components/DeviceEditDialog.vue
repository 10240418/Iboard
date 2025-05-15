<script setup lang="ts">
import { ref, defineProps, watch, defineEmits } from 'vue';
import { Device } from '@/model/deviceModel';
import { useDeviceData } from '../useDevice';
import { getBuildings } from '@/httpapis/building';
import { Building } from '@/model/buildingModel';

const props = defineProps({
  visible: Boolean,
  mode: String,
  deviceData: Object as () => Device | undefined,
});
const emit = defineEmits(['update:visible', 'cancel', 'created', 'updated']);

const { update, create } = useDeviceData();
const buildings = ref<Building[]>([]);

const formRef = ref();
const formData = ref<Device>(new Device());

watch(() => props.visible, async (newVisible) => {
  if (!newVisible) {
    return;
  }

  if (props.mode === 'edit' && props.deviceData) {
    formData.value = new Device({ ...props.deviceData });
  } else {
    formData.value = new Device();
  }

  getBuildings().then((response) => {
    buildings.value = response.data.data;
  });
}, { immediate: true });

const handleOk = () => {
  formRef.value.validate().then(() => {
    if (props.mode === 'edit') {
      update(formData.value).then(() => {
        emit('update:visible', false);
        emit('updated');
      });
    } else {
      create(formData.value).then((response) => {
        emit('update:visible', false);
        emit('created', response);
      });
    }
  });
};

const handleCancel = () => {
  emit('update:visible', false);
  emit('cancel');
};
</script>

<template>
  <a-modal :open="props.visible" @ok="handleOk" @cancel="handleCancel"
    :title="props.mode === 'create' ? 'Add Device' : 'Edit Device'" :mask-closable="false">
    <a-form ref="formRef" :model="formData" :label-col="{ style: { width: '200px' } }">
      <a-form-item label="Device ID" name="deviceId" :rules="[{ required: true, message: 'Please input device ID!' }]">
        <a-input v-model:value="formData.deviceId" />
      </a-form-item>
      <a-form-item label="Building" name="buildingId"
        :rules="[{ required: true, message: 'Please select a building!' }]">
        <a-select v-model:value="formData.buildingId">
          <a-select-option v-for="building in buildings" :key="building.id" :value="building.id">
            {{ building.name }}
          </a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item label="Arrearage Update Duration" name="settings.arrearageUpdateDuration">
        <a-input-number v-model:value="formData.settings!.arrearageUpdateDuration" :min="1" addon-after="seconds" />
      </a-form-item>
      <a-form-item label="Notice Update Duration" name="settings.noticeUpdateDuration">
        <a-input-number v-model:value="formData.settings!.noticeUpdateDuration" :min="1" addon-after="seconds" />
      </a-form-item>
      <a-form-item label="Advertisement Update Duration" name="settings.advertisementUpdateDuration">
        <a-input-number v-model:value="formData.settings!.advertisementUpdateDuration" :min="1" addon-after="seconds" />
      </a-form-item>
      <a-form-item label="Advertisement Play Duration" name="settings.advertisementPlayDuration">
        <a-input-number v-model:value="formData.settings!.advertisementPlayDuration" :min="1" addon-after="seconds" />
      </a-form-item>
      <a-form-item label="Notice Play Duration" name="settings.noticePlayDuration">
        <a-input-number v-model:value="formData.settings!.noticePlayDuration" :min="1" addon-after="seconds" />
      </a-form-item>
      <a-form-item label="Spare Duration" name="settings.spareDuration">
        <a-input-number v-model:value="formData.settings!.spareDuration" :min="1" addon-after="seconds" />
      </a-form-item>
      <a-form-item label="Notice Stay Duration" name="settings.noticeStayDuration">
        <a-input-number v-model:value="formData.settings!.noticeStayDuration" :min="1" addon-after="seconds" />
      </a-form-item>
    </a-form>
  </a-modal>
</template>
