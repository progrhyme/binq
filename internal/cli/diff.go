package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/binqry/binq/internal/erron"
	"github.com/gookit/color"
	"github.com/mattn/go-isatty"
	"github.com/pmezard/go-difflib/difflib"
)

type diffArgs struct {
	textA, textB, fileA, fileB string
}

func getDiff(args diffArgs) (diff string, err error) {
	obj := difflib.UnifiedDiff{
		A:        difflib.SplitLines(args.textA),
		B:        difflib.SplitLines(args.textB),
		FromFile: args.fileA,
		ToFile:   args.fileB,
		Context:  3,
	}
	diff, _err := difflib.GetUnifiedDiffString(obj)
	if _err != nil {
		return diff, erron.Errorwf(_err, "Failed to get diff")
	}

	return diff, nil
}

func fprintDiff(out io.Writer, diff string) {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Fprintln(out, colorizeDiff(diff))
	} else {
		fmt.Fprintln(out, diff)
	}
}

func colorizeDiff(diff string) (colored string) {
	lines := strings.Split(diff, "\n")
	for i, s := range lines {
		switch {
		case strings.HasPrefix(s, "---"):
			lines[i] = color.Danger.Render(s)
		case strings.HasPrefix(s, "+++"):
			lines[i] = color.Success.Render(s)
		case strings.HasPrefix(s, "-"):
			lines[i] = color.Red.Render(s)
		case strings.HasPrefix(s, "+"):
			lines[i] = color.Green.Render(s)
		case strings.HasPrefix(s, "@@"):
			lines[i] = color.Note.Render(s)
		default:
			// Nothing to do
		}
	}
	return strings.Join(lines, "\n")
}
