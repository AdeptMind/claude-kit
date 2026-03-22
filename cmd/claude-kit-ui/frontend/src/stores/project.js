import { writable } from 'svelte/store'

const STORAGE_KEY = 'ck-recent-projects'
const MAX_RECENTS = 10

export const currentProject = writable(null)
export const recentProjects = writable([])

export function setProject(project) {
  currentProject.set(project)
  recentProjects.update(list => {
    const filtered = list.filter(p => p.path !== project.path)
    const updated = [project, ...filtered].slice(0, MAX_RECENTS)
    try { localStorage.setItem(STORAGE_KEY, JSON.stringify(updated)) } catch {}
    return updated
  })
}

export function removeProject(path) {
  recentProjects.update(list => {
    const updated = list.filter(p => p.path !== path)
    try { localStorage.setItem(STORAGE_KEY, JSON.stringify(updated)) } catch {}
    return updated
  })
  // If we removed the current project, switch to the first remaining or null
  currentProject.update(current => {
    if (current?.path === path) {
      let first = null
      recentProjects.subscribe(list => { first = list[0] || null })()
      return first
    }
    return current
  })
}

export function loadRecents() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw)
      recentProjects.set(parsed)
      if (parsed.length > 0) currentProject.set(parsed[0])
    }
  } catch {}
}
