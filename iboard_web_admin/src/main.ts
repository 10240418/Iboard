import { createApp } from 'vue'
import { createPinia } from 'pinia';
import './style.css'
import Layout from '@/App.vue'
import 'ant-design-vue/dist/reset.css';
import router from '@/router/router';

const pinia = createPinia();

createApp(Layout)
    .use(router)
    .use(pinia)
    .mount('#app');
