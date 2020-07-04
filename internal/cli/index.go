package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/progrhyme/binq/internal/erron"
	"github.com/progrhyme/binq/schema"
	"golang.org/x/crypto/ssh/terminal"
)

type indexRunner interface {
	runner
	getIndexOpts() indexFlavor
	getPrevRawIndex() []byte
	setPrevRawIndex([]byte)
}

type indexFlavor interface {
	getYes() *bool
}

type indexCmd struct {
	prevRawIndex []byte
	*commonCmd
	option *indexOpts
}

type indexOpts struct {
	yes *bool
	*commonOpts
}

func (cmd *indexCmd) getIndexOpts() indexFlavor {
	return cmd.option
}

func (cmd *indexCmd) getPrevRawIndex() (b []byte) {
	return cmd.prevRawIndex
}

func (cmd *indexCmd) setPrevRawIndex(b []byte) {
	cmd.prevRawIndex = b
}

func (opt *indexOpts) getYes() (y *bool) {
	return opt.yes
}

func resolveIndexPathByArg(arg string) (pathIndex string, err error) {
	pathIndex = arg
	if strings.HasSuffix(pathIndex, ".json") {
		if filepath.Base(pathIndex) != "index.json" {
			err = fmt.Errorf("INDEX JSON filename must be \"index.json\". Given: %s", pathIndex)
			return pathIndex, err
		}
	} else {
		pathIndex = filepath.Join(pathIndex, "index.json")
	}
	return pathIndex, err
}

func decodeIndex(cmd indexRunner, file string) (idx *schema.Index, err error) {
	if _, _err := os.Stat(file); os.IsNotExist(_err) {
		err = fmt.Errorf("Index file not found: %s", file)
		return idx, err
	}

	raw, _err := ioutil.ReadFile(file)
	if _err != nil {
		err = erron.Errorwf(_err, "Error! Can't read item file: %s", file)
		return idx, err
	}

	idx, _err = schema.DecodeIndexJSON(raw)
	if _err != nil {
		err = erron.Errorwf(_err, "Error! Can't decode Index JSON: %s", file)
		return idx, err
	}

	cmd.setPrevRawIndex(raw)
	return idx, nil
}

func writeNewIndex(cmd indexRunner, idx *schema.Index, fileIndex string) (err error) {
	newRawIndex, _err := idx.Print(true)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to encode new Index")
	}

	fromFile := "<Null>"
	if len(cmd.getPrevRawIndex()) > 0 {
		fromFile = fileIndex
	}

	diff, err := getDiff(diffArgs{
		textA: strings.TrimRight(string(cmd.getPrevRawIndex()), "\r\n"),
		textB: string(newRawIndex),
		fileA: fromFile,
		fileB: fileIndex,
	})
	if err != nil {
		return err
	}
	if diff == "" {
		fmt.Fprintln(cmd.getErrs(), "Index has no change")
		return nil
	}

	yes := *(cmd.getIndexOpts().getYes())
	if !yes {
		fprintDiff(cmd.getOuts(), diff)
	}
	if terminal.IsTerminal(0) && !yes {
		fmt.Fprintf(cmd.getErrs(), "Write %s. Okay? (Y/n) ", fileIndex)
		stdin := bufio.NewScanner(os.Stdin)
		stdin.Scan()
		ans := stdin.Text()
		if strings.HasPrefix(ans, "n") || strings.HasPrefix(ans, "N") {
			fmt.Fprintln(cmd.getErrs(), "Canceled")
			return errCanceled
		}
	}

	return writeFile(fileIndex, newRawIndex, func() {
		fmt.Fprintf(cmd.getOuts(), "Saved %s\n", fileIndex)
	})
}