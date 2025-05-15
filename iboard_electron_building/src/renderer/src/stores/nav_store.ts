import { defineStore } from 'pinia'
export type routerState = 'urgentNotice' | 'generalNotice' | 'legalNotice' | 'governmentNotice'
export const useRouterStore = defineStore('router', {
  state: () => {
    return {
      routerStates: [
        'urgentNotice',
        'generalNotice',
        'legalNotice',
        'governmentNotice'
      ] as routerState[],
      currentRouter: 'urgentNotice' as routerState
    }
  },
  getters: {
    getCurrentRouter: (state) => state.currentRouter,
    getRouterStates: (state) => state.routerStates
  },
  actions: {
    setCurrentRouter(router: routerState) {
      this.currentRouter = router
    }
  }
  //persist: true,//是否开启持久化存储
})
