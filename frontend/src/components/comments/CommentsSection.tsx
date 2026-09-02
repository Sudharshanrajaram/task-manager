import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import ReactMarkdown from 'react-markdown'
import { commentsApi } from '../../api/comments'
import { MessageSquare, Send, Trash2 } from 'lucide-react'

interface CommentsSectionProps {
  taskId: string
}

export default function CommentsSection({ taskId }: CommentsSectionProps) {
  const queryClient = useQueryClient()
  const [content, setContent] = useState('')

  const { data: comments = [], isLoading } = useQuery({
    queryKey: ['comments', taskId],
    queryFn: () => commentsApi.listByTask(taskId),
    enabled: !!taskId,
  })

  const { mutate: addComment, isPending: isAdding } = useMutation({
    mutationFn: (text: string) => commentsApi.create(taskId, text),
    onSuccess: () => {
      setContent('')
      queryClient.invalidateQueries({ queryKey: ['comments', taskId] })
    },
  })

  const { mutate: deleteComment } = useMutation({
    mutationFn: (commentId: string) => commentsApi.delete(taskId, commentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments', taskId] })
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const trimmed = content.trim()
    if (!trimmed) return
    addComment(trimmed)
  }

  const formatTimestamp = (dateStr: string) => {
    try {
      const date = new Date(dateStr)
      return date.toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    } catch {
      return dateStr
    }
  }

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm mb-6">
      <div className="flex items-center justify-between gap-2 mb-4 pb-3 border-b border-slate-100 dark:border-slate-800/60">
        <div className="flex items-center gap-2">
          <MessageSquare className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
          <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100">
            Comments & Activity
          </h2>
          <span className="text-xs text-slate-400 font-mono">({comments.length})</span>
        </div>
      </div>

      {/* Comment Form */}
      <form onSubmit={handleSubmit} className="mb-6">
        <div className="border border-slate-200 dark:border-slate-700 rounded-lg overflow-hidden focus-within:border-indigo-500 focus-within:ring-1 focus-within:ring-indigo-500 transition-all">
          <textarea
            rows={2}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Add a comment... (Markdown supported, Shift+Enter for newline)"
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                handleSubmit(e)
              }
            }}
            className="w-full p-3 text-sm bg-slate-50/50 dark:bg-slate-800/50 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 outline-none resize-none"
          />
          <div className="flex items-center justify-between px-3 py-2 bg-slate-50 dark:bg-slate-800 border-t border-slate-200 dark:border-slate-700">
            <span className="text-[11px] text-slate-400">Press Enter to send</span>
            <button
              type="submit"
              disabled={!content.trim() || isAdding}
              className="flex items-center gap-1.5 px-3 py-1 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed rounded-md transition-colors shadow-sm"
            >
              <Send className="w-3 h-3" />
              {isAdding ? 'Posting...' : 'Comment'}
            </button>
          </div>
        </div>
      </form>

      {/* Chronological Comments Feed */}
      {isLoading ? (
        <div className="space-y-3">
          <div className="h-16 bg-slate-100 dark:bg-slate-800 rounded-lg animate-pulse" />
          <div className="h-16 bg-slate-100 dark:bg-slate-800 rounded-lg animate-pulse" />
        </div>
      ) : comments.length === 0 ? (
        <div className="py-6 text-center border-2 border-dashed border-slate-200 dark:border-slate-800/80 rounded-lg">
          <p className="text-xs text-slate-400">No comments yet. Leave a log or note for this ticket.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {comments.map((comment) => (
            <div
              key={comment.id}
              className="group p-4 rounded-xl bg-slate-50 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60 transition-all hover:border-slate-300 dark:hover:border-slate-600"
            >
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-medium text-slate-500 dark:text-slate-400">
                  {formatTimestamp(comment.created_at)}
                </span>
                <button
                  onClick={() => {
                    if (confirm('Delete this comment?')) deleteComment(comment.id)
                  }}
                  className="opacity-0 group-hover:opacity-100 p-1 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950/40 rounded transition-all"
                  title="Delete comment"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
              <div className="prose prose-sm dark:prose-invert max-w-none text-slate-800 dark:text-slate-200">
                <ReactMarkdown>{comment.content}</ReactMarkdown>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
