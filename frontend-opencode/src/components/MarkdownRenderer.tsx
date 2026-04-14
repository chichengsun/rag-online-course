import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import type { CSSProperties } from 'react'

interface MarkdownRendererProps {
  content: string
  className?: string
}

const MarkdownRenderer = ({ content, className = '' }: MarkdownRendererProps) => {
  return (
    <div className={`markdown-body ${className}`}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          code({ className, children, ...props }) {
            const match = /language-(\w+)/.exec(className || '')
            const isInline = !match
            
            if (isInline) {
              return (
                <code 
                  className="rounded bg-muted px-1.5 py-0.5 font-mono text-sm text-primary" 
                  {...props}
                >
                  {children}
                </code>
              )
            }
            
            return (
              <SyntaxHighlighter
                style={oneDark as { [key: string]: CSSProperties }}
                language={match[1]}
                PreTag="div"
                customStyle={{
                  margin: '1em 0',
                  borderRadius: '0.5rem',
                  fontSize: '0.875rem',
                }}
              >
                {String(children).replace(/\n$/, '')}
              </SyntaxHighlighter>
            )
          },
          a({ href, children, ...props }) {
            return (
              <a 
                href={href} 
                target="_blank" 
                rel="noopener noreferrer" 
                className="text-primary underline underline-offset-4 hover:text-primary/80 transition-colors"
                {...props}
              >
                {children}
              </a>
            )
          },
        }}
      >
        {content}
      </ReactMarkdown>

      <style>{`
        .markdown-body {
          color: hsl(var(--foreground));
          line-height: 1.75;
          font-size: 1rem;
        }
        .markdown-body h1 {
          font-size: 2em;
          font-weight: 700;
          margin-top: 0;
          margin-bottom: 0.75em;
          padding-bottom: 0.3em;
          border-bottom: 1px solid hsl(var(--border));
        }
        .markdown-body h2 {
          font-size: 1.5em;
          font-weight: 600;
          margin-top: 1.5em;
          margin-bottom: 0.5em;
          padding-bottom: 0.3em;
          border-bottom: 1px solid hsl(var(--border));
        }
        .markdown-body h3 {
          font-size: 1.25em;
          font-weight: 600;
          margin-top: 1.25em;
          margin-bottom: 0.5em;
        }
        .markdown-body h4 {
          font-size: 1em;
          font-weight: 600;
          margin-top: 1em;
          margin-bottom: 0.5em;
        }
        .markdown-body p {
          margin-top: 0;
          margin-bottom: 1em;
        }
        .markdown-body ul, .markdown-body ol {
          margin-top: 0;
          margin-bottom: 1em;
          padding-left: 2em;
        }
        .markdown-body li {
          margin-bottom: 0.25em;
        }
        .markdown-body li > ul, .markdown-body li > ol {
          margin-top: 0.25em;
          margin-bottom: 0;
        }
        .markdown-body blockquote {
          margin: 0 0 1em 0;
          padding: 0.5em 1em;
          border-left: 4px solid hsl(var(--primary) / 0.5);
          background-color: hsl(var(--muted));
          color: hsl(var(--muted-foreground));
        }
        .markdown-body blockquote p:last-child {
          margin-bottom: 0;
        }
        .markdown-body table {
          border-collapse: collapse;
          width: 100%;
          margin-bottom: 1em;
          overflow: hidden;
          border-radius: 0.5rem;
          border: 1px solid hsl(var(--border));
        }
        .markdown-body th, .markdown-body td {
          padding: 0.75em 1em;
          border: 1px solid hsl(var(--border));
        }
        .markdown-body th {
          background-color: hsl(var(--muted));
          font-weight: 600;
        }
        .markdown-body tr:nth-child(even) {
          background-color: hsl(var(--muted) / 0.5);
        }
        .markdown-body img {
          max-width: 100%;
          height: auto;
          border-radius: 0.5rem;
        }
        .markdown-body hr {
          height: 1px;
          background-color: hsl(var(--border));
          border: none;
          margin: 2em 0;
        }
        .markdown-body input[type="checkbox"] {
          margin-right: 0.5em;
          width: 1em;
          height: 1em;
        }
      `}</style>
    </div>
  )
}

export default MarkdownRenderer