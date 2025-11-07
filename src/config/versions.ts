export const VERSIONS = {
  'v0.29.0': { branch: 'v0.29.0', label: 'v0.29.0 (Latest)', isDefault: true },
  'v0.29.0-rc.1': { branch: 'v0.29.0-rc.1', label: 'v0.29.0-rc.1' },
  'v0.28.0': { branch: 'v0.28.0', label: 'v0.28.0' },
  'v0.27.2': { branch: 'v0.27.2', label: 'v0.27.2' },
  'v0.27.1': { branch: 'v0.27.1', label: 'v0.27.1' },
  'v0.27.0': { branch: 'v0.27.0', label: 'v0.27.0' },
  'v0.26.0': { branch: 'release-0.26.0', label: 'v0.26.0' },
  'v0.25.1': { branch: 'release-0.25.1', label: 'v0.25.1' },
  'main': { branch: 'main', label: 'Development (main)' },
} as const

export type VersionKey = keyof typeof VERSIONS

export function getDefaultVersion(): VersionKey {
  return Object.entries(VERSIONS).find(([_, v]) => v.isDefault)?.[0] as VersionKey || 'v0.29.0'
}

export function getBranchForVersion(version: VersionKey): string {
  return VERSIONS[version]?.branch || 'main'
}

export function getAllVersions() {
  return Object.entries(VERSIONS).map(([key, value]) => ({
    key: key as VersionKey,
    ...value
  }))
}