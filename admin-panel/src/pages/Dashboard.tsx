import { useEffect, useState } from 'react'
import {
  ResponsiveContainer,
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts'
import Layout from '../components/Layout'
import { getDashboardOverview } from '../api/admin'
import type { DashboardOverview } from '../types/admin'

const COLORS = ['#34d399', '#fbbf24', '#38bdf8', '#a78bfa', '#f87171', '#94a3b8', '#fb923c', '#22d3ee']

const STATUS_COLORS: Record<string, string> = {
  pending: '#fbbf24',
  confirmed: '#38bdf8',
  shipped: '#a78bfa',
  delivered: '#34d399',
  cancelled: '#f87171',
  returned: '#fb923c',
  open: '#fbbf24',
  in_progress: '#38bdf8',
  resolved: '#34d399',
  closed: '#94a3b8',
}

function fmtMoney(n: number) {
  return '\u20b9' + Number(n ?? 0).toLocaleString('en-IN', { maximumFractionDigits: 0 })
}

function fmtDate(d: string) {
  const parts = d.split('-')
  if (parts.length !== 3) return d
  return `${parts[2]}/${parts[1]}`
}

function ChartCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-slate-900 rounded-xl p-5 border border-slate-800">
      <p className="text-sm text-slate-400 mb-4">{title}</p>
      <div style={{ width: '100%', height: 260 }}>{children}</div>
    </div>
  )
}

const tooltipStyle = {
  contentStyle: { background: '#0f172a', border: '1px solid #1e293b', borderRadius: 8, fontSize: 12 },
  labelStyle: { color: '#94a3b8' },
}

