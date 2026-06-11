import test from 'node:test'
import assert from 'node:assert/strict'
import { validateModelProfiles, modelValidationSummary } from './modelsValidation.js'

test('validateModelProfiles blocks invalid saves before API calls', () => {
  const results = validateModelProfiles([
    { var_name: 'MODEL_1', model: 'gpt-4o', apibase: 'https://api.example.test', max_retries: 2, read_timeout: 300 },
    { var_name: 'MODEL_1', model: '', apibase: 'localhost:8000', max_retries: -1, read_timeout: 0 }
  ])

  assert.equal(results[0].ok, false)
  assert.deepEqual(results[0].warnings, ['apiKeyEmpty'])
  assert.equal(results[1].ok, false)
  assert.match(results[1].errors.join(','), /varNameDuplicate/)
  assert.match(results[1].errors.join(','), /modelRequired/)
  assert.match(results[1].errors.join(','), /maxRetriesInvalid/)
  assert.match(results[1].errors.join(','), /readTimeoutInvalid/)
  assert.match(results[1].warnings.join(','), /apiBaseProtocol/)
})

test('modelValidationSummary marks valid profile sets ready', () => {
  const summary = modelValidationSummary(validateModelProfiles([
    { var_name: 'MODEL_A', model: 'gpt-4o-mini', apibase: 'http://127.0.0.1:8000', apikey: 'set', max_retries: 0, read_timeout: 1 }
  ]))

  assert.deepEqual(summary, { errors: 0, warnings: 0, ready: true })
})
