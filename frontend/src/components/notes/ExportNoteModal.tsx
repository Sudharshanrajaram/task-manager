import { useEffect, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import { Printer, X, Copy, Check } from 'lucide-react'

interface ExportNoteModalProps {
  content: string
  title: string
  ticketKey?: string
  ticketTitle?: string
  onClose: () => void
}

export default function ExportNoteModal({
  content,
  title,
  ticketKey,
  ticketTitle,
  onClose,
}: ExportNoteModalProps) {
  const [copied, setCopied] = useState(false)

  // Esc to close
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose])

  const handleCopy = () => {
    navigator.clipboard.writeText(content)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handlePrint = () => {
    window.print()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm print:p-0 print:bg-white animate-in fade-in duration-150">
      <div className="w-full max-w-3xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh] print:max-h-none print:border-0 print:shadow-none">
        {/* Modal Toolbar (hidden in print) */}
        <div className="flex items-center justify-between p-4 border-b border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-900 print:hidden shrink-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-slate-900 dark:text-slate-100">
              Formatted Document Export
            </span>
            {ticketKey && (
              <span className="font-mono text-xs px-2 py-0.5 rounded bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
                {ticketKey}
              </span>
            )}
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={handleCopy}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors shadow-sm"
              title="Copy markdown text"
            >
              {copied ? <Check className="w-3.5 h-3.5 text-green-500" /> : <Copy className="w-3.5 h-3.5" />}
              <span>{copied ? 'Copied' : 'Copy'}</span>
            </button>

            <button
              onClick={handlePrint}
              className="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-colors shadow-sm"
              title="Print document or save to PDF"
            >
              <Printer className="w-3.5 h-3.5" />
              <span>Print / Save PDF</span>
            </button>

            <button
              onClick={onClose}
              className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors ml-1"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Printable Document Paper */}
        <div className="p-8 md:p-12 overflow-y-auto bg-white text-slate-900 print:p-0 print:overflow-visible">
          {/* Document Header */}
          <div className="border-b border-slate-200 pb-6 mb-6">
            <div className="flex items-center justify-between text-xs text-slate-500 mb-2">
              <span className="font-semibold uppercase tracking-wider text-slate-400">TaskFlow Document</span>
              <span>{new Date().toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' })}</span>
            </div>
            <h1 className="text-2xl font-bold text-slate-900 mb-1">
              {ticketKey ? `${ticketKey}: ${ticketTitle || title}` : title}
            </h1>
            {ticketKey && (
              <p className="text-xs text-slate-500 font-mono">
                Ticket Ref: {ticketKey}
              </p>
            )}
          </div>

          {/* Rendered Markdown Body */}
          <div className="prose prose-slate max-w-none text-slate-900 leading-relaxed">
            <ReactMarkdown>
              {content || '_No notes recorded._'}
            </ReactMarkdown>
          </div>

          {/* Document Footer */}
          <div className="mt-12 pt-4 border-t border-slate-100 text-[11px] text-slate-400 flex items-center justify-between font-mono">
            <span>Generated via TaskFlow Focus Notes</span>
            <span>Confidential & Private</span>
          </div>
        </div>
      </div>
    </div>
  )
}
