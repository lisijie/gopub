package libs

import (
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type ServerConn struct {
	addr       string // 192.168.1.1:22
	user       string
	key        string
	conn       *ssh.Client
	sftpClient *sftp.Client
}

func NewServerConn(addr, user, key string) *ServerConn {
	key = RealPath(key)
	return &ServerConn{
		addr: addr,
		user: user,
		key:  key,
	}
}

// 连接ssh服务器
func (s *ServerConn) getSshConnect() (*ssh.Client, error) {
	if s.conn != nil {
		return s.conn, nil
	}
	config := ssh.ClientConfig{
		User: s.user,
	}

	keys := []ssh.Signer{}
	if pk, err := readPrivateKey(s.key); err == nil {
		keys = append(keys, pk)
	}
	config.Auth = append(config.Auth, ssh.PublicKeys(keys...))

	conn, err := ssh.Dial("tcp", s.addr, &config)
	if err != nil {
		return nil, fmt.Errorf("无法连接到服务器 [%s]: %v", s.addr, err)
	}
	s.conn = conn
	return s.conn, nil
}

// 返回sftp连接
func (s *ServerConn) getSftpConnect() (*sftp.Client, error) {
	if s.sftpClient != nil {
		return s.sftpClient, nil
	}

	conn, err := s.getSshConnect()
	if err != nil {
		return nil, err
	}

	s.sftpClient, err = sftp.NewClient(conn, sftp.MaxPacket(1<<15))
	return s.sftpClient, err
}

// 关闭连接
func (s *ServerConn) Close() {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	if s.sftpClient != nil {
		s.sftpClient.Close()
		s.sftpClient = nil
	}
}

// 尝试连接服务器
func (s *ServerConn) TryConnect() error {
	_, err := s.getSshConnect()
	if err != nil {
		return err
	}
	s.Close()
	return nil
}

// 在远程服务器执行命令
func (s *ServerConn) RunCmd(cmd string) (string, error) {

	conn, err := s.getSshConnect()
	if err != nil {
		return "", err
	}

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建会话失败: %v", err)
	}
	defer session.Close()

	var buf bytes.Buffer

	session.Stdout = &buf
	session.Stdin = &buf

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("执行命令失败: %v", err)
	}

	return buf.String(), nil

}

// 拷贝本机文件到远程服务器
func (s *ServerConn) CopyFile(srcFile, dstFile string) error {
	client, err := s.getSftpConnect()
	if err != nil {
		return err
	}

	toPath := path.Dir(dstFile)
	if _, err := s.RunCmd("mkdir -p " + toPath); err != nil {
		return fmt.Errorf("创建目录失败：%v", err)
	}

	f, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %v", err)
	}
	defer f.Close()

	w, err := client.Create(dstFile)
	if err != nil {
		return fmt.Errorf("创建文件失败 [%s]: %v", dstFile, err)
	}
	defer w.Close()

	n, err := io.Copy(w, f)
	if err != nil {
		return fmt.Errorf("拷贝文件失败: %v", err)
	}

	fstat, _ := f.Stat()
	if fstat.Size() != n {
		return fmt.Errorf("写入文件大小错误，源文件大小：%d, 写入大小：%d", fstat.Size(), n)
	}

	return nil
}

func readPrivateKey(path string) (ssh.Signer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(b)
}
