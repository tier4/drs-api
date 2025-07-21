import { useState, useCallback, useRef } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import type { Module } from '@/components/ModuleStatus'
import type { PtpStatus } from '@/components/PtpSyncStatus'
import type { ModuleTopicStatus } from '@/components/TopicRateStatus'
import { PowerControl } from '@/components/PowerControl'
import { RecordingSwitch } from '@/components/RecordingSwitch'
import { Navigation } from '@/components/Navigation'
import { ModulesPage } from '@/pages/ModulesPage'
import { TimeSyncPage } from '@/pages/TimeSyncPage'
import { TopicRatesPage } from '@/pages/TopicRatesPage'
import { ApiService } from '@/services/api'
import { useAutoRefresh } from '@/hooks/useAutoRefresh'

// Mock data based on API design - matching the visual from docs/new_ui.md
const mockModules: Module[] = [
  {
    hostname: 'ecu0',
    moduleId: 'a3b2c1d4aaa',
    services: {
      drs_sensor: 'active',
      drs_recorder: 'active',
    },
    recordingStatus: 'recording',
    dataStatus: 'OK',
    diskUsagePercentage: 75.0,
    diskFreeBytes: 268435456000, // 250GB
    diskTotalBytes: 1073741824000, // 1TB
  },
  {
    hostname: 'ecu1',
    moduleId: 'f5e6d7c8aaa',
    services: {
      drs_sensor: 'active',
      drs_recorder: 'failed',
    },
    recordingStatus: 'stopped',
    dataStatus: 'WARN',
    diskUsagePercentage: 90.0,
    diskFreeBytes: 107374182400, // 100GB
    diskTotalBytes: 1073741824000, // 1TB
  },
  {
    hostname: 'nas',
    dataStatus: 'OK',
    diskUsagePercentage: 25.0,
    diskFreeBytes: 805306368000, // 750GB
    diskTotalBytes: 1073741824000, // 1TB
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
  const [modules, setModules] = useState<Module[]>(mockModules)
  const [ptpStatuses, setPtpStatuses] = useState<PtpStatus[]>(mockPtpStatuses)
  const [topicStatuses, setTopicStatuses] = useState<ModuleTopicStatus[]>(mockModuleTopicStatuses)
  const [isLoading, setIsLoading] = useState(false)
  const [lastUpdated, setLastUpdated] = useState<Date>(new Date())
  const [apiError, setApiError] = useState<string | null>(null)
  const [isRecording, setIsRecording] = useState(false)
  const [isRecordingLoading, setIsRecordingLoading] = useState(false)
  const recordingToggleTimestamp = useRef<number>(0)

  const apiService = ApiService.getInstance()

  // Convert API data to component format
  const convertToModule = (apiModule: any): Module => {
    const module: Module = {
      hostname: apiModule.hostname,
      moduleId: apiModule.environment?.module_id,
      dataStatus: 'OK', // Default value
      diskUsagePercentage: apiModule.disk?.usage_percentage || 0,
      diskFreeBytes: apiModule.disk?.free_bytes || 0,
      diskTotalBytes: apiModule.disk?.total_bytes || 0,
    }
    
    // Only add services and recording for ecu modules, not for nas
    if (apiModule.hostname.startsWith('ecu')) {
      module.services = apiModule.status_detail?.services
      module.recordingStatus = apiModule.status_detail?.recording?.status
      // For ECU modules, use recording data_status if available, otherwise use module status
      module.dataStatus = apiModule.status_detail?.recording?.data_status || apiModule.status || 'OK'
    } else {
      // For NAS, set empty string since it doesn't have recording capability
      module.dataStatus = '' as any
    }
    
    return module
  }

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
      let convertedModules: Module[] = []
      if (modulesData.status === 'fulfilled') {
        convertedModules = modulesData.value.map(convertToModule)
        allFailed = false
      } else {
        errorMessages.push('Modules API failed')
      }

      // Update PTP status
      let validPtpStatuses: PtpStatus[] = []
      if (ptpData.status === 'fulfilled') {
        validPtpStatuses = ptpData.value
          .filter((apiPtp: any) => apiPtp.local_status && apiPtp.local_status.clock_id) // Only include modules with PTP enabled
          .map(convertToPtpStatus)
          .filter((status): status is PtpStatus => status !== null)
        setPtpStatuses(validPtpStatuses)
        allFailed = false
      } else {
        errorMessages.push('PTP API failed')
      }

      // Merge PTP status into modules
      if (convertedModules.length > 0) {
        const ptpStatusMap = new Map(validPtpStatuses.map(ptp => [ptp.hostname, ptp]))
        convertedModules = convertedModules.map(module => {
          const ptpStatus = ptpStatusMap.get(module.hostname)
          if (ptpStatus) {
            return {
              ...module,
              ptpStatus: {
                gmPresent: ptpStatus.localStatus.gmPresent,
                offsetNs: ptpStatus.localStatus.masterOffsetNs
              }
            }
          }
          return module
        })
        
        // Sort modules alphabetically by hostname
        convertedModules.sort((a, b) => a.hostname.localeCompare(b.hostname))
        setModules(convertedModules)
        
        // Check if any module is recording
        const anyModuleRecording = convertedModules.some(module => 
          module.recordingStatus === 'recording'
        )
        
        // Don't override recording state if recently toggled (within 10 seconds)
        const timeSinceToggle = Date.now() - recordingToggleTimestamp.current
        if (timeSinceToggle > 10000) {
          setIsRecording(anyModuleRecording)
        }
      }

      // Sort PTP statuses alphabetically by hostname
      validPtpStatuses.sort((a, b) => a.hostname.localeCompare(b.hostname))

      // Recording status is now merged with module data
      if (recordingData.status === 'rejected') {
        errorMessages.push('Recording API failed')
      }

      if (allFailed) {
        setApiError(`API Gateway is not accessible. ${errorMessages.join(', ')}. Using mock data.`)
      }

      // Fetch topic statuses for each module
      if (modulesData.status === 'fulfilled') {
        const topicPromises = modulesData.value
          .filter(module => module.hostname.startsWith('ecu'))
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
        // Sort topic statuses alphabetically by hostname
        topicResults.sort((a, b) => a.hostname.localeCompare(b.hostname))
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

  const handleRecordingToggle = async (enabled: boolean) => {
    if (isRecordingLoading) return
    
    // Optimistic update - immediately update UI
    setIsRecording(enabled)
    setIsRecordingLoading(true)
    recordingToggleTimestamp.current = Date.now()
    
    try {
      if (enabled) {
        const success = await apiService.startRecording()
        if (!success) {
          // Revert only on failure
          setIsRecording(false)
          console.error('Failed to start recording')
        } else {
          console.log('Recording started')
        }
      } else {
        const success = await apiService.stopRecording()
        if (!success) {
          // Revert only on failure
          setIsRecording(true)
          console.error('Failed to stop recording')
        } else {
          console.log('Recording stopped')
        }
      }
    } catch (error) {
      console.error('Failed to toggle recording:', error)
      // Revert the state on error
      setIsRecording(!enabled)
    } finally {
      setIsRecordingLoading(false)
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
            <RecordingSwitch 
              isRecording={isRecording}
              isLoading={isRecordingLoading}
              onToggle={handleRecordingToggle}
            />
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
        <Navigation />
        <Routes>
          <Route 
            path="/" 
            element={
              <ModulesPage 
                modules={modules}
                onRestartSensors={handleRestartSensors}
                onRestartMachine={handleRestartMachine}
                onShutdownMachine={handleShutdownMachine}
              />
            } 
          />
          <Route 
            path="/time-sync" 
            element={<TimeSyncPage ptpStatuses={ptpStatuses} />} 
          />
          <Route 
            path="/topic-rates" 
            element={<TopicRatesPage moduleTopicStatuses={topicStatuses} />} 
          />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </div>
    </div>
  )
}

export default App
