import { useState } from 'react'
import { EcuStatusTable } from '@/components/EcuStatusTable'
import type { EcuModule } from '@/components/EcuStatusTable'
import { PtpSyncStatus } from '@/components/PtpSyncStatus'
import type { PtpStatus } from '@/components/PtpSyncStatus'
import { RecordingControl } from '@/components/RecordingControl'
import type { RecordingStatus } from '@/components/RecordingControl'
import { TopicRateStatus } from '@/components/TopicRateStatus'
import type { ModuleTopicStatus } from '@/components/TopicRateStatus'
import { PowerControl } from '@/components/PowerControl'

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

// Mock PTP data based on API design
const mockPtpStatuses: PtpStatus[] = [
  {
    hostname: 'ecu0',
    localStatus: {
      clockId: '001122.fffe.334455',
      masterOffsetNs: 1234,
      gmPresent: true,
    },
    remoteStatuses: [
      {
        deviceName: 'ecu1',
        ipAddress: '192.168.20.2',
        isReachable: true,
        offsetNs: 2345,
      },
      {
        deviceName: 'nas',
        ipAddress: '192.168.10.100',
        isReachable: false,
      },
    ],
  },
  {
    hostname: 'ecu1',
    localStatus: {
      clockId: '001122.fffe.665544',
      masterOffsetNs: -890,
      gmPresent: true,
    },
    remoteStatuses: [
      {
        deviceName: 'ecu0',
        ipAddress: '192.168.20.1',
        isReachable: true,
        offsetNs: -1234,
      },
    ],
  },
  // NAS does not have PTP functionality - only system and disk monitoring
]

// Mock Recording data based on API design
const mockRecordingStatuses: RecordingStatus[] = [
  {
    hostname: 'ecu0',
    status: 'recording',
    active: true,
    hardwareId: 'hw-001',
  },
  {
    hostname: 'ecu1',
    status: 'paused',
    active: true,
    hardwareId: 'hw-002',
  },
  // NAS does not have recording functionality
]

// Mock Topic Rate data based on API design
const mockModuleTopicStatuses: ModuleTopicStatus[] = [
  {
    hostname: 'ecu0',
    topics: [
      {
        topicName: '/camera/image_raw',
        rateHz: 30.0,
        status: 'OK',
      },
      {
        topicName: '/lidar/pointcloud',
        rateHz: 9.8,
        status: 'WARN',
      },
      {
        topicName: '/imu/data',
        rateHz: 100.0,
        status: 'OK',
      },
      {
        topicName: '/gps/fix',
        rateHz: 0.0,
        status: 'ERROR',
      },
    ],
  },
  {
    hostname: 'ecu1',
    topics: [
      {
        topicName: '/camera/image_raw',
        rateHz: 29.5,
        status: 'OK',
      },
      {
        topicName: '/radar/tracks',
        rateHz: 4.2,
        status: 'WARN',
      },
      {
        topicName: '/vehicle/velocity',
        rateHz: 50.0,
        status: 'OK',
      },
    ],
  },
  // NAS does not have ROS2 topics
]

function App() {
  const [globalRecordingEnabled, setGlobalRecordingEnabled] = useState(true)
  const [recordingStatuses, setRecordingStatuses] = useState<RecordingStatus[]>(mockRecordingStatuses)

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

  const handleGlobalRecordingToggle = (enabled: boolean) => {
    setGlobalRecordingEnabled(enabled)
    console.log(`Global recording ${enabled ? 'enabled' : 'disabled'}`)
    
    if (enabled) {
      // Start recording
      console.log('Starting recording...')
      setRecordingStatuses(prev => prev.map(r => ({ ...r, status: 'recording' as const })))
      // TODO: Implement API call to start recording
    } else {
      // Stop recording
      console.log('Stopping recording...')
      setRecordingStatuses(prev => prev.map(r => ({ ...r, status: 'stopped' as const })))
      // TODO: Implement API call to stop recording
    }
  }

  const handleSystemRestart = () => {
    console.log('System restart requested')
    // TODO: Implement API call to restart all systems
  }

  const handleSystemShutdown = () => {
    console.log('System shutdown requested')
    // TODO: Implement API call to shutdown all systems
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-3xl font-bold">DRS Dashboard</h1>
          <PowerControl 
            onSystemRestart={handleSystemRestart}
            onSystemShutdown={handleSystemShutdown}
          />
        </div>
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
          
          <div>
            <RecordingControl
              recordingStatuses={recordingStatuses}
              globalRecordingEnabled={globalRecordingEnabled}
              onGlobalRecordingToggle={handleGlobalRecordingToggle}
            />
          </div>
          
          <div>
            <PtpSyncStatus ptpStatuses={mockPtpStatuses} />
          </div>
          
          <div>
            <TopicRateStatus moduleTopicStatuses={mockModuleTopicStatuses} />
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
