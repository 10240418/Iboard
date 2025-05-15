<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useLogin } from './useLogin';
import { LoginModel } from '@/model/authModel';

const router = useRouter();

const formRef = ref();
const formData = ref<LoginModel>({
  email: '',
  password: '',
});
const { isLoading, errorMessage, handleLogin } = useLogin();

const onSubmit = async () => {
  await handleLogin(formData.value);
  if (!errorMessage.value) {
    router.push({ name: 'Dashboard' });
  }
};
</script>

<template>
  <div class="login-container">
    <a-form ref="formRef" @submit.prevent="onSubmit" :label-col="{ style: { width: '100px' } }" style="width:300px;">
      <a-form-item label="Email" name="email" :rules="[{ required: true, message: 'Please input email!' }]">
        <a-input v-model:value="formData.email" id="email" />
      </a-form-item>
      <a-form-item label="Password" name="password" :rules="[{ required: true, message: 'Please input password!' }]">
        <a-input type="password" v-model:value="formData.password" id="password" />
      </a-form-item>
      <a-form-item>
        <a-button type="primary" html-type="submit" :loading="isLoading">Login</a-button>
      </a-form-item>
      <p v-if="errorMessage">{{ errorMessage }}</p>
    </a-form>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}
</style>
