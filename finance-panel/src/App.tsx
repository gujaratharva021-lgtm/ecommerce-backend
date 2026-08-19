import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import ProtectedRoute from './components/ProtectedRoute'
import Layout from './components/Layout'
import Login from './pages/Login'
import Revenue from './pages/Revenue'
import Expenses from './pages/Expenses'
import Payroll from './pages/Payroll'
import Invoices from './pages/Invoices'
import Reports from './pages/Reports'
import Settings from './pages/Settings'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<Login />} />

          <Route
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route path="/revenue" element={<Revenue />} />
            <Route path="/expenses" element={<Expenses />} />
            <Route path="/payroll" element={<Payroll />} />
            <Route path="/invoices" element={<Invoices />} />
            <Route path="/reports" element={<Reports />} />
            <Route path="/settings" element={<Settings />} />
          </Route>

          <Route path="*" element={<Navigate to="/revenue" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
