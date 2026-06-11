import { Save, Search } from 'lucide-react'
import { Panel } from '../components/common'
import { fileEditorDirty, saveReviewText } from '../lib/filesSafety'

export function FilesPage({
  t,
  filePath,
  setFilePath,
  fileList,
  fileContent,
  loadedFileContent = '',
  loadedFilePath = '',
  setFileContent,
  fileSearch,
  setFileSearch,
  searchHits,
  tailLines,
  setTailLines,
  loadFiles,
  readFile,
  tailFile,
  saveFile,
  runSearch,
}) {
  const dirty = fileEditorDirty(fileContent, loadedFileContent)
  const retargeted = Boolean(loadedFilePath && filePath && loadedFilePath !== filePath)
  const saveReview = saveReviewText({ path: filePath, loadedPath: loadedFilePath, dirty })
  const saveDisabled = !filePath || !dirty
  const fileListEmpty = !fileList?.length
  const searchEmpty = !searchHits?.length
  const searchHint = fileSearch ? 'No matches found.' : 'Enter search text, then run search.'
  return (
    <section>
      <div className="workspace">
        <Panel title={t.lists.fileList}>
          <div className="inline-form">
            <input value={filePath} onChange={e => setFilePath(e.target.value)} placeholder={t.hints.filePath}/>
            <button onClick={() => loadFiles(filePath)}>{t.read}</button>
          </div>
          <div className="inline-form">
            <input value={fileSearch} onChange={e => setFileSearch(e.target.value)} placeholder={t.hints.searchText}/>
            <button onClick={runSearch}><Search size={14}/>{t.search}</button>
          </div>
          <div className="inline-form">
            <input type="number" value={tailLines} onChange={e => setTailLines(Number(e.target.value))}/>
            <span>{t.hints.tailLines}</span>
            <button onClick={() => tailFile(filePath)}>{t.tail || 'Tail'}</button>
            <button onClick={saveFile} disabled={saveDisabled} title={saveReview}><Save size={14}/>{t.save}</button>
          </div>
          <div className="file-list">
            {fileListEmpty && <p className="muted">{t.hints?.fileListEmpty || 'No files found. Choose a folder and read again.'}</p>}
            {fileList.map(e => <button key={e.path} onClick={() => e.kind === 'dir' ? loadFiles(e.path) : readFile(e.path)}>{e.kind === 'dir' ? '📁' : '📄'} {e.path}</button>)}
          </div>
          <h4>{t.lists.searchResults}</h4>
          {searchEmpty && <p className="muted">{t.hints?.searchEmpty || searchHint}</p>}
          {searchHits.map(h => <button className="hit" key={`${h.path}:${h.line}`} onClick={() => readFile(h.path)}>{h.path}:{h.line} · {h.preview}</button>)}
        </Panel>
        <Panel title={t.lists.filePreview} className="log-panel">
          <div className="file-editor-toolbar">
            <span className={dirty ? 'status-pill warn' : 'status-pill ok'}>{dirty ? 'Unsaved changes' : 'Saved/clean'}</span>
            {loadedFilePath && <span className="muted">Loaded: {loadedFilePath}</span>}
            {retargeted && <span className="status-pill bad">Save target changed</span>}
          </div>
          <div className={`file-save-review ${retargeted ? 'bad' : dirty ? 'warn' : 'ok'}`} role="status" aria-live="polite">
            {saveReview}
          </div>
          <textarea className="file-editor" value={fileContent} onChange={e => setFileContent(e.target.value)} placeholder={t.empty}/>
        </Panel>
      </div>
    </section>
  )
}
