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
      setError('Enter a valid 10-digit phone number (no +91, no spaces).')
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
      setError(
        err.response?.data?.error ?? err.message ?? 'Failed to verify OTP. Please try again.'
      )
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div style={styles.wrapper}>
      <div style={styles.card}>
        <h1 style={styles.title}>Admin Login</h1>

        {step === 'phone' && (
          <form onSubmit={handleSendOtp} style={styles.form}>
            <label style={styles.label} htmlFor="phone">
              Phone number
            </label>
            <input
              id="phone"
              type="tel"
              inputMode="numeric"
              maxLength={10}
              placeholder="9876543210"
              value={phone}
              onChange={(e) => setPhone(e.target.value.replace(/\D/g, ''))}
              style={styles.input}
              autoFocus
            />
            {error && <p style={styles.error}>{error}</p>}
            <button type="submit" disabled={isSubmitting} style={styles.button}>
              {isSubmitting ? 'Sending...' : 'Send OTP'}
            </button>
          </form>
        )}

        {step === 'otp' && (
          <form onSubmit={handleVerifyOtp} style={styles.form}>
            <p style={styles.hint}>OTP sent to {phone}</p>
            <label style={styles.label} htmlFor="otp">
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
              style={styles.input}
              autoFocus
            />
            {error && <p style={styles.error}>{error}</p>}
            <button type="submit" disabled={isSubmitting} style={styles.button}>
              {isSubmitting ? 'Verifying...' : 'Verify & Login'}
            </button>
            <button
              type="button"
              onClick={() => {
                setStep('phone')
                setOtp('')
                setError(null)
              }}
              style={styles.linkButton}
            >
              Use a different number
            </button>
          </form>
        )}
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  wrapper: {
    minHeight: '100vh',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    background: '#f4f5f7',
  },
  card: {
    width: 360,
    padding: '32px',
    borderRadius: 12,
    background: '#fff',
    boxShadow: '0 4px 24px rgba(0,0,0,0.08)',
  },
  title: {
    fontSize: 20,
    fontWeight: 600,
    marginBottom: 24,
    textAlign: 'center',
  },
  form: {
    display: 'flex',
    flexDirection: 'column',
    gap: 8,
  },
  label: {
    fontSize: 13,
    fontWeight: 500,
    color: '#374151',
  },
  input: {
    padding: '10px 12px',
    borderRadius: 8,
    border: '1px solid #d1d5db',
    fontSize: 15,
    marginBottom: 8,
  },
  button: {
    padding: '10px 12px',
    borderRadius: 8,
    border: 'none',
    background: '#111827',
    color: '#fff',
    fontWeight: 600,
    cursor: 'pointer',
    marginTop: 8,
  },
  linkButton: {
    background: 'none',
    border: 'none',
    color: '#6b7280',
    fontSize: 13,
    cursor: 'pointer',
    marginTop: 8,
  },
  hint: {
    fontSize: 13,
    color: '#6b7280',
    marginBottom: 8,
  },
  error: {
    color: '#dc2626',
    fontSize: 13,
    margin: '4px 0',
  },
}
