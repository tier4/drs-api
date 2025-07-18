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
      <ModuleStatusTable 
        modules={modules}
        onRestartSensors={onRestartSensors}
        onRestartMachine={onRestartMachine}
        onShutdownMachine={onShutdownMachine}
      />
    </div>
  )
}