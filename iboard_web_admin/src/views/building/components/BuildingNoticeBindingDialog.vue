<script setup lang="ts">
import { ref, defineProps, defineEmits, watch } from 'vue';
import { getNotices } from '@/httpapis/notice';
import { Building } from '@/model/buildingModel';
import { Notice } from '@/model/noticeModel';
import { useBuildingData } from '../useBuilding';
import { getBuildingsNotice } from '@/httpapis/building';

const { bindNotices, unbindNotices, isLoading } = useBuildingData();

const props = defineProps({
  visible: Boolean,
  building: Object as () => Building | undefined,
});
const emit = defineEmits(['update:visible', 'success', 'cancel']);

const initialNotices = ref<number[]>([]);
const selectedNotices = ref<number[]>([]);
const notices = ref<Notice[]>([]);

watch(() => props.visible, () => {
  if (!props.visible) return;
  if (props.building) {
    getNotices({ pageSize: 999, pageNum: 1 }).then((response) => {
      notices.value = response.data.data;
    }).catch((error) => {
      console.log('Failed to fetch notices:', error);
    });
    getBuildingsNotice(props.building.id!).then((response) => {
      selectedNotices.value = response.data.data.map((notice: Notice) => notice.id!);
      initialNotices.value = selectedNotices.value;
    }).catch((error) => {
      console.log('Failed to fetch building notices:', error);
    });
  }
}, { immediate: true });

const handleOk = () => {
  if (props.building) {
    console.log(_getDeletedNotices());
    unbindNotices(props.building.id!, _getDeletedNotices()).then(() => {
      bindNotices(props.building!.id!, selectedNotices.value).then(() => {
        emit('update:visible', false);
        emit('success');
      })
    })
  }
};

const handleCancel = () => {
  emit('update:visible', false);
  emit('cancel');
};

const _getDeletedNotices = () => {
  return initialNotices.value.filter((id) => !selectedNotices.value.includes(id));
};
</script>

<template>
  <a-modal :open="props.visible" @ok="handleOk" @cancel="handleCancel" title="Bind Notices" :mask-closable="false"
    :confirm-loading="isLoading">
    <p>Select notices to bind with building: {{ props.building?.name }}</p>
    <a-select v-model:value="selectedNotices" mode="multiple" style="width: 100%" option-label-prop="label">
      <a-select-option v-for="notice in notices" :key="notice.id" :value="notice.id" :label="notice.title">
        <div>
          <strong>{{ notice.title }}</strong><span style="color: gray;"> ID:{{ notice.id }}</span><br>
          <span style="color: gray;"> {{ notice.description }}</span><br>
        </div>
      </a-select-option>
    </a-select>
  </a-modal>
</template>

<style scoped>
/* Add any necessary styles here */
</style>
