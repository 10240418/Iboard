<script setup lang="ts">
import { ref, defineProps, defineEmits, watch } from 'vue';
import { getAdvertisements } from '@/httpapis/advertisement';
import { Building } from '@/model/buildingModel';
import { Advertisement } from '@/model/advertisementModel';
import { useBuildingData } from '../useBuilding';
import { getBuildingsAdvertisement } from '@/httpapis/building';

const { bindAdvertisements, unbindAdvertisements, isLoading } = useBuildingData();

const props = defineProps({
  visible: Boolean,
  building: Object as () => Building | undefined,
});
const emit = defineEmits(['update:visible', 'success', 'cancel']);

const initialAdvertisements = ref<number[]>([]);
const selectedAdvertisements = ref<number[]>([]);
const advertisements = ref<Advertisement[]>([]);

watch(() => props.visible, () => {
  if (!props.visible) return;
  if (props.building) {
    getAdvertisements({ pageSize: 999, pageNum: 1 }).then((response) => {
      advertisements.value = response.data.data;
    }).catch((error) => {
      console.log('Failed to fetch advertisements:', error);
    });
    getBuildingsAdvertisement(props.building.id!).then((response) => {
      selectedAdvertisements.value = response.data.data.map((ad: Advertisement) => ad.id!);
      initialAdvertisements.value = selectedAdvertisements.value;
    }).catch((error) => {
      console.log('Failed to fetch building advertisements:', error);
    });
  }
}, { immediate: true });

const handleOk = () => {
  if (props.building) {
    unbindAdvertisements(props.building.id!, _getDeletedAdvertisements()).then(() => {
      bindAdvertisements(props.building!.id!, selectedAdvertisements.value).then(() => {
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

const _getDeletedAdvertisements = () => {
  return initialAdvertisements.value.filter((id) => !selectedAdvertisements.value.includes(id));
};
</script>

<template>
  <a-modal :open="props.visible" @ok="handleOk" @cancel="handleCancel" title="Bind Advertisements"
    :mask-closable="false" :confirm-loading="isLoading">
    <p>Select advertisements to bind with building: {{ props.building?.name }}</p>
    <a-select v-model:value="selectedAdvertisements" mode="multiple" style="width: 100%" option-label-prop="label">
      <a-select-option v-for="ad in advertisements" :key="ad.id" :value="ad.id" :label="ad.title">
        <div>
          <strong>{{ ad.title }}</strong><span style="color: gray;"> ID:{{ ad.id }}</span><br>
          <span style="color: gray;"> {{ ad.description }}</span><br>
        </div>
      </a-select-option>
    </a-select>
  </a-modal>
</template>

<style scoped>
/* Add any necessary styles here */
</style>
