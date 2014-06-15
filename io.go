package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
)

type File interface {
	Id() int
	Name() string
	ReadLine() (string, error)
	ReadChar() (rune, error)
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

	d, err := os.Stat(rootPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(rootPath, 0666)
		}
		if err != nil {
			panic(err)
		}
	}
	if !d.IsDir() {
		panic(errorNotDir(rootPath))
	}

	f := &Files{
		rootPath,
		1,
		make(map[string]File),
		df, df, df, nil}

	return f
}

func (this *Files) writeLine(s string) {
	this.writer.Write(s)
	this.writer.Write("\n")
}

func (this *Files) write(s string) {
	this.writer.Write(s)
}

func (this *Files) normPath(p string) string {
	p = path.Clean(p)
	if !path.IsAbs(p) {
		p = path.Join(this.rootPath, p)
	}

	return p
}

func (this *Files) SetPrefix(prefix string) error {

	p := this.normPath(prefix)

	f, err := os.Stat(p)
	if err != nil {
		return err
	}

	if !f.IsDir() {
		return errorNotDir(p)
	}

	this.rootPath = p
	return nil
}

func (this *Files) CreateDir(path string) error {

	p := this.normPath(path)

	_, err := os.Stat(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.Mkdir(p, 0666)
}

func (this *Files) Catalog() error {

	this.writeLine(this.rootPath)
	fs, err := ioutil.ReadDir(this.rootPath)
	if err != nil {
		return err
	}
	for _, f := range fs {
		this.write("  ")
		if f.IsDir() {
			this.write(f.Name())
			this.writeLine("/")
		} else {
			this.writeLine(f.Name())
		}
	}

	return nil
}

func (this *Files) EraseFile(name string) error {

	p := this.normPath(name)

	_, err := os.Stat(p)
	if err != nil {
		return err
	}

	return os.Remove(p)
}

func (this *Files) IsFile(name string) bool {
	p := this.normPath(name)

	f, err := os.Stat(p)
	if err != nil {
		return false
	}

	return !f.IsDir()
}

func (this *Files) Rename(from, to string) error {
	fp := this.normPath(from)
	tp := this.normPath(to)

	return os.Rename(fp, tp)
}

func (this *Files) OpenFile(name string) error {
	name = this.normPath(name)

	_, exists := this.openFiles[name]
	if exists {
		return nil
	}

	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	this.openFiles[name] = &NormalFile{
		this.nextId, name, f, bufio.NewReader(f)}

	this.nextId++

	return nil
}

func (this *Files) GetFile(name string) (File, error) {

	if name == "" {
		return this.defaultFile, nil
	}

	name = this.normPath(name)
	f, ok := this.openFiles[name]
	if !ok {
		return nil, errorFileNotOpen(name)
	}

	return f, nil
}

func (this *Files) CloseFile(name string) error {
	if name == "" {
		return nil
	}

	name = this.normPath(name)
	f, ok := this.openFiles[name]
	if !ok {
		return errorFileNotOpen(name)
	}

	delete(this.openFiles, name)

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

	name = this.normPath(name)
	f, ok := this.openFiles[name]
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

	name = this.normPath(name)
	f, ok := this.openFiles[name]
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

func (this *NormalFile) ReadChar() (rune, error) {
	r, _, err := this.r.ReadRune()
	return r, err
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

func (this *StdIOFile) ReadChar() (rune, error) {
	r := bufio.NewReader(this.stdin)
	c, _, err := r.ReadRune()
	return c, err
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
