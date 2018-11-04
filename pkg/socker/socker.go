// Copyright (c) 2018 China-HPC.

// Package socker implements a secure runner for docker containers.
package socker

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/China-HPC/go-socker/pkg/su"
	log "github.com/Sirupsen/logrus"
	flags "github.com/jessevdk/go-flags"
	"github.com/kr/pty"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"
	yaml "gopkg.in/yaml.v2"
)

const (
	cmdDocker     = "docker"
	cmdCgclassify = "cgclassify"
	cmdPgrep      = "pgrep"
	sepColon      = ":"
	sepPipe       = "|"
	lineBrk       = "\n"
	envSlurmJobID = "SLURM_JOBID"

	containerRunTimeout = time.Second * 30
	epilogDir           = "/var/lib/socker/epilog"
	permEpilogDir       = 0700
	permRecordFile      = 0600

	dftImageConfigFile = "/var/lib/socker/images.yaml"
	layoutImageFormat  = `{{.ID}}|{{.Repository}}|{{.Tag}}|{{.CreatedSince}}|{{.CreatedAt}}|{{.Size}}`
)

// Socker provides a runner for docker.
type Socker struct {
	dockerUID     string
	dockerGID     string
	CurrentUID    string
	currentUser   string
	currentGID    string
	currentGroup  string
	homeDir       string
	containerUUID string
	isInsideJob   bool
	slurmJobID    string
	EpilogEnabled bool
	Insecure      bool
}

// Opts represents the socker supported docker options.
type Opts struct {
	Volumes     []string `short:"v" long:"volume"`
	TTY         bool     `short:"t" long:"tty"`
	Interactive bool     `short:"i" long:"interactive"`
	Detach      bool     `short:"d" long:"detach"`
	Runtime     string   `long:"runtime"`
	Network     string   `long:"network"`
	Name        string   `long:"name"`
	Hostname    string   `short:"h" long:"hostname"`
	User        string   `short:"u" long:"user"`
	StorageOpt  string   `long:"storage-opt"`
}

