export const API_BASE = import.meta.env.VITE_API_BASE || ''

import type { Mod } from '@/types/mod'

export const fetchWithAuth = async (url: string, options: RequestInit = {}): Promise<Response> => {
  const username = localStorage.getItem('username') || ''
  const password = localStorage.getItem('password') || ''
  const headers = new Headers(options.headers || {})

  if (!headers.has('Authorization')) {
    if (username === '' && password === '') {
      throw Error('No logged in')
    }
    headers.set('Authorization', 'Basic ' + btoa(`${username}:${password}`))
  }
  if (options.body && !headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  return fetch(url, {
    ...options,
    headers,
  })
}

// ----------------------------------------------------------------------------
// FSM Admins
// ----------------------------------------------------------------------------

export const addAdmin = async (username: string, password: string) => {
  const res = await fetchWithAuth(`${API_BASE}/admins`, {
    method: 'POST',
    body: JSON.stringify({ username: encodeURIComponent(username), password }),
  })
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to add admin user.' + (message ? `\n${message}` : ''))
  }
}

export const deleteAdmin = async (username: string) => {
  const res = await fetchWithAuth(`${API_BASE}/admins/${encodeURIComponent(username)}`, {
    method: 'DELETE',
  })
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to delete admin user.' + (message ? `\n${message}` : ''))
  }
}

export const fetchAdmins = async () => {
  const res = await fetchWithAuth(`${API_BASE}/admins`)
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to get admin users.' + (message ? `\n${message}` : ''))
  }
  return await res.json() || {}
}

export const updateAdmin = async (username: string, password: string) => {
  const res = await fetchWithAuth(`${API_BASE}/admins/${username}`, {
    method: 'POST',
    body: JSON.stringify({ password }),
  })
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to update admin user.' + (message ? `\n${message}` : ''))
  }
}

// ----------------------------------------------------------------------------
// Saves
// ----------------------------------------------------------------------------

export const deleteSave = async (name: string) => {
  const res = await fetchWithAuth(`${API_BASE}/saves/${encodeURIComponent(name)}`, {
    method: 'DELETE',
  })
  if (!res.ok) throw new Error('Failed to delete save')
}

export const fetchCurrentSave = async () => {
  const res = await fetchWithAuth(`${API_BASE}/settings`)
  if (!res.ok) throw new Error('Failed to load current save')
  return await res.json()
}

export const fetchSaves = async () => {
  const res = await fetchWithAuth(`${API_BASE}/saves`)
  if (!res.ok) throw new Error('Failed to load saves')
  return await res.json()
}

export const updateSave = async (save: string) => {
  const res = await fetchWithAuth(`${API_BASE}/settings/save`, {
    method: 'POST',
    body: JSON.stringify({ save }),
  })
  if (!res.ok) throw new Error('Failed to update save')
}

export const uploadSave = async (file: File) => {
  const formData = new FormData()
  formData.append('save', file)

  const res = await fetchWithAuth(`${API_BASE}/saves`, {
    method: 'POST',
    body: formData,
  })
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to upload save.' + (message ? `\n${message}` : ''))
  }
}

// ----------------------------------------------------------------------------
// Factorio User
// ----------------------------------------------------------------------------

export const getFactorioUser = async () => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-user`)
  if (!res.ok) throw new Error('Failed to update factorio user')
  return await res.json()
}

export const updateFactorioUser = async (username: string, token: string) => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-user`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: JSON.stringify({ username, token }),
  })
  if (!res.ok) throw new Error('Failed to update factorio user')
}

// ----------------------------------------------------------------------------
// RCON
// ----------------------------------------------------------------------------

export const sendRconCommand = async (command: string): Promise<string> => {
  const res = await fetchWithAuth(`${API_BASE}/rcon`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: new URLSearchParams({ command }),
  })

  if (!res.ok) {
    const error = await res.text()
    throw new Error(`RCON error: ${error}`)
  }

  const data = await res.json()
  return data.output || data.error || 'No response'
}

// ----------------------------------------------------------------------------
// Mods
// ----------------------------------------------------------------------------

export const fetchMods = async () => {
  const res = await fetchWithAuth(`${API_BASE}/mods`)
  const data = await res.json()
  return data.mods || []
}

export const toggleMod = async (mod: Mod) => {
  const url = `${API_BASE}/toggle-mod?mod=${mod.name}&enabled=${!mod.enabled}`
  const res = await fetchWithAuth(url, { method: 'POST' })
  const data = await res.json()
  return data.mods || []
}

