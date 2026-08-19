import { useState } from 'react'
import type { FormEvent } from 'react'
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
      setError(err.response?.data?.error ?? 'Failed to verify OTP. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex items-center justify-center px-4 relative overflow-hidden">
      <div className="absolute top-[-10%] left-[-10%] w-[500px] h-[500px] bg-red-500/20 rounded-full blur-[120px] pointer-events-none" />
      <div className="absolute bottom-[-10%] right-[-10%] w-[500px] h-[500px] bg-orange-600/10 rounded-full blur-[120px] pointer-events-none" />

      <div className="w-full max-w-sm relative">
        <div className="flex flex-col items-center mb-8">
          <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-red-400 to-orange-600 flex items-center justify-center mb-4 shadow-lg shadow-red-500/20">
            <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4Z" />
              <path d="M3 6h18" />
              <path d="M16 10a4 4 0 0 1-8 0" />
            </svg>
          </div>
          <p className="font-mono text-[10px] tracking-widest text-red-500 uppercase mb-1">
            Store Staff App
          </p>
          <h1 className="font-display text-2xl font-semibold">Welcome back</h1>
          <p className="text-xs text-slate-500 mt-1">Sign in to manage your store operations</p>
        </div>

        <div className="border border-slate-800 rounded-2xl p-8 bg-slate-900/80 backdrop-blur-sm shadow-xl">
          {step === 'phone' && (
            <form onSubmit={handleSendOtp} className="space-y-4">
              <div>
                <label className="text-xs text-slate-400 block mb-1.5">Phone number</label>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 text-sm">+91</span>
                  <input
                    type="tel"
                    inputMode="numeric"
                    maxLength={10}
                    placeholder="9876543210"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value.replace(/\D/g, ''))}
                    autoFocus
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg pl-11 pr-3 py-2.5 text-sm focus:outline-none focus:border-red-500 focus:ring-1 focus:ring-red-500/50 transition-colors"
                  />
                </div>
              </div>
              {error && (
                <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-xs rounded-lg px-3 py-2">
                  {error}
                </div>
              )}
              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full py-2.5 rounded-lg bg-gradient-to-r from-red-500 to-orange-600 hover:from-red-400 hover:to-orange-500 text-white text-sm font-semibold transition-all shadow-lg shadow-red-500/20 disabled:opacity-50"
              >
                {isSubmitting ? 'Sending...' : 'Send OTP'}
              </button>
            </form>
          )}

          {step === 'otp' && (
            <form onSubmit={handleVerifyOtp} className="space-y-4">
              <p className="text-xs text-slate-400">
                OTP sent to <span className="text-red-400 font-medium">{phone}</span>
              </p>
              <div>
                <label className="text-xs text-slate-400 block mb-1.5">Enter OTP</label>
                <input
                  type="text"
                  inputMode="numeric"
                  maxLength={6}
                  placeholder="â€¢â€¢â€¢â€¢â€¢â€¢"
                  value={otp}
                  onChange={(e) => setOtp(e.target.value.replace(/\D/g, ''))}
                  autoFocus
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2.5 text-sm tracking-[0.5em] text-center focus:outline-none focus:border-red-500 focus:ring-1 focus:ring-red-500/50 transition-colors"
                />
              </div>
              {error && (
                <div className="border border-rose-900 bg-rose-950/40 text-rose-300 text-xs rounded-lg px-3 py-2">
                  {error}
                </div>
              )}
              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full py-2.5 rounded-lg bg-gradient-to-r from-red-500 to-orange-600 hover:from-red-400 hover:to-orange-500 text-white text-sm font-semibold transition-all shadow-lg shadow-red-500/20 disabled:opacity-50"
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
                className="w-full text-center text-xs text-slate-400 hover:text-red-400 transition-colors"
              >
                Use a different number
              </button>
            </form>
          )}
        </div>
      </div>
    </div>
  )
}
