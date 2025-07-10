import React from 'react'
import { Outlet } from 'react-router-dom'
import { Box, AppBar, Toolbar, Typography } from '@mui/material'

export default function Layout() {
  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar position="fixed">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            IBC Relayer Dashboard
          </Typography>
        </Toolbar>
      </AppBar>
      <Box component="main" sx={{ flexGrow: 1, p: 3, mt: 8 }}>
        <Outlet />
      </Box>
    </Box>
  )
}