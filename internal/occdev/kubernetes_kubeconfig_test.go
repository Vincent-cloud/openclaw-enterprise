package occdev

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteKubeconfigsNormalizesDockerDesktopHost(t *testing.T) {
	state := &developmentState{
		Cluster:   "occ-dev-kubeconfig",
		directory: t.TempDir(),
	}
	host := filepath.Join(state.directory, "kubeconfig")
	hostDocker := "https" + "://host.docker.internal:6443"
	hostLoopback := "https" + "://127.0.0.1:6443"

	commands := fakeProfileCommands(t, map[string]string{
		"k3d":     `"kubeconfig get occ-dev-kubeconfig") printf '%s\n' "contexts:" "- name: k3d-occ-dev-kubeconfig" "  context:" "    cluster: owned" "clusters:" "- name: owned" "  cluster:" "    server: ` + hostDocker + `" ;;`,
		"kubectl": `"--kubeconfig "*" --context k3d-occ-dev-kubeconfig get --raw=/version") grep -Fq '` + hostLoopback + `' "$2" || exit 91 ;;`,
	})
	r := newRunner(Options{})

	if err := r.writeKubeconfigs(context.Background(), state); err != nil {
		t.Fatalf("write kubeconfigs: %v\ncommands:\n%s", err, strings.Join(commands(), "\n"))
	}

	assertKubeconfigFileMode(t, host, 0o600)
	hostData, err := os.ReadFile(host)
	if err != nil {
		t.Fatal(err)
	}
	hostText := string(hostData)
	if !strings.Contains(hostText, hostLoopback) ||
		strings.Contains(hostText, "host.docker.internal") {
		t.Fatalf("unexpected host kubeconfig:\n%s", hostText)
	}

	container := filepath.Join(state.directory, "container-kubeconfig")
	assertKubeconfigFileMode(t, container, 0o644)
	containerData, err := os.ReadFile(container)
	if err != nil {
		t.Fatal(err)
	}
	containerText := string(containerData)
	containerServer := "https" + "://k3d-occ-dev-kubeconfig-serverlb:6443"
	if !strings.Contains(containerText, containerServer) ||
		!strings.Contains(containerText, "tls-server-name: k3d-occ-dev-kubeconfig-serverlb") {
		t.Fatalf("unexpected container kubeconfig:\n%s", containerText)
	}
}

func assertKubeconfigFileMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("%s mode: got %o, want %o", path, got, want)
	}
}
