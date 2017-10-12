package cluster

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	corev1 "k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
)

const (
	LEVEL_0 = iota
	LEVEL_1
	LEVEL_2
	LEVEL_3
)

func DescribeEvents(el *corev1.EventList, w *PrefixWriter) {
	if len(el.Items) == 0 {
		w.Write(LEVEL_0, "No events.\n")
		return
	}
	sort.Sort(SortableEvents(el.Items))
	w.Write(LEVEL_0, "Events:\n  FirstSeen\tLastSeen\tCount\tFrom\tSubObjectPath\tType\tReason\tMessage\n")
	w.Write(LEVEL_1, "---------\t--------\t-----\t----\t-------------\t--------\t------\t-------\n")
	for _, e := range el.Items {
		w.Write(LEVEL_1, "%s\t%s\t%d\t%v\t%v\t%v\t%v\t%v\n",
			e.FirstTimestamp,
			e.LastTimestamp,
			e.Count,
			e.Source,
			e.InvolvedObject.FieldPath,
			e.Type,
			e.Reason,
			e.Message)
	}
}

type SortableResourceNames []corev1.ResourceName

func (list SortableResourceNames) Len() int {
	return len(list)
}

func (list SortableResourceNames) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list SortableResourceNames) Less(i, j int) bool {
	return list[i] < list[j]
}

// SortedResourceNames returns the sorted resource names of a resource list.
func SortedResourceNames(list corev1.ResourceList) []corev1.ResourceName {
	resources := make([]corev1.ResourceName, 0, len(list))
	for res := range list {
		resources = append(resources, res)
	}
	sort.Sort(SortableResourceNames(resources))
	return resources
}

type SortableEvents []corev1.Event

func (list SortableEvents) Len() int {
	return len(list)
}

func (list SortableEvents) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list SortableEvents) Less(i, j int) bool {
	return list[i].LastTimestamp.Time.Before(list[j].LastTimestamp.Time)
}

/*job*/

type SortableJobs []*batchv1.Job

func (list SortableJobs) Len() int {
	return len(list)
}

func (list SortableJobs) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

//按距离当前的创建时间进行排序
func (list SortableJobs) Less(i, j int) bool {
	return list[i].CreationTimestamp.Time.After(list[j].CreationTimestamp.Time)
}

type PrefixWriter struct {
	out io.Writer
}

func (pw *PrefixWriter) Write(level int, format string, a ...interface{}) {
	levelSpace := "  "
	prefix := ""
	for i := 0; i < level; i++ {
		prefix += levelSpace
	}
	fmt.Fprintf(pw.out, prefix+format, a...)
}

func (pw *PrefixWriter) WriteLine(a ...interface{}) {
	fmt.Fprintln(pw.out, a...)
}

func tabbedString(f func(io.Writer) error) (string, error) {
	out := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	out.Init(buf, 0, 8, 1, '\t', 0)

	err := f(out)
	if err != nil {
		return "", err
	}

	out.Flush()
	str := string(buf.String())
	return str, nil
}
