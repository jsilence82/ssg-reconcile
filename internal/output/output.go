package output

import (
	"github.com/jsilence82/ssg-reconcile/internal/model"
)

// Renderer writes a ReconciliationReport to some destination.
type Renderer interface {
	Render(report *model.ReconciliationReport) error
}
