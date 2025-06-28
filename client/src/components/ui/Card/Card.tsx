import React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const cardVariants = cva(
  'rounded-xl transition-all',
  {
    variants: {
      variant: {
        default: 'bg-white border border-slate-200',
        interactive: 'bg-white border-2 border-blue-600 hover:shadow-lg cursor-pointer',
        status: 'bg-white border border-slate-200 relative overflow-hidden',
      },
      padding: {
        none: '',
        small: 'p-3',
        medium: 'p-4',
        large: 'p-6',
      },
    },
    defaultVariants: {
      variant: 'default',
      padding: 'medium',
    },
  }
)

export interface CardProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof cardVariants> {
  statusColor?: 'green' | 'red' | 'amber' | 'blue'
}

const Card = React.forwardRef<HTMLDivElement, CardProps>(
  ({ className, variant, padding, statusColor, children, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(cardVariants({ variant, padding, className }))}
        {...props}
      >
        {variant === 'status' && statusColor && (
          <div
            className={cn(
              'absolute left-0 top-0 bottom-0 w-1',
              {
                'bg-green-500': statusColor === 'green',
                'bg-red-500': statusColor === 'red',
                'bg-amber-500': statusColor === 'amber',
                'bg-blue-600': statusColor === 'blue',
              }
            )}
          />
        )}
        {children}
      </div>
    )
  }
)
Card.displayName = 'Card'

const CardHeader = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn('flex flex-col space-y-1.5', className)}
    {...props}
  />
))
CardHeader.displayName = 'CardHeader'

const CardTitle = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => (
  <h3
    ref={ref}
    className={cn('text-sm font-medium leading-none tracking-tight text-slate-900', className)}
    {...props}
  />
))
CardTitle.displayName = 'CardTitle'

const CardDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <p
    ref={ref}
    className={cn('text-xs text-slate-500', className)}
    {...props}
  />
))
CardDescription.displayName = 'CardDescription'

const CardContent = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div ref={ref} className={cn('pt-0', className)} {...props} />
))
CardContent.displayName = 'CardContent'

export { Card, CardHeader, CardTitle, CardDescription, CardContent, cardVariants }