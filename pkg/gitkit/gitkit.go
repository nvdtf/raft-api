package gitkit

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/google/go-github/v45/github"
	"github.com/onflow/flow-cli/pkg/flowkit"
)

func test() {
	var _ flowkit.ReaderWriter = (*GitReaderWriter)(nil)
}

type GitReaderWriter struct {
	owner string
	repo  string
}

func NewGitReaderWriter(owner string, repo string) flowkit.ReaderWriter {
	return &GitReaderWriter{
		owner: owner,
		repo:  repo,
	}
}

func (grw *GitReaderWriter) ReadFile(source string) ([]byte, error) {
	ctx := context.Background()

	fmt.Printf("Reading %s ...\n", source)

	client := github.NewClient(nil)

	result, _, _, err := client.Repositories.GetContents(ctx, grw.owner, grw.repo, source, nil)
	if err != nil {
		return nil, err
	}

	// fmt.Println(*result.Content)

	contents, err := base64.StdEncoding.DecodeString(*result.Content)
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(contents))

	return contents, nil
}

func (grw *GitReaderWriter) WriteFile(filename string, data []byte, perm os.FileMode) error {
	fmt.Printf("Write not implemented: %s (%d)\n", filename, len(data))
	return nil
}