export const fetchBookmarkedMods = async () => {
  const res = await fetchWithAuth(`${API_BASE}/mods/bookmarked`)
  const data = await res.json()
  return data || {}
}

export const downloadMod = async (mod: string, version: string) => {
  try {
    const res = await fetchWithAuth(`${API_BASE}/mods/download/${mod}/${version}`)
    if (!res.ok) {
      const message = (await res.json())?.message || null
      throw Error(message ? message : 'Failed to download mod.')
    }
  } catch (e) {
    throw e
  }
}

export const installMod = async (mod: string, version: string) => {
  try {
    const res = await fetchWithAuth(`${API_BASE}/mods/install/${mod}/${version}`, { method: 'PUT' })
    if (!res.ok) {
      const message = (await res.json())?.message || null
      throw Error(message ? message : 'Failed to install mod.')
    }
  } catch (e) {
    throw e
  }
}

export const uninstallMod = async (mod: string, version: string) => {
  try {
    const res = await fetchWithAuth(`${API_BASE}/mods/uninstall/${mod}/${version}`, { method: 'DELETE' })
    if (!res.ok) {
      const message = (await res.json())?.message || null
      throw Error(message ? message : 'Failed to uninstall mod.')
    }
  } catch (e) {
    throw e
  }
}

export const deleteMod = async (mod: string, version: string) => {
  try {
    const res = await fetchWithAuth(`${API_BASE}/mods/${mod}/${version}`, { method: 'DELETE' })
    if (!res.ok) {
      const message = (await res.json())?.message || null
      throw Error(message ? message : 'Failed to delete mod.')
    }
  } catch (e) {
    throw e
  }
}

// ----------------------------------------------------------------------------
// Factorio Admins
// ----------------------------------------------------------------------------

export const addFactorioAdmin = async (username: string) => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-admins`, {
    method: 'POST',
    body: JSON.stringify({ username }),
  })
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to add admin.' + (message ? `\n${message}` : ''))
  }
}

export const loadFactorioAdmins = async () => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-admins`)
  if (!res.ok) throw Error('Failed to get admins')
  const data = await res.json()
  return data || []
}

export const removeFactorioAdmin = async (username: string) => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-admins/${encodeURIComponent(username)}`, { method: 'DELETE' })
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to remove admin.' + (message ? `\n${message}` : ''))
  }
}

// ----------------------------------------------------------------------------
// Factorio Bans
// ----------------------------------------------------------------------------

export const addFactorioBan = async (username: string) => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-bans`, {
    method: 'POST',
    body: JSON.stringify({ username }),
  })
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to add ban.' + (message ? `\n${message}` : ''))
  }
}

export const loadFactorioBans = async () => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-bans`)
  if (!res.ok) throw Error('Failed to get bans')
  const data = await res.json()
  return data || []
}

export const removeFactorioBan = async (username: string) => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-bans/${encodeURIComponent(username)}`, { method: 'DELETE' })
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to remove ban.' + (message ? `\n${message}` : ''))
  }

}

// ----------------------------------------------------------------------------
// Factorio Whitelist
// ----------------------------------------------------------------------------

export const addFactorioWhitelistUser = async (username: string) => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-whitelist`, {
    method: 'POST',
    body: JSON.stringify({ username }),
  })
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to add whitelist user.' + (message ? `\n${message}` : ''))
  }
}

export const loadFactorioWhitelist = async () => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-whitelist`)
  if (!res.ok) throw Error('Failed to get whitelist')
  const data = await res.json()
  return data || []
}

export const removeFactorioWhitelistUser = async (username: string) => {
  const res = await fetchWithAuth(`${API_BASE}/factorio-whitelist/${encodeURIComponent(username)}`, { method: 'DELETE' })
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to remove whitelist user.' + (message ? `\n${message}` : ''))
  }
}

// ----------------------------------------------------------------------------
// Server
// ----------------------------------------------------------------------------

export const serverStatus = async () => {
  const res = await fetchWithAuth(`${API_BASE}/status`)
  if (res.ok) {
    try {
      const data = await res.json()
      return { loggedIn: res.ok, ...data }
    } catch (e) {
      console.log('Unable to fetch status', e)
    }
  }
  return { loggedIn: false }
}

export const startServer = async () => {
  const res = await fetchWithAuth(`${API_BASE}/start`)
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to start the server.' + (message ? `\n${message}` : ''))
  }
}

export const stopServer = async () => {
  const res = await fetchWithAuth(`${API_BASE}/stop`)
  if (!res.ok) {
    const message = (await res.json())?.message || null
    throw Error('Failed to stop the server.' + (message ? `\n${message}` : ''))
  }
}
