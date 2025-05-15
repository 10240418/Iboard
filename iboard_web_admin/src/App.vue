<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import router from './router/router';

const selectedKeys = ref<string[]>(['Dashboard']);
const openKeys = ref<string[]>(['sub1']);

const menuItems = [
  { key: 'Dashboard', label: 'Dashboard' },
  { key: 'Device', label: "Device" },
  { key: 'Building', label: 'Building' },
  { key: 'File', label: 'File' },
  { key: 'Advertisement', label: 'Advertisement' },
  { key: 'Notice', label: 'Notice' },
  // { key: 'API', label: 'API' },
  { key: 'BuildingAdmin', label: 'Building Admin' },
  { key: 'SuperAdmin', label: 'Super Admin' },
];

const onSiderClick = ({ key }: { key: string }) => {
  router.push({ name: key });
};

const route = useRoute();
const breadcrumbItems = computed(() => {
  const matched = route.matched;
  return matched.map((m) => m.name);
});

const showSidebar = computed(() => {
  return route.name !== 'Login';
});

watch(route, () => {
  selectedKeys.value = [route.name as string];
});
</script>

<template>
  <a-layout class="layout">
    <a-layout>
      <a-layout-sider width="200" style="background: #fff" v-if="showSidebar">
        <div class="logo" />

        <!-- Sidebar menu -->
        <a-menu v-model:selectedKeys="selectedKeys" v-model:openKeys="openKeys" mode="inline"
          :style="{ height: '100%', borderRight: 0 }" @click="onSiderClick">

          <!-- Sidebar menu items -->
          <a-menu-item v-for="item in menuItems" :key="item.key">
            <span>{{ item.label }}</span>
          </a-menu-item>

          <!-- Submenu example -->
          <!-- <a-sub-menu key="sub3">
            <template #title>
              <span>
                <notification-outlined />
                subnav 3
              </span>
            </template>
<a-menu-item key="9">option9</a-menu-item>
<a-menu-item key="10">option10</a-menu-item>
<a-menu-item key="11">option11</a-menu-item>
<a-menu-item key="12">option12</a-menu-item>
</a-sub-menu> -->
        </a-menu>
      </a-layout-sider>

      <!-- Main content -->
      <a-layout style="padding: 0 24px 24px">
        <a-breadcrumb style="margin: 16px 0">
          <a-breadcrumb-item v-for="item in breadcrumbItems" :key="item">{{ item }}</a-breadcrumb-item>
        </a-breadcrumb>
        <a-layout-content :style="{ background: '#fff', padding: '24px', margin: 0, minHeight: '280px' }">
          <RouterView />
        </a-layout-content>
      </a-layout>
    </a-layout>
  </a-layout>
</template>

<style scoped>
.layout {
  width: 100vw;
  min-height: 100vh;
}
</style>
