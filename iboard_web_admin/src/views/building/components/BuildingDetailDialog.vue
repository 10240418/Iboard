<script setup lang="ts">
import { computed } from 'vue';
import dayjs from 'dayjs';
import { Building } from '@/model/buildingModel'; // Assuming Building interface includes optional location: string
import { DeviceStatus } from '@/model/deviceModel';

const props = defineProps<{
  visible: boolean;
  building?: Building;
}>();

const emits = defineEmits<{
  (e: 'update:visible', value: boolean): void;
}>();

const closeDialog = () => {
  emits('update:visible', false);
};

const notices = computed(() => props.building?.notices || []);
const advertisements = computed(() => props.building?.advertisements || []);
const devices = computed(() => props.building?.devices || []);

const formatDateTime = (date: string) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss');
};

const getFileUrl = (file: any) => {
  return file?.path || '';
};

const getStatusBadgeProps = (status?: DeviceStatus) => {
  switch (status) {
    case DeviceStatus.Online:
      return { status: 'success', text: 'Online' };
    case DeviceStatus.Offline:
      return { status: 'error', text: 'Offline' };
    default:
      return { status: 'default', text: 'Unknown' };
  }
};

const formatSeconds = (seconds?: number) => {
  if (!seconds) return '-';
  return `${seconds} seconds`;
};
</script>

<template>
  <a-modal :open="props.visible" title="Building Details" width="1000px" @cancel="closeDialog" :footer="null">
    <template v-if="building">
      <a-descriptions bordered>
        <a-descriptions-item label="ID" :span="1">{{ building.id }}</a-descriptions-item>
        <a-descriptions-item label="Name" :span="2">{{ building.name }}</a-descriptions-item>
        <a-descriptions-item label="Ismart ID" :span="1">{{ building.ismartId }}</a-descriptions-item>
        <a-descriptions-item label="Location" :span="2">{{ building.location || '-' }}</a-descriptions-item> {/* Added
        Location */}
        <a-descriptions-item label="Remark" :span="3">{{ building.remark }}</a-descriptions-item> {/* Adjusted span */}
      </a-descriptions>

      <a-divider />

      <a-tabs>
        <a-tab-pane key="notices" tab="Notices">
          <a-table :dataSource="notices" :pagination="false" rowKey="id"> {/* Added rowKey */}
            <a-table-column key="id" title="ID" data-index="id" />
            <a-table-column key="title" title="Title" data-index="title" />
            <a-table-column key="type" title="Type" data-index="type" />
            <a-table-column key="status" title="Status" data-index="status" />
            <a-table-column key="startTime" title="Start Time">
              <template #default="{ record }">
                {{ formatDateTime(record.startTime) }}
              </template>
            </a-table-column>
            <a-table-column key="endTime" title="End Time">
              <template #default="{ record }">
                {{ formatDateTime(record.endTime) }}
              </template>
            </a-table-column>
            <a-table-column key="file" title="File">
              <template #default="{ record }">
                <a v-if="record.file" :href="getFileUrl(record.file)" target="_blank">View File</a>
              </template>
            </a-table-column>
          </a-table>
        </a-tab-pane>

        <a-tab-pane key="advertisements" tab="Advertisements">
          <a-table :dataSource="advertisements" :pagination="false" rowKey="id"> {/* Added rowKey */}
            <a-table-column key="id" title="ID" data-index="id" />
            <a-table-column key="title" title="Title" data-index="title" />
            <a-table-column key="type" title="Type" data-index="type" />
            <a-table-column key="status" title="Status" data-index="status" />
            <a-table-column key="display" title="Display" data-index="display" />
            <a-table-column key="duration" title="Duration" data-index="duration" />
            <a-table-column key="startTime" title="Start Time">
              <template #default="{ record }">
                {{ formatDateTime(record.startTime) }}
              </template>
            </a-table-column>
            <a-table-column key="endTime" title="End Time">
              <template #default="{ record }">
                {{ formatDateTime(record.endTime) }}
              </template>
            </a-table-column>
            <a-table-column key="file" title="File">
              <template #default="{ record }">
                <a v-if="record.file" :href="getFileUrl(record.file)" target="_blank">View File</a>
              </template>
            </a-table-column>
          </a-table>
        </a-tab-pane>

        <a-tab-pane key="devices" tab="Devices">
          <a-table :dataSource="devices" :pagination="false" rowKey="id">
            <a-table-column key="id" title="ID" data-index="id" />
            <a-table-column key="deviceId" title="Device ID" data-index="deviceId" />
            <a-table-column key="status" title="Status">
              <template #default="{ record }">
                <a-badge v-bind="getStatusBadgeProps(record.status)" />
              </template>
            </a-table-column>
            <a-table-column key="settings" title="Settings" :ellipsis="true">
              <template #default="{ record }">
                <div v-if="record.settings">
                  <p>Notice Update: {{ formatSeconds(record.settings.noticeUpdateDuration) }}</p>
                  <p>Notice Play: {{ formatSeconds(record.settings.noticePlayDuration) }}</p>
                  <p>Notice Stay: {{ formatSeconds(record.settings.noticeStayDuration) }}</p>
                  <p>Ad Update: {{ formatSeconds(record.settings.advertisementUpdateDuration) }}</p>
                  <p>Ad Play: {{ formatSeconds(record.settings.advertisementPlayDuration) }}</p>
                  <p>Spare Duration: {{ formatSeconds(record.settings.spareDuration) }}</p>
                  <p>Arrearage Update: {{ formatSeconds(record.settings.arrearageUpdateDuration) }}</p>
                </div>
              </template>
            </a-table-column>
          </a-table>
        </a-tab-pane>
      </a-tabs>
    </template>
  </a-modal>
</template>

<style scoped>
.ant-descriptions-item-content p {
  margin-bottom: 4px;
}
</style>
