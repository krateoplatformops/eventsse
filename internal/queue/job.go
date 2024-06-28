package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const msgpackContentType = "application/msgpack"

// Job contains the information for a job to be published to a queue.
type Job struct {
	// ID of the job.
	ID string
	// Priority is the priority level.
	Priority Priority
	// Timestamp is the time of creation.
	Timestamp time.Time
	// Retries is the number of times this job can be processed before being rejected.
	Retries int32
	// ErrorType is the kind of error that made the job fail.
	ErrorType string
	// ContentType of the job
	ContentType string
	// Raw content of the Job
	Raw []byte
	// Acknowledger is the acknowledgement management system for the job.
	Acknowledger Acknowledger
}

// Acknowledger represents the object in charge of acknowledgement
// management for a job. When a job is acknowledged using any of the
// functions in this interface, it will be considered delivered by the
// queue.
type Acknowledger interface {
	// Ack is called when the Job has finished.
	Ack() error
	// Reject is called if the job has errored. The parameter indicates
	// whether the job should be put back in the queue or not.
	Reject(requeue bool) error
}

// NewJob creates a new Job with default values, a new unique ID and current
// timestamp.
func NewJob(id string) (*Job, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("id cannot be empty")
	}

	return &Job{
		ID:          id,
		Priority:    PriorityNormal,
		Timestamp:   time.Now(),
		ContentType: msgpackContentType,
	}, nil
}

// SetPriority sets job priority
func (j *Job) SetPriority(priority Priority) {
	j.Priority = priority
}

// Encode encodes the payload to the wire format used.
func (j *Job) Encode(payload interface{}) error {
	var err error
	j.Raw, err = json.Marshal(&payload)
	if err != nil {
		return err
	}

	return nil
}

// Decode decodes the payload from the wire format.
func (j *Job) Decode(payload interface{}) error {
	return json.Unmarshal(j.Raw, &payload)
}

// ErrCantAck is the error returned when the Job does not come from a queue
var ErrCantAck = errors.New("can't acknowledge this message, it does not come from a queue")

// Ack is called when the job is finished.
func (j *Job) Ack() error {
	if j.Acknowledger == nil {
		return ErrCantAck
	}
	return j.Acknowledger.Ack()
}

// Reject is called when the job errors. The parameter is true if and only if the
// job should be put back in the queue.
func (j *Job) Reject(requeue bool) error {
	if j.Acknowledger == nil {
		return ErrCantAck
	}
	return j.Acknowledger.Reject(requeue)
}

// Size returns the size of the message.
func (j *Job) Size() int {
	return len(j.Raw)
}
