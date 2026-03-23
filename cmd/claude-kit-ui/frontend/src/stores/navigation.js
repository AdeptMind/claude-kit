import { writable } from 'svelte/store'

export const currentPage = writable('dashboard')
export const currentRole = writable('po')

const MANAGEMENT_ROLES = new Set([
  'po', 'business-analyst', 'project-manager', 'scrum-master',
  'people-ops', 'change-manager', 'architect', 'ux-designer',
  'finops', 'soc2-compliance', 'access-review', 'data-governance',
])

export function navigateTo(page) {
  currentPage.set(page)
}

export function setRole(role, projectPath) {
  currentRole.set(role)
  currentPage.set('dashboard')
  // Persist role to project settings
  if (projectPath) {
    import('../../wailsjs/go/main/RoleService.js')
      .then(({ SetRole }) => SetRole(projectPath, role))
      .catch(() => {})
  }
}

export function isManagementRole(role) {
  return MANAGEMENT_ROLES.has(role)
}

export async function refreshRole(projectPath) {
  if (!projectPath) return
  try {
    const { GetCurrent } = await import('../../wailsjs/go/main/RoleService.js')
    const info = await GetCurrent(projectPath)
    if (info?.name) {
      currentRole.set(info.name)
    }
  } catch {}
}
