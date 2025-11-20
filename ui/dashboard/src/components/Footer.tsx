export function Footer() {
  const version = __APP_VERSION__
  const buildTime = __BUILD_TIME__
  const gitCommit = __GIT_COMMIT__

  const formatBuildTime = (isoString: string) => {
    try {
      const date = new Date(isoString)
      return date.toLocaleString('ja-JP', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
      })
    } catch (error) {
      console.error('Failed to parse build time:', error)
      return 'N/A'
    }
  }

  const formatCommit = (commit: string) => {
    if (commit === 'dev' || !commit) return commit
    return commit.substring(0, 7)
  }

  return (
    <footer className="mt-16 py-6 border-t border-border">
      <div className="container mx-auto px-4">
        <div className="flex flex-col sm:flex-row items-center justify-center gap-2 sm:gap-6 text-sm text-muted-foreground">
          <div className="flex items-center gap-2">
            <span className="font-semibold">Version:</span>
            <span className="font-mono">{version}</span>
          </div>
          <div className="hidden sm:block text-border">|</div>
          <div className="flex items-center gap-2">
            <span className="font-semibold">Built:</span>
            <span className="font-mono">{formatBuildTime(buildTime)}</span>
          </div>
          <div className="hidden sm:block text-border">|</div>
          <div className="flex items-center gap-2">
            <span className="font-semibold">Commit:</span>
            <span className="font-mono">{formatCommit(gitCommit)}</span>
          </div>
        </div>
      </div>
    </footer>
  )
}
