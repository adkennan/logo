package logo

import (
	"bufio"
	"os"
	"path"
	"strings"
)

type File interface {
	Id() int
	Name() string
	ReadLine() (string, error)
	Write(text string) error
	Close() error
	IsInteractive() bool
}

type Files struct {
	rootPath    string
	nextId      int
	openFiles   map[string]File
	defaultFile File
	reader      File
	writer      File
	dribble     File
}

func CreateFiles(rootPath string) *Files {
	df := NewStdIOFile()

	f := &Files{
		rootPath,
		1,
		make(map[string]File),
		df, df, df, nil}

	return f
}

func (this *Files) OpenFile(name string) error {
	uname := strings.ToUpper(name)
	_, exists := this.openFiles[uname]
	if exists {
		return nil
	}

	p := path.Join(this.rootPath, uname)
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	this.openFiles[uname] = &NormalFile{
		this.nextId, uname, f, bufio.NewReader(f)}

	this.nextId++

	return nil
}

func (this *Files) CloseFile(name string) error {
	if name == "" {
		return nil
	}

	uname := strings.ToUpper(name)
	f, ok := this.openFiles[uname]
	if !ok {
		return errorFileNotOpen(name)
	}

	delete(this.openFiles, uname)

	if this.reader == f {
		this.reader = this.defaultFile
	}
	if this.writer == f {
		this.writer = this.defaultFile
	}

	f.Close()
	return nil
}

func (this *Files) SetReader(name string) error {
	if name == "" {
		this.reader = this.defaultFile
	}

	uname := strings.ToUpper(name)
	f, ok := this.openFiles[uname]
	if !ok {
		return errorFileNotOpen(name)
	}

	this.reader = f
	return nil
}

func (this *Files) SetWriter(name string) error {

	if name == "" {
		this.writer = this.defaultFile
	}

	uname := strings.ToUpper(name)
	f, ok := this.openFiles[uname]
	if !ok {
		return errorFileNotOpen(name)
	}

	this.writer = f
	return nil
}

type NormalFile struct {
	id   int
	name string
	f    *os.File
	r    *bufio.Reader
}

func (this *NormalFile) Id() int {
	return this.id
}

func (this *NormalFile) Name() string {
	return ""
}

func (this *NormalFile) ReadLine() (string, error) {
	line, _, err := this.r.ReadLine()
	if err != nil {
		return "", err
	}
	return string(line), nil
}

func (this *NormalFile) Write(text string) error {

	_, err := this.f.WriteString(text)
	return err
}

func (this *NormalFile) Close() error {
	return this.f.Close()
}

func (this *NormalFile) IsInteractive() bool {
	return false
}

type StdIOFile struct {
	id     int
	stdin  *os.File
	stdout *os.File
}

func NewStdIOFile() *StdIOFile {
	return &StdIOFile{0, os.Stdin, os.Stdout}
}

func (this *StdIOFile) Id() int {
	return this.id
}

func (this *StdIOFile) Name() string {
	return ""
}

func (this *StdIOFile) ReadLine() (string, error) {
	r := bufio.NewReader(this.stdin)
	line, _, err := r.ReadLine()
	if err != nil {
		return "", err
	}
	return string(line), nil
}

func (this *StdIOFile) Write(text string) error {

	_, err := this.stdout.WriteString(text)
	return err
}

func (this *StdIOFile) Close() error {
	return nil
}

func (this *StdIOFile) IsInteractive() bool {
	return true
}
