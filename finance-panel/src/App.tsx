import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import ProtectedRoute from './components/ProtectedRoute'
import Layout from './components/Layout'
import Login from './pages/Login'
import Revenue from './pages/Revenue'
import Payments from './pages/Payments'
import Expenses from './pages/Expenses'
import Payroll from './pages/Payroll'
import ProfitLoss from './pages/ProfitLoss'
import GST from './pages/GST'
import Invoices from './pages/Invoices'
import Reports from './pages/Reports'
import RangeReport from './pages/RangeReport'
import Settings from './pages/Settings'
import Vendors from './pages/Vendors'
import VendorBills from './pages/VendorBills'
import Accounts from './pages/Accounts'
import Ledger from './pages/Ledger'
import BankReconciliation from './pages/BankReconciliation'

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
            <Route path="/payments" element={<Payments />} />
            <Route path="/expenses" element={<Expenses />} />
            <Route path="/payroll" element={<Payroll />} />
            <Route path="/profit-loss" element={<ProfitLoss />} />
            <Route path="/gst" element={<GST />} />
            <Route path="/invoices" element={<Invoices />} />
            <Route path="/reports" element={<Reports />} />
            <Route path="/reports/range" element={<RangeReport />} />
            <Route path="/settings" element={<Settings />} />
            <Route path="/accounting/vendors" element={<Vendors />} />
            <Route path="/accounting/vendor-bills" element={<VendorBills />} />
            <Route path="/accounting/accounts" element={<Accounts />} />
            <Route path="/accounting/ledger" element={<Ledger />} />
            <Route path="/accounting/bank-reconciliation" element={<BankReconciliation />} />
          </Route>

          <Route path="*" element={<Navigate to="/revenue" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
