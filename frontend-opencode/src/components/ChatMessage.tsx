import { useState, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import MarkdownRenderer from './MarkdownRenderer'
import { ChevronDown, ChevronRight, FileText } from 'lucide-react'

/**
 * 引用块类型定义
 */
export interface Reference {
  citationNo?: number
  resourceTitle: string
  chunkIndex?: number
  snippet: string
  fullContent?: string
}

/**
 * 消息类型定义
 */
export interface ChatMessageData {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  createdAt?: string
  references?: Reference[]
  isStreaming?: boolean
}

/**
 * ChatMessage 组件属性
 */
export interface ChatMessageProps {
  message: ChatMessageData
  className?: string
  onReferenceClick?: (reference: Reference) => void
}

function ReferenceBlock({ 
  reference, 
  index,
  onExpand 
}: { 
  reference: Reference
  index: number
  onExpand?: (reference: Reference) => void 
}) {
  const [isExpanded, setIsExpanded] = useState(false)
  const citationNo = reference.citationNo ?? index + 1
  
  return (
    <div className="border border-border rounded-lg overflow-hidden bg-muted/30">
      <button
        type="button"
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center gap-2 px-3 py-2 hover:bg-muted/50 transition-colors text-left"
      >
        {isExpanded ? (
          <ChevronDown className="size-4 text-muted-foreground" />
        ) : (
          <ChevronRight className="size-4 text-muted-foreground" />
        )}
        <FileText className="size-4 text-primary" />
        <span className="flex-1 text-sm font-medium">
          [{citationNo}] {reference.resourceTitle}
          {reference.chunkIndex !== undefined && (
            <span className="text-muted-foreground ml-1">
              #{reference.chunkIndex + 1}
            </span>
          )}
        </span>
      </button>
      
      {isExpanded && (
        <div className="px-3 py-2 border-t border-border bg-background">
          <p className="text-sm text-muted-foreground leading-relaxed">
            {reference.snippet}
          </p>
          {reference.fullContent && (
            <button
              type="button"
              onClick={() => onExpand?.(reference)}
              className="mt-2 text-xs text-primary hover:underline"
            >
              查看完整内容
            </button>
          )}
        </div>
      )}
    </div>
  )
}

function StreamingCursor() {
  return (
    <span className="inline-block w-2 h-4 bg-primary animate-pulse ml-0.5" />
  )
}

export function ChatMessage({ 
  message, 
  className,
  onReferenceClick 
}: ChatMessageProps) {
  const messageEndRef = useRef<HTMLDivElement>(null)
  const isUser = message.role === 'user'
  const isAssistant = message.role === 'assistant'
  const isSystem = message.role === 'system'
  
  useEffect(() => {
    if (message.isStreaming) {
      messageEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
  }, [message.content, message.isStreaming])
  
  const formattedTime = message.createdAt 
    ? new Date(message.createdAt).toLocaleString('zh-CN', {
        hour: '2-digit',
        minute: '2-digit',
      })
    : ''
  
  return (
    <article
      className={cn(
        'flex flex-col gap-2',
        isUser && 'items-end',
        (isAssistant || isSystem) && 'items-start',
        className
      )}
    >
      <div className={cn(
        'flex items-center gap-2 text-xs text-muted-foreground',
        isUser && 'flex-row-reverse'
      )}>
        <span className="font-medium">
          {isUser ? '你' : isAssistant ? '助手' : '系统'}
        </span>
        {formattedTime && (
          <span className="opacity-60">{formattedTime}</span>
        )}
        {message.isStreaming && (
          <span className="flex items-center gap-1">
            <span className="w-1.5 h-1.5 bg-primary rounded-full animate-pulse" />
            <span>正在输入...</span>
          </span>
        )}
      </div>
      
      <div
        className={cn(
          'max-w-[80%] rounded-2xl px-4 py-3',
          isUser && 'bg-primary text-primary-foreground rounded-br-md',
          isAssistant && 'bg-muted text-foreground rounded-bl-md',
          isSystem && 'bg-secondary text-secondary-foreground rounded-bl-md'
        )}
      >
        {isUser ? (
          <pre className="whitespace-pre-wrap font-sans text-sm leading-relaxed">
            {message.content}
          </pre>
        ) : isAssistant ? (
          <div className="text-sm leading-relaxed">
            <MarkdownRenderer content={message.content} />
            {message.isStreaming && <StreamingCursor />}
          </div>
        ) : (
          <p className="text-sm leading-relaxed">{message.content}</p>
        )}
      </div>
      
      {isAssistant && message.references && message.references.length > 0 && (
        <div className="max-w-[80%] w-full space-y-2">
          <div className="text-xs font-medium text-muted-foreground mb-2">
            引用内容
          </div>
          <div className="space-y-2">
            {message.references.map((ref, idx) => (
              <ReferenceBlock
                key={idx}
                reference={ref}
                index={idx}
                onExpand={onReferenceClick}
              />
            ))}
          </div>
        </div>
      )}
      
      <div ref={messageEndRef} />
    </article>
  )
}

export function ChatMessageList({ 
  messages,
  className,
  onReferenceClick 
}: { 
  messages: ChatMessageData[]
  className?: string
  onReferenceClick?: (reference: Reference) => void
}) {
  const listEndRef = useRef<HTMLDivElement>(null)
  
  useEffect(() => {
    listEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])
  
  return (
    <div className={cn('flex flex-col gap-4 p-4 overflow-y-auto', className)}>
      {messages.length === 0 ? (
        <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
          开始提问吧
        </div>
      ) : (
        messages.map((msg) => (
          <ChatMessage
            key={msg.id}
            message={msg}
            onReferenceClick={onReferenceClick}
          />
        ))
      )}
      <div ref={listEndRef} />
    </div>
  )
}

export default ChatMessage