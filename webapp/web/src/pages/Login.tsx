import React from 'react'
import { useNavigate } from 'react-router-dom'
import { Box, Paper, TextField, Button, Typography } from '@mui/material'
import { useAuth } from '../contexts/AuthContext'

export default function Login() {
  const navigate = useNavigate()
  const { login } = useAuth()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await login('admin', 'admin123')
    navigate('/dashboard')
  }

  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
      }}
    >
      <Paper sx={{ p: 4, maxWidth: 400, width: '100%' }}>
        <Typography variant="h4" gutterBottom align="center">
          IBC Relayer Dashboard
        </Typography>
        <form onSubmit={handleSubmit}>
          <TextField
            fullWidth
            label="Username"
            margin="normal"
            defaultValue="admin"
          />
          <TextField
            fullWidth
            label="Password"
            type="password"
            margin="normal"
            defaultValue="admin123"
          />
          <Button
            fullWidth
            variant="contained"
            size="large"
            sx={{ mt: 2 }}
            type="submit"
          >
            Login
          </Button>
        </form>
      </Paper>
    </Box>
  )
}