import { useState, useCallback } from 'react'
import { EcuStatusTable } from '@/components/EcuStatusTable'
import type { EcuModule } from '@/components/EcuStatusTable'
import { PtpSyncStatus } from '@/components/PtpSyncStatus'
import type { PtpStatus } from '@/components/PtpSyncStatus'
import { RecordingControl } from '@/components/RecordingControl'
import type { RecordingStatus } from '@/components/RecordingControl'
import { TopicRateStatus } from '@/components/TopicRateStatus'
import type { ModuleTopicStatus } from '@/components/TopicRateStatus'
import { PowerControl } from '@/components/PowerControl'
import { ApiService } from '@/services/api'
import { useAutoRefresh } from '@/hooks/useAutoRefresh'

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
  const [modules, setModules] = useState<EcuModule[]>(mockModules)
  const [ptpStatuses, setPtpStatuses] = useState<PtpStatus[]>(mockPtpStatuses)
  const [recordingStatuses, setRecordingStatuses] = useState<RecordingStatus[]>(mockRecordingStatuses)
  const [topicStatuses, setTopicStatuses] = useState<ModuleTopicStatus[]>(mockModuleTopicStatuses)
  const [globalRecordingEnabled, setGlobalRecordingEnabled] = useState(true)
  const [isLoading, setIsLoading] = useState(false)
  const [lastUpdated, setLastUpdated] = useState<Date>(new Date())
  const [apiError, setApiError] = useState<string | null>(null)

  const apiService = ApiService.getInstance()

  // Convert API data to component format
  const convertToEcuModule = (apiModule: any): EcuModule => ({
    hostname: apiModule.hostname,
    status: apiModule.status,
    diskUsagePercentage: apiModule.disk?.usage_percentage || 0,
    diskFreeBytes: apiModule.disk?.free_bytes || 0,
    diskTotalBytes: apiModule.disk?.total_bytes || 0,
  })

  const convertToPtpStatus = (apiPtp: any): PtpStatus | null => {
    if (!apiPtp || !apiPtp.hostname) return null
    
    return {
      hostname: apiPtp.hostname,
      localStatus: {
        clockId: apiPtp.local_status?.clock_id || '',
        masterOffsetNs: apiPtp.local_status?.master_offset_ns || 0,
        gmPresent: apiPtp.local_status?.gm_present || false,
      },
      remoteStatuses: (apiPtp.remote_statuses || []).map((remote: any) => ({
        deviceName: remote.device_name,
        ipAddress: remote.ip_address,
        isReachable: remote.is_reachable,
        offsetNs: remote.status?.master_offset_ns,
      })),
    }
  }

  const convertToRecordingStatus = (apiRec: any): RecordingStatus => ({
    hostname: apiRec.hostname,
    status: apiRec.status,
    active: apiRec.active,
    hardwareId: apiRec.hardware_id,
  })


  // Data fetching function
  const fetchAllData = useCallback(async () => {
    if (isLoading) return
    
    setIsLoading(true)
    setApiError(null)
    try {
      // Fetch all data in parallel
      const [modulesData, ptpData, recordingData] = await Promise.allSettled([
        apiService.getModules(),
        apiService.getPtpStatus(),
        apiService.getRecordingStatus(),
      ])

      // Check if all API calls failed
      let allFailed = true
      let errorMessages: string[] = []

      // Update modules
      if (modulesData.status === 'fulfilled') {
        setModules(modulesData.value.map(convertToEcuModule))
        allFailed = false
      } else {
        errorMessages.push('Modules API failed')
      }

      // Update PTP status
      if (ptpData.status === 'fulfilled') {
        const validPtpStatuses = ptpData.value
          .filter((apiPtp: any) => apiPtp.local_status && apiPtp.local_status.clock_id) // Only include modules with PTP enabled
          .map(convertToPtpStatus)
          .filter((status): status is PtpStatus => status !== null)
        setPtpStatuses(validPtpStatuses)
        allFailed = false
      } else {
        errorMessages.push('PTP API failed')
      }

      // Update recording status
      if (recordingData.status === 'fulfilled') {
        setRecordingStatuses(recordingData.value.map(convertToRecordingStatus))
        allFailed = false
      } else {
        errorMessages.push('Recording API failed')
      }

      if (allFailed) {
        setApiError(`API Gateway is not accessible. ${errorMessages.join(', ')}. Using mock data.`)
      }

      // Fetch topic statuses for each module
      if (modulesData.status === 'fulfilled') {
        const topicPromises = modulesData.value
          .filter(module => module.enabled_services && module.enabled_services.includes('ros2'))
          .map(async (module) => {
            try {
              const topics = await apiService.getTopicStatus(module.hostname)
              return {
                hostname: module.hostname,
                topics: topics.map(topic => ({
                  topicName: topic.topic_name,
                  rateHz: topic.rate_hz,
                  status: topic.status,
                })),
              }
            } catch (error) {
              console.error(`Failed to fetch topics for ${module.hostname}:`, error)
              return {
                hostname: module.hostname,
                topics: [],
              }
            }
          })

        const topicResults = await Promise.all(topicPromises)
        setTopicStatuses(topicResults)
      }

      setLastUpdated(new Date())
    } catch (error) {
      console.error('Failed to fetch data:', error)
      setApiError('Failed to connect to API Gateway. Using mock data.')
    } finally {
      setIsLoading(false)
    }
  }, [apiService, isLoading])

  // Set up auto-refresh
  useAutoRefresh(fetchAllData, { enabled: true, interval: 5000 })

  const handleRestartSensors = async (hostname: string) => {
    try {
      await apiService.restartModuleSensors(hostname)
      console.log(`Restarted sensors for ${hostname}`)
    } catch (error) {
      console.error(`Failed to restart sensors for ${hostname}:`, error)
    }
  }

  const handleRestartMachine = async (hostname: string) => {
    try {
      await apiService.restartModule(hostname)
      console.log(`Restarted machine ${hostname}`)
    } catch (error) {
      console.error(`Failed to restart machine ${hostname}:`, error)
    }
  }

  const handleShutdownMachine = async (hostname: string) => {
    try {
      await apiService.shutdownModule(hostname)
      console.log(`Shut down machine ${hostname}`)
    } catch (error) {
      console.error(`Failed to shutdown machine ${hostname}:`, error)
    }
  }

  const handleGlobalRecordingToggle = async (enabled: boolean) => {
    setGlobalRecordingEnabled(enabled)
    
    try {
      if (enabled) {
        await apiService.startRecording()
        console.log('Started recording')
      } else {
        await apiService.stopRecording()
        console.log('Stopped recording')
      }
      // Refresh data immediately after recording operation
      fetchAllData()
    } catch (error) {
      console.error('Failed to toggle recording:', error)
      // Revert the switch state on error
      setGlobalRecordingEnabled(!enabled)
    }
  }

  const handleSystemRestart = async () => {
    try {
      await apiService.restartSystem()
      console.log('System restart requested')
    } catch (error) {
      console.error('Failed to restart system:', error)
    }
  }

  const handleSystemShutdown = async () => {
    try {
      await apiService.shutdownSystem()
      console.log('System shutdown requested')
    } catch (error) {
      console.error('Failed to shutdown system:', error)
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center space-x-4">
            <h1 className="text-3xl font-bold">DRS Dashboard</h1>
            {isLoading && (
              <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin"></div>
                <span>Updating...</span>
              </div>
            )}
          </div>
          <div className="flex items-center space-x-4">
            <span className="text-sm text-muted-foreground">
              Last updated: {lastUpdated.toLocaleTimeString()}
            </span>
            <PowerControl 
              onSystemRestart={handleSystemRestart}
              onSystemShutdown={handleSystemShutdown}
            />
          </div>
        </div>
        {apiError && (
          <div className="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
            <div className="flex items-center">
              <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
              <span className="font-medium">API Connection Error:</span>
              <span className="ml-2">{apiError}</span>
            </div>
          </div>
        )}
        <div className="space-y-6">
          <div>
            <h2 className="text-xl font-semibold mb-4">Module Status</h2>
            <EcuStatusTable 
              modules={modules}
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
            <PtpSyncStatus ptpStatuses={ptpStatuses} />
          </div>
          
          <div>
            <TopicRateStatus moduleTopicStatuses={topicStatuses} />
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
