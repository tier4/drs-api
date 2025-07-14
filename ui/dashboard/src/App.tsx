import { EcuStatusTable } from '@/components/EcuStatusTable'
import type { EcuModule } from '@/components/EcuStatusTable'

// Mock data based on API design
const mockModules: EcuModule[] = [
  {
    hostname: 'ecu0',
    status: 'OK',
    diskUsagePercentage: 75.5,
    diskFreeBytes: 1073741824, // 1GB
    diskTotalBytes: 4294967296, // 4GB
  },
  {
    hostname: 'ecu1',
    status: 'WARN',
    diskUsagePercentage: 89.2,
    diskFreeBytes: 536870912, // 0.5GB
    diskTotalBytes: 4294967296, // 4GB
  },
  {
    hostname: 'nas',
    status: 'ERROR',
    diskUsagePercentage: 95.8,
    diskFreeBytes: 214748364, // 0.2GB
    diskTotalBytes: 5368709120, // 5GB
  },
]

function App() {
  const handleRestartSensors = (hostname: string) => {
    console.log(`Restarting sensors for ${hostname}`)
    // TODO: Implement API call to restart sensors
  }

  const handleRestartMachine = (hostname: string) => {
    console.log(`Restarting machine ${hostname}`)
    // TODO: Implement API call to restart machine
  }

  const handleShutdownMachine = (hostname: string) => {
    console.log(`Shutting down machine ${hostname}`)
    // TODO: Implement API call to shutdown machine
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">DRS Dashboard</h1>
        <div className="space-y-6">
          <div>
            <h2 className="text-xl font-semibold mb-4">Module Status</h2>
            <EcuStatusTable 
              modules={mockModules}
              onRestartSensors={handleRestartSensors}
              onRestartMachine={handleRestartMachine}
              onShutdownMachine={handleShutdownMachine}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
