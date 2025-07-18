import { ModuleStatus } from '@/components/ModuleStatus'
import type { Module } from '@/components/ModuleStatus'

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
    <ModuleStatus 
      modules={modules}
      onRestartSensors={onRestartSensors}
      onRestartMachine={onRestartMachine}
      onShutdownMachine={onShutdownMachine}
    />
  )
}