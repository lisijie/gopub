package service

import (
	"errors"
	"github.com/lisijie/gopub/app/libs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type repositoryService struct{}

// 返回一个仓库对象
func (this *repositoryService) GetRepoByProjectId(projectId int) (*Repository, error) {
	project, err := ProjectService.GetProject(projectId)
	if err != nil {
		return nil, err
	}
	return OpenRepository(project.Domain)
}

// 获取某项目代码库的标签列表
func (this *repositoryService) GetTags(projectId int, limit int) ([]string, error) {
	repo, err := this.GetRepoByProjectId(projectId)
	if err != nil {
		return nil, err
	}
	repo.Pull()
	list, err := repo.GetTags()
	if err != nil {
		return nil, err
	}
	if len(list) > limit {
		list = list[0:limit]
	}
	return list, nil
}

// 克隆git仓库
func (this *repositoryService) CloneRepo(url string, dst string) error {
	out, stderr, err := libs.ExecCmd("git", "clone", url, dst)
	debug("out", out)
	debug("stderr", stderr)
	debug("err", err)
	if err != nil {
		return concatenateError(err, stderr)
	}
	return nil
}

type SortTag struct {
	data []string
}

func (t *SortTag) Len() int {
	return len(t.data)
}
func (t *SortTag) Swap(i, j int) {
	t.data[i], t.data[j] = t.data[j], t.data[i]
}
func (t *SortTag) Less(i, j int) bool {
	return libs.VerCompare(t.data[i], t.data[j]) == 1
}
func (t *SortTag) Sort() []string {
	sort.Sort(t)
	return t.data
}

type Repository struct {
	Path string
}

func OpenRepository(repoPath string) (*Repository, error) {
	repoPath = GetProjectPath(repoPath)
	repoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, err
	} else if !libs.IsDir(repoPath) {
		return nil, errors.New("no such file or directory")
	}

	return &Repository{Path: repoPath}, nil
}

// 拉取代码
func (repo *Repository) Pull() error {
	_, stderr, err := libs.ExecCmdDir(repo.Path, "git", "pull")
	if err != nil {
		return concatenateError(err, stderr)
	}
	return nil
}

// 获取tag列表
func (repo *Repository) GetTags() ([]string, error) {
	stdout, stderr, err := libs.ExecCmdDir(repo.Path, "git", "tag", "-l")
	if err != nil {
		return nil, concatenateError(err, stderr)
	}
	tags := strings.Split(stdout, "\n")
	tags = tags[:len(tags)-1]

	so := &SortTag{data: tags}
	return so.Sort(), nil
}

// 获取两个版本之间的修改日志
func (repo *Repository) GetChangeLogs(startVer, endVer string) ([]string, error) {
	// git log --pretty=format:"%cd %cn: %s" --date=iso v1.8.0...v1.9.0
	stdout, stderr, err := libs.ExecCmdDir(repo.Path, "git", "log", "--pretty=format:%cd %cn: %s", "--date=iso", startVer+"..."+endVer)
	if err != nil {
		return nil, concatenateError(err, stderr)
	}

	logs := strings.Split(stdout, "\n")
	return logs, nil
}

// 获取两个版本之间的差异文件列表
func (repo *Repository) GetChangeFiles(startVer, endVer string, onlyFile bool) ([]string, error) {
	// git diff --name-status -b v1.8.0 v1.9.0
	param := "--name-status"
	if onlyFile {
		param = "--name-only"
	}
	stdout, stderr, err := libs.ExecCmdDir(repo.Path, "git", "diff", param, "-b", startVer, endVer)
	if err != nil {
		return nil, concatenateError(err, stderr)
	}
	lines := strings.Split(stdout, "\n")
	return lines[:len(lines)-1], nil
}

// 获取两个版本间的新增或修改的文件数量
func (repo *Repository) GetDiffFileCount(startVer, endVer string) (int, error) {
	cmd := "git diff --name-status -b " + startVer + " " + endVer + " |grep -v ^D |wc -l"
	stdout, stderr, err := libs.ExecCmdDir(repo.Path, "/bin/bash", "-c", cmd)
	if err != nil {
		return 0, concatenateError(err, stderr)
	}
	count, _ := strconv.Atoi(strings.TrimSpace(stdout))
	return count, nil
}

// 导出版本到tar包
func (repo *Repository) Export(startVer, endVer string, filename string) error {
	// git archive --format=tar.gz $endVer $(git diff --name-status -b $beginVer $endVer |grep -v ^D |grep -v Upgrade/ |awk '{print $2}') -o $tmpFile

	cmd := ""
	if startVer == "" {
		cmd = "git archive --format=tar " + endVer + " | gzip > " + filename
	} else {
		cmd = "git archive --format=tar " + endVer + " $(git diff --name-status -b " + startVer + " " + endVer + "|grep -v ^D |awk '{print $2}') | gzip > " + filename
	}

	_, stderr, err := libs.ExecCmdDir(repo.Path, "/bin/bash", "-c", cmd)

	if err != nil {
		return concatenateError(err, stderr)
	}
	return nil
}
