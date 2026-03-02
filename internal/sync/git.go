package sync

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitClient wraps Git command-line operations.
type GitClient struct {
	repoPath string
}

// NewGitClient creates a new GitClient for the specified repository path.
func NewGitClient(repoPath string) *GitClient {
	return &GitClient{repoPath: repoPath}
}

// runGit executes a git command and returns its output.
func (g *GitClient) runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath
	// Inherit environment variables (important for SSH_AUTH_SOCK)
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}

	return stdout.String(), nil
}

// runGitWithInput executes a git command with stdin input.
func (g *GitClient) runGitWithInput(input string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath
	cmd.Stdin = strings.NewReader(input)
	// Inherit environment variables (important for SSH_AUTH_SOCK)
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}

	return stdout.String(), nil
}

// IsGitInstalled checks if git is available on the system.
func (g *GitClient) IsGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// IsRepo checks if the path is a Git repository.
func (g *GitClient) IsRepo() bool {
	gitDir := filepath.Join(g.repoPath, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// Init initializes a new Git repository.
func (g *GitClient) Init() error {
	// Ensure directory exists
	if err := os.MkdirAll(g.repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create repo directory: %w", err)
	}

	_, err := g.runGit("init")
	if err != nil {
		return err
	}

	// Configure default user for commits
	if _, err := g.runGit("config", "user.email", "sync@ltools.local"); err != nil {
		return err
	}
	if _, err := g.runGit("config", "user.name", "LTools Sync"); err != nil {
		return err
	}

	// Set default branch name to main
	if _, err := g.runGit("branch", "-M", "main"); err != nil {
		// Non-fatal, might already be on main
		fmt.Printf("[GitClient] Warning: could not rename branch to main: %v\n", err)
	}

	return nil
}

// Clone clones a repository to the client's path.
func (g *GitClient) Clone(url string) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(g.repoPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Clone to the repo path
	cmd := exec.Command("git", "clone", url, g.repoPath)
	// Inherit environment variables (important for SSH_AUTH_SOCK)
	cmd.Env = os.Environ()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\nstderr: %s", err, stderr.String())
	}

	// Configure user
	if _, err := g.runGit("config", "user.email", "sync@ltools.local"); err != nil {
		return err
	}
	if _, err := g.runGit("config", "user.name", "LTools Sync"); err != nil {
		return err
	}

	return nil
}

// SetRemote sets the remote URL for the repository.
func (g *GitClient) SetRemote(url string) error {
	// Check if remote exists
	_, err := g.runGit("remote", "get-url", "origin")
	if err != nil {
		// Remote doesn't exist, add it
		_, err = g.runGit("remote", "add", "origin", url)
		return err
	}

	// Remote exists, update it
	_, err = g.runGit("remote", "set-url", "origin", url)
	return err
}

// GetRemoteURL returns the current remote URL.
func (g *GitClient) GetRemoteURL() (string, error) {
	output, err := g.runGit("remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// AddAll stages all changes.
func (g *GitClient) AddAll() error {
	_, err := g.runGit("add", "-A")
	return err
}

// Commit creates a commit with the given message.
func (g *GitClient) Commit(message string) error {
	_, err := g.runGit("commit", "-m", message)
	return err
}

// Push pushes to the remote repository.
// If force is true, uses --force-with-lease for safer force push.
func (g *GitClient) Push(force bool) error {
	// First, try normal push
	args := []string{"push", "-u", "origin", "main"}
	_, err := g.runGit(args...)
	if err == nil {
		return nil
	}

	// If normal push fails and force is requested, try force-with-lease
	if force {
		fmt.Printf("[GitClient] Normal push failed, trying force-with-lease: %v\n", err)
		args = []string{"push", "--force-with-lease", "-u", "origin", "main"}
		_, err = g.runGit(args...)
		return err
	}

	return err
}

// Pull pulls from the remote repository.
// Uses rebase with 'ours' strategy to auto-resolve conflicts (local-first).
func (g *GitClient) Pull() error {
	// Try to pull from main branch first with auto-conflict resolution
	_, err := g.runGit("pull", "origin", "main", "--rebase", "-X", "ours")
	if err == nil {
		return nil
	}

	// If main fails, try master (for older repositories)
	fmt.Printf("[GitClient] Pull from main failed, trying master: %v\n", err)
	_, err = g.runGit("pull", "origin", "master", "--rebase", "-X", "ours")

	// If rebase still fails, try to abort and fall back to merge
	if err != nil {
		fmt.Printf("[GitClient] Rebase pull failed, aborting and trying merge: %v\n", err)
		g.runGit("rebase", "--abort")
		_, err = g.runGit("pull", "origin", "main", "--no-rebase")
		if err != nil {
			_, err = g.runGit("pull", "origin", "master", "--no-rebase")
		}
	}

	return err
}

// HasChanges checks if there are any changes to commit.
func (g *GitClient) HasChanges() (bool, error) {
	// Check for staged changes
	output, err := g.runGit("diff", "--cached", "--name-only")
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(output) != "" {
		return true, nil
	}

	// Check for unstaged changes
	output, err = g.runGit("diff", "--name-only")
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(output) != "" {
		return true, nil
	}

	// Check for untracked files
	output, err = g.runGit("ls-files", "--others", "--exclude-standard")
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(output) != "", nil
}

// GetCommitHash returns the current commit hash.
func (g *GitClient) GetCommitHash() (string, error) {
	output, err := g.runGit("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// GetShortHash returns the short commit hash.
func (g *GitClient) GetShortHash() (string, error) {
	output, err := g.runGit("rev-parse", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// GetStatus returns a summary of the repository status.
func (g *GitClient) GetStatus() (string, error) {
	output, err := g.runGit("status", "--short")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// GetFileCount returns the number of tracked files.
func (g *GitClient) GetFileCount() (int, error) {
	output, err := g.runGit("ls-files")
	if err != nil {
		return 0, err
	}
	files := strings.Split(strings.TrimSpace(output), "\n")
	count := 0
	for _, f := range files {
		if f != "" {
			count++
		}
	}
	return count, nil
}

// TestConnection tests if we can connect to the remote repository.
func (g *GitClient) TestConnection(url string) (*ConnectionTestResult, error) {
	result := &ConnectionTestResult{
		AuthMethod: AuthMethodAuto,
	}

	// Determine URL type (SSH or HTTPS)
	if strings.HasPrefix(url, "git@") {
		result.AuthMethod = AuthMethodSSH
	} else if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		result.AuthMethod = AuthMethodHTTPS
	}

	// Test connection using git ls-remote
	cmd := exec.Command("git", "ls-remote", "--exit-code", url)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Inherit environment variables (important for SSH_AUTH_SOCK)
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("连接失败: %s", strings.TrimSpace(stderr.String()))
		return result, fmt.Errorf("connection test failed: %w", err)
	}

	result.Success = true
	result.Message = "连接成功"
	return result, nil
}

// WriteGitignore writes a .gitignore file to the repository.
func (g *GitClient) WriteGitignore(content string) error {
	gitignorePath := filepath.Join(g.repoPath, ".gitignore")
	return os.WriteFile(gitignorePath, []byte(content), 0644)
}

// GetLastCommitTime returns the time of the last commit.
func (g *GitClient) GetLastCommitTime() (time.Time, error) {
	output, err := g.runGit("log", "-1", "--format=%ct")
	if err != nil {
		return time.Time{}, err
	}

	var timestamp int64
	_, err = fmt.Sscanf(strings.TrimSpace(output), "%d", &timestamp)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp, 0), nil
}

// Fetch fetches from the remote without merging.
func (g *GitClient) Fetch() error {
	_, err := g.runGit("fetch", "origin")
	return err
}

// GetBehindCount returns the number of commits behind the remote.
func (g *GitClient) GetBehindCount() (int, error) {
	// First fetch to get latest remote state
	if err := g.Fetch(); err != nil {
		return 0, err
	}

	output, err := g.runGit("rev-list", "--count", "HEAD..origin/main")
	if err != nil {
		return 0, err
	}

	var count int
	_, err = fmt.Sscanf(strings.TrimSpace(output), "%d", &count)
	return count, err
}

// GetAheadCount returns the number of commits ahead of the remote.
func (g *GitClient) GetAheadCount() (int, error) {
	output, err := g.runGit("rev-list", "--count", "origin/main..HEAD")
	if err != nil {
		return 0, err
	}

	var count int
	_, err = fmt.Sscanf(strings.TrimSpace(output), "%d", &count)
	return count, err
}
