import React from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { ThemeProvider, createTheme } from '@mui/material/styles'
import CssBaseline from '@mui/material/CssBaseline'
import { AuthProvider } from './contexts/AuthContext'
import { WebSocketProvider } from './contexts/WebSocketContext'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Chains from './pages/Chains'
import Channels from './pages/Channels'
import Packets from './pages/Packets'
import Settings from './pages/Settings'
import Login from './pages/Login'
import PrivateRoute from './components/PrivateRoute'

const darkTheme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#90caf9',
    },
    secondary: {
      main: '#f48fb1',
    },
    background: {
      default: '#0a0a0a',
      paper: '#121212',
    },
  },
})

export default function App() {
  return (
    <ThemeProvider theme={darkTheme}>
      <CssBaseline />
      <AuthProvider>
        <WebSocketProvider>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route
              path="/"
              element={
                <PrivateRoute>
                  <Layout />
                </PrivateRoute>
              }
            >
              <Route index element={<Navigate to="/dashboard" replace />} />
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="chains" element={<Chains />} />
              <Route path="channels" element={<Channels />} />
              <Route path="packets" element={<Packets />} />
              <Route path="settings" element={<Settings />} />
            </Route>
          </Routes>
        </WebSocketProvider>
      </AuthProvider>
    </ThemeProvider>
  )
}