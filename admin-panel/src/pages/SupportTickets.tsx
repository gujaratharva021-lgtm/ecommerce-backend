import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import {
  listSupportTickets,
  getSupportTicketMessages,
  replyToSupportTicket,
  updateSupportTicketStatus,
} from '../api/admin'
import type { SupportTicket, SupportMessage } from '../types/admin'

const STATUS_OPTIONS = ['open', 'in_progress', 'resolved', 'closed']

const STATUS_STYLES: Record<string, string> = {
  open: 'bg-amber-500/15 text-amber-300',
  in_progress: 'bg-indigo-500/15 text-indigo-300',
  resolved: 'bg-emerald-500/15 text-emerald-300',
  closed: 'bg-slate-700/50 text-slate-400',
}

export default function SupportTickets() {
  const [tickets, setTickets] = useState<SupportTicket[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState('')

  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [messages, setMessages] = useState<SupportMessage[]>([])
  const [isLoadingThread, setIsLoadingThread] = useState(false)
  const [replyText, setReplyText] = useState('')
  const [isSending, setIsSending] = useState(false)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listSupportTickets(statusFilter || undefined)
      setTickets(res ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load tickets.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [statusFilter])

  async function openTicket(id: number) {
    setSelectedId(id)
    setIsLoadingThread(true)
    try {
      const res = await getSupportTicketMessages(id)
      setMessages(res.messages ?? [])
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to load ticket thread.')
    } finally {
      setIsLoadingThread(false)
    }
  }

  async function handleReply() {
    if (!selectedId || !replyText.trim()) return
    setIsSending(true)
    try {
      const msg = await replyToSupportTicket(selectedId, replyText.trim())
      setMessages((prev) => [...prev, msg])
      setReplyText('')
      setTickets((prev) =>
        prev.map((t) => (t.id === selectedId && t.status === 'open' ? { ...t, status: 'in_progress' } : t))
      )
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to send reply.')
    } finally {
      setIsSending(false)
    }
  }

  async function handleStatusChange(id: number, status: string) {
    try {
      await updateSupportTicketStatus(id, status)
      setTickets((prev) => prev.map((t) => (t.id === id ? { ...t, status: status as any } : t)))
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to update ticket status.')
    }
  }

  const selectedTicket = tickets.find((t) => t.id === selectedId)

  return (
    <Layout>
      <div className="p-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold">Customer Support</h1>
            <p className="text-sm text-slate-400 mt-1">Customer tickets and support threads</p>
          </div>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
          >
            <option value="">All statuses</option>
            {STATUS_OPTIONS.map((s) => (
              <option key={s} value={s}>
                {s.replace('_', ' ')}
              </option>
            ))}
          </select>
        </div>

        <div className="grid grid-cols-5 gap-6">
          <div className="col-span-2">
            {isLoading && <p className="text-slate-400">Loading...</p>}
            {error && <p className="text-red-400">{error}</p>}

            {!isLoading && !error && tickets.length === 0 && (
              <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
                No tickets found.
              </div>
            )}

            {!isLoading && tickets.length > 0 && (
              <div className="border border-slate-800 rounded-xl overflow-hidden divide-y divide-slate-800">
                {tickets.map((t) => (
                  <button
                    key={t.id}
                    onClick={() => openTicket(t.id)}
                    className={
                      'w-full text-left px-4 py-3 hover:bg-slate-900 transition-colors ' +
                      (selectedId === t.id ? 'bg-slate-900' : '')
                    }
                  >
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm text-slate-200 font-medium truncate">{t.subject}</span>
                      <span
                        className={
                          'px-2 py-0.5 rounded-md text-xs font-medium shrink-0 ml-2 ' +
                          (STATUS_STYLES[t.status] ?? '')
                        }
                      >
                        {t.status.replace('_', ' ')}
                      </span>
                    </div>
                    <div className="text-xs text-slate-500">
                      Ticket #{t.id} &middot; {new Date(t.created_at).toLocaleDateString()}
                      {t.order_id ? ` \u00b7 Order #${t.order_id}` : ''}
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>

          <div className="col-span-3">
            {!selectedId && (
              <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500 h-full flex items-center justify-center">
                Select a ticket to view the conversation
              </div>
            )}

            {selectedId && (
              <div className="border border-slate-800 rounded-xl overflow-hidden flex flex-col h-full">
                <div className="px-4 py-3 bg-slate-900 border-b border-slate-800 flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-slate-200">{selectedTicket?.subject}</p>
                    <p className="text-xs text-slate-500">Ticket #{selectedId}</p>
                  </div>
                  <select
                    value={selectedTicket?.status ?? ''}
                    onChange={(e) => handleStatusChange(selectedId, e.target.value)}
                    className="bg-slate-800 border border-slate-700 rounded-lg px-2 py-1.5 text-xs"
                  >
                    {STATUS_OPTIONS.map((s) => (
                      <option key={s} value={s}>
                        {s.replace('_', ' ')}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="p-4 space-y-3 max-h-96 overflow-y-auto">
                  {isLoadingThread && <p className="text-slate-400 text-sm">Loading thread...</p>}
                  {!isLoadingThread &&
                    messages.map((m) => (
                      <div
                        key={m.id}
                        className={
                          'max-w-[80%] rounded-lg px-3 py-2 text-sm ' +
                          (m.sender_type === 'admin'
                            ? 'ml-auto bg-indigo-500/15 text-indigo-100'
                            : 'bg-slate-800 text-slate-200')
                        }
                      >
                        <p>{m.message}</p>
                        <p className="text-xs text-slate-500 mt-1">
                          {m.sender_type === 'admin' ? 'You' : 'Customer'} &middot;{' '}
                          {new Date(m.created_at).toLocaleString()}
                        </p>
                      </div>
                    ))}
                </div>

                <div className="p-4 border-t border-slate-800 flex gap-2">
                  <input
                    type="text"
                    value={replyText}
                    onChange={(e) => setReplyText(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleReply()}
                    placeholder="Type a reply..."
                    className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-slate-500"
                  />
                  <button
                    onClick={handleReply}
                    disabled={isSending || !replyText.trim()}
                    className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium disabled:opacity-40 transition-colors"
                  >
                    Send
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </Layout>
  )
}
