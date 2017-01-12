package repository

import (
    "strings"
    "fmt"
    "net/url"
    "github.com/lisijie/gopub/app/libs/command"
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
    cmdStr := fmt.Sprintf("git clone --mirror -q %s %s", repoUrl, r.Path)
    cmd := command.NewCommand(cmdStr)
    err := cmd.Run()
    return err
}

func (r *GitRepository) Update() error {
    cmd := command.NewCommand("git remote update")
    err := cmd.RunInDir(r.Path)
    return err
}

func (r *GitRepository) GetTags() ([]string, error) {
    cmd := command.NewCommand("git tag --sort=-v:refname")
    err := cmd.RunInDir(r.Path)
    if err != nil {
        return nil, err
    }
    tags := make([]string, 0)
    for _, v := range strings.Split(string(cmd.Stdout()), "\n") {
        v = strings.TrimSpace(v)
        if v != "" {
            tags = append(tags, v)
        }
    }
    return tags, nil
}

func (r *GitRepository) GetBranches() ([]string, error) {
    cmd := command.NewCommand("git show-ref --heads")
    if err := cmd.RunInDir(r.Path); err != nil {
        return nil, err
    }
    out := string(cmd.Stdout())
    lines := strings.Split(out, "\n")
    branches := make([]string, 0, len(lines))
    prefixLen := len("refs/heads/")
    for _, v := range lines {
        fields := strings.Fields(strings.TrimSpace(v))
        if len(fields) == 2 {
            branches = append(branches, fields[1][prefixLen:])
        }
    }
    return branches, nil
}

func (r *GitRepository) Export(branch, filename string) error {
    cmdStr := "git archive --format=tar " + branch + " | gzip > " + filename
    cmd := command.NewCommand(cmdStr)
    return cmd.RunInDir(r.Path)
}

func (r *GitRepository) ExportDiffFiles(fromVer, toVer, filename string) error {
    cmdStr := "git archive --format=tar " + toVer + " $(git diff --name-status -b " + fromVer + "..." + toVer + " | grep -v ^D | awk '{print $2}') | gzip > " + filename
    cmd := command.NewCommand(cmdStr)
    return cmd.RunInDir(r.Path)
}

func (r *GitRepository) GetChangeLogs(fromVer, toVer string) ([]string, error) {
    return nil, nil
}

func (r *GitRepository) GetChangeFiles(fromVer, toVer string) ([]ChangeFile, error) {
    cmdStr := "git diff --name-status -b " + fromVer + "..." + toVer
    cmd := command.NewCommand(cmdStr)
    if err := cmd.RunInDir(r.Path); err != nil {
        return nil, err
    }
    lines := strings.Split(string(cmd.Stdout()), "\n")
    result := make([]ChangeFile, 0, len(lines))
    for _, v := range lines {
        v = strings.TrimSpace(v)
        if v == "" {
            continue
        }
        fields := strings.Fields(v)
        result = append(result, ChangeFile{Flag:fields[0], Filename:fields[1]})
    }
    return result, nil
}