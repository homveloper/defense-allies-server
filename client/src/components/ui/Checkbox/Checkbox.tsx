import React from 'react'
import { cn } from '@/lib/utils'

export interface CheckboxProps {
  checked?: boolean
  onChange?: (checked: boolean) => void
  disabled?: boolean
  label?: string
  className?: string
}

const Checkbox = React.forwardRef<HTMLButtonElement, CheckboxProps>(
  ({ checked = false, onChange, disabled = false, label, className }, ref) => {
    const handleClick = () => {
      if (!disabled && onChange) {
        onChange(!checked)
      }
    }

    return (
      <div className={cn('flex items-center gap-2.5', className)}>
        <button
          ref={ref}
          type="button"
          role="checkbox"
          aria-checked={checked}
          disabled={disabled}
          onClick={handleClick}
          className={cn(
            'h-5 w-5 rounded border transition-all focus:outline-none focus:ring-2 focus:ring-blue-600 focus:ring-offset-2',
            checked 
              ? 'bg-blue-600 border-blue-600' 
              : 'bg-white border-slate-200 hover:border-slate-300',
            disabled && 'opacity-50 cursor-not-allowed'
          )}
        >
          {checked && (
            <svg
              className="h-full w-full text-white"
              fill="none"
              viewBox="0 0 20 20"
              stroke="currentColor"
              strokeWidth={3}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M5 10l3 3L15 6"
              />
            </svg>
          )}
        </button>
        {label && (
          <label 
            className={cn(
              'text-sm text-slate-900 select-none cursor-pointer',
              disabled && 'opacity-50 cursor-not-allowed'
            )}
            onClick={handleClick}
          >
            {label}
          </label>
        )}
      </div>
    )
  }
)
Checkbox.displayName = 'Checkbox'

export { Checkbox }