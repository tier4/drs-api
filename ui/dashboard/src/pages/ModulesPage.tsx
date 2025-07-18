import { ModuleStatusTable } from '@/components/ModuleStatusTable'
import type { Module } from '@/components/ModuleStatusTable'

interface ModulesPageProps {
  modules: Module[]
  onRestartSensors: (hostname: string) => Promise<void>
  onRestartMachine: (hostname: string) => Promise<void>
  onShutdownMachine: (hostname: string) => Promise<void>
}

export function ModulesPage({
  modules,
  onRestartSensors,
  onRestartMachine,
  onShutdownMachine,
}: ModulesPageProps) {
  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Module Status</h2>
      <ModuleStatusTable 
        modules={modules}
        onRestartSensors={onRestartSensors}
        onRestartMachine={onRestartMachine}
        onShutdownMachine={onShutdownMachine}
      />
    </div>
  )
}