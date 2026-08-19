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
      setError(err.response?.data?.error ?? 'Failed to send OTP.')
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
      navigate('/revenue')
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to verify OTP, or you do not have finance access.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex items-center justify-center px-4">
      <div className="w-full max-w-sm border border-slate-800 rounded-xl p-8 bg-slate-900">
        <p className="font-mono text-[10px] tracking-widest text-emerald-500 uppercase mb-1 text-center">
          Finance & Accounting
        </p>
        <h1 className="text-xl font-semibold text-center mb-6">Finance Panel Login</h1>

        {step === 'phone' && (
          <form onSubmit={handleSendOtp} className="space-y-3">
            <div>
              <label className="text-xs text-slate-400 block mb-1">Phone number</label>
              <input
                type="tel"
                inputMode="numeric"
                maxLength={10}
                placeholder="9876543210"
                value={phone}
                onChange={(e) => setPhone(e.target.value.replace(/\D/g, ''))}
                autoFocus
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>
            {error && <p className="text-red-400 text-xs">{error}</p>}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium transition-colors disabled:opacity-50"
            >
              {isSubmitting ? 'Sending...' : 'Send OTP'}
            </button>
          </form>
        )}

        {step === 'otp' && (
          <form onSubmit={handleVerifyOtp} className="space-y-3">
            <p className="text-xs text-slate-400">OTP sent to {phone}</p>
            <div>
              <label className="text-xs text-slate-400 block mb-1">Enter OTP</label>
              <input
                type="text"
                inputMode="numeric"
                maxLength={6}
                placeholder="123456"
                value={otp}
                onChange={(e) => setOtp(e.target.value.replace(/\D/g, ''))}
                autoFocus
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>
            {error && <p className="text-red-400 text-xs">{error}</p>}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium transition-colors disabled:opacity-50"
            >
              {isSubmitting ? 'Verifying...' : 'Verify & Login'}
            </button>
            <button
              type="button"
              onClick={() => { setStep('phone'); setOtp(''); setError(null) }}
              className="w-full text-center text-xs text-slate-400 hover:text-slate-200"
            >
              Use a different number
            </button>
          </form>
        )}
      </div>
    </div>
  )
}