export default function Dashboard() {
  const [data, setData] = useState<DashboardOverview | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const res = await getDashboardOverview()
        if (!cancelled) setData(res)
      } catch (err: any) {
        if (!cancelled) setError(err.response?.data?.error ?? 'Failed to load dashboard.')
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }

    load()
    return () => {
      cancelled = true
    }
  }, [])

  const s = data?.stats
  const c = data?.charts

  const cards = s
    ? [
        { label: 'Total Revenue', value: fmtMoney(s.total_sales) },
        { label: "Today's Revenue", value: fmtMoney(s.revenue_today) },
        { label: 'Avg Order Value', value: fmtMoney(s.avg_order_value) },
        { label: 'Total Orders', value: s.total_orders },
        { label: "Today's Orders", value: s.orders_today },
        { label: 'Pending Orders', value: s.pending_orders },
        { label: 'Confirmed Orders', value: s.confirmed_orders },
        { label: 'Shipped Orders', value: s.shipped_orders },
        { label: 'Delivered Orders', value: s.delivered_orders },
        { label: 'Cancelled Orders', value: s.cancelled_orders },
        { label: 'Returned Orders', value: s.returned_orders },
        { label: 'Total Users', value: s.total_users },
        { label: 'New Users Today', value: s.new_users_today },
        { label: 'Total Products', value: s.total_products },
        { label: 'Low Stock Products', value: s.low_stock_products, warn: s.low_stock_products > 0 },
        { label: 'Out of Stock', value: s.out_of_stock_products, warn: s.out_of_stock_products > 0 },
        { label: 'Active Delivery Partners', value: s.active_delivery_partners },
        { label: 'Total Warehouses', value: s.total_warehouses },
        { label: 'Open Support Tickets', value: s.open_support_tickets, warn: s.open_support_tickets > 0 },
        { label: 'Pending Payment Amount', value: fmtMoney(s.pending_payment_amount), warn: s.pending_payment_amount > 0 },
      ]
    : []

  return (
    <Layout>
      <div className="p-8">
        <h1 className="text-xl font-semibold mb-6">Overview</h1>

        {isLoading && <p className="text-slate-400">Loading dashboard...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {s && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
            {cards.map((card) => (
              <div
                key={card.label}
                className="bg-slate-900 rounded-xl p-5 border border-slate-800"
              >
                <p className="text-sm text-slate-400 mb-1">{card.label}</p>
                <p className={`text-2xl font-bold ${card.warn ? 'text-amber-400' : ''}`}>{card.value}</p>
              </div>
            ))}
          </div>
        )}

        {c && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <ChartCard title="Revenue trend (last 14 days)">
              <ResponsiveContainer>
                <LineChart data={c.sales_trend}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                  <XAxis dataKey="date" tickFormatter={fmtDate} stroke="#64748b" fontSize={12} />
                  <YAxis stroke="#64748b" fontSize={12} />
                  <Tooltip {...tooltipStyle} formatter={(v) => fmtMoney(Number(v))} labelFormatter={(l) => fmtDate(String(l ?? ""))} />
                  <Line type="monotone" dataKey="revenue" stroke="#34d399" strokeWidth={2} dot={false} name="Revenue" />
                </LineChart>
              </ResponsiveContainer>
            </ChartCard>

            <ChartCard title="Orders trend (last 14 days)">
              <ResponsiveContainer>
                <BarChart data={c.sales_trend}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                  <XAxis dataKey="date" tickFormatter={fmtDate} stroke="#64748b" fontSize={12} />
                  <YAxis stroke="#64748b" fontSize={12} allowDecimals={false} />
                  <Tooltip {...tooltipStyle} labelFormatter={(l) => fmtDate(String(l ?? ""))} />
                  <Bar dataKey="orders" fill="#38bdf8" name="Orders" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </ChartCard>

            <ChartCard title="New users (last 14 days)">
              <ResponsiveContainer>
                <LineChart data={c.user_growth}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                  <XAxis dataKey="date" tickFormatter={fmtDate} stroke="#64748b" fontSize={12} />
                  <YAxis stroke="#64748b" fontSize={12} allowDecimals={false} />
                  <Tooltip {...tooltipStyle} labelFormatter={(l) => fmtDate(String(l ?? ""))} />
                  <Line type="monotone" dataKey="count" stroke="#a78bfa" strokeWidth={2} dot={false} name="New users" />
                </LineChart>
              </ResponsiveContainer>
            </ChartCard>

            <ChartCard title="Orders by status">
              <ResponsiveContainer>
                <PieChart>
                  <Pie data={c.orders_by_status} dataKey="count" nameKey="status" innerRadius={55} outerRadius={90} paddingAngle={2}>
                    {c.orders_by_status.map((entry, i) => (
                      <Cell key={entry.status} fill={STATUS_COLORS[entry.status] ?? COLORS[i % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip {...tooltipStyle} />
                  <Legend wrapperStyle={{ fontSize: 12, color: '#94a3b8' }} />
                </PieChart>
              </ResponsiveContainer>
            </ChartCard>

            <ChartCard title="Payment method split (paid)">
              <ResponsiveContainer>
                <PieChart>
                  <Pie data={c.payment_split} dataKey="revenue" nameKey="method" innerRadius={55} outerRadius={90} paddingAngle={2}>
                    {c.payment_split.map((entry, i) => (
                      <Cell key={entry.method} fill={COLORS[i % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip {...tooltipStyle} formatter={(v) => fmtMoney(Number(v))} />
                  <Legend wrapperStyle={{ fontSize: 12, color: '#94a3b8' }} />
                </PieChart>
              </ResponsiveContainer>
            </ChartCard>

            <ChartCard title="Top 5 products by revenue">
              <ResponsiveContainer>
                <BarChart data={c.top_products} layout="vertical" margin={{ left: 24 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                  <XAxis type="number" stroke="#64748b" fontSize={12} />
                  <YAxis type="category" dataKey="product_name" stroke="#64748b" fontSize={12} width={110} />
                  <Tooltip {...tooltipStyle} formatter={(v) => fmtMoney(Number(v))} />
                  <Bar dataKey="total_revenue" fill="#fbbf24" name="Revenue" radius={[0, 4, 4, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </ChartCard>

            <ChartCard title="Revenue by warehouse">
              <ResponsiveContainer>
                <BarChart data={c.revenue_by_warehouse}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                  <XAxis dataKey="warehouse_name" stroke="#64748b" fontSize={12} />
                  <YAxis stroke="#64748b" fontSize={12} />
                  <Tooltip {...tooltipStyle} formatter={(v) => fmtMoney(Number(v))} />
                  <Bar dataKey="revenue" fill="#22d3ee" name="Revenue" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </ChartCard>

            <ChartCard title="Support tickets by status">
              <ResponsiveContainer>
                <PieChart>
                  <Pie data={c.tickets_by_status} dataKey="count" nameKey="status" innerRadius={55} outerRadius={90} paddingAngle={2}>
                    {c.tickets_by_status.map((entry, i) => (
                      <Cell key={entry.status} fill={STATUS_COLORS[entry.status] ?? COLORS[i % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip {...tooltipStyle} />
                  <Legend wrapperStyle={{ fontSize: 12, color: '#94a3b8' }} />
                </PieChart>
              </ResponsiveContainer>
            </ChartCard>
          </div>
        )}
      </div>
    </Layout>
  )
}
