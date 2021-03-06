import { connectRouter } from 'connected-react-router'
import { combineReducers } from 'redux'

import { toasterReducers as toaster } from './App/App.components/Toaster/Toaster.reducers'
import { metricsReducers as posts } from './pages/Dashboard/Dashboard.reducers'
import { drawerReducers as drawer } from './App/App.components/Drawer/Drawer.reducers'
import { progressBarReducers as progressBar } from './App/App.components/ProgressBar/ProgressBar.reducers'

export const rootReducer = (history: any) =>
  combineReducers({
    router: connectRouter(history),
    toaster,
    posts,
    drawer,
    progressBar
  })
