package types

// EventSource contains information for an event.
type EventSource struct {
	// Component from which the event is generated.
	// +optional
	Component string `json:"component,omitempty"`
	// Node name on which the event is generated.
	// +optional
	Host string `json:"host,omitempty"`
}

// Event is a report of an event somewhere in the cluster.  Events
// have a limited retention time and triggers and messages may evolve
// with time.  Event consumers should not rely on the timing of an event
// with a given Reason reflecting a consistent underlying trigger, or the
// continued existence of events with that Reason.  Events should be
// treated as informative, best-effort, supplemental data.
type Event struct {
	TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	ObjectMeta `json:"metadata"`

	// The object that this event is about.
	InvolvedObject ObjectReference `json:"involvedObject"`

	// This should be a short, machine understandable string that gives the reason
	// for the transition into the object's current status.
	// TODO: provide exact specification for format.
	// +optional
	Reason string `json:"reason,omitempty"`

	// A human-readable description of the status of this operation.
	// TODO: decide on maximum length.
	// +optional
	Message string `json:"message,omitempty"`

	// The component reporting this event. Should be a short machine understandable string.
	// +optional
	Source EventSource `json:"source,omitempty"`

	// The time at which the event was first recorded. (Time of server receipt is in TypeMeta.)
	// +optional
	FirstTimestamp Time `json:"firstTimestamp,omitempty"`

	// The time at which the most recent occurrence of this event was recorded.
	// +optional
	LastTimestamp Time `json:"lastTimestamp,omitempty"`

	// The number of times this event has occurred.
	// +optional
	Count int32 `json:"count,omitempty"`

	// Type of this event (Normal, Warning), new types could be added in the future
	// +optional
	Type string `json:"type,omitempty"`

	// What action was taken/failed regarding to the Regarding object.
	// +optional
	Action string `json:"action,omitempty"`

	// Optional secondary object for more complex actions.
	// +optional
	Related *ObjectReference `json:"related,omitempty"`

	// Name of the controller that emitted this Event, e.g. `kubernetes.io/kubelet`.
	// +optional
	ReportingController string `json:"reportingComponent"`

	// ID of the controller instance, e.g. `kubelet-xyzf`.
	// +optional
	ReportingInstance string `json:"reportingInstance"`
}
