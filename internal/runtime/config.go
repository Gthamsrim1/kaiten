package runtime

type Config struct {
	Repo     string
	Snapshot string

	Command []string

	Hostname string

	WorkDir string

	Env []string

	ReadOnly bool
}

const ChildCommand = "__child__"
