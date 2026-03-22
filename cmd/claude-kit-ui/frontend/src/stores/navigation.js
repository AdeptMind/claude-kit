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

export function setRole(role) {
  currentRole.set(role)
  currentPage.set('dashboard')
}

export function isManagementRole(role) {
  return MANAGEMENT_ROLES.has(role)
}
