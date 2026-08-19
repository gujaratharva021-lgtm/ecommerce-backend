const STYLES: Record<string, string> = {
  confirmed: 'bg-sky-500/15 text-sky-300',
  picking: 'bg-amber-500/15 text-amber-300',
  picked: 'bg-amber-500/15 text-amber-300',
  packing: 'bg-violet-500/15 text-violet-300',
  packed: 'bg-violet-500/15 text-violet-300',
  ready_for_dispatch: 'bg-emerald-500/15 text-emerald-300',
  handed_over: 'bg-cyan-500/15 text-cyan-300',
  shipped: 'bg-emerald-500/15 text-emerald-300',
  delivered: 'bg-emerald-500/15 text-emerald-300',
  cancelled: 'bg-rose-500/15 text-rose-300',
  returned: 'bg-rose-500/15 text-rose-300',
  pending: 'bg-slate-700 text-slate-300',
  in_progress: 'bg-amber-500/15 text-amber-300',
  completed: 'bg-emerald-500/15 text-emerald-300',
  picked_item: 'bg-emerald-500/15 text-emerald-300',
  unavailable: 'bg-rose-500/15 text-rose-300',
  short: 'bg-amber-500/15 text-amber-300',
  // Exception statuses
  open: 'bg-rose-500/15 text-rose-300',
  investigating: 'bg-amber-500/15 text-amber-300',
  resolved: 'bg-emerald-500/15 text-emerald-300',
  closed: 'bg-slate-700 text-slate-300',
  // Exception priorities
  high: 'bg-rose-500/15 text-rose-300',
  medium: 'bg-amber-500/15 text-amber-300',
  low: 'bg-slate-700 text-slate-300',
  // Substitution statuses
  approved: 'bg-emerald-500/15 text-emerald-300',
  rejected: 'bg-rose-500/15 text-rose-300',
}

export default function StatusBadge({ status }: { status: string }) {
  const style = STYLES[status] ?? 'bg-slate-700 text-slate-300'
  const label = status.replace(/_/g, ' ')
  return (
    <span className={`inline-block px-2 py-0.5 rounded-md text-xs font-medium capitalize ${style}`}>
      {label}
    </span>
  )
}
