import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ChatMessage, ChatMessageList } from './ChatMessage'
import type { ChatMessageData } from './ChatMessage'

// Mock scrollIntoView for jsdom
beforeEach(() => {
  Element.prototype.scrollIntoView = () => {}
})

describe('ChatMessage', () => {
  const userMessage: ChatMessageData = {
    id: '1',
    role: 'user',
    content: 'Hello, this is a test message',
    createdAt: '2024-01-01T12:00:00Z',
  }

  const assistantMessage: ChatMessageData = {
    id: '2',
    role: 'assistant',
    content: 'This is an **AI response** with *markdown*',
    createdAt: '2024-01-01T12:01:00Z',
  }

  const assistantMessageWithRefs: ChatMessageData = {
    id: '3',
    role: 'assistant',
    content: 'Here is the answer with references',
    createdAt: '2024-01-01T12:02:00Z',
    references: [
      {
        citationNo: 1,
        resourceTitle: 'Test Resource',
        chunkIndex: 0,
        snippet: 'This is a test snippet',
        fullContent: 'Full content here',
      },
    ],
  }

  it('renders user message correctly', () => {
    render(<ChatMessage message={userMessage} />)
    
    expect(screen.getByText('你')).toBeInTheDocument()
    expect(screen.getByText('Hello, this is a test message')).toBeInTheDocument()
  })

  it('renders assistant message correctly', () => {
    render(<ChatMessage message={assistantMessage} />)
    
    expect(screen.getByText('助手')).toBeInTheDocument()
  })

  it('renders message with references', () => {
    render(<ChatMessage message={assistantMessageWithRefs} />)
    
    expect(screen.getByText('引用内容')).toBeInTheDocument()
    expect(screen.getByText(/\[1\] Test Resource/)).toBeInTheDocument()
  })

  it('shows streaming indicator when isStreaming is true', () => {
    const streamingMessage: ChatMessageData = {
      ...assistantMessage,
      isStreaming: true,
    }
    
    render(<ChatMessage message={streamingMessage} />)
    
    expect(screen.getByText('正在输入...')).toBeInTheDocument()
  })
})

describe('ChatMessageList', () => {
  const messages: ChatMessageData[] = [
    {
      id: '1',
      role: 'user',
      content: 'First message',
      createdAt: '2024-01-01T12:00:00Z',
    },
    {
      id: '2',
      role: 'assistant',
      content: 'Second message',
      createdAt: '2024-01-01T12:01:00Z',
    },
  ]

  it('renders empty state when no messages', () => {
    render(<ChatMessageList messages={[]} />)
    
    expect(screen.getByText('开始提问吧')).toBeInTheDocument()
  })

  it('renders all messages', () => {
    render(<ChatMessageList messages={messages} />)
    
    expect(screen.getByText('First message')).toBeInTheDocument()
    expect(screen.getByText('Second message')).toBeInTheDocument()
  })
})