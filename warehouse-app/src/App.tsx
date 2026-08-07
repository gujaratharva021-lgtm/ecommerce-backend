import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import ProtectedRoute from './components/ProtectedRoute'
import Login from './pages/Login'
import StockTransfers from './pages/StockTransfers'

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route
            path="/stock-transfers"
            element={
              <ProtectedRoute>
                <StockTransfers />
              </ProtectedRoute>
            }
          />
          <Route path="*" element={<Navigate to="/stock-transfers" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App