// New creates a socker instance.
func New(verbose, epilogEnabled, insecure bool) (*Socker, error) {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
	log.SetOutput(os.Stdout)
	s := &Socker{
		EpilogEnabled: epilogEnabled,
		Insecure:      insecure,
	}
	err := s.checkPrerequisite()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Image represents the socker/socker availible image format
type Image struct {
	ID            string `yaml:"id"`
	Desc          string `yaml:"desc"`
	Repository    string `yaml:"repository"`
	Tag           string `yaml:"tag"`
	CreatedScince string `yaml:"created_since"`
	CreatedAt     string `yaml:"created_at"`
	Size          string `yaml:"size"`
}

// FormatImages lists all available images from registry by map.
func (s *Socker) FormatImages(config string) (map[string]Image, error) {
	data, err := listImagesData(config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var images map[string]Image
	err = yaml.Unmarshal(data, &images)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return images, nil
}

// PrintImages prints available images for CLI.
func (s *Socker) PrintImages(config string) error {
	images, err := s.FormatImages(config)
	if err != nil {
		log.Fatal(err)
		return err
	}
	for k := range images {
		fmt.Println(k)
	}
	return nil
}

// SyncImages syncs available images for CLI.
func (s *Socker) SyncImages(configFile, repoFilter, filter string) error {
	if configFile == "" {
		configFile = dftImageConfigFile
	}
	images, err := ParseImages(repoFilter, filter)
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(images)
	if err != nil {
		log.Errorf("marshal yaml data failed: %v", err)
		return err
	}
	return ioutil.WriteFile(configFile, data, permRecordFile)
}

// ParseImages parses images from docker.
func ParseImages(repoFilter, filter string) (map[string]Image, error) {
	args := []string{"images", "--format", layoutImageFormat}
	if filter != "" {
		args = append(args, fmt.Sprintf("--filter=%s", filter))
	}
	out, err := exec.Command(cmdDocker, args...).CombinedOutput()
	if err != nil {
		log.Errorf("list images from Docker failed: %v:%s", err, out)
		return nil, err
	}
	images := make(map[string]Image)
	for _, line := range strings.Split(strings.
		TrimSpace(string(out)), lineBrk) {
		image, err := parseImage(line)
		if err != nil {
			log.Errorf("parse image failed: %v", err)
			return nil, err
		}
		if repoFilter == "" {
			images[fmt.Sprintf("%s:%s", image.Repository, image.Tag)] = *image
			continue
		}
		if strings.Contains(image.Repository, repoFilter) {
			images[fmt.Sprintf("%s:%s", image.Repository, image.Tag)] = *image
		}
	}
	return images, nil
}

func parseImage(text string) (*Image, error) {
	fields := strings.Split(text, sepPipe)
	if len(fields) != 6 {
		return nil, fmt.Errorf("parse image failed due to fields mismatch")
	}
	return &Image{
		ID:            fields[0],
		Repository:    fields[1],
		Tag:           fields[2],
		CreatedScince: fields[3],
		CreatedAt:     fields[4],
		Size:          fields[5],
	}, nil
}

func listImagesData(config string) ([]byte, error) {
	if config == "" {
		config = dftImageConfigFile
	}
	info, err := os.Stat(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var data []byte
	if !info.IsDir() {
		data, err = ioutil.ReadFile(config)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		return data, nil
	}
	files, err := ioutil.ReadDir(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		content, err := ioutil.ReadFile(path.Join(config, file.Name()))
		if err != nil {
			return nil, err
		}
		data = append(data, content...)
	}
	return data, nil
}

// RunImage runs container.
func (s *Socker) RunImage(command []string) error {
	opts := Opts{}
	_, err := flags.ParseArgs(&opts, command)
	if err != nil {
		log.Errorf("parse command args failed: %v", err)
		return err
	}
	// specified name has a higher priority, uniqueness is guaranteed by the
	// user, it will automatically generate UUID as the name if it is empty.
	if opts.Name != "" {
		s.containerUUID = opts.Name
	} else {
		s.containerUUID = uuid.NewV4().String()
	}
	args := []string{"run", "--name", s.containerUUID}
	// refuse to mount a directory that is not authorized to access
	if err := s.isVolumePermit(opts.Volumes); err != nil {
		return err
	}
	// create security swap directory and mount into container.
	if !s.Insecure {
		swapDir := path.Join(s.homeDir, "container")
		args = append(args, "-v", fmt.Sprintf("%s:%s", swapDir, swapDir))
		err = os.MkdirAll(swapDir, 0777)
		if err != nil {
			return err
		}
		err = os.Chmod(swapDir, 0777)
		if err != nil {
			return err
		}
		err = os.Chmod(s.homeDir, 0755)
		if err != nil {
			return err
		}
	}
	go s.containerMonitor()

	log.Debugf("epilog enabled: %t", s.EpilogEnabled)
	if s.EpilogEnabled {
		err := ioutil.WriteFile(path.Join(epilogDir, s.slurmJobID),
			[]byte(s.containerUUID), permEpilogDir)
		if err != nil {
			return err
		}
	}
	args = append(args, command...)
	log.Debugf("docker run args: %v", args)
	cmd, err := su.Command(s.dockerUID, cmdDocker, args...)
	if err != nil {
		return err
	}
	if opts.TTY {
		return s.runWithPty(cmd)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%s", output)
	return nil
}

func isContainerRan(containerName string) (bool, error) {
	cmd := exec.Command(cmdDocker, "events",
		"--filter", "event=start",
		"--filter", fmt.Sprintf("container=%s", containerName))
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}
	defer reader.Close()
	err = cmd.Start()
	if err != nil {
		return false, err
	}
	b := bufio.NewScanner(reader)
	isStarted := make(chan bool, 1)
	select {
	case isStarted <- b.Scan():
		log.Debugf("container started")
		return true, nil
	case <-time.After(containerRunTimeout):
		log.Errorf("container start timeout")
		return false, fmt.Errorf("container start timeout")
	}
}

func queryContainerPID(containerName string) (string, error) {
	args := []string{"inspect", "-f", "'{{ .State.Pid }}'", containerName}
	output, err := exec.Command(cmdDocker, args...).CombinedOutput()
	if err != nil {
		log.Errorf("query container pid failed: %v:%s", err, output)
		return "", err
	}
	containerPID := strings.Trim(string(output), "\r\n'")
	log.Debugf("container PID is: %s", containerPID)
	return containerPID, nil
}

func (s *Socker) containerMonitor() error {
	started, err := isContainerRan(s.containerUUID)
	if err != nil {
		log.Errorf("detect container status failed: %v", err)
		return err
	}
	if started {
		// container has ran, change user's home dir permission.
		defer changeDirPerm(s.homeDir, 0750)
	}
	if !s.isInsideJob {
		log.Debugf("not inside of job")
		return nil
	}
	err = s.enforceLimit()
	if err != nil {
		log.Errorf("enforce limit failed: %v", err)
	}
	return nil
}

func changeDirPerm(dir string, perm os.FileMode) error {
	err := os.Chmod(dir, perm)
	if err != nil {
		log.Errorf("change home dir permission error: %v", err)
	}
	return nil
}

func (s *Socker) enforceLimit() error {
	containerPID, err := queryContainerPID(s.containerUUID)
	if err != nil {
		log.Errorf("query container pid error: %v", err)
		return err
	}
	cgroupID := fmt.Sprintf("slurm/uid_%s/job_%s/", s.CurrentUID, s.slurmJobID)
	log.Debugf("target cgroup id is: %s", cgroupID)
	pids, err := QueryChildPIDs(containerPID)
	if err != nil {
		log.Errorf("query child process ids failed: %v", err)
	}
	return s.setCgroupLimit(append(pids, containerPID), cgroupID)
}

func (s *Socker) setCgroupLimit(pids []string, cgroupID string) error {
	for _, pid := range pids {
		// frees process from the docker cgroups.
		output, err := exec.Command(cmdCgclassify, "-g",
			"blkio,net_cls,devices,cpu:/", pid).CombinedOutput()
		log.Debugf("frees container cgroups limit")
		if err != nil {
			log.Errorf("frees container cgroups limit failed: %v:%s", err, output)
			return err
		}
		// add process into slurm job cgroups.
		output, err = exec.Command(cmdCgclassify, "-g",
			fmt.Sprintf("memory,cpu,freezer,devices:/%s", cgroupID),
			pid).CombinedOutput()
		log.Debugf("enforcing slurm limit to container: %s", s.containerUUID)
		if err != nil {
			log.Errorf("enforces Slurm job limit failed: %v:%s", err, output)
			return err
		}
	}
	return nil
}

// QueryChildPIDs lookups child process ids of specified parent process.
func QueryChildPIDs(parentID string) ([]string, error) {
	out, err := exec.Command(cmdPgrep, "-P", parentID).CombinedOutput()
	if err != nil {
		// if no processes were matched pgrep exit with 1
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, nil
		}
		log.Errorf("query child pids failed: %v:%s", err, out)
		return nil, err
	}
	pids := strings.Split(strings.TrimSpace(string(out)), lineBrk)
	return pids, nil
}

func (s *Socker) isVolumePermit(vols []string) error {
	for _, vol := range vols {
		if strings.Contains(vol, sepColon) {
			vol = strings.Split(vol, sepColon)[0]
		}
		if err := unix.Access(vol, unix.W_OK); err != nil {
			log.Debugf("volume %s permissin denined: %v", vol, err)
			return fmt.Errorf("volume %s permissin denined: %v", vol, err)
		}
	}
	return nil
}

func (s *Socker) runWithPty(cmd *exec.Cmd) error {
	tty, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("docker command exec failed: %v", err)
	}
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }()
	go func() { io.Copy(os.Stdout, tty) }()
	go func() { io.Copy(tty, os.Stdin) }()
	return cmd.Wait()
}

