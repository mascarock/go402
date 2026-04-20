import { useEffect, useState } from 'react'

export type Theme = 'dark' | 'light' | 'system'

const KEY = 'theme'

function apply(theme: Theme) {
  const root = document.documentElement
  root.classList.remove('theme-dark', 'theme-light', 'theme-system')
  root.classList.add('theme-' + theme)
}

export function useTheme(): [Theme, (t: Theme) => void] {
  const [theme, setThemeState] = useState<Theme>(() => {
    return (localStorage.getItem(KEY) as Theme) || 'dark'
  })

  useEffect(() => {
    apply(theme)
    localStorage.setItem(KEY, theme)
  }, [theme])

  return [theme, setThemeState]
}
