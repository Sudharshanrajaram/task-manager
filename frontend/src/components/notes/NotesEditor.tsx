import { useState, useEffect, useRef } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { notesApi } from '../../api/notes'
import { Link } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import { Edit3, Eye, CheckCircle2, Loader2 } from 'lucide-react'

interface NotesEditorProps {
  taskId?: string // If undefined, edits global scratchpad
  title?: string
}

export default function NotesEditor({ taskId, title = 'Notes' }: NotesEditorProps) {
  const [content, setContent] = useState('')
  const [activeTab, setActiveTab] = useState<'edit' | 'preview'>('edit')
  const [isSaved, setIsSaved] = useState(true)
  const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Fetch initial note content
  const { data: noteData, isLoading } = useQuery({
    queryKey: ['notes', taskId || 'scratchpad'],
    queryFn: () => (taskId ? notesApi.getTaskNote(taskId) : notesApi.getScratchpad()),
  })

  useEffect(() => {
    if (noteData?.content !== undefined) {
      setContent(noteData.content)
      setIsSaved(true)
    }
  }, [noteData])

  // Mutation to persist note
  const { mutate: saveNote, isPending: isSaving } = useMutation({
    mutationFn: (newContent: string) =>
      taskId ? notesApi.saveTaskNote(taskId, newContent) : notesApi.saveScratchpad(newContent),
    onSuccess: () => {
      setIsSaved(true)
    },
  })

  // Debounced auto-save handler
  const handleContentChange = (newContent: string) => {
    setContent(newContent)
    setIsSaved(false)

    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current)
    }

    debounceTimerRef.current = setTimeout(() => {
      saveNote(newContent)
    }, 1000)
  }

  // Backlink transformer: converts [[KEY-1]] to markdown link
  const renderableMarkdown = content.replace(/\[\[([A-Z0-9]+-\d+)\]\]/g, '[$1](/tasks/$1)')

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm flex flex-col h-full">
      <div className="flex items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800 mb-3">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-slate-900 dark:text-slate-100">{title}</span>
          <span className="text-[11px] text-slate-400 font-mono">Obsidian Markdown</span>
        </div>

        <div className="flex items-center gap-3">
          {/* Save status */}
          <div className="flex items-center gap-1 text-[11px] text-slate-400">
            {isSaving ? (
              <>
                <Loader2 className="w-3 h-3 animate-spin text-indigo-500" />
                <span>Saving...</span>
              </>
            ) : isSaved ? (
              <>
                <CheckCircle2 className="w-3 h-3 text-green-500" />
                <span>Saved</span>
              </>
            ) : (
              <span className="text-amber-500">Unsaved changes...</span>
            )}
          </div>

          {/* Edit / Preview Tabs */}
          <div className="flex gap-1 p-0.5 bg-slate-100 dark:bg-slate-800 rounded-lg text-xs">
            <button
              onClick={() => setActiveTab('edit')}
              className={`flex items-center gap-1 px-2.5 py-1 rounded-md transition-colors ${
                activeTab === 'edit'
                  ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm'
                  : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'
              }`}
            >
              <Edit3 className="w-3 h-3" />
              <span>Edit</span>
            </button>
            <button
              onClick={() => setActiveTab('preview')}
              className={`flex items-center gap-1 px-2.5 py-1 rounded-md transition-colors ${
                activeTab === 'preview'
                  ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm'
                  : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'
              }`}
            >
              <Eye className="w-3 h-3" />
              <span>Preview</span>
            </button>
          </div>
        </div>
      </div>

      {isLoading ? (
        <div className="h-40 bg-slate-100 dark:bg-slate-800/40 rounded-lg animate-pulse" />
      ) : activeTab === 'edit' ? (
        <textarea
          value={content}
          onChange={(e) => handleContentChange(e.target.value)}
          placeholder={`Write markdown notes here...\n\n- Bullet points\n- [[TICKET-KEY]] backlinks to other tickets\n- Code snippets`}
          className="w-full flex-1 min-h-[160px] p-3 text-sm font-mono bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-800 rounded-lg text-slate-800 dark:text-slate-200 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none leading-relaxed"
        />
      ) : (
        <div className="flex-1 min-h-[160px] p-3 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 overflow-y-auto prose dark:prose-invert prose-sm max-w-none">
          {content.trim() ? (
            <ReactMarkdown
              components={{
                a: ({ href, children }) => (
                  <Link
                    to={href || '#'}
                    className="text-indigo-600 dark:text-indigo-400 underline font-semibold hover:text-indigo-700"
                  >
                    {children}
                  </Link>
                ),
              }}
            >
              {renderableMarkdown}
            </ReactMarkdown>
          ) : (
            <span className="text-xs text-slate-400 italic">No notes written yet. Switch to Edit to write notes.</span>
          )}
        </div>
      )}
    </div>
  )
}
