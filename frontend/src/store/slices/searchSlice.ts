import { createSlice, PayloadAction } from '@reduxjs/toolkit'

interface SearchResult {
  id: string
  title: string
  content: string
  highlightedText: string
  similarity: number
  docType: string
  department: string
  uploadDate: string
  url: string
}

interface SearchState {
  query: string
  results: SearchResult[]
  answer: string
  loading: boolean
  error: string | null
  filters: {
    docType: string
    department: string
    dateRange: any[]
  }
}

const initialState: SearchState = {
  query: '',
  results: [],
  answer: '',
  loading: false,
  error: null,
  filters: {
    docType: '',
    department: '',
    dateRange: []
  }
}

const searchSlice = createSlice({
  name: 'search',
  initialState,
  reducers: {
    setSearchQuery: (state, action: PayloadAction<string>) => {
      state.query = action.payload
    },
    setSearchResults: (state, action: PayloadAction<SearchResult[]>) => {
      state.results = action.payload
    },
    setAnswer: (state, action: PayloadAction<string>) => {
      state.answer = action.payload
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload
    },
    setFilters: (state, action: PayloadAction<Partial<SearchState['filters']>>) => {
      state.filters = { ...state.filters, ...action.payload }
    },
    clearSearch: (state) => {
      state.query = ''
      state.results = []
      state.answer = ''
      state.error = null
    }
  },
})

export const { 
  setSearchQuery, 
  setSearchResults, 
  setAnswer, 
  setLoading, 
  setError, 
  setFilters,
  clearSearch 
} = searchSlice.actions
export default searchSlice.reducer