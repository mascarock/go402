import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Tenant from './pages/Tenant'
import PaymentDemo from './pages/PaymentDemo'
import Settlements from './pages/Settlements'

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/tenants/:id" element={<Tenant />} />
        <Route path="/demo" element={<PaymentDemo />} />
        <Route path="/settlements" element={<Settlements />} />
      </Routes>
    </Layout>
  )
}
