import test from 'node:test'
import assert from 'node:assert/strict'
import { buildChatRoute, buildRoute, chatReturnRoute, normalizeAdminReturnRoute, parseRoute } from './routing.js'

const setLocation = (url) => {
  globalThis.window = { location: new URL(url) }
}

test('parseRoute maps aliases and keeps top-level goals routable', () => {
  setLocation('http://localhost/goals')
  assert.deepEqual(parseRoute(), { tab: 'goals', taskSubTab: 'services' })
  setLocation('http://localhost/runs')
  assert.deepEqual(parseRoute(), { tab: 'tasks', taskSubTab: 'runs' })
  setLocation('http://localhost/tmwd')
  assert.deepEqual(parseRoute(), { tab: 'control', taskSubTab: 'services' })
})

test('parseRoute prefers hash routes', () => {
  setLocation('http://localhost/settings#/tasks/reports')
  assert.deepEqual(parseRoute(), { tab: 'tasks', taskSubTab: 'reports' })
})

test('buildRoute normalizes invalid tabs and task sub tabs', () => {
  assert.equal(buildRoute('missing'), '/overview')
  assert.equal(buildRoute('tasks', 'missing'), '/tasks/services')
})


test('chat routes preserve and sanitize admin return targets', () => {
  setLocation('http://localhost/models')
  assert.equal(buildChatRoute('/tasks/scheduled'), '/chat?from=%2Ftasks%2Fscheduled')
  assert.equal(normalizeAdminReturnRoute('/tasks/scheduled'), '/tasks/scheduled')
  assert.equal(normalizeAdminReturnRoute('https://evil.example/phish'), '/overview')
  assert.equal(normalizeAdminReturnRoute('/not-a-console-route'), '/overview')
  assert.equal(normalizeAdminReturnRoute('/chat?from=%2Fmodels'), '/overview')
})

test('chatReturnRoute prefers safe from query and falls back to overview', () => {
  globalThis.window = { location: new URL('http://localhost/chat?from=%2Fmodels'), sessionStorage: { getItem: () => '', setItem: () => {} } }
  assert.equal(chatReturnRoute(), '/models')
  globalThis.window = { location: new URL('http://localhost/chat?from=https%3A%2F%2Fevil.example%2F'), sessionStorage: { getItem: () => '/tasks/runs', setItem: () => {} } }
  assert.equal(chatReturnRoute(), '/overview')
})
