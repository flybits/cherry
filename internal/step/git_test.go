package step

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		name          string
		output        string
		expectedOwner string
		expectedName  string
		expectedError string
	}{
		{
			name:          "Empty",
			output:        ``,
			expectedError: "failed to get git repository url",
		},
		{
			name: "Invalid",
			output: `
			origin	flybits/cherry (fetch)
			origin	flybits/cherry (push)
			`,
			expectedError: "failed to get git repository name",
		},
		{
			name: "SSH#1",
			output: `
			origin	git@github.com:flybits/cherry (fetch)
			origin	git@github.com:flybits/cherry (push)
			`,
			expectedOwner: "flybits",
			expectedName:  "cherry",
		},
		{
			name: "SSH#2",
			output: `
			origin	git@github.com:flybits/cherry.git (fetch)
			origin	git@github.com:flybits/cherry.git (push)
			`,
			expectedOwner: "flybits",
			expectedName:  "cherry",
		},
		{
			name: "HTTPS#1",
			output: `
			origin	https://github.com/flybits/cherry (fetch)
			origin	https://github.com/flybits/cherry (push)
			`,
			expectedOwner: "flybits",
			expectedName:  "cherry",
		},
		{
			name: "HTTPS#2",
			output: `
			origin	https://github.com/flybits/cherry.git (fetch)
			origin	https://github.com/flybits/cherry.git (push)
			`,
			expectedOwner: "flybits",
			expectedName:  "cherry",
		},
		{
			name: "CustomSSH#1",
			output: `
			origin	ssh://git@github.com/flybits/cherry (fetch)
			origin	ssh://git@github.com/flybits/cherry (push)
			`,
			expectedOwner: "flybits",
			expectedName:  "cherry",
		},
		{
			name: "CustomSSH#2",
			output: `
			origin	ssh://git@github.com/flybits/cherry.git (fetch)
			origin	ssh://git@github.com/flybits/cherry.git (push)
			`,
			expectedOwner: "flybits",
			expectedName:  "cherry",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			owner, name, err := parseGitURL(tc.output)

			if tc.expectedError != "" {
				assert.Equal(t, tc.expectedError, err.Error())
				assert.Empty(t, owner)
				assert.Empty(t, name)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOwner, owner)
				assert.Equal(t, tc.expectedName, name)
			}
		})
	}
}

func TestGitStatusMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitStatus{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
		})
	}
}

func TestGitStatusDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitStatus.Dry: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitStatus{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitStatusRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitStatus.Run: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitStatus{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				// TODO: Test results (IsClean)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitStatusRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitStatus{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetRepoMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetRepo{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
		})
	}
}

func TestGitGetRepoDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitGetRepo.Dry: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetRepo{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetRepoRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitGetRepo.Run: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetRepo{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.NotEmpty(t, step.Result.Owner)
				assert.NotEmpty(t, step.Result.Name)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetRepoRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetRepo{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetBranchMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetBranch{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
		})
	}
}

func TestGitGetBranchDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitGetBranch.Dry: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetBranch{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetBranchRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitGetBranch.Run: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetBranch{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.NotEmpty(t, step.Result.Name)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetBranchRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetBranch{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetHEADMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetHEAD{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
		})
	}
}

func TestGitGetHEADDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitGetHEAD.Dry: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "FullSHA",
			workDir: ".",
		},
		{
			name:    "ShortSHA",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetHEAD{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetHEADRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitGetHEAD.Run: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "FullSHA",
			workDir: ".",
		},
		{
			name:    "ShortSHA",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetHEAD{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Len(t, step.Result.SHA, 40)
				assert.Len(t, step.Result.ShortSHA, 7)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetHEADRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetHEAD{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitAddMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitAdd{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
		})
	}
}

func TestGitAddDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		files         []string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			files:         []string{"."},
			expectedError: `GitAdd.Dry: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitAdd{
				WorkDir: tc.workDir,
				Files:   tc.files,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitAddRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		files         []string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			files:         []string{"."},
			expectedError: `GitAdd.Run: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitAdd{
				WorkDir: tc.workDir,
				Files:   tc.files,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitAddRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		files         []string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			files:         []string{"."},
			expectedError: `GitAdd.Revert: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitAdd{
				WorkDir: tc.workDir,
				Files:   tc.files,
			}

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitCommitMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitCommit{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
		})
	}
}

func TestGitCommitDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		message       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			message:       "test message",
			expectedError: `GitCommit.Dry: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitCommit{
				WorkDir: tc.workDir,
				Message: tc.message,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitCommitRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		message       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			message:       "test message",
			expectedError: `GitCommit.Run: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitCommit{
				WorkDir: tc.workDir,
				Message: tc.message,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitCommitRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		message       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			message:       "test message",
			expectedError: `GitCommit.Revert: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitCommit{
				WorkDir: tc.workDir,
				Message: tc.message,
			}

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitTagMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitTag{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
		})
	}
}

func TestGitTagDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		annotation    string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "",
			expectedError: "GitTag.Dry: exit status 128 fatal: not a git repository (or any of the parent directories): .git",
		},
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "annotation message",
			expectedError: `GitTag.Dry: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitTag{
				WorkDir:    tc.workDir,
				Tag:        tc.tag,
				Annotation: tc.annotation,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitTagRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		annotation    string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "",
			expectedError: "GitTag.Run: exit status 128 fatal: not a git repository (or any of the parent directories): .git",
		},
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "annotation message",
			expectedError: `GitTag.Run: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitTag{
				WorkDir:    tc.workDir,
				Tag:        tc.tag,
				Annotation: tc.annotation,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitTagRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		annotation    string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "",
			expectedError: "GitTag.Revert: exit status 128 fatal: not a git repository (or any of the parent directories): .git",
		},
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "annotation message",
			expectedError: `GitTag.Revert: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitTag{
				WorkDir:    tc.workDir,
				Tag:        tc.tag,
				Annotation: tc.annotation,
			}

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitPushMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPush{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
		})
	}
}

func TestGitPushDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitPush.Dry: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPush{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitPushRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitPush.Run: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPush{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitPushRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `cannot revert git push`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPush{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitPushTagMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPushTag{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
		})
	}
}

func TestGitPushTagDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "v0.1.0",
			expectedError: `GitPushTag.Dry: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPushTag{
				WorkDir: tc.workDir,
				Tag:     tc.tag,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitPushTagRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "v0.1.0",
			expectedError: `GitPushTag.Run: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPushTag{
				WorkDir: tc.workDir,
				Tag:     tc.tag,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitPushTagRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "v0.1.0",
			expectedError: `cannot revert git push`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPushTag{
				WorkDir: tc.workDir,
				Tag:     tc.tag,
			}

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitPullMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPull{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
		})
	}
}

func TestGitPullDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitPull.Dry: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPull{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitPullRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `GitPull.Run: exit status 128 fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPull{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitPullRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `cannot revert git pull`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPull{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
