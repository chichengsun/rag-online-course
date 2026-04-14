import * as React from "react"

import { cn } from "@/lib/utils"

interface ProgressProps extends React.ComponentProps<"div"> {
  value?: number
}

/**
 * 进度条组件
 * 用于显示任务完成进度
 */
function Progress({
  className,
  value = 0,
  ...props
}: ProgressProps) {
  return (
    <div
      data-slot="progress"
      role="progressbar"
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={value}
      className={cn(
        "relative h-2 w-full overflow-hidden rounded-full bg-primary/20",
        className
      )}
      {...props}
    >
      <div
        className="h-full bg-primary transition-all duration-300 ease-in-out"
        style={{ width: `${Math.min(100, Math.max(0, value))}%` }}
      />
    </div>
  )
}

export { Progress }