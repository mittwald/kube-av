package engine

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	avv1beta1 "github.com/mittwald/kube-av/pkg/apis/av/v1beta1"
)

var matchRE *regexp.Regexp

func init() {
	matchRE = regexp.MustCompile(`^(.*): (.*) FOUND$`)
}

func (c *clamAVEngine) Execute(ctx context.Context, _ *avv1beta1.VirusScan, scanDirs []string) (*ScanReport, error) {
	stdout := bytes.Buffer{}

	args := []string{"-i", "--no-summary"}
	args = append(args, scanDirs...)

	tee := io.MultiWriter(&stdout, os.Stdout)

	cmd := exec.CommandContext(ctx, "clamscan", args...)
	cmd.Stdout = tee
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		// don't return an error in case of a non-zero exit code. In case ClamAV
		// matches a file, this is to be expected.
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, err
		}
	}

	report := ScanReport{}

	// TODO: conditional?, if exitcode == 0, nothing has been found?
	outputLines := strings.Split(stdout.String(), "\n")
	for i := range outputLines {
		match := matchRE.FindStringSubmatch(outputLines[i])
		if len(match) != 3 {
			continue
		}

		item := ScanReportItem{
			FilePath:         match[1],
			MatchedSignature: match[2],
		}

		report.InfectedFiles = append(report.InfectedFiles, item)
	}

	return &report, nil
}
