package output

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/aelnahas/pomo/sessions"
	"github.com/aelnahas/pomo/task"
)

var header = []string{"id", "title", "status", "sessions"}
var sessionsHeader = []string{"current", "next", "count"}

func Printlist(tasks ...task.Task) {
	if len(tasks) == 0 {
		return
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	defer writer.Flush()

	fmt.Fprintln(writer, strings.Join(header, "\t"))
	for _, entry := range tasks {
		fmt.Fprintln(writer, entry.Format())
	}
}

func PrintSession(current sessions.Type, next sessions.Type, count int) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	defer writer.Flush()

	fmt.Fprintln(writer, strings.Join(sessionsHeader, "\t"))
	fmt.Fprintln(writer, strings.Join([]string{string(current), string(next), fmt.Sprintf("%d", count)}, "\t"))
}
