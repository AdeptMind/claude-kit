import { writable } from 'svelte/store'

export const currentPage = writable('dashboard')

export function navigateTo(page) {
  currentPage.set(page)
}
