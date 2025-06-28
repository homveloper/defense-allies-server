import React from 'react'
import { cn } from '@/lib/utils'

export interface ToggleProps {
  checked?: boolean
  onChange?: (checked: boolean) => void
  disabled?: boolean
  label?: string
  className?: string
}

const Toggle = React.forwardRef<HTMLButtonElement, ToggleProps>(
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
          role="switch"
          aria-checked={checked}
          disabled={disabled}
          onClick={handleClick}
          className={cn(
            'relative inline-block w-10 h-6 rounded-[34px] transition-colors duration-300 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:ring-offset-2 cursor-pointer',
            checked ? 'bg-blue-600' : 'bg-slate-300',
            disabled && 'opacity-50 cursor-not-allowed'
          )}
        >
          <span
            className={cn(
              'absolute w-5 h-5 bg-white rounded-full transition-transform duration-300',
              'shadow-[0px_2px_5px_0px_rgba(0,0,0,0.3)] top-0.5 left-0.5',
              checked ? 'transform translate-x-4' : 'transform translate-x-0'
            )}
          />
        </button>
        {label && (
          <label 
            className={cn(
              'text-sm text-slate-900 select-none cursor-pointer',
              disabled && 'opacity-50 cursor-not-allowed'
            )}
            onClick={!disabled ? handleClick : undefined}
          >
            {label}
          </label>
        )}
      </div>
    )
  }
)
Toggle.displayName = 'Toggle'

export { Toggle }