package version

var (
    // Version is the semantic version
    Version = "dev"
    // GitCommit is the git commit hash
    GitCommit = "unknown"
    // BuildDate is the build timestamp
    BuildDate = "unknown"
)

type Info struct {
    Version   string `json:"version"`
    GitCommit string `json:"gitCommit"`
    BuildDate string `json:"buildDate"`
}

func GetInfo() Info {
    return Info{
        Version:   Version,
        GitCommit: GitCommit,
        BuildDate: BuildDate,
    }
}
