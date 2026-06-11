import { AlertTriangle, CheckCircle2, Eye, RefreshCw, UploadCloud } from 'lucide-react'
import { emptyProfile } from '../lib/format'
import { modelValidationSummary, validateModelProfiles } from '../lib/modelsValidation'
import { Panel, SecretInput } from '../components/common'

const VALIDATION_COPY = {
  valid: 'Validation passed',
  blocked: 'Fix blocking issues before writing',
  warnings: 'Warnings',
  errors: 'Blocking issues',
  saveDisabled: 'Fix red validation items before writing mykey.py.',
  keys: {
    varNameRequired: 'Variable name is required',
    varNameInvalid: 'Variable name must be a Python identifier',
    varNameDuplicate: 'Variable name is duplicated',
    modelRequired: 'Model is required',
    apiBaseProtocol: 'API Base should start with http:// or https://',
    maxRetriesInvalid: 'Retries must be zero or greater',
    readTimeoutInvalid: 'Timeout must be greater than zero',
    apiKeyEmpty: 'API Key is empty; use only for local or unauthenticated endpoints'
  }
}

function ValidationList({ title, items, copy, tone }) {
  if (!items?.length) return null
  return <div className={`validation-list ${tone}`}><span>{title}</span><ul>{items.map((key) => <li key={key}>{copy.keys[key] || key}</li>)}</ul></div>
}

export function Models({ t, profiles, setProfiles, patchProfile, importModels, previewModels, saveModels, modelPreview }) {
  const validation = validateModelProfiles(profiles)
  const summary = modelValidationSummary(validation)
  const copy = VALIDATION_COPY
  const hasErrors = summary.errors > 0

  return <section>
    <div className="model-top">
      <div><h3>{t.nav.models}</h3><p>{t.hints.previewHelp}</p></div>
      <div className="actions">
        <button onClick={importModels}><RefreshCw size={14}/>{t.hints.modelSource}</button>
        <button onClick={() => setProfiles([...profiles, emptyProfile(profiles.length)])}>{t.hints.addProfile}</button>
        <button onClick={previewModels}><Eye size={14}/>{t.hints.preview}</button>
        <button onClick={saveModels} disabled={hasErrors} title={hasErrors ? copy.saveDisabled : t.hints.writeMykey}><UploadCloud size={14}/>{t.hints.writeMykey}</button>
      </div>
    </div>
    <div className={`model-validation-summary ${hasErrors ? 'has-errors' : 'ok'}`} role={hasErrors ? 'alert' : 'status'}>
      {hasErrors ? <AlertTriangle size={16}/> : <CheckCircle2 size={16}/>}
      <b>{hasErrors ? copy.blocked : copy.valid}</b>
      <span>{copy.errors}: {summary.errors}</span>
      <span>{copy.warnings}: {summary.warnings}</span>
    </div>
    <div className="models-layout">
      <div className="profiles">{profiles.map((p, idx) => {
        const item = validation[idx] || { errors: [], warnings: [], ok: true }
        return <div className={`profile ${item.errors.length ? 'has-errors' : item.warnings.length ? 'has-warnings' : ''}`} key={idx}>
          <div className="profile-head">
            <b>#{idx + 1} {p.name || p.var_name || t.nav.models}</b>
            <span className={`profile-status ${item.errors.length ? 'bad' : item.warnings.length ? 'warn' : 'ok'}`}>{item.errors.length ? `${copy.errors}: ${item.errors.length}` : item.warnings.length ? `${copy.warnings}: ${item.warnings.length}` : copy.valid}</span>
          </div>
          <ValidationList title={copy.errors} items={item.errors} copy={copy} tone="error"/>
          <ValidationList title={copy.warnings} items={item.warnings} copy={copy} tone="warning"/>
          <div className="form-grid">
            <label>{t.fields.varName}<input value={p.var_name || ''} onChange={(e) => patchProfile(idx, { var_name: e.target.value })}/></label>
            <label>{t.fields.type}<input value={p.type || ''} onChange={(e) => patchProfile(idx, { type: e.target.value })}/></label>
            <label>{t.fields.name}<input value={p.name || ''} onChange={(e) => patchProfile(idx, { name: e.target.value })}/></label>
            <label>{t.fields.model}<input value={p.model || ''} onChange={(e) => patchProfile(idx, { model: e.target.value })}/></label>
            <label className="span2">{t.fields.apiBase}<input value={p.apibase || ''} onChange={(e) => patchProfile(idx, { apibase: e.target.value })}/></label>
            <label className="span2">{t.fields.apiKey}<SecretInput value={p.apikey} onChange={(v) => patchProfile(idx, { apikey: v })} t={t}/></label>
            <label>{t.fields.stream}<select value={String(!!p.stream)} onChange={(e) => patchProfile(idx, { stream: e.target.value === 'true' })}><option value="true">true</option><option value="false">false</option></select></label>
            <label>{t.fields.maxRetries}<input type="number" value={p.max_retries ?? 3} onChange={(e) => patchProfile(idx, { max_retries: Number(e.target.value) })}/></label>
            <label>{t.fields.readTimeout}<input type="number" value={p.read_timeout ?? 300} onChange={(e) => patchProfile(idx, { read_timeout: Number(e.target.value) })}/></label>
            <label>{t.fields.reasoningEffort}<input value={p.reasoning_effort || ''} onChange={(e) => patchProfile(idx, { reasoning_effort: e.target.value })}/></label>
          </div>
        </div>
      })}</div>
      <Panel title={t.lists.generatedPreview} className="preview"><pre>{modelPreview || t.empty}</pre></Panel>
    </div>
  </section>
}
