package instruction

import (
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Label struct {
	instructions.LabelCommand
}

func NewLabel(i instructions.LabelCommand) *Label {
	return &Label{LabelCommand: i}
}

func (i *Label) UsesBuildContext() bool {
	return false
}

func (i *Label) LabelsAsList() []string {
	var labels []string
	for _, item := range i.Labels {
		labels = append(labels, fmt.Sprintf("%s=%s", item.Key, item.Value))
	}
	return labels
}
