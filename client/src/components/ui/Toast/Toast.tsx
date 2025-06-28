import React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const toastVariants = cva(
  'flex items-start gap-3 w-full rounded-lg p-4',
  {
    variants: {
      variant: {
        success: 'bg-green-500 text-white',
        error: 'bg-red-500 text-white',
        warning: 'bg-amber-500 text-white',
        info: 'bg-blue-600 text-white',
      },
    },
    defaultVariants: {
      variant: 'info',
    },
  }
)

export interface ToastProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof toastVariants> {
  title: string
  description?: string
  icon?: React.ReactNode
}

const Toast = React.forwardRef<HTMLDivElement, ToastProps>(
  ({ className, variant, title, description, icon, ...props }, ref) => {
    const getDefaultIcon = () => {
      switch (variant) {
        case 'success':
          return (
            <div className="flex-shrink-0 w-6 h-6 bg-white rounded-full flex items-center justify-center">
              <svg className="w-4 h-4 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </div>
          )
        case 'error':
          return (
            <div className="flex-shrink-0 w-6 h-6 bg-white rounded-full flex items-center justify-center">
              <span className="text-red-500 font-bold text-sm">!</span>
            </div>
          )
        case 'warning':
          return (
            <div className="flex-shrink-0 w-6 h-6 bg-white rounded-full flex items-center justify-center">
              <span className="text-amber-500 font-bold text-sm">!</span>
            </div>
          )
        default:
          return null
      }
    }

    return (
      <div
        ref={ref}
        className={cn(toastVariants({ variant }), className)}
        {...props}
      >
        {icon || getDefaultIcon()}
        <div className="flex-1">
          <h3 className="text-sm font-medium">{title}</h3>
          {description && (
            <p className="text-xs mt-0.5 opacity-90">{description}</p>
          )}
        </div>
      </div>
    )
  }
)
Toast.displayName = 'Toast'

export { Toast, toastVariants }