// Multi-project versions config
// Supports KubeStellar, a2a, and kubeflex with independent versioning
//
// Versioning Strategy:
// - Each project has its own version scheme
// - Branch naming convention:
//   - KubeStellar: main (latest), docs/{version} (e.g., docs/0.28.0)
//   - a2a: main (latest), docs/a2a/{version} (e.g., docs/a2a/0.1.0)
//   - kubeflex: main (latest), docs/kubeflex/{version} (e.g., docs/kubeflex/0.8.0)
// - The main branch always contains the latest version for all projects

// Netlify site name for branch deploys
export const NETLIFY_SITE_NAME = "kubestellar-docs"

// Production URL for latest version
export const PRODUCTION_URL = "https://kubestellar.io"

// Project identifiers
export type ProjectId = "kubestellar" | "a2a" | "kubeflex" | "multi-plugin" | "kubectl-claude"

// Version info structure
export interface VersionInfo {
  label: string
  branch: string
  isDefault: boolean
  externalUrl?: string
  isDev?: boolean // marks development/unreleased versions
}

// Project configuration
export interface ProjectConfig {
  id: ProjectId
  name: string
  basePath: string // '' for kubestellar, 'a2a' for a2a, etc.
  currentVersion: string
  contentPath: string
  versions: Record<string, VersionInfo>
}

// KubeStellar versions (existing)
const KUBESTELLAR_VERSIONS: Record<string, VersionInfo> = {
  latest: {
    label: "v0.29.0 (Latest)",
    branch: "docs/0.29.0",
    isDefault: true,
  },
  main: {
    label: "main (dev)",
    branch: "main",
    isDefault: false,
    isDev: true,
  },
  "0.28.0": {
    label: "v0.28.0",
    branch: "docs/0.28.0",
    isDefault: false,
  },
  "0.27.2": {
    label: "v0.27.2",
    branch: "docs/0.27.2",
    isDefault: false,
  },
  "0.27.1": {
    label: "v0.27.1",
    branch: "docs/0.27.1",
    isDefault: false,
  },
  "0.27.0": {
    label: "v0.27.0",
    branch: "docs/0.27.0",
    isDefault: false,
  },
  "0.26.0": {
    label: "v0.26.0",
    branch: "docs/0.26.0",
    isDefault: false,
  },
  "0.25.1": {
    label: "v0.25.1",
    branch: "docs/0.25.1",
    isDefault: false,
  },
  "0.25.0": {
    label: "v0.25.0",
    branch: "docs/0.25.0",
    isDefault: false,
  },
  "0.24.0": {
    label: "v0.24.0",
    branch: "docs/0.24.0",
    isDefault: false,
  },
  "0.23.1": {
    label: "v0.23.1",
    branch: "docs/0.23.1",
    isDefault: false,
  },
  "0.23.0": {
    label: "v0.23.0",
    branch: "docs/0.23.0",
    isDefault: false,
  },
  "0.22.0": {
    label: "v0.22.0",
    branch: "docs/0.22.0",
    isDefault: false,
  },
  "0.21.2": {
    label: "v0.21.2",
    branch: "docs/0.21.2",
    isDefault: false,
  },
  "0.21.1": {
    label: "v0.21.1",
    branch: "docs/0.21.1",
    isDefault: false,
  },
  "0.21.0": {
    label: "v0.21.0",
    branch: "docs/0.21.0",
    isDefault: false,
  },
  legacy: {
    label: "Older Versions",
    branch: "legacy",
    isDefault: false,
    externalUrl: "https://kubestellar.github.io/kubestellar",
  },
}

// a2a versions
const A2A_VERSIONS: Record<string, VersionInfo> = {
  latest: {
    label: "v0.1.0 (Latest)",
    branch: "main",
    isDefault: true,
  },
  main: {
    label: "main (dev)",
    branch: "main",
    isDefault: false,
    isDev: true,
  },
}

// kubeflex versions
const KUBEFLEX_VERSIONS: Record<string, VersionInfo> = {
  latest: {
    label: "v0.9.3 (Latest)",
    branch: "main",
    isDefault: true,
  },
  main: {
    label: "main (dev)",
    branch: "main",
    isDefault: false,
    isDev: true,
  },
  "0.8.0": {
    label: "v0.8.0",
    branch: "docs/kubeflex/0.8.0",
    isDefault: false,
  },
  "0.7.0": {
    label: "v0.7.0",
    branch: "docs/kubeflex/0.7.0",
    isDefault: false,
  },
}

// multi-plugin versions
const MULTI_PLUGIN_VERSIONS: Record<string, VersionInfo> = {
  latest: {
    label: "v0.1.0 (Latest)",
    branch: "main",
    isDefault: true,
  },
  main: {
    label: "main (dev)",
    branch: "main",
    isDefault: false,
    isDev: true,
  },
}

// kubectl-claude versions
const KUBECTL_CLAUDE_VERSIONS: Record<string, VersionInfo> = {
  latest: {
    label: "v0.4.0 (Latest)",
    branch: "main",
    isDefault: true,
  },
  main: {
    label: "main (dev)",
    branch: "main",
    isDefault: false,
    isDev: true,
  },
  "0.4.0": {
    label: "v0.4.0",
    branch: "docs/kubectl-claude/0.4.0",
    isDefault: false,
  },
}