func (s *Socker) checkPrerequisite() error {
	if !isCommandAvailable(cmdDocker) {
		return cli.NewExitError("docker command not found, make sure Docker is installed...", 127)
	}
	u, err := user.Lookup("dockerroot")
	if err != nil {
		return cli.NewExitError("there must exist a user 'dockerroot' and a group 'docker'", 1)
	}
	s.dockerUID = u.Uid
	g, err := user.LookupGroup("docker")
	if err != nil {
		return cli.NewExitError("there must exist a user 'dockerroot' and a group 'docker'", 1)
	}
	s.dockerGID = g.Gid
	gids, err := u.GroupIds()
	if err != nil && isMemberOfGroup(gids, u.Gid) {
		return cli.NewExitError("the user 'dockerroot' must be a member of the 'docker' group", 2)
	}
	current, err := user.Current()
	if err != nil {
		return cli.NewExitError("can't get current user info", 2)
	}
	s.CurrentUID = current.Uid
	s.currentUser = current.Username
	s.currentGID = current.Gid
	currentGroup, err := user.LookupGroupId(s.currentGID)
	if err != nil {
		return cli.NewExitError("can't get current user's group info", 2)
	}
	s.currentGroup = currentGroup.Name
	s.homeDir = current.HomeDir
	if jobID := os.Getenv(envSlurmJobID); jobID != "" {
		log.Debugf("slurm job id: %s", jobID)
		s.isInsideJob = true
		s.slurmJobID = jobID
	}
	return os.MkdirAll(epilogDir, permRecordFile)
}

func isMemberOfGroup(gids []string, gid string) bool {
	for _, id := range gids {
		if id == gid {
			return true
		}
	}
	return false
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	if err != nil {
		return false
	}
	return true
}
