import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

type Step = 'phone' | 'otp'

export default function Login() {
  const navigate = useNavigate()
  const { sendOtp, verifyOtp } = useAuth()

  const [step, setStep] = useState<Step>('phone')
  const [phone, setPhone] = useState('')
  const [otp, setOtp] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const phoneIsValid = /^\d{10}$/.test(phone)
  const otpIsValid = /^\d{4,6}$/.test(otp)

  async function handleSendOtp(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (!phoneIsValid) {
      setError('Enter a valid 10-digit phone number.')
      return
    }
    setIsSubmitting(true)
    try {
      await sendOtp(phone)
      setStep('otp')
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to send OTP. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleVerifyOtp(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (!otpIsValid) {
      setError('Enter the OTP you received.')
      return
    }
    setIsSubmitting(true)
    try {
      await verifyOtp(phone, otp)
      navigate('/dashboard')
    } catch (err: any) {
      setError(err.response?.data?.error ?? err.message ?? 'Failed to verify OTP. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen flex bg-slate-50">
      {/* Left branding panel */}
      <div className="hidden lg:flex lg:w-1/2 relative overflow-hidden bg-gradient-to-br from-green-800 to-green-950 text-white flex-col justify-between p-12">
        {/* dot pattern */}
        <div
          className="absolute top-8 right-8 w-40 h-40 opacity-30"
          style={{
            backgroundImage: 'radial-gradient(circle, rgba(255,255,255,0.4) 1.5px, transparent 1.5px)',
            backgroundSize: '18px 18px',
          }}
        />
        <div className="absolute top-1/3 right-0 w-72 h-72 rounded-full bg-green-700/20 blur-3xl" />

        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-16">
            <div className="w-12 h-12 rounded-xl bg-white flex items-center justify-center text-2xl">
              🛒
            </div>
            <div>
              <p className="text-xl font-bold leading-tight">GoFresh</p>
              <p className="text-sm text-green-200 leading-tight">Admin Panel</p>
            </div>
          </div>

          <h1 className="text-4xl font-extrabold leading-tight mb-1">Welcome Back,</h1>
          <h1 className="text-4xl font-extrabold leading-tight text-green-400 mb-4">Admin!</h1>
          <div className="w-14 h-1 bg-green-400 rounded-full mb-5" />
          <p className="text-green-100 max-w-sm">
            Login to access and manage your store operations seamlessly.
          </p>

          {/* dashboard mock */}
          <div className="mt-10 bg-slate-900 rounded-xl border border-white/10 shadow-2xl overflow-hidden max-w-md">
            <div className="flex">
              <div className="w-28 bg-green-900/60 p-3 space-y-2 text-xs">
                <div className="px-2 py-1.5 rounded bg-green-600 font-medium">Dashboard</div>
                <div className="px-2 py-1.5 text-green-200/80">Orders</div>
                <div className="px-2 py-1.5 text-green-200/80">Products</div>
                <div className="px-2 py-1.5 text-green-200/80">Users</div>
                <div className="px-2 py-1.5 text-green-200/80">Reports</div>
              </div>
              <div className="flex-1 p-4 bg-white text-slate-900">
                <p className="text-sm font-semibold mb-3">Overview</p>
                <div className="grid grid-cols-3 gap-2 mb-3">
                  <div className="bg-slate-50 rounded-lg p-2">
                    <p className="text-[10px] text-slate-500">Total Orders</p>
                    <p className="text-sm font-bold">1,245</p>
                    <p className="text-[10px] text-green-600">+12.5%</p>
                  </div>
                  <div className="bg-slate-50 rounded-lg p-2">
                    <p className="text-[10px] text-slate-500">Total Sales</p>
                    <p className="text-sm font-bold">₹2,45,000</p>
                    <p className="text-[10px] text-green-600">+15.2%</p>
                  </div>
                  <div className="bg-slate-50 rounded-lg p-2">
                    <p className="text-[10px] text-slate-500">Total Users</p>
                    <p className="text-sm font-bold">856</p>
                    <p className="text-[10px] text-green-600">+8.3%</p>
                  </div>
                </div>
                <div className="bg-slate-50 rounded-lg h-16" />
              </div>
            </div>
          </div>
        </div>

        <div className="relative z-10 grid grid-cols-3 gap-4 mt-10">
          <div>
            <p className="text-sm font-semibold">Secure Access</p>
            <p className="text-xs text-green-200">100% Protected</p>
          </div>
          <div>
            <p className="text-sm font-semibold">Real-time Control</p>
            <p className="text-xs text-green-200">Manage in Real-time</p>
          </div>
          <div>
            <p className="text-sm font-semibold">Smart Insights</p>
            <p className="text-xs text-green-200">Data-driven Decisions</p>
          </div>
        </div>
      </div>

      {/* Right login panel */}
      <div className="flex-1 flex items-center justify-center p-6">
        <div className="w-full max-w-md bg-white rounded-2xl shadow-xl p-8">
          <div className="flex flex-col items-center text-center mb-6">
            <div className="w-16 h-16 rounded-full bg-green-50 flex items-center justify-center mb-4">
              <span className="text-3xl">🛡️</span>
            </div>
            <h2 className="text-2xl font-bold text-slate-900">Admin Login</h2>
            <p className="text-sm text-slate-500 mt-1">
              {step === 'phone'
                ? 'Enter your registered phone number to continue'
                : `OTP sent to +91 ${phone}`}
            </p>
          </div>

          {step === 'phone' && (
            <form onSubmit={handleSendOtp} className="space-y-4">
              <div className="flex items-center border border-green-600 rounded-xl overflow-hidden focus-within:ring-2 focus-within:ring-green-200">
                <span className="px-4 py-3 text-slate-700 font-medium border-r border-slate-200 bg-slate-50">
                  +91
                </span>
                <span className="pl-3 text-green-600">📞</span>
                <input
                  id="phone"
                  type="tel"
                  inputMode="numeric"
                  maxLength={10}
                  placeholder="Enter phone number"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value.replace(/\D/g, ''))}
                  className="flex-1 px-3 py-3 outline-none text-slate-900 placeholder:text-slate-400"
                  autoFocus
                />
              </div>

              {error && <p className="text-sm text-red-600">{error}</p>}

              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full py-3 rounded-xl bg-green-700 hover:bg-green-800 disabled:opacity-60 text-white font-semibold flex items-center justify-center gap-2 transition-colors"
              >
                {isSubmitting ? 'Sending...' : 'Send OTP'}
                <span>→</span>
              </button>

              <div className="flex items-center gap-3 text-xs text-slate-400 py-1">
                <div className="flex-1 h-px bg-slate-200" />
                or
                <div className="flex-1 h-px bg-slate-200" />
              </div>

              <div className="border border-green-200 bg-green-50 rounded-xl py-3 text-center">
                <p className="text-green-700 font-medium text-sm flex items-center justify-center gap-1.5">
                  🛡️ Secure Login
                </p>
                <p className="text-xs text-slate-500 mt-0.5">We never share your number with anyone</p>
              </div>
            </form>
          )}

          {step === 'otp' && (
            <form onSubmit={handleVerifyOtp} className="space-y-4">
              <div>
                <label htmlFor="otp" className="block text-sm font-medium text-slate-700 mb-1.5">
                  Enter OTP
                </label>
                <input
                  id="otp"
                  type="text"
                  inputMode="numeric"
                  maxLength={6}
                  placeholder="123456"
                  value={otp}
                  onChange={(e) => setOtp(e.target.value.replace(/\D/g, ''))}
                  className="w-full px-4 py-3 rounded-xl border border-green-600 outline-none focus:ring-2 focus:ring-green-200 text-slate-900 tracking-widest text-lg text-center"
                  autoFocus
                />
              </div>

              {error && <p className="text-sm text-red-600">{error}</p>}

              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full py-3 rounded-xl bg-green-700 hover:bg-green-800 disabled:opacity-60 text-white font-semibold transition-colors"
              >
                {isSubmitting ? 'Verifying...' : 'Verify & Login'}
              </button>

              <button
                type="button"
                onClick={() => {
                  setStep('phone')
                  setOtp('')
                  setError(null)
                }}
                className="w-full text-sm text-slate-500 hover:text-slate-700 py-1"
              >
                Use a different number
              </button>
            </form>
          )}

          <p className="text-center text-xs text-slate-400 mt-6">
            © 2026 <span className="font-medium text-slate-500">GoFresh</span> Admin Panel. All rights reserved.
          </p>
        </div>
      </div>
    </div>
  )
}
