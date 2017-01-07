package git

import (
    "strings"
    "fmt"
    "net/url"
    "github.com/lisijie/gopub/app/libs/command"
    "github.com/lisijie/gopub/app/libs/repository"
)

type GitRepository struct {
    Username  string `json:"username"`
    Password  string `json:"password"`
    Path      string `json:"path"`
    RemoteUrl string `json:"url"`
}

func (r *GitRepository) Clone() error {
    repoUrl := r.RemoteUrl
    if r.Username != "" && r.Password != "" {
        auth := url.QueryEscape(r.Username) + ":" + url.QueryEscape(r.Password)
        if strings.HasPrefix(repoUrl, "https://") {
            repoUrl = "https://" + auth + "@" + repoUrl[8:]
        } else if strings.HasPrefix(repoUrl, "http://") {
            repoUrl = "http://" + auth + "@" + repoUrl[7:]
        }
    }
    cmdStr := fmt.Sprintf("git clone -q %s %s", repoUrl, r.Path)
    cmd := command.NewCommand(cmdStr)
    err := cmd.Run()
    return err
}

func (r *GitRepository) Update() error {
    cmd := command.NewCommand("git pull")
    err := cmd.RunInDir(r.Path)
    return err
}

func (r *GitRepository) GetTags() ([]string, error) {
    cmd := command.NewCommand("git tag --sort=-v:refname")
    err := cmd.RunInDir(r.Path)
    if err != nil {
        return nil, err
    }
    return strings.Split(string(cmd.Stdout()), "\n"), nil
}

func (r *GitRepository) GetBranches() ([]string, error) {
    cmd := command.NewCommand("git branch -r --no-color")
    if err := cmd.RunInDir(r.Path); err != nil {
        return nil, err
    }
    out := string(cmd.Stdout())
    lines := strings.Split(out, "\n")
    branches := make([]string, 0, len(lines))
    for _, v := range lines {
        if v == "" || strings.Contains(v, " -> ") {
            continue
        }
        branches = append(branches, strings.SplitN(v, "/", 2)[1])
    }
    return branches, nil
}

func (r *GitRepository) Export(branch string, filename string) error {
    cmdStr := "git archive --format=tar " + branch + " | gzip > " + filename
    cmd := command.NewCommand(cmdStr)
    return cmd.RunInDir(r.Path)
}

func (r *GitRepository) ExportDiffFiles(ver1 string, ver2 string, filename string) error {
    cmdStr := "git archive --format=tar " + ver1 + " $(git diff --name-status -b " + ver1 + " " + ver2 + " | grep -v ^D | awk '{print $2}') | gzip > " + filename
    cmd := command.NewCommand(cmdStr)
    return cmd.RunInDir(r.Path)
}

func (r *GitRepository) GetChangeList() (*repository.ChangeList, error) {
    return nil, nil
}