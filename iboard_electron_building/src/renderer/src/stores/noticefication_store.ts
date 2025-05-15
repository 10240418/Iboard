// src/renderer/src/stores/notificationStore.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Notification {
  id: number
  message: string
  type: 'success' | 'error' | 'info' | 'warning'
  duration?: number // 持续时间（毫秒），可选
}

export const useNotificationStore = defineStore('notification', () => {
  const notifications = ref<Notification[]>([])

  const addNotification = (
    message: string,
    type: Notification['type'] = 'info',
    duration: number = 3000
  ) => {
    const id = Date.now()
    notifications.value.push({ id, message, type, duration })

    // 自动移除通知
    setTimeout(() => {
      removeNotification(id)
    }, duration)
  }

  const removeNotification = (id: number) => {
    notifications.value = notifications.value.filter((notification) => notification.id !== id)
  }

  return {
    notifications,
    addNotification,
    removeNotification
  }
})
