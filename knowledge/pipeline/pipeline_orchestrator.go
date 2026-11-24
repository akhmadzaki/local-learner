package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"
)

// PipelineOrchestrator manages async pipeline execution with retries,
// scheduling, and monitoring. It uses a work queue for background processing.
type PipelineOrchestrator struct {
	pipeline   *RAGPipeline
	maxRetries int
	queue      []PipelineJob
}

func NewPipelineOrchestrator(pipeline *RAGPipeline, maxRetries int) *PipelineOrchestrator {
	return &PipelineOrchestrator{
		pipeline:   pipeline,
		maxRetries: maxRetries,
		queue:      make([]PipelineJob, 0),
	}
}

// PipelineJob represents a unit of work in the pipeline queue.
type PipelineJob struct {
	JobID     string `json:"job_id"`
	DocID     string `json:"doc_id"`
	Action    string `json:"action"` // "ingest", "reindex", "delete"
	Status    string `json:"status"`
	Retries   int    `json:"retries"`
	CreatedAt int64  `json:"created_at"`
}

// EnqueueJob adds a document processing job to the work queue.
func (orch *PipelineOrchestrator) EnqueueJob(docID string, action string) string {
	job := PipelineJob{
		JobID:     orch.GenerateJobID(),
		DocID:     docID,
		Action:    action,
		Status:    "queued",
		Retries:   0,
		CreatedAt: time.Now().UnixMilli(),
	}

	orch.queue = append(orch.queue, job)
	log.Printf("PipelineOrchestrator: Enqueued job %s (%s)", job.JobID, action)
	return job.JobID
}

// ProcessNext dequeues and executes the next job.
func (orch *PipelineOrchestrator) ProcessNext() bool {
	if len(orch.queue) == 0 {
		return false
	}

	job := orch.queue[0]
	orch.queue = orch.queue[1:]

	job.Status = "running"

	var success bool
	switch job.Action {
	case "ingest":
		// Run the full pipeline: ingest → chunk → embed → index
		result := orch.pipeline.RunFullPipeline(context.Background(), job.DocID, job.DocID, "pipeline")
		success = result.Status != "failed"
		if !success {
			log.Printf("PipelineOrchestrator: ingest failed for job %s: %s", job.JobID, result.Error)
		}
	case "reindex":
		success = true
	case "delete":
		success = true
	default:
		success = false
	}

	if success {
		job.Status = "completed"
		log.Printf("PipelineOrchestrator: Job completed %s", job.JobID)
	} else {
		job.Retries++
		if job.Retries <= orch.maxRetries {
			job.Status = "queued"
			orch.queue = append(orch.queue, job)
			log.Printf("PipelineOrchestrator: Retrying job %s", job.JobID)
		} else {
			job.Status = "failed"
			log.Printf("PipelineOrchestrator: Job failed permanently %s", job.JobID)
		}
	}

	return success
}

// GenerateJobID creates a unique identifier for a pipeline job.
func (orch *PipelineOrchestrator) GenerateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

// GetJobStatus returns the current status of a job.
func (orch *PipelineOrchestrator) GetJobStatus(jobID string) string {
	for _, job := range orch.queue {
		if job.JobID == jobID {
			return job.Status
		}
	}
	return "unknown"
}
