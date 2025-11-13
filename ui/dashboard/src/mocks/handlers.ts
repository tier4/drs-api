// MSW handlers for DRS API endpoints

import { http, HttpResponse, delay } from 'msw'
import {
  mockModules,
  mockRecordingStatus,
  mockPtpStatus,
  mockTopicStatus,
  toggleRecording,
  simulateRecordingStart,
  simulateRecordingStop,
} from './data'

const API_BASE = '/api/v1'

export const handlers = [
  // Health check
  http.get('/health', () => {
    return HttpResponse.json({
      status: 'healthy',
      time: new Date().toISOString(),
    })
  }),

  // Get all modules
  http.get(`${API_BASE}/modules`, async () => {
    await delay(300) // Simulate network delay
    return HttpResponse.json({
      modules: mockModules,
    })
  }),

  // Get single module
  http.get(`${API_BASE}/modules/:hostname`, async ({ params }) => {
    await delay(200)
    const { hostname } = params
    const module = mockModules.find((m) => m.hostname === hostname)

    if (!module) {
      return HttpResponse.json(
        {
          error: 'module_not_found',
          message: 'Module not found or unreachable',
          details: { hostname },
        },
        { status: 404 }
      )
    }

    return HttpResponse.json(module)
  }),

  // Get recording status
  http.get(`${API_BASE}/recording/status`, async () => {
    await delay(250)
    return HttpResponse.json({
      recording_status: mockRecordingStatus,
    })
  }),

  // Start recording
  http.post(`${API_BASE}/recording/start`, async () => {
    await delay(500)
    toggleRecording(true)
    simulateRecordingStart()

    return HttpResponse.json({
      success: true,
      message: 'Recording started successfully on all modules',
    })
  }),

  // Stop recording
  http.post(`${API_BASE}/recording/stop`, async () => {
    await delay(500)
    toggleRecording(false)
    simulateRecordingStop()

    return HttpResponse.json({
      success: true,
      message: 'Recording stopped successfully on all modules',
    })
  }),

  // Pause recording
  http.post(`${API_BASE}/recording/pause`, async () => {
    await delay(400)
    return HttpResponse.json({
      success: true,
      message: 'Recording paused successfully on all modules',
    })
  }),

  // Resume recording
  http.post(`${API_BASE}/recording/resume`, async () => {
    await delay(400)
    toggleRecording(true)

    return HttpResponse.json({
      success: true,
      message: 'Recording resumed successfully on all modules',
    })
  }),

  // Get PTP status
  http.get(`${API_BASE}/ptp/status`, async () => {
    await delay(300)
    return HttpResponse.json({
      ptp_status: mockPtpStatus,
    })
  }),

  // Get topic status for a module
  http.get(`${API_BASE}/modules/:hostname/topics/status`, async ({ params }) => {
    await delay(250)
    const { hostname } = params
    const topics = mockTopicStatus[hostname as string]

    if (!topics) {
      return HttpResponse.json(
        {
          error: 'module_not_found',
          message: 'Module not found or topics unavailable',
        },
        { status: 404 }
      )
    }

    return HttpResponse.json({
      topics,
    })
  }),

  // Get services for a module
  http.get(`${API_BASE}/modules/:hostname/services`, async ({ params }) => {
    await delay(200)
    const { hostname } = params
    const module = mockModules.find((m) => m.hostname === hostname)

    if (!module) {
      return HttpResponse.json(
        { error: 'Module not found' },
        { status: 404 }
      )
    }

    if (!module.enabled_services.includes('services')) {
      return HttpResponse.json(
        { error: 'Service management disabled' },
        { status: 403 }
      )
    }

    return HttpResponse.json({
      services: [
        {
          name: 'services/drs_sensor',
          state: module.status_detail.services.drs_sensor,
          description: 'DRS Sensor Service',
        },
        {
          name: 'services/drs_recorder',
          state: module.status_detail.services.drs_recorder,
          description: 'DRS Recorder Service',
        },
      ],
    })
  }),

  // Start service
  http.post(`${API_BASE}/modules/:hostname/services/:service_name/start`, async ({ params }) => {
    await delay(800)
    const { hostname, service_name } = params
    const module = mockModules.find((m) => m.hostname === hostname)

    if (!module) {
      return HttpResponse.json({ error: 'Module not found' }, { status: 404 })
    }

    // Update mock data
    if (service_name === 'drs_sensor') {
      module.status_detail.services.drs_sensor = 'active'
    } else if (service_name === 'drs_recorder') {
      module.status_detail.services.drs_recorder = 'active'
    }

    return HttpResponse.json({
      success: true,
      message: 'Service operation completed successfully',
      service: {
        name: `services/${service_name}`,
        state: 'active',
      },
    })
  }),

  // Stop service
  http.post(`${API_BASE}/modules/:hostname/services/:service_name/stop`, async ({ params }) => {
    await delay(800)
    const { hostname, service_name } = params
    const module = mockModules.find((m) => m.hostname === hostname)

    if (!module) {
      return HttpResponse.json({ error: 'Module not found' }, { status: 404 })
    }

    // Update mock data
    if (service_name === 'drs_sensor') {
      module.status_detail.services.drs_sensor = 'inactive'
    } else if (service_name === 'drs_recorder') {
      module.status_detail.services.drs_recorder = 'inactive'
    }

    return HttpResponse.json({
      success: true,
      message: 'Service operation completed successfully',
      service: {
        name: `services/${service_name}`,
        state: 'inactive',
      },
    })
  }),

  // Restart service
  http.post(`${API_BASE}/modules/:hostname/services/:service_name/restart`, async ({ params }) => {
    await delay(1500)
    const { hostname, service_name } = params
    const module = mockModules.find((m) => m.hostname === hostname)

    if (!module) {
      return HttpResponse.json({ error: 'Module not found' }, { status: 404 })
    }

    // Update mock data
    if (service_name === 'drs_sensor') {
      module.status_detail.services.drs_sensor = 'active'
    } else if (service_name === 'drs_recorder') {
      module.status_detail.services.drs_recorder = 'active'
    }

    return HttpResponse.json({
      success: true,
      message: 'Service operation completed successfully',
      service: {
        name: `services/${service_name}`,
        state: 'active',
      },
    })
  }),

  // Restart services (sensor only)
  http.post(`${API_BASE}/modules/:hostname/services/restart`, async ({ params }) => {
    await delay(1500)
    const { hostname } = params
    const module = mockModules.find((m) => m.hostname === hostname)

    if (!module) {
      return HttpResponse.json({ error: 'Module not found' }, { status: 404 })
    }

    module.status_detail.services.drs_sensor = 'active'

    return HttpResponse.json({
      success: true,
      message: 'Sensor service restarted successfully',
    })
  }),

  // System restart
  http.post(`${API_BASE}/system/restart`, async () => {
    await delay(1000)
    console.log('[Mock] System restart requested')

    return HttpResponse.json({
      success: true,
      message: 'System operation completed successfully',
      delay_seconds: 5,
    })
  }),

  // System shutdown
  http.post(`${API_BASE}/system/shutdown`, async () => {
    await delay(1000)
    console.log('[Mock] System shutdown requested')

    return HttpResponse.json({
      success: true,
      message: 'System operation completed successfully',
      delay_seconds: 5,
    })
  }),

  // Module restart
  http.post(`${API_BASE}/modules/:hostname/restart`, async ({ params }) => {
    await delay(800)
    const { hostname } = params
    console.log(`[Mock] Module ${hostname} restart requested`)

    return HttpResponse.json({
      success: true,
      message: `Module ${hostname} restart scheduled`,
      delay_seconds: 5,
    })
  }),

  // Module shutdown
  http.post(`${API_BASE}/modules/:hostname/shutdown`, async ({ params }) => {
    await delay(800)
    const { hostname } = params
    console.log(`[Mock] Module ${hostname} shutdown requested`)

    return HttpResponse.json({
      success: true,
      message: `Module ${hostname} shutdown scheduled`,
      delay_seconds: 5,
    })
  }),
]
