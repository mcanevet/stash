package cli

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	api "github.com/appscode/stash/apis/stash/v1alpha1"
	shell "github.com/codeskyblue/go-sh"
)

const (
	Exe = "/bin/restic"
)

type BackupWrapper struct {
	sh          *shell.Session
	scratchDir  string
	enableCache bool
	hostname    string
	cacertFile  string
}

func New(scratchDir string, enableCache bool, hostname string) *BackupWrapper {
	ctrl := &BackupWrapper{
		sh:          shell.NewSession(),
		scratchDir:  scratchDir,
		enableCache: enableCache,
		hostname:    hostname,
	}
	ctrl.sh.SetDir(scratchDir)
	ctrl.sh.ShowCMD = true
	return ctrl
}

type Snapshot struct {
	ID       string    `json:"id"`
	Time     time.Time `json:"time"`
	Tree     string    `json:"tree"`
	Paths    []string  `json:"paths"`
	Hostname string    `json:"hostname"`
	Username string    `json:"username"`
	UID      int       `json:"uid"`
	Gid      int       `json:"gid"`
}

func (w *BackupWrapper) ListSnapshots() ([]Snapshot, error) {
	result := make([]Snapshot, 0)
	args := w.appendCacheDirFlag([]interface{}{"snapshots", "--json"})
	args = w.appendCaCertFlag(args)

	err := w.sh.Command(Exe, args...).UnmarshalJSON(&result)
	return result, err
}

func (w *BackupWrapper) InitRepositoryIfAbsent() error {
	args := w.appendCacheDirFlag([]interface{}{"snapshots", "--json"})
	args = w.appendCaCertFlag(args)
	if err := w.run(Exe, args); err != nil {
		args = w.appendCacheDirFlag([]interface{}{"init"})
		args = w.appendCaCertFlag(args)

		return w.run(Exe, args)
	}
	return nil
}

func (w *BackupWrapper) Backup(resource *api.Backup, fg api.FileGroup) error {
	args := []interface{}{"backup", fg.Path, "--force"}
	if w.hostname != "" {
		args = append(args, "--hostname")
		args = append(args, w.hostname)
	}
	// add tags if any
	for _, tag := range fg.Tags {
		args = append(args, "--tag")
		args = append(args, tag)
	}
	args = w.appendCacheDirFlag(args)
	args = w.appendCaCertFlag(args)

	return w.run(Exe, args)
}

func (w *BackupWrapper) Forget(resource *api.Backup, fg api.FileGroup) error {
	// Get retentionPolicy for fileGroup, ignore if not found
	retentionPolicy := api.RetentionPolicy{}
	for _, policy := range resource.Spec.RetentionPolicies {
		if policy.Name == fg.RetentionPolicyName {
			retentionPolicy = policy
			break
		}
	}

	args := []interface{}{"forget"}
	if retentionPolicy.KeepLast > 0 {
		args = append(args, string(api.KeepLast))
		args = append(args, strconv.Itoa(retentionPolicy.KeepLast))
	}
	if retentionPolicy.KeepHourly > 0 {
		args = append(args, string(api.KeepHourly))
		args = append(args, strconv.Itoa(retentionPolicy.KeepHourly))
	}
	if retentionPolicy.KeepDaily > 0 {
		args = append(args, string(api.KeepDaily))
		args = append(args, strconv.Itoa(retentionPolicy.KeepDaily))
	}
	if retentionPolicy.KeepWeekly > 0 {
		args = append(args, string(api.KeepWeekly))
		args = append(args, strconv.Itoa(retentionPolicy.KeepWeekly))
	}
	if retentionPolicy.KeepMonthly > 0 {
		args = append(args, string(api.KeepMonthly))
		args = append(args, strconv.Itoa(retentionPolicy.KeepMonthly))
	}
	if retentionPolicy.KeepYearly > 0 {
		args = append(args, string(api.KeepYearly))
		args = append(args, strconv.Itoa(retentionPolicy.KeepYearly))
	}
	for _, tag := range retentionPolicy.KeepTags {
		args = append(args, string(api.KeepTag))
		args = append(args, tag)
	}
	if retentionPolicy.Prune {
		args = append(args, "--prune")
	}
	if retentionPolicy.DryRun {
		args = append(args, "--dry-run")
	}
	if len(args) > 1 {
		args = w.appendCacheDirFlag(args)
		args = w.appendCaCertFlag(args)

		return w.run(Exe, args)
	}
	return nil
}

func (w *BackupWrapper) Restore(path, host string) error {
	args := []interface{}{"restore"}
	args = append(args, "latest") // TODO @ Dipta: Add support for specific snapshotID
	args = append(args, "--path")
	args = append(args, path) // source-path specified in restic fileGroup
	args = append(args, "--host")
	args = append(args, host)
	args = append(args, "--target")
	args = append(args, path) // restore in same path as source-path
	args = w.appendCacheDirFlag(args)
	args = w.appendCaCertFlag(args)

	return w.run(Exe, args)
}

func (w *BackupWrapper) Check() error {
	args := w.appendCacheDirFlag([]interface{}{"check"})
	args = w.appendCaCertFlag(args)

	return w.run(Exe, args)
}

func (w *BackupWrapper) appendCacheDirFlag(args []interface{}) []interface{} {
	if w.enableCache {
		cacheDir := filepath.Join(w.scratchDir, "restic-cache")
		return append(args, "--cache-dir", cacheDir)
	}
	return append(args, "--no-cache")
}

func (w *BackupWrapper) appendCaCertFlag(args []interface{}) []interface{} {
	if w.cacertFile != "" {
		return append(args, "--cacert", w.cacertFile)
	}
	return args
}

func (w *BackupWrapper) run(cmd string, args []interface{}) error {
	out, err := w.sh.Command(cmd, args...).CombinedOutput()
	if err != nil {
		parts := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
		if len(parts) > 1 {
			parts = parts[len(parts)-1:]
			return errors.New(parts[0])
		}
	}
	return err
}
