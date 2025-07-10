import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react'

interface WebSocketContextType {
  isConnected: boolean
  lastMessage: any
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined)

export function WebSocketProvider({ children }: { children: ReactNode }) {
  const [isConnected, setIsConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState<any>(null)

  useEffect(() => {
    // TODO: Implement WebSocket connection
    setIsConnected(true)
  }, [])

  return (
    <WebSocketContext.Provider value={{ isConnected, lastMessage }}>
      {children}
    </WebSocketContext.Provider>
  )
}

export function useWebSocket() {
  const context = useContext(WebSocketContext)
  if (!context) {
    throw new Error('useWebSocket must be used within WebSocketProvider')
  }
  return context
}