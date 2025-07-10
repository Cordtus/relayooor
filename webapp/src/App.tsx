import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Channels from './pages/Channels'
import PacketClearing from './pages/PacketClearing'
import Settings from './pages/Settings'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="channels" element={<Channels />} />
        <Route path="packet-clearing" element={<PacketClearing />} />
        <Route path="settings" element={<Settings />} />
      </Route>
    </Routes>
  )
}