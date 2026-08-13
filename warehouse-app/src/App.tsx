import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import ProtectedRoute from './components/ProtectedRoute'
import Layout from './components/Layout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Orders from './pages/Orders'
import Picking from './pages/Picking'
import Packing from './pages/Packing'
import StockTransfers from './pages/StockTransfers'
import Exceptions from './pages/Exceptions'
import Performance from './pages/Performance'
import Locations from './pages/Locations'
import StockOperations from './pages/StockOperations'

function Protected({ children }: { children: React.ReactNode }) {
  return (
    <ProtectedRoute>
      <Layout>{children}</Layout>
    </ProtectedRoute>
  )
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/dashboard" element={<Protected><Dashboard /></Protected>} />
          <Route path="/orders" element={<Protected><Orders /></Protected>} />
          <Route path="/picking/:orderId" element={<Protected><Picking /></Protected>} />
          <Route path="/packing/:orderId" element={<Protected><Packing /></Protected>} />
          <Route path="/stock-transfers" element={<Protected><StockTransfers /></Protected>} />
          <Route path="/exceptions" element={<Protected><Exceptions /></Protected>} />
          <Route path="/performance" element={<Protected><Performance /></Protected>} />
          <Route path="/locations" element={<Protected><Locations /></Protected>} />
          <Route path="/stock-operations" element={<Protected><StockOperations /></Protected>} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App
