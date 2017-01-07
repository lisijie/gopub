package git

import (
    "testing"
    "os"
    "fmt"
)

func getPath() string {
    path, _ := os.Getwd()
    path = path + "/tmp"
    return path
}

func TestClone(t *testing.T) {
    os.RemoveAll(getPath())
    repo := &GitRepository{}
    repo.Init(fmt.Sprintf(`{"path":"%s"}`, getPath()))
    err := repo.CloneFrom("https://github.com/lisijie/cron.git")
    if err != nil {
        t.Error(err)
    }
}

func TestUpdate(t *testing.T) {
    repo := &GitRepository{Path:getPath()}
    if err := repo.Update(); err != nil {
        t.Error(err)
    }
}

func TestGetTags(t *testing.T) {
    repo := &GitRepository{Path:getPath()}
    _, err := repo.GetTags()
    if err != nil {
        t.Error(err)
    }
}
