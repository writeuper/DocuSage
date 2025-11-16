import { createSlice, PayloadAction } from '@reduxjs/toolkit'

interface UserInfo {
  id: string
  username: string
  role: string
  department: string
  permissions: string[]
}

interface UserState {
  userInfo: UserInfo | null
  isAuthenticated: boolean
  token: string | null
}

const initialState: UserState = {
  userInfo: null,
  isAuthenticated: false,
  token: localStorage.getItem('token') || null,
}

const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    login: (state, action: PayloadAction<UserInfo & { token: string }>) => {
      const { token, ...userInfo } = action.payload
      state.userInfo = userInfo
      state.isAuthenticated = true
      state.token = token
      localStorage.setItem('token', token)
    },
    logout: (state) => {
      state.userInfo = null
      state.isAuthenticated = false
      state.token = null
      localStorage.removeItem('token')
    },
    updateUserInfo: (state, action: PayloadAction<Partial<UserInfo>>) => {
      if (state.userInfo) {
        state.userInfo = { ...state.userInfo, ...action.payload }
      }
    },
  },
})

export const { login, logout, updateUserInfo } = userSlice.actions
export default userSlice.reducer