// All projects configuration
export const PROJECTS: Record<ProjectId, ProjectConfig> = {
  kubestellar: {
    id: "kubestellar",
    name: "KubeStellar",
    basePath: "",
    currentVersion: "0.29.0",
    contentPath: "docs/content",
    versions: KUBESTELLAR_VERSIONS,
  },
  a2a: {
    id: "a2a",
    name: "A2A",
    basePath: "a2a",
    currentVersion: "0.1.0",
    contentPath: "docs/content/a2a",
    versions: A2A_VERSIONS,
  },
  kubeflex: {
    id: "kubeflex",
    name: "KubeFlex",
    basePath: "kubeflex",
    currentVersion: "0.9.3",
    contentPath: "docs/content/kubeflex",
    versions: KUBEFLEX_VERSIONS,
  },
  "multi-plugin": {
    id: "multi-plugin",
    name: "Multi Plugin",
    basePath: "multi-plugin",
    currentVersion: "0.1.0",
    contentPath: "docs/content/multi-plugin",
    versions: MULTI_PLUGIN_VERSIONS,
  },
  "kubectl-claude": {
    id: "kubectl-claude",
    name: "kubectl-claude",
    basePath: "kubectl-claude",
    currentVersion: "0.3.0",
    contentPath: "docs/content/kubectl-claude",
    versions: KUBECTL_CLAUDE_VERSIONS,
  },
}

// Get project from URL pathname
export function getProjectFromPath(pathname: string): ProjectConfig {
  if (pathname.startsWith("/docs/a2a")) {
    return PROJECTS.a2a
  }
  if (pathname.startsWith("/docs/kubeflex")) {
    return PROJECTS.kubeflex
  }
  if (pathname.startsWith("/docs/multi-plugin")) {
    return PROJECTS["multi-plugin"]
  }
  if (pathname.startsWith("/docs/kubectl-claude") || pathname.startsWith("/docs/related-projects/kubectl-claude")) {
    return PROJECTS["kubectl-claude"]
  }
  return PROJECTS.kubestellar
}

// Get project by ID
export function getProject(projectId: ProjectId): ProjectConfig {
  return PROJECTS[projectId]
}

// Get all projects
export function getAllProjects(): ProjectConfig[] {
  return Object.values(PROJECTS)
}

// ============================================
// Backwards-compatible exports for KubeStellar
// ============================================

export const CURRENT_VERSION = PROJECTS.kubestellar.currentVersion
export const VERSIONS = KUBESTELLAR_VERSIONS

export type VersionKey = keyof typeof KUBESTELLAR_VERSIONS

export function getDefaultVersion(): VersionKey {
  return "latest"
}

export function getCurrentVersion(): string {
  return CURRENT_VERSION
}

export function getBranchForVersion(version: VersionKey): string {
  return KUBESTELLAR_VERSIONS[version]?.branch ?? "main"
}

export function getVersionFromBranch(branch: string): VersionKey | null {
  // Check if branch matches docs/{version} pattern
  const match = branch.match(/^docs\/(.+)$/)
  if (match) {
    const versionNum = match[1]
    // Find version entry with matching branch
    for (const [key, value] of Object.entries(KUBESTELLAR_VERSIONS)) {
      if (value.branch === branch || key === versionNum) {
        return key as VersionKey
      }
    }
  }

  // Check for main branch
  if (branch === "main" || branch === "master") {
    return "latest"
  }

  return null
}

export function getAllVersions(): Array<{ key: VersionKey } & VersionInfo> {
  return Object.entries(KUBESTELLAR_VERSIONS).map(([key, value]) => ({
    key: key as VersionKey,
    ...value,
  }))
}

// Helper to validate if a branch name follows version convention
export function isVersionBranch(branch: string): boolean {
  return branch === "main" || branch.startsWith("docs/")
}

// Get the URL for a specific version (project-aware)
export function getVersionUrl(
  versionKey: string,
  pathname: string = "/docs",
  projectId: ProjectId = "kubestellar"
): string {
  const project = PROJECTS[projectId]
  const version = project.versions[versionKey]

  if (!version) {
    return `${PRODUCTION_URL}${pathname}`
  }

  // If it has an external URL (like legacy), use that
  if ("externalUrl" in version && version.externalUrl) {
    return version.externalUrl
  }

  // Latest version uses production URL
  if (versionKey === "latest" || version.isDefault) {
    return `${PRODUCTION_URL}${pathname}`
  }

  // Other versions use Netlify branch deploys
  // Netlify converts branch names: docs/0.28.0 -> docs-0-28-0
  const branchSlug = version.branch.replace(/\//g, "-").replace(/\./g, "-")
  return `https://${branchSlug}--${NETLIFY_SITE_NAME}.netlify.app${pathname}`
}

// Get versions for a specific project
export function getProjectVersions(
  projectId: ProjectId
): Array<{ key: string } & VersionInfo> {
  const project = PROJECTS[projectId]
  return Object.entries(project.versions).map(([key, value]) => ({
    key,
    ...value,
  }))
}

// Check if a version has been migrated (branch exists)
export function isVersionMigrated(
  versionKey: string,
  projectId: ProjectId = "kubestellar"
): boolean {
  const project = PROJECTS[projectId]

  // Latest is always available
  if (versionKey === "latest") return true

  // Legacy links externally, so it's "available"
  if (versionKey === "legacy") return true

  // For other versions, assume they exist if in the versions list
  return versionKey in project.versions
}
