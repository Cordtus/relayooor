import React from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

export default function PrivateRoute({ children }: { children: React.ReactElement }) {
  const { isAuthenticated } = useAuth()
  
  // For now, allow access without authentication for development
  return isAuthenticated || true ? children : <Navigate to="/login" />
}