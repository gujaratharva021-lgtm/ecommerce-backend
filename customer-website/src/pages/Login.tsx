import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { sendOTP, verifyOTP } from '../api/auth'
import { useAuth } from '../context/AuthContext'

export default function Login() {
  const [step, setStep] = useState<'phone' | 'otp'>('phone')
  const [phone, setPhone] = useState('')
  const [otp, setOtp] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { login } = useAuth()
  const navigate = useNavigate()

  async function handleSendOTP(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (!/^\d{10}$/.test(phone)) {
      setError('Enter a valid 10-digit phone number.')
      return
    }
    setIsSubmitting(true)
    try {
      await sendOTP(phone)
      setStep('otp')
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to send OTP.')
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleVerifyOTP(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    if (!/^\d{6}$/.test(otp)) {
      setError('Enter the 6-digit code.')
      return
    }
    setIsSubmitting(true)
    try {
      const res = await verifyOTP(phone, otp)
      login(res.token, res.user)
      navigate('/')
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Invalid or expired code.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="max-w-md mx-auto px-6 py-20">
      <h1 className="font-display text-4xl font-600 mb-2">Welcome back</h1>
      <p className="text-ink/60 mb-8">
        {step === 'phone'
          ? "We'll text you a code — no password needed."
          : `Enter the code sent to ${phone}`}
      </p>

      {step === 'phone' ? (
        <form onSubmit={handleSendOTP} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1.5">Phone number</label>
            <div className="flex items-center border border-line rounded-lg overflow-hidden focus-within:border-ink">
              <span className="px-3 py-2.5 bg-line/40 text-sm text-ink/60 font-mono">+91</span>
              <input
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value.replace(/\D/g, '').slice(0, 10))}
                placeholder="98765 43210"
                className="flex-1 px-3 py-2.5 outline-none bg-transparent"
                autoFocus
              />
            </div>
          </div>

          {error && <p className="text-clay text-sm">{error}</p>}

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full bg-ink text-paper font-medium py-3 rounded-lg hover:bg-marigold transition-colors disabled:opacity-50"
          >
            {isSubmitting ? 'Sending...' : 'Send code'}
          </button>
        </form>
      ) : (
        <form onSubmit={handleVerifyOTP} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1.5">6-digit code</label>
            <input
              type="text"
              inputMode="numeric"
              value={otp}
              onChange={(e) => setOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
              placeholder="123456"
              className="w-full border border-line rounded-lg px-3 py-2.5 outline-none focus:border-ink font-mono text-lg tracking-widest"
              autoFocus
            />
          </div>

          {error && <p className="text-clay text-sm">{error}</p>}

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full bg-ink text-paper font-medium py-3 rounded-lg hover:bg-marigold transition-colors disabled:opacity-50"
          >
            {isSubmitting ? 'Verifying...' : 'Verify & continue'}
          </button>

          <button
            type="button"
            onClick={() => setStep('phone')}
            className="w-full text-sm text-ink/60 hover:text-ink py-1"
          >
            Use a different number
          </button>
        </form>
      )}
    </div>
  )
}
