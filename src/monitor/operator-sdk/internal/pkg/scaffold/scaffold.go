// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Modified from github.com/kubernetes-sigs/controller-tools/pkg/scaffold/scaffold.go

package scaffold

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/operator-framework/operator-sdk/internal/pkg/scaffold/input"
	"github.com/operator-framework/operator-sdk/internal/util/fileutil"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"golang.org/x/tools/imports"
)

// Scaffold writes Templates to scaffold new files
type Scaffold struct {
	// Repo is the go project package
	Repo string
	// AbsProjectPath is the absolute path to the project root, including the project directory.
	AbsProjectPath string
	// ProjectName is the operator's name, ex. app-operator
	ProjectName string
	// Fs is the filesystem GetWriter uses to write scaffold files.
	Fs afero.Fs
	// GetWriter returns a writer for writing scaffold files.
	GetWriter func(path string, mode os.FileMode) (io.Writer, error)
	// BoilerplatePath is the path to a file containing Go boilerplate text.
	BoilerplatePath string

	// boilerplateBytes are bytes of Go boilerplate text.
	boilerplateBytes []byte
}

func (s *Scaffold) setFieldsAndValidate(t input.File) error {
	if b, ok := t.(input.Repo); ok {
		b.SetRepo(s.Repo)
	}
	if b, ok := t.(input.AbsProjectPath); ok {
		b.SetAbsProjectPath(s.AbsProjectPath)
	}
	if b, ok := t.(input.ProjectName); ok {
		b.SetProjectName(s.ProjectName)
	}

	// Validate the template is ok
	if v, ok := t.(input.Validate); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scaffold) configure(cfg *input.Config) {
	s.Repo = cfg.Repo
	s.AbsProjectPath = cfg.AbsProjectPath
	s.ProjectName = cfg.ProjectName
}

func validateBoilerplateBytes(b []byte) error {
	// Append a 'package main' so we can parse the file.
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", append([]byte("package main\n"), b...), parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse boilerplate comments: %v", err)
	}
	if len(f.Comments) == 0 {
		return fmt.Errorf("boilerplate does not contain comments")
	}
	var cb []byte
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			cb = append(cb, []byte(strings.TrimSpace(c.Text)+"\n")...)
		}
	}
	var tb []byte
	tb, cb = bytes.TrimSpace(b), bytes.TrimSpace(cb)
	if bytes.Compare(tb, cb) != 0 {
		return fmt.Errorf(`boilerplate contains text other than comments:\n"%s"\n`, tb)
	}
	return nil
}

func wrapBoilerplateErr(err error, bp string) error {
	return errors.Wrapf(err, `boilerplate file "%s"`, bp)
}

func (s *Scaffold) setBoilerplate() (err error) {
	// If we've already set boilerplate bytes, don't overwrite them.
	if len(s.boilerplateBytes) == 0 {
		bp := s.BoilerplatePath
		if bp == "" {
			i, err := (&Boilerplate{}).GetInput()
			if err != nil {
				return wrapBoilerplateErr(err, i.Path)
			}
			if _, err := s.Fs.Stat(i.Path); err == nil {
				bp = i.Path
			}
		}
		if bp != "" {
			b, err := afero.ReadFile(s.Fs, bp)
			if err != nil {
				return wrapBoilerplateErr(err, bp)
			}
			if err = validateBoilerplateBytes(b); err != nil {
				return wrapBoilerplateErr(err, bp)
			}
			s.boilerplateBytes = append(bytes.TrimSpace(b), '\n', '\n')
		}
	}
	return nil
}

// Execute executes scaffolding the Files
func (s *Scaffold) Execute(cfg *input.Config, files ...input.File) error {
	if s.Fs == nil {
		s.Fs = afero.NewOsFs()
	}
	if s.GetWriter == nil {
		s.GetWriter = fileutil.NewFileWriterFS(s.Fs).WriteCloser
	}

	// Generate boilerplate file first so new Go files get headers.
	if err := s.setBoilerplate(); err != nil {
		return err
	}

	// Configure s using common fields from cfg.
	s.configure(cfg)

	for _, f := range files {
		if err := s.doFile(f); err != nil {
			return err
		}
	}
	return nil
}

// doFile scaffolds a single file
func (s *Scaffold) doFile(e input.File) error {
	// Set common fields
	err := s.setFieldsAndValidate(e)
	if err != nil {
		return err
	}

	// Get the template input params
	i, err := e.GetInput()
	if err != nil {
		return err
	}

	// Ensure we use the absolute file path; i.Path is relative to the project root.
	absFilePath := filepath.Join(s.AbsProjectPath, i.Path)

	// Check if the file to write already exists
	if _, err := s.Fs.Stat(absFilePath); err == nil || os.IsExist(err) {
		switch i.IfExistsAction {
		case input.Overwrite:
		case input.Skip:
			return nil
		case input.Error:
			return fmt.Errorf("%s already exists", absFilePath)
		}
	}

	return s.doRender(i, e, absFilePath)
}

const goFileExt = ".go"

func isGoFile(p string) bool {
	return filepath.Ext(p) == goFileExt
}

func (s *Scaffold) doRender(i input.Input, e input.File, absPath string) error {
	var mode os.FileMode = fileutil.DefaultFileMode
	if i.IsExec {
		mode = fileutil.DefaultExecFileMode
	}
	f, err := s.GetWriter(absPath, mode)
	if err != nil {
		return err
	}
	if c, ok := f.(io.Closer); ok {
		defer func() {
			if err := c.Close(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	var b []byte
	if c, ok := e.(CustomRenderer); ok {
		c.SetFS(s.Fs)
		// CustomRenderers have a non-template method of file rendering.
		if b, err = c.CustomRender(); err != nil {
			return err
		}
	} else {
		// All other files are rendered via their templates.
		temp, err := newTemplate(i)
		if err != nil {
			return err
		}

		out := &bytes.Buffer{}
		if err = temp.Execute(out, e); err != nil {
			return err
		}
		b = out.Bytes()
	}

	// gofmt the imports
	if isGoFile(absPath) {
		b, err = imports.Process(absPath, b, nil)
		if err != nil {
			return err
		}
	}

	// Files being overwritten must be trucated to len 0 so no old bytes remain.
	if _, err = s.Fs.Stat(absPath); err == nil && i.IfExistsAction == input.Overwrite {
		if file, ok := f.(afero.File); ok {
			if err = file.Truncate(0); err != nil {
				return err
			}
		}
	}

	if isGoFile(absPath) && len(s.boilerplateBytes) != 0 {
		if _, err = f.Write(s.boilerplateBytes); err != nil {
			return err
		}
	}
	_, err = f.Write(b)
	log.Infoln("Created", i.Path)
	return err
}

// newTemplate returns a new template named by i.Path with common functions and
// the input's TemplateFuncs.
func newTemplate(i input.Input) (*template.Template, error) {
	t := template.New(i.Path).Funcs(template.FuncMap{
		"title": strings.Title,
		"lower": strings.ToLower,
	})
	if len(i.TemplateFuncs) > 0 {
		t.Funcs(i.TemplateFuncs)
	}
	if i.Delims[0] != "" && i.Delims[1] != "" {
		t.Delims(i.Delims[0], i.Delims[1])
	}
	return t.Parse(i.TemplateBody)
}
