const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'
console.log('API Base URL:', API_BASE_URL)

export interface ModuleStatus {
  hostname: string
  address: string
  status: 'OK' | 'WARN' | 'ERROR'
  status_detail: {
    services: {
      drs_sensor: string
      drs_recorder: string
    }
    recording: {
      status: string
      data_status: string
    }
    ptp: {
      offset_ns: number
    }
  }
  disk: {
    usage_percentage: number
    free_bytes: number
    total_bytes: number
  }
  environment: {
    sensing_system_id: string
    module_id: string
  }
  enabled_services: string[]
}

export interface RecordingStatus {
  hostname: string
  recording_status: 'recording' | 'stopped'
  data_status: 'OK' | 'WARN' | 'ERROR'
  hardware_id?: string
}

export interface PtpStatus {
  hostname: string
  local_status: {
    clock_id: string
    master_offset_ns: number
    gm_present: boolean
  }
  remote_statuses: {
    device_name: string
    ip_address: string
    is_reachable: boolean
    status?: {
      clock_id: string
      master_offset_ns: number
      gm_present: boolean
    }
  }[]
}

export interface TopicStatus {
  topic_name: string
  rate_hz: number
  status: 'OK' | 'WARN' | 'ERROR'
}

export class ApiService {
  private static instance: ApiService

  public static getInstance(): ApiService {
    if (!ApiService.instance) {
      ApiService.instance = new ApiService()
    }
    return ApiService.instance
  }

  private async fetchWithTimeout(
    url: string,
    timeout = 5000,
    options: RequestInit = {},
  ): Promise<Response> {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), timeout)

    try {
      const response = await fetch(url, {
        ...options,
        signal: controller.signal,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        // Remove credentials to avoid CORS issues with credentials
        // credentials: 'include',
      })
      clearTimeout(timeoutId)
      return response
    } catch (error) {
      clearTimeout(timeoutId)
      throw error
    }
  }

  async getModules(): Promise<ModuleStatus[]> {
    const url = `${API_BASE_URL}/modules`
    console.log('Fetching modules from:', url)
    try {
      const response = await this.fetchWithTimeout(url)
      console.log('Response status:', response.status)
      if (!response.ok) {
        const errorText = await response.text()
        console.error('Error response:', errorText)
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      console.log('Modules data:', data)
      return data.modules || []
    } catch (error) {
      console.error('Failed to fetch modules:', error)
      if (error instanceof TypeError && error.message.includes('Failed to fetch')) {
        console.error('Network error - API may not be accessible at:', url)
      }
      throw error
    }
  }

  async getRecordingStatus(): Promise<RecordingStatus[]> {
    try {
      const response = await this.fetchWithTimeout(`${API_BASE_URL}/recording/status`)
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.recording_status || []
    } catch (error) {
      console.error('Failed to fetch recording status:', error)
      throw error
    }
  }

  async getPtpStatus(): Promise<PtpStatus[]> {
    try {
      const response = await this.fetchWithTimeout(`${API_BASE_URL}/ptp/status`)
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.ptp_status || []
    } catch (error) {
      console.error('Failed to fetch PTP status:', error)
      throw error
    }
  }

  async getTopicStatus(hostname: string): Promise<TopicStatus[]> {
    try {
      const response = await this.fetchWithTimeout(
        `${API_BASE_URL}/modules/${hostname}/topics/status`,
      )
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.topics || []
    } catch (error) {
      console.error(`Failed to fetch topic status for ${hostname}:`, error)
      throw error
    }
  }

  async startRecording(): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(`${API_BASE_URL}/recording/start`, 10000, {
        method: 'POST',
      })
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error('Failed to start recording:', error)
      throw error
    }
  }

  async stopRecording(): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(`${API_BASE_URL}/recording/stop`, 10000, {
        method: 'POST',
      })
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error('Failed to stop recording:', error)
      throw error
    }
  }

  async restartSystem(): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(`${API_BASE_URL}/system/restart`, 10000, {
        method: 'POST',
        body: JSON.stringify({}),
      })
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error('Failed to restart system:', error)
      throw error
    }
  }

  async shutdownSystem(): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(`${API_BASE_URL}/system/shutdown`, 10000, {
        method: 'POST',
        body: JSON.stringify({}),
      })
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error('Failed to shutdown system:', error)
      throw error
    }
  }

  async startModuleSensor(hostname: string): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(
        `${API_BASE_URL}/modules/${hostname}/services/drs_sensor/start`,
        10000,
        {
          method: 'POST',
        },
      )
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error(`Failed to start sensor for ${hostname}:`, error)
      throw error
    }
  }

  async stopModuleSensor(hostname: string): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(
        `${API_BASE_URL}/modules/${hostname}/services/drs_sensor/stop`,
        10000,
        {
          method: 'POST',
        },
      )
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error(`Failed to stop sensor for ${hostname}:`, error)
      throw error
    }
  }

  async restartModuleSensor(hostname: string): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(
        `${API_BASE_URL}/modules/${hostname}/services/drs_sensor/restart`,
        10000,
        {
          method: 'POST',
        },
      )
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error(`Failed to restart sensor for ${hostname}:`, error)
      throw error
    }
  }

  async startModuleRecorder(hostname: string): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(
        `${API_BASE_URL}/modules/${hostname}/services/drs_recorder/start`,
        10000,
        {
          method: 'POST',
        },
      )
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error(`Failed to start recorder for ${hostname}:`, error)
      throw error
    }
  }

  async stopModuleRecorder(hostname: string): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(
        `${API_BASE_URL}/modules/${hostname}/services/drs_recorder/stop`,
        10000,
        {
          method: 'POST',
        },
      )
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error(`Failed to stop recorder for ${hostname}:`, error)
      throw error
    }
  }

  async restartModuleRecorder(hostname: string): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(
        `${API_BASE_URL}/modules/${hostname}/services/drs_recorder/restart`,
        10000,
        {
          method: 'POST',
        },
      )
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error(`Failed to restart recorder for ${hostname}:`, error)
      throw error
    }
  }

  async restartModuleSensors(hostname: string): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(
        `${API_BASE_URL}/modules/${hostname}/services/restart`,
        10000,
        {
          method: 'POST',
        },
      )
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error(`Failed to restart sensors for ${hostname}:`, error)
      throw error
    }
  }

  async restartModule(hostname: string): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(
        `${API_BASE_URL}/modules/${hostname}/restart`,
        10000,
        {
          method: 'POST',
          body: JSON.stringify({}),
        },
      )
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error(`Failed to restart module ${hostname}:`, error)
      throw error
    }
  }

  async shutdownModule(hostname: string): Promise<boolean> {
    try {
      const response = await this.fetchWithTimeout(
        `${API_BASE_URL}/modules/${hostname}/shutdown`,
        10000,
        {
          method: 'POST',
          body: JSON.stringify({}),
        },
      )
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      return data.success || false
    } catch (error) {
      console.error(`Failed to shutdown module ${hostname}:`, error)
      throw error
    }
  }
}
