import { configureStore } from '@reduxjs/toolkit'
import userReducer from './slices/userSlice'
import searchReducer from './slices/searchSlice'

export const store = configureStore({
  reducer: {
    user: userReducer,
    search: searchReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // 忽略这些 action 的序列化检查
        ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE'],
      },
    }),
})

// 导出类型
export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch

export default store