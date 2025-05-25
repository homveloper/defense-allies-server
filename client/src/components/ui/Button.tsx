import React from 'react';
import { cn } from '@/lib/utils';

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  size?: 'small' | 'medium' | 'large';
  fullWidth?: boolean;
  loading?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className,
      variant = 'primary',
      size = 'medium',
      fullWidth = false,
      loading = false,
      disabled,
      children,
      ...props
    },
    ref
  ) => {
    const baseStyles = [
      'inline-flex items-center justify-center rounded-full font-medium',
      'transition-colors focus-visible:outline-none focus-visible:ring-2',
      'focus-visible:ring-primary focus-visible:ring-offset-2',
      'disabled:pointer-events-none disabled:opacity-50',
      'touch-manipulation', // Better touch response
    ];

    const variants = {
      primary: [
        'bg-primary text-white hover:bg-primary/90',
        'active:bg-primary/80',
      ],
      secondary: [
        'bg-surface border border-border text-text-primary',
        'hover:bg-background active:bg-border/20',
      ],
      danger: [
        'bg-warning text-white hover:bg-warning/90',
        'active:bg-warning/80',
      ],
      ghost: [
        'text-text-primary hover:bg-background',
        'active:bg-border/20',
      ],
    };

    const sizes = {
      small: 'h-8 px-3 text-sm min-w-[2rem]',
      medium: 'h-10 px-4 text-sm min-w-[2.5rem]',
      large: 'h-12 px-6 text-base min-w-[3rem]',
    };

    const widthClass = fullWidth ? 'w-full' : '';

    return (
      <button
        className={cn(
          baseStyles,
          variants[variant],
          sizes[size],
          widthClass,
          className
        )}
        ref={ref}
        disabled={disabled || loading}
        {...props}
      >
        {loading && (
          <svg
            className="mr-2 h-4 w-4 animate-spin"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        )}
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';

export { Button };
