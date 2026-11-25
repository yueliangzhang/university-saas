import http from 'k6/http'
import { check, sleep } from 'k6'

const base = (__ENV.GVA_BASE_URL || 'http://localhost:8080').replace(/\/$/, '')
const token = __ENV.GVA_ADMIN_TOKEN
const studentToken = __ENV.GVA_STUDENT_TOKEN
const target = Number(__ENV.GVA_VUS || '25')
const requestOptions = {
  timeout: '15s',
  headers: { 'x-token': token || '' }
}

if (!token) throw new Error('Set GVA_ADMIN_TOKEN for a test university administrator')
if (!Number.isInteger(target) || target < 1) throw new Error('GVA_VUS must be a positive integer')

export const options = {
  stages: [
    { duration: '30s', target },
    { duration: '2m', target },
    { duration: '30s', target: 0 }
  ],
  thresholds: {
    'http_req_duration{endpoint:listing}': ['p(95)<=200'],
    'http_req_duration{endpoint:dashboard}': ['p(95)<=500'],
    http_req_failed: ['rate<0.001'],
    checks: ['rate>=0.999']
  }
}

function query(path, endpoint, config = requestOptions) {
  const result = http.get(base + path, { ...config, tags: { endpoint, name: endpoint } })
  check(result, { 'authorized request succeeds': res => {
    if (res.status !== 200) return false
    try { return res.json('code') === 0 } catch { return false }
  } })
}

export default function () {
  query('/api/uni/students?page=1&pageSize=20', 'listing')
  query('/api/uni/monitor/analytics?days=7', 'dashboard')
  if (studentToken) {
    query('/api/uni/student-portal/leaves?page=1&pageSize=10', 'listing', {
      timeout: '15s', headers: { Authorization: `Bearer ${studentToken}` }
    })
  }
  sleep(1)
}
