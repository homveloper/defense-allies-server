'use client'

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'

type Theme = 'light' | 'dark'

interface ThemeContextType {
  theme: Theme
  toggleTheme: () => void
  colors: {
    // Background colors
    bg: {
      primary: string
      secondary: string
      tertiary: string
      overlay: string
    }
    // Text colors
    text: {
      primary: string
      secondary: string
      accent: string
    }
    // Border colors
    border: {
      primary: string
      secondary: string
    }
    // Game canvas colors
    game: {
      background: string
      grid: string
      path: string
      pathHover: string
      hover: string
    }
  }
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined)

const lightTheme = {
  bg: {
    primary: '#ffffff',
    secondary: '#f8fafc',
    tertiary: '#f1f5f9',
    overlay: 'rgba(255, 255, 255, 0.9)'
  },
  text: {
    primary: '#0f172a',
    secondary: '#64748b',
    accent: '#2563eb'
  },
  border: {
    primary: '#e2e8f0',
    secondary: '#cbd5e1'
  },
  game: {
    background: '#f8fafc',
    grid: '#e2e8f0',
    path: '#ef4444',
    pathHover: '#dc2626',
    hover: 'rgba(37, 99, 235, 0.3)'
  }
}

const darkTheme = {
  bg: {
    primary: '#1f2937',
    secondary: '#111827',
    tertiary: '#374151',
    overlay: 'rgba(31, 41, 55, 0.9)'
  },
  text: {
    primary: '#f9fafb',
    secondary: '#d1d5db',
    accent: '#60a5fa'
  },
  border: {
    primary: '#374151',
    secondary: '#4b5563'
  },
  game: {
    background: '#1f2937',
    grid: '#374151',
    path: '#dc2626',
    pathHover: '#b91c1c',
    hover: 'rgba(96, 165, 250, 0.3)'
  }
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState<Theme>('light')

  useEffect(() => {
    // Load theme from localStorage
    const savedTheme = localStorage.getItem('theme') as Theme
    if (savedTheme && (savedTheme === 'light' || savedTheme === 'dark')) {
      setTheme(savedTheme)
    }
  }, [])

  useEffect(() => {
    // Save theme to localStorage
    localStorage.setItem('theme', theme)
  }, [theme])

  const toggleTheme = () => {
    setTheme(prev => prev === 'light' ? 'dark' : 'light')
  }

  const colors = theme === 'light' ? lightTheme : darkTheme

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme, colors }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const context = useContext(ThemeContext)
  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return context
}