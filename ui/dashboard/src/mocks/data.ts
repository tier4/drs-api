// Mock data generators for DRS API

import type { ModuleStatus, RecordingStatus, PtpStatus, TopicStatus } from '../services/api'

export const mockModules: ModuleStatus[] = [
  {
    hostname: 'ecu0',
    address: '192.168.20.1:50051',
    status: 'OK',
    status_detail: {
      services: {
        drs_sensor: 'active',
        drs_recorder: 'active',
      },
      recording: {
        status: 'recording',
        data_status: 'OK',
      },
      ptp: {
        offset_ns: 12345,
      },
    },
    disk: {
      usage_percentage: 45.5,
      free_bytes: 550000000000,
      total_bytes: 1000000000000,
    },
    environment: {
      sensing_system_id: 'sys001',
      module_id: 'ecu0',
    },
    enabled_services: ['system', 'ptp', 'services', 'disk'],
  },
  {
    hostname: 'ecu1',
    address: '192.168.20.2:50051',
    status: 'WARN',
    status_detail: {
      services: {
        drs_sensor: 'active',
        drs_recorder: 'inactive',
      },
      recording: {
        status: 'stopped',
        data_status: 'WARN',
      },
      ptp: {
        offset_ns: 850000,
      },
    },
    disk: {
      usage_percentage: 82.3,
      free_bytes: 177000000000,
      total_bytes: 1000000000000,
    },
    environment: {
      sensing_system_id: 'sys001',
      module_id: 'ecu1',
    },
    enabled_services: ['system', 'ptp', 'services', 'disk'],
  },
  {
    hostname: 'nas',
    address: '192.168.10.100:50051',
    status: 'OK',
    status_detail: {
      services: {
        drs_sensor: 'unknown',
        drs_recorder: 'unknown',
      },
      recording: {
        status: 'stopped',
        data_status: 'OK',
      },
      ptp: {
        offset_ns: 0,
      },
    },
    disk: {
      usage_percentage: 23.8,
      free_bytes: 7620000000000,
      total_bytes: 10000000000000,
    },
    environment: {
      sensing_system_id: 'sys001',
      module_id: 'nas',
    },
    enabled_services: ['system', 'disk'],
  },
]

export const mockRecordingStatus: RecordingStatus[] = [
  {
    hostname: 'ecu0',
    recording_status: 'recording',
    data_status: 'OK',
    hardware_id: 'ecu0',
  },
  {
    hostname: 'ecu1',
    recording_status: 'stopped',
    data_status: 'WARN',
    hardware_id: 'ecu1',
  },
]

export const mockPtpStatus: PtpStatus[] = [
  {
    hostname: 'ecu0',
    local_status: {
      clock_id: '0xffffffffffff001',
      master_offset_ns: 12345,
      gm_present: true,
    },
    remote_statuses: [
      {
        device_name: 'ecu1',
        ip_address: '192.168.20.2',
        is_reachable: true,
        status: {
          clock_id: '0xffffffffffff002',
          master_offset_ns: 850000,
          gm_present: true,
        },
      },
      {
        device_name: 'nas',
        ip_address: '192.168.10.100',
        is_reachable: false,
        error_message: 'Connection timeout',
      },
    ],
  },
  {
    hostname: 'ecu1',
    local_status: {
      clock_id: '0xffffffffffff002',
      master_offset_ns: 850000,
      gm_present: true,
    },
    remote_statuses: [
      {
        device_name: 'ecu0',
        ip_address: '192.168.20.1',
        is_reachable: true,
        status: {
          clock_id: '0xffffffffffff001',
          master_offset_ns: 12345,
          gm_present: true,
        },
      },
    ],
  },
]

export const mockTopicStatus: Record<string, TopicStatus[]> = {
  ecu0: [
    {
      topic_name: '/camera/front/image_raw',
      rate_hz: 30.2,
      status: 'OK',
    },
    {
      topic_name: '/camera/rear/image_raw',
      rate_hz: 29.8,
      status: 'OK',
    },
    {
      topic_name: '/lidar/points',
      rate_hz: 10.1,
      status: 'OK',
    },
    {
      topic_name: '/imu/data',
      rate_hz: 100.5,
      status: 'OK',
    },
  ],
  ecu1: [
    {
      topic_name: '/camera/left/image_raw',
      rate_hz: 5.2,
      status: 'WARN',
    },
    {
      topic_name: '/camera/right/image_raw',
      rate_hz: 0.0,
      status: 'ERROR',
    },
    {
      topic_name: '/gps/fix',
      rate_hz: 1.0,
      status: 'OK',
    },
  ],
}

// State management for mock data
let isRecording = true
let recordingStoppedCount = 0

export function toggleRecording(start: boolean) {
  isRecording = start

  // Update recording status in modules and recording status
  mockModules.forEach((module) => {
    if (module.hostname !== 'nas') {
      module.status_detail.recording.status = start ? 'recording' : 'stopped'
    }
  })

  mockRecordingStatus.forEach((status) => {
    status.recording_status = start ? 'recording' : 'stopped'
  })
}

export function getCurrentRecordingStatus() {
  return isRecording
}

export function simulateRecordingStop() {
  recordingStoppedCount++
  console.log(`[Mock] Recording stopped ${recordingStoppedCount} times`)
}

export function simulateRecordingStart() {
  console.log('[Mock] Recording started')
}
