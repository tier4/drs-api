import { ModuleStatus } from '@/components/ModuleStatus'
import type { Module } from '@/components/ModuleStatus'

interface ModulesPageProps {
  modules: Module[]
  onStartSensor: (hostname: string) => Promise<void>
  onStopSensor: (hostname: string) => Promise<void>
  onRestartSensor: (hostname: string) => Promise<void>
  onStartRecorder: (hostname: string) => Promise<void>
  onStopRecorder: (hostname: string) => Promise<void>
  onRestartRecorder: (hostname: string) => Promise<void>
  onRestartMachine: (hostname: string) => Promise<void>
  onShutdownMachine: (hostname: string) => Promise<void>
}

export function ModulesPage({
  modules,
  onStartSensor,
  onStopSensor,
  onRestartSensor,
  onStartRecorder,
  onStopRecorder,
  onRestartRecorder,
  onRestartMachine,
  onShutdownMachine,
}: ModulesPageProps) {
  return (
    <ModuleStatus
      modules={modules}
      onStartSensor={onStartSensor}
      onStopSensor={onStopSensor}
      onRestartSensor={onRestartSensor}
      onStartRecorder={onStartRecorder}
      onStopRecorder={onStopRecorder}
      onRestartRecorder={onRestartRecorder}
      onRestartMachine={onRestartMachine}
      onShutdownMachine={onShutdownMachine}
    />
  )
}
