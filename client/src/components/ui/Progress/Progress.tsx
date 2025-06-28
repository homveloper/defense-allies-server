import React from 'react'
import { cn } from '@/lib/utils'

export interface ProgressProps {
  value: number
  max?: number
  variant?: 'default' | 'health' | 'experience'
  label?: string
  showValue?: boolean
  className?: string
}

const Progress = React.forwardRef<HTMLDivElement, ProgressProps>(
  ({ 
    value, 
    max = 100, 
    variant = 'default', 
    label, 
    showValue = false,
    className 
  }, ref) => {
    const percentage = Math.min(Math.max((value / max) * 100, 0), 100)
    
    const getBarColor = () => {
      switch (variant) {
        case 'health':
          if (percentage > 60) return 'bg-green-500'
          if (percentage > 30) return 'bg-amber-500'
          return 'bg-red-500'
        case 'experience':
          return 'bg-blue-600'
        default:
          return 'bg-blue-600'
      }
    }

    return (
      <div ref={ref} className={cn('w-full', className)}>
        <div 
          className={cn(
            'h-2 w-full bg-slate-200 rounded',
            variant === 'health' && 'h-3'
          )}
        >
          <div
            className={cn(
              'h-full rounded transition-all duration-300',
              getBarColor()
            )}
            style={{ width: `${percentage}%` }}
          />
        </div>
        {(label || showValue) && (
          <p className="mt-1.5 text-xs text-slate-500">
            {label && <span>{label}: </span>}
            {showValue && <span>{value}/{max}</span>}
            {percentage !== Math.round(percentage) ? ` (${percentage.toFixed(1)}%)` : ` (${percentage}%)`}
          </p>
        )}
      </div>
    )
  }
)
Progress.displayName = 'Progress'

export { Progress }