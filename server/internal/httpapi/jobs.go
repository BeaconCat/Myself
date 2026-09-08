package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// 压缩任务：请求立即返回任务 ID，后台串行处理，前端轮询进度。
// 单进程内存态，重启即丢；同一时刻只允许一个任务运行。

type compressJob struct {
	ID        string           `json:"id"`
	Total     int              `json:"total"`
	Done      int              `json:"done"`
	Running   bool             `json:"running"`
	Current   string           `json:"current"`
	Results   []compressResult `json:"results"`
	StartedAt string           `json:"startedAt"`
	EndedAt   string           `json:"endedAt,omitempty"`
}

type jobRegistry struct {
	mu   sync.Mutex
	jobs map[string]*compressJob
	// order 记录创建顺序，用于淘汰旧任务
	order []string
}

const maxKeptJobs = 10

func newJobRegistry() *jobRegistry {
	return &jobRegistry{jobs: map[string]*compressJob{}}
}

func (r *jobRegistry) running() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, j := range r.jobs {
		if j.Running {
			return true
		}
	}
	return false
}

func (r *jobRegistry) create(total int) *compressJob {
	raw := make([]byte, 8)
	_, _ = rand.Read(raw)
	job := &compressJob{
		ID:        hex.EncodeToString(raw),
		Total:     total,
		Running:   true,
		Results:   []compressResult{},
		StartedAt: nowISO(),
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.ID] = job
	r.order = append(r.order, job.ID)
	for len(r.order) > maxKeptJobs {
		delete(r.jobs, r.order[0])
		r.order = r.order[1:]
	}
	return job
}

// snapshot 返回任务副本（避免把内部指针交给 JSON 编码时被并发修改）。
func (r *jobRegistry) snapshot(id string) (compressJob, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[id]
	if !ok {
		return compressJob{}, false
	}
	cp := *j
	cp.Results = append([]compressResult(nil), j.Results...)
	return cp, true
}

func (r *jobRegistry) update(id string, fn func(j *compressJob)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if j, ok := r.jobs[id]; ok {
		fn(j)
	}
}

// GET /admin/quality/jobs/{id} 任务进度
func (s *Server) compressJobStatus(w http.ResponseWriter, r *http.Request) {
	job, ok := s.jobs.snapshot(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// runCompressJob 后台串行压缩；每张图完成即更新进度。
func (s *Server) runCompressJob(id string, names []string, quality int) {
	for _, name := range names {
		s.jobs.update(id, func(j *compressJob) { j.Current = name })
		res := s.compressNamed(name, quality)
		s.jobs.update(id, func(j *compressJob) {
			j.Done++
			if res != nil {
				j.Results = append(j.Results, *res)
			}
		})
	}
	s.jobs.update(id, func(j *compressJob) {
		j.Running = false
		j.Current = ""
		j.EndedAt = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	})
}